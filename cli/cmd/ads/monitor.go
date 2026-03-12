package ads

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

// --- overview ---

type monitorOverviewOpts struct {
	f         *internal.Factory
	TenantID  string
	Status    string
	AlertOnly bool
	SortBy    string
	Search    string
	Limit     int
	Open      bool
}

func newCmdMonitorOverview(f *internal.Factory) *cobra.Command {
	o := &monitorOverviewOpts{f: f, Limit: 50}
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "投放概览（项目列表 + 当日指标 + 告警状态）",
		Example: `  ahcli ads monitor overview
  ahcli ads monitor overview --status active --sort-by cpa
  ahcli ads monitor overview --alert-only
  ahcli ads monitor overview --search "Smart Watch"
  ahcli -c ads monitor overview | jq '.projects[] | select(.alert_level != "normal")'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return monitorOverviewRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.Status, "status", "all", "项目状态过滤 (active|paused|all)")
	cmd.Flags().BoolVar(&o.AlertOnly, "alert-only", false, "仅显示告警项目")
	cmd.Flags().StringVar(&o.SortBy, "sort-by", "spend", "排序字段 (spend|cpa|ctr|roas|budget_pct)")
	cmd.Flags().StringVar(&o.Search, "search", "", "项目名称/ID 搜索")
	cmd.Flags().IntVar(&o.Limit, "limit", 50, "返回数量")
	cmd.Flags().BoolVar(&o.Open, "open", false, "同时在浏览器中打开盯盘页面")
	return cmd
}

func monitorOverviewRun(o *monitorOverviewOpts) error {
	client := o.f.HUIClient()
	params := url.Values{
		"tenant_id": {o.TenantID},
		"status":    {o.Status},
		"sort_by":   {o.SortBy},
		"limit":     {fmt.Sprintf("%d", o.Limit)},
	}
	if o.AlertOnly {
		params.Set("alert_only", "true")
	}
	if o.Search != "" {
		params.Set("search", o.Search)
	}

	resp, err := client.Get("/api/v1/monitor/overview", params)
	if err != nil {
		return err
	}

	// 非 compact 模式时在 stderr 输出人性化摘要 + 项目表格
	if !o.f.Compact {
		var overview struct {
			Summary struct {
				Total    int     `json:"total"`
				Active   int     `json:"active"`
				Paused   int     `json:"paused"`
				Alerting int     `json:"alerting"`
				Spend    float64 `json:"spend_today"`
			} `json:"summary"`
			Projects []struct {
				ProjectID   string  `json:"project_id"`
				Name        string  `json:"name"`
				Status      string  `json:"status"`
				SpendToday  float64 `json:"spend_today"`
				Conversions int64   `json:"conversions"`
				CPA         float64 `json:"cpa"`
				ROAS        float64 `json:"roas"`
				BudgetPct   float64 `json:"budget_pct"`
				AlertLevel  string  `json:"alert_level"`
			} `json:"projects"`
		}
		if err := json.Unmarshal(resp.Data, &overview); err != nil {
			fmt.Fprintf(internal.Stderr, "  ⚠ 数据格式异常，无法显示摘要: %v\n", err)
		} else {
			s := overview.Summary
			fmt.Fprintf(internal.Stderr, "\n")
			fmt.Fprintf(internal.Stderr, "  投放概览: %d 个项目 (投放中 %d / 已暂停 %d / 告警 %d)\n", s.Total, s.Active, s.Paused, s.Alerting)
			fmt.Fprintf(internal.Stderr, "  今日总花费: $%.2f\n\n", s.Spend)

			if len(overview.Projects) > 0 {
				// Table header
				fmt.Fprintf(internal.Stderr, "  %-4s %-20s %-8s %10s %6s %10s %8s %6s\n",
					"", "项目", "状态", "花费", "转化", "CPA", "ROAS", "预算%")
				fmt.Fprintf(internal.Stderr, "  %s\n", "────────────────────────────────────────────────────────────────────────────────")

				for _, p := range overview.Projects {
					alert := "  "
					if p.AlertLevel == "critical" {
						alert = "🔴"
					} else if p.AlertLevel == "warning" {
						alert = "🟡"
					}

					name := p.Name
					nameRunes := []rune(name)
					if len(nameRunes) > 20 {
						name = string(nameRunes[:17]) + "..."
					}

					fmt.Fprintf(internal.Stderr, "  %-4s %-20s %-8s %10s %6d %10s %7sx %5.0f%%\n",
						alert,
						name,
						p.Status,
						fmt.Sprintf("$%.2f", p.SpendToday),
						p.Conversions,
						fmt.Sprintf("$%.2f", p.CPA),
						fmt.Sprintf("%.2f", p.ROAS),
						p.BudgetPct,
					)
				}
				fmt.Fprintf(internal.Stderr, "\n")
			}
		}
	}

	if o.Open {
		openMonitorPage(o.f, "")
	}

	return o.f.Print(resp.Data)
}

// --- alerts (shortcut for overview --alert-only) ---

func newCmdMonitorAlerts(f *internal.Factory) *cobra.Command {
	o := &monitorOverviewOpts{f: f, AlertOnly: true, Limit: 50}
	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "仅显示告警项目",
		Example: `  ahcli ads monitor alerts
  ahcli ads monitor alerts --sort-by cpa`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return monitorOverviewRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.SortBy, "sort-by", "spend", "排序")
	return cmd
}

// --- config ---

func newCmdMonitorConfig(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "告警配置管理",
	}
	cmd.AddCommand(newCmdMonitorConfigGet(f))
	cmd.AddCommand(newCmdMonitorConfigSet(f))
	cmd.AddCommand(newCmdMonitorConfigReset(f))
	return cmd
}

// config get
type monitorConfigGetOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdMonitorConfigGet(f *internal.Factory) *cobra.Command {
	o := &monitorConfigGetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "查看告警配置",
		Example: `  ahcli ads monitor config get`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			client := f.HUIClient()
			params := url.Values{"tenant_id": {o.TenantID}}
			resp, err := client.Get("/api/v1/monitor/alerts/config", params)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// config set
type monitorConfigSetOpts struct {
	f        *internal.Factory
	TenantID string
	File     string
	Stdin    bool
}

func newCmdMonitorConfigSet(f *internal.Factory) *cobra.Command {
	o := &monitorConfigSetOpts{f: f}
	cmd := &cobra.Command{
		Use:   "set",
		Short: "更新告警配置",
		Example: `  ahcli ads monitor config set --file thresholds.json
  echo '{"thresholds":[...]}' | ahcli ads monitor config set --stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return monitorConfigSetRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func monitorConfigSetRun(o *monitorConfigSetOpts) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}
	if jsonInput == nil {
		return fmt.Errorf("请提供 --file 或 --stdin 输入 JSON")
	}

	bodyBytes, err := json.Marshal(jsonInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	client := o.f.HUIClient()
	params := url.Values{"tenant_id": {o.TenantID}}
	resp, err := client.PutWithParams("/api/v1/monitor/alerts/config", params, bodyBytes)
	if err != nil {
		return err
	}

	fmt.Fprintln(internal.Stderr, "告警配置已更新")
	return o.f.Print(resp.Data)
}

// config reset
type monitorConfigResetOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdMonitorConfigReset(f *internal.Factory) *cobra.Command {
	o := &monitorConfigResetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "reset",
		Short:   "重置告警配置为默认值",
		Example: `  ahcli ads monitor config reset`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			client := f.HUIClient()
			params := url.Values{"tenant_id": {o.TenantID}}
			_, err := client.Delete("/api/v1/monitor/alerts/config", params)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "告警配置已重置为默认值")
			return nil
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// NewCmdMonitor 创建盯盘子命令组
func NewCmdMonitor(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "投放盯盘",
		Long: `投放盯盘命令组，查看投放概览、告警状态、持续监控。

子命令:
  overview    投放概览（项目列表 + 当日指标 + 告警）
  detail      项目详情（指标 + 趋势折线图）
  alerts      仅显示告警项目
  watch       持续监控（定时轮询）
  config      告警配置管理
  open        在浏览器中打开盯盘页面`,
	}

	cmd.AddCommand(newCmdMonitorOverview(f))
	cmd.AddCommand(newCmdMonitorDetail(f))
	cmd.AddCommand(newCmdMonitorAlerts(f))
	cmd.AddCommand(newCmdMonitorWatch(f))
	cmd.AddCommand(newCmdMonitorConfig(f))
	cmd.AddCommand(newCmdMonitorOpen(f))

	return cmd
}

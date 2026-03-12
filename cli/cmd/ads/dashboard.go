package ads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

// NewCmdDashboard 创建 NL Dashboard 子命令组。
func NewCmdDashboard(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "NL Dashboard 数据查询与可视化",
		Long: `NL Dashboard 允许通过自然语言生成的 SQL 查询 ClickHouse 数据，
并在浏览器中以图表形式展示结果。

子命令:
  schema    获取可用的数据表 Schema
  execute   执行 SQL 查询并创建 Session
  open      在浏览器中打开 Dashboard`,
	}

	cmd.AddCommand(newCmdDashboardSchema(f))
	cmd.AddCommand(newCmdDashboardExecute(f))
	cmd.AddCommand(newCmdDashboardOpen(f))

	return cmd
}

func newCmdDashboardSchema(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "schema",
		Short: "获取 ClickHouse 表 Schema（供 LLM 生成 SQL）",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := f.AdsClient().Get("/api/v1/nl-dashboard/schema", nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

func newCmdDashboardExecute(f *internal.Factory) *cobra.Command {
	var (
		sql       string
		chartType string
		xField    string
		yField    string
		title     string
		sessionID string
	)

	cmd := &cobra.Command{
		Use:   "execute",
		Short: "执行 SQL 查询并创建/更新 Session",
		Example: `  ahcli ads dashboard execute --sql "SELECT date, SUM(spend) FROM ad_metrics_daily GROUP BY date"
  ahcli ads dashboard execute --sql "SELECT project_name, SUM(spend) FROM ad_metrics_daily GROUP BY project_name" --chart bar --x project_name --y spend`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sql == "" {
				return fmt.Errorf("--sql 为必填参数")
			}

			tenantID := f.TenantID()

			chartConfig := map[string]interface{}{
				"type": chartType,
			}
			if xField != "" {
				chartConfig["x"] = xField
			}
			if yField != "" {
				chartConfig["y"] = []string{yField}
			}
			if title != "" {
				chartConfig["title"] = title
			}

			req := map[string]interface{}{
				"sql":          sql,
				"tenant_id":    tenantID,
				"chart_config": chartConfig,
			}
			if sessionID != "" {
				req["session_id"] = sessionID
			}

			body, _ := json.Marshal(req)
			resp, err := f.AdsClient().Post("/api/v1/nl-dashboard/execute", bytes.NewReader(body))
			if err != nil {
				return err
			}

			// Parse response to show session ID
			var result struct {
				SessionID string `json:"session_id"`
				RowCount  int    `json:"row_count"`
				ExpiresAt string `json:"expires_at"`
			}
			if err := json.Unmarshal(resp.Data, &result); err == nil {
				fmt.Fprintf(internal.Stderr, "✓ 查询完成: %d 行, Session: %s\n", result.RowCount, result.SessionID)
				fmt.Fprintf(internal.Stderr, "  过期时间: %s\n", result.ExpiresAt)
				fmt.Fprintf(internal.Stderr, "  在浏览器中查看: ahcli ads dashboard open %s\n", result.SessionID)
			}

			return f.Print(resp.Data)
		},
	}

	cmd.Flags().StringVar(&sql, "sql", "", "要执行的 SQL 查询 (必填)")
	cmd.Flags().StringVar(&chartType, "chart", "table", "图表类型: bar, line, pie, table, number")
	cmd.Flags().StringVar(&xField, "x", "", "X 轴字段")
	cmd.Flags().StringVar(&yField, "y", "", "Y 轴字段")
	cmd.Flags().StringVar(&title, "title", "", "图表标题")
	cmd.Flags().StringVar(&sessionID, "session", "", "更新已有 Session（传入 session ID）")

	return cmd
}

func newCmdDashboardOpen(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "open <session-id>",
		Short: "在浏览器中打开 Dashboard Session",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads dashboard open ses_abc123
  ahcli ads dashboard execute --sql "..." && ahcli ads dashboard open <session-id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessionID := args[0]

			feURL, ok := internal.FrontendURLs[f.Env]
			if !ok {
				return fmt.Errorf("未知环境: %s", f.Env)
			}

			target := fmt.Sprintf("%s/open/nl-dashboard?session=%s", feURL, url.QueryEscape(sessionID))

			fmt.Fprintf(internal.Stderr, "正在打开 NL Dashboard...\n")
			if err := internal.OpenBrowser(target); err != nil {
				fmt.Fprintf(internal.Stderr, "无法自动打开浏览器，请手动访问：\n%s\n", target)
				return nil
			}

			fmt.Fprintf(internal.Stderr, "✓ 已打开 Dashboard\n")
			return nil
		},
	}
}

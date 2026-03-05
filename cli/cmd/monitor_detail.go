package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/guptarohit/asciigraph"
	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

type monitorDetailOpts struct {
	f         *internal.Factory
	TenantID  string
	ProjectID string
	Days      int
	Open      bool
}

func newCmdMonitorDetail(f *internal.Factory) *cobra.Command {
	o := &monitorDetailOpts{f: f, Days: 7}
	cmd := &cobra.Command{
		Use:   "detail <project-id>",
		Short: "项目盯盘详情（指标 + 趋势图）",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads monitor detail proj-xxx
  ahcli ads monitor detail proj-xxx --days 14`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.ProjectID = args[0]
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return monitorDetailRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().IntVar(&o.Days, "days", 7, "趋势天数 (1-90)")
	cmd.Flags().BoolVar(&o.Open, "open", false, "同时在浏览器中打开盯盘页面")
	return cmd
}

type timeSeriesPoint struct {
	Time        string  `json:"time"`
	Spend       float64 `json:"spend"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	CTR         float64 `json:"ctr"`
	CPC         float64 `json:"cpc"`
	CPA         float64 `json:"cpa"`
	ROAS        float64 `json:"roas"`
}

func monitorDetailRun(o *monitorDetailOpts) error {
	client := o.f.HUIClient()

	// 1. Fetch overview to get project summary
	overviewParams := url.Values{
		"tenant_id": {o.TenantID},
		"search":    {o.ProjectID},
		"limit":     {"1"},
	}
	overviewResp, err := client.Get("/api/v1/monitor/overview", overviewParams)
	if err != nil {
		return fmt.Errorf("获取项目概览失败: %w", err)
	}

	var overview struct {
		Projects []struct {
			ProjectID   string  `json:"project_id"`
			Name        string  `json:"name"`
			Channel     string  `json:"channel"`
			Status      string  `json:"status"`
			DailyBudget float64 `json:"daily_budget"`
			SpendToday  float64 `json:"spend_today"`
			Impressions int64   `json:"impressions"`
			Clicks      int64   `json:"clicks"`
			Conversions int64   `json:"conversions"`
			CTR         float64 `json:"ctr"`
			CPA         float64 `json:"cpa"`
			ROAS        float64 `json:"roas"`
			BudgetPct   float64 `json:"budget_pct"`
			AlertLevel  string  `json:"alert_level"`
		} `json:"projects"`
	}
	json.Unmarshal(overviewResp.Data, &overview)

	if len(overview.Projects) == 0 {
		return fmt.Errorf("未找到项目: %s", o.ProjectID)
	}

	p := overview.Projects[0]

	// 2. Print project header
	alertTag := ""
	if p.AlertLevel == "critical" {
		alertTag = " 🔴 CRITICAL"
	} else if p.AlertLevel == "warning" {
		alertTag = " 🟡 WARNING"
	}

	fmt.Fprintf(internal.Stderr, "\n")
	fmt.Fprintf(internal.Stderr, "  %-12s %s%s\n", "项目:", p.Name, alertTag)
	fmt.Fprintf(internal.Stderr, "  %-12s %s\n", "ID:", p.ProjectID)
	fmt.Fprintf(internal.Stderr, "  %-12s %s | %s\n", "状态:", strings.ToUpper(p.Status), strings.ToUpper(p.Channel))
	fmt.Fprintf(internal.Stderr, "  %-12s $%.2f / $%.2f (%.0f%%)\n", "预算:", p.SpendToday, p.DailyBudget, p.BudgetPct)
	fmt.Fprintf(internal.Stderr, "\n")

	// 3. Print today's KPI
	fmt.Fprintf(internal.Stderr, "  ┌─────── 今日指标 ───────┐\n")
	fmt.Fprintf(internal.Stderr, "  │ 花费     %12s  │\n", fmt.Sprintf("$%.2f", p.SpendToday))
	fmt.Fprintf(internal.Stderr, "  │ 展示     %12s  │\n", fmtInt(p.Impressions))
	fmt.Fprintf(internal.Stderr, "  │ 点击     %12s  │\n", fmtInt(p.Clicks))
	fmt.Fprintf(internal.Stderr, "  │ 转化     %12s  │\n", fmtInt(p.Conversions))
	fmt.Fprintf(internal.Stderr, "  │ CTR      %11s%%  │\n", fmt.Sprintf("%.2f", p.CTR))
	fmt.Fprintf(internal.Stderr, "  │ CPA      %12s  │\n", fmt.Sprintf("$%.2f", p.CPA))
	fmt.Fprintf(internal.Stderr, "  │ ROAS     %12s  │\n", fmt.Sprintf("%.2fx", p.ROAS))
	fmt.Fprintf(internal.Stderr, "  └─────────────────────────┘\n")

	// 4. Fetch time-series metrics
	metricsParams := url.Values{
		"tenant_id": {o.TenantID},
		"days":      {fmt.Sprintf("%d", o.Days)},
	}
	metricsResp, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/metrics", o.ProjectID), metricsParams)
	if err != nil {
		fmt.Fprintf(internal.Stderr, "\n  ⚠ 趋势数据获取失败: %v\n", err)
		return o.f.Print(overviewResp.Data)
	}

	var points []timeSeriesPoint
	json.Unmarshal(metricsResp.Data, &points)

	if len(points) < 2 {
		fmt.Fprintf(internal.Stderr, "\n  趋势数据不足（需至少 2 天）\n")
		return o.f.Print(overviewResp.Data)
	}

	// 5. Render trend charts
	fmt.Fprintf(internal.Stderr, "\n")
	renderChart(points, "花费 ($)", func(pt timeSeriesPoint) float64 { return pt.Spend })
	renderChart(points, "CPA ($)", func(pt timeSeriesPoint) float64 { return pt.CPA })
	renderChart(points, "ROAS (x)", func(pt timeSeriesPoint) float64 { return pt.ROAS })
	renderChart(points, "转化", func(pt timeSeriesPoint) float64 { return float64(pt.Conversions) })

	// 6. Open browser if requested
	if o.Open {
		openMonitorPage(o.f, o.ProjectID)
	}

	// 7. JSON to stdout
	return o.f.Print(overviewResp.Data)
}

func renderChart(points []timeSeriesPoint, label string, extract func(timeSeriesPoint) float64) {
	values := make([]float64, len(points))
	dates := make([]string, len(points))
	for i, pt := range points {
		values[i] = extract(pt)
		dates[i] = pt.Time
	}

	// Date axis label
	axisLabel := ""
	if len(dates) >= 2 {
		axisLabel = fmt.Sprintf("%s → %s", dates[0], dates[len(dates)-1])
	}

	chart := asciigraph.Plot(values,
		asciigraph.Height(8),
		asciigraph.Width(50),
		asciigraph.Caption(fmt.Sprintf("  %s  (%s)  spark: %s", label, axisLabel, internal.Sparkline(values))),
	)
	fmt.Fprintf(internal.Stderr, "%s\n\n", indent(chart, "  "))
}

func indent(s string, prefix string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}

func fmtInt(v int64) string {
	// Simple thousands formatting
	s := fmt.Sprintf("%d", v)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}

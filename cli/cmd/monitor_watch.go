package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os/signal"
	"syscall"
	"time"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

type monitorWatchOpts struct {
	f         *internal.Factory
	TenantID  string
	Interval  int
	AlertOnly bool
	Count     int
}

func newCmdMonitorWatch(f *internal.Factory) *cobra.Command {
	o := &monitorWatchOpts{f: f, Interval: 60}
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "持续监控（定时轮询，Ctrl+C 退出）",
		Example: `  ahcli ads monitor watch
  ahcli ads monitor watch --interval 120 --alert-only
  ahcli ads monitor watch --count 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return monitorWatchRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().IntVar(&o.Interval, "interval", 60, "轮询间隔（秒）")
	cmd.Flags().BoolVar(&o.AlertOnly, "alert-only", false, "仅在有告警时输出")
	cmd.Flags().IntVar(&o.Count, "count", 0, "最大轮询次数（0=无限）")
	return cmd
}

func monitorWatchRun(o *monitorWatchOpts) error {
	client := o.f.HUIClient()
	params := url.Values{
		"tenant_id": {o.TenantID},
	}
	if o.AlertOnly {
		params.Set("alert_only", "true")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Fprintf(internal.Stderr, "开始监控（每 %ds 刷新，Ctrl+C 退出）\n", o.Interval)

	iteration := 0
	for {
		iteration++
		now := time.Now().Format("15:04:05")

		resp, err := client.Get("/api/v1/monitor/overview", params)
		if err != nil {
			fmt.Fprintf(internal.Stderr, "[%s] #%d 查询失败: %v\n", now, iteration, err)
		} else {
			var overview struct {
				Summary struct {
					Total    int     `json:"total"`
					Active   int     `json:"active"`
					Alerting int     `json:"alerting"`
					Spend    float64 `json:"spend_today"`
				} `json:"summary"`
			}
			if err := json.Unmarshal(resp.Data, &overview); err != nil {
				fmt.Fprintf(internal.Stderr, "[%s] #%d 数据格式异常: %v\n", now, iteration, err)
			} else {
				alertCount := overview.Summary.Alerting

				if o.AlertOnly && alertCount == 0 {
					fmt.Fprintf(internal.Stderr, "[%s] #%d 无告警 (投放中 %d, 花费 $%.2f)\n",
						now, iteration, overview.Summary.Active, overview.Summary.Spend)
				} else {
					fmt.Fprintf(internal.Stderr, "[%s] #%d 投放 %d / 告警 %d / 花费 $%.2f\n",
						now, iteration, overview.Summary.Active, alertCount, overview.Summary.Spend)
					o.f.Print(resp.Data)
				}
			}
		}

		if o.Count > 0 && iteration >= o.Count {
			fmt.Fprintf(internal.Stderr, "已完成 %d 次轮询\n", iteration)
			return nil
		}

		select {
		case <-ctx.Done():
			fmt.Fprintf(internal.Stderr, "\n监控已停止\n")
			return nil
		case <-time.After(time.Duration(o.Interval) * time.Second):
		}
	}
}

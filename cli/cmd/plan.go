package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"sandwichlab_core/tools/ahcli/internal"

	"github.com/spf13/cobra"
)

// --- list ---

type planListOpts struct {
	f         *internal.Factory
	ProjectID string
	TenantID  string
	Limit     int
	Offset    int
}

func newCmdPlanList(f *internal.Factory) *cobra.Command {
	o := &planListOpts{f: f}
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "列出投放计划",
		Example: `  ahcli ads plan list --project-id proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)
			return planListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func planListRun(o *planListOpts) error {
	client := o.f.AdsClient()
	params := url.Values{"project_id": {o.ProjectID}}
	if o.TenantID != "" {
		params.Set("tenant_id", o.TenantID)
	}
	params.Set("limit", fmt.Sprintf("%d", o.Limit))
	params.Set("offset", fmt.Sprintf("%d", o.Offset))

	resp, err := client.Get("/api/v1/plans", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- get ---

func newCmdPlanGet(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <plan-id>",
		Short:   "获取计划详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads plan get plan-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.Get(fmt.Sprintf("/api/v1/plans/%s", args[0]), nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- create ---

type planCreateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdPlanCreate(f *internal.Factory) *cobra.Command {
	o := &planCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "创建投放计划",
		Example: `  ahcli ads plan create --file plan.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供计划 JSON")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/plans", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

// --- update ---

type planUpdateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdPlanUpdate(f *internal.Factory) *cobra.Command {
	o := &planUpdateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "update <plan-id>",
		Short:   "更新投放计划",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads plan update plan-xxx --file updates.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供更新内容")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Put(fmt.Sprintf("/api/v1/plans/%s", args[0]), bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

// --- delete ---

func newCmdPlanDelete(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <plan-id>",
		Short:   "删除投放计划",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads plan delete plan-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.Delete(fmt.Sprintf("/api/v1/plans/%s", args[0]), nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- activate ---

func newCmdPlanActivate(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "activate <plan-id>",
		Short:   "激活投放计划",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads plan activate plan-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/plans/%s/activate", args[0]),
				nil,
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- deactivate ---

func newCmdPlanDeactivate(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "deactivate <plan-id>",
		Short:   "停用投放计划",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads plan deactivate plan-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/plans/%s/deactivate", args[0]),
				nil,
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- epoch get ---

func newCmdPlanEpochGet(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <plan-id> <epoch-number>",
		Short:   "获取指定 Epoch",
		Args:    cobra.ExactArgs(2),
		Example: `  ahcli ads plan epoch get plan-xxx 3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.Get(
				fmt.Sprintf("/api/v1/plans/epochs/plan/%s/%s", args[0], args[1]),
				nil,
			)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- epoch latest ---

func newCmdPlanEpochLatest(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "latest <plan-id>",
		Short:   "获取最新已完成 Epoch",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads plan epoch latest plan-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.Get(
				fmt.Sprintf("/api/v1/plans/epochs/%s/latest-completed", args[0]),
				nil,
			)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- epoch retry ---

func newCmdPlanEpochRetry(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "retry <plan-id> <epoch-number>",
		Short:   "重试 Epoch",
		Args:    cobra.ExactArgs(2),
		Example: `  ahcli ads plan epoch retry plan-xxx 3`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/plans/epochs/%s/retry/%s", args[0], args[1]),
				nil,
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// NewCmdEpoch 创建 Epoch 子命令组。
func NewCmdEpoch(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch",
		Short: "Epoch 管理",
	}

	cmd.AddCommand(newCmdPlanEpochGet(f))
	cmd.AddCommand(newCmdPlanEpochLatest(f))
	cmd.AddCommand(newCmdPlanEpochRetry(f))

	return cmd
}

// NewCmdPlan 创建投放计划子命令组。
func NewCmdPlan(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "投放计划管理",
		Long: `管理投放计划的生命周期和 Epoch 执行。

示例:
  ahcli ads plan list --project-id proj-xxx
  ahcli ads plan create --file plan.json
  ahcli ads plan activate plan-xxx
  ahcli ads plan epoch get <plan-id> <epoch-number>
  ahcli ads plan epoch latest <plan-id>`,
	}

	cmd.AddCommand(newCmdPlanList(f))
	cmd.AddCommand(newCmdPlanGet(f))
	cmd.AddCommand(newCmdPlanCreate(f))
	cmd.AddCommand(newCmdPlanUpdate(f))
	cmd.AddCommand(newCmdPlanDelete(f))
	cmd.AddCommand(newCmdPlanActivate(f))
	cmd.AddCommand(newCmdPlanDeactivate(f))
	cmd.AddCommand(NewCmdEpoch(f))

	return cmd
}

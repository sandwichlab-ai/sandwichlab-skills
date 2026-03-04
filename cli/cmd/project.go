package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"sandwichlab_core/tools/ahcli/internal"

	"github.com/spf13/cobra"
)

// --- list ---

type projectListOpts struct {
	f        *internal.Factory
	TenantID string
	Status   string
	Limit    int
	Offset   int
}

func newCmdProjectList(f *internal.Factory) *cobra.Command {
	o := &projectListOpts{f: f}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出项目",
		Example: `  ahcli ads project list
  ahcli ads project list --status active --limit 20
  ahcli -c ads project list | jq '.projects[].name'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return projectListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVarP(&o.Status, "status", "s", "",
		"按状态过滤 (draft|processing|active|failed|paused|completed|archived)")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func projectListRun(o *projectListOpts) error {
	client := o.f.AdsClient()
	params := url.Values{
		"tenant_id": {o.TenantID},
		"limit":     {fmt.Sprintf("%d", o.Limit)},
		"offset":    {fmt.Sprintf("%d", o.Offset)},
	}
	if o.Status != "" {
		params.Set("status", o.Status)
	}

	resp, err := client.Get("/api/v1/projects", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- get ---

type projectGetOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjectGet(f *internal.Factory) *cobra.Command {
	o := &projectGetOpts{f: f}
	cmd := &cobra.Command{
		Use:   "get <project-id>",
		Short: "获取项目详情",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads project get proj-xxx
  ahcli --env prod ads project get proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			client := o.f.AdsClient()
			params := url.Values{"tenant_id": {o.TenantID}}
			resp, err := client.Get(fmt.Sprintf("/api/v1/projects/%s", args[0]), params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- create ---

type projectCreateOpts struct {
	f              *internal.Factory
	File           string
	Stdin          bool
	TenantID       string
	UserID         string
	Name           string
	Description    string
	SourcePlatform string
}

func newCmdProjectCreate(f *internal.Factory) *cobra.Command {
	o := &projectCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建项目（JSON 文件或 flags）",
		Example: `  ahcli ads project create --file project.json
  ahcli ads project create --name "Smart Watch Pro"
  echo '{"name":"test"}' | ahcli ads project create --stdin`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectCreateRun(o)
		},
	}
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.UserID, "user-id", "", "用户 ID")
	cmd.Flags().StringVar(&o.Name, "name", "", "项目名称（必填）")
	cmd.Flags().StringVar(&o.Description, "description", "", "项目描述")
	cmd.Flags().StringVar(&o.SourcePlatform, "source-platform", "", "来源平台")
	return cmd
}

func projectCreateRun(o *projectCreateOpts) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}

	var bodyBytes []byte

	if jsonInput != nil {
		bodyBytes, err = json.Marshal(jsonInput)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON input: %w", err)
		}
	} else {
		if o.Name == "" {
			return fmt.Errorf("--name 为必填参数（或使用 --file / --stdin 提供 JSON）")
		}

		if requireErr := o.f.RequireTenantID(&o.TenantID); requireErr != nil {
			return requireErr
		}
		if o.UserID == "" {
			o.UserID = o.f.UserID()
		}
		if o.UserID == "" {
			return fmt.Errorf("--user-id 为必填参数（或通过登录设置）")
		}

		// 自动生成 project_id（API 要求必传）
		projectID := fmt.Sprintf("hui-%d", time.Now().UnixMilli())

		reqBody := map[string]interface{}{
			"project_id":      projectID,
			"name":            o.Name,
			"source_platform": "hui",
		}

		// 只在非空时才添加（空值由 client.Post() 自动注入）
		if o.TenantID != "" {
			reqBody["tenant_id"] = o.TenantID
		}
		if o.UserID != "" {
			reqBody["user_id"] = o.UserID
		}

		if o.Description != "" {
			reqBody["description"] = o.Description
		}
		if o.SourcePlatform != "" {
			reqBody["source_platform"] = o.SourcePlatform
		}

		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	client := o.f.AdsClient()
	resp, err := client.Post("/api/v1/projects", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	fmt.Fprintln(internal.Stderr, "项目创建成功")
	return o.f.Print(resp.Data)
}

// --- update ---

type projectUpdateOpts struct {
	f           *internal.Factory
	File        string
	Stdin       bool
	Name        string
	Description string
}

func newCmdProjectUpdate(f *internal.Factory) *cobra.Command {
	o := &projectUpdateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "update <project-id>",
		Short: "更新项目",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads project update proj-xxx --file update.json
  ahcli ads project update proj-xxx --name "New Name"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return projectUpdateRun(o, args[0])
		},
	}
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	cmd.Flags().StringVar(&o.Name, "name", "", "项目名称")
	cmd.Flags().StringVar(&o.Description, "description", "", "项目描述")
	return cmd
}

func projectUpdateRun(o *projectUpdateOpts, id string) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}

	var bodyBytes []byte

	if jsonInput != nil {
		bodyBytes, err = json.Marshal(jsonInput)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON input: %w", err)
		}
	} else {
		reqBody := map[string]interface{}{}

		if o.Name != "" {
			reqBody["name"] = o.Name
		}
		if o.Description != "" {
			reqBody["description"] = o.Description
		}

		if len(reqBody) == 0 {
			return fmt.Errorf("请指定要更新的字段（--name, --description）或使用 --file / --stdin")
		}

		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	client := o.f.AdsClient()
	resp, err := client.Put(fmt.Sprintf("/api/v1/projects/%s", id), bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- delete ---

type projectDeleteOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjectDelete(f *internal.Factory) *cobra.Command {
	o := &projectDeleteOpts{f: f}
	cmd := &cobra.Command{
		Use:     "delete <project-id>",
		Short:   "删除项目",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project delete proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			client := o.f.AdsClient()
			params := url.Values{"tenant_id": {o.TenantID}}
			resp, err := client.Delete(fmt.Sprintf("/api/v1/projects/%s", args[0]), params)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "项目已删除")
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- activate ---

type projectActivateOpts struct {
	f                *internal.Factory
	TenantID         string
	SkipInitialEpoch bool
}

func newCmdProjectActivate(f *internal.Factory) *cobra.Command {
	o := &projectActivateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "activate <project-id>",
		Short: "激活项目",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads project activate proj-xxx
  ahcli ads project activate proj-xxx --skip-initial-epoch`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			params := url.Values{"tenant_id": {o.TenantID}}
			if o.SkipInitialEpoch {
				params.Set("skip_initial_epoch", "true")
			}
			client := o.f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/projects/%s/activate", args[0]),
				params,
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "项目已激活")
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().BoolVar(&o.SkipInitialEpoch, "skip-initial-epoch", false, "跳过 Epoch 0")
	return cmd
}

// --- pause ---

type projectPauseOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjectPause(f *internal.Factory) *cobra.Command {
	o := &projectPauseOpts{f: f}
	cmd := &cobra.Command{
		Use:     "pause <project-id>",
		Short:   "暂停项目",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project pause proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			params := url.Values{"tenant_id": {o.TenantID}}
			client := o.f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/projects/%s/pause", args[0]),
				params,
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "项目已暂停")
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- archive ---

type projectArchiveOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjectArchive(f *internal.Factory) *cobra.Command {
	o := &projectArchiveOpts{f: f}
	cmd := &cobra.Command{
		Use:     "archive <project-id>",
		Short:   "归档项目",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project archive proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			params := url.Values{"tenant_id": {o.TenantID}}
			client := o.f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/projects/%s/archive", args[0]),
				params,
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "项目已归档")
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- diagnose ---

type projectDiagnoseOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjectDiagnose(f *internal.Factory) *cobra.Command {
	o := &projectDiagnoseOpts{f: f}
	cmd := &cobra.Command{
		Use:     "diagnose <project-id>",
		Short:   "项目激活诊断",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project diagnose proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			client := o.f.AdsClient()
			params := url.Values{"tenant_id": {o.TenantID}}
			resp, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/activation-diagnosis", args[0]), params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- metrics ---

type projectMetricsOpts struct {
	f        *internal.Factory
	TenantID string
	Days     int
}

func newCmdProjectMetrics(f *internal.Factory) *cobra.Command {
	o := &projectMetricsOpts{f: f}
	cmd := &cobra.Command{
		Use:   "metrics <project-id>",
		Short: "查看项目指标时序数据",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads project metrics proj-xxx
  ahcli ads project metrics proj-xxx --days 30`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			client := o.f.AdsClient()
			params := url.Values{
				"tenant_id": {o.TenantID},
				"days":      {fmt.Sprintf("%d", o.Days)},
			}
			resp, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/metrics/time-series", args[0]), params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().IntVar(&o.Days, "days", 7, "时间范围（天）")
	return cmd
}

// --- adjust-budget ---

type projectAdjustBudgetOpts struct {
	f           *internal.Factory
	TenantID    string
	DailyBudget int
	File        string
	Stdin       bool
}

func newCmdProjectAdjustBudget(f *internal.Factory) *cobra.Command {
	o := &projectAdjustBudgetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "adjust-budget <project-id>",
		Short:   "调整项目预算",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project adjust-budget proj-xxx --file adj.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return projectAdjustBudgetRun(o, args[0])
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().IntVar(&o.DailyBudget, "daily-budget", 0, "日预算")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func projectAdjustBudgetRun(o *projectAdjustBudgetOpts, id string) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}

	var bodyBytes []byte
	if jsonInput != nil {
		bodyBytes, err = json.Marshal(jsonInput)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON input: %w", err)
		}
	} else {
		if o.DailyBudget <= 0 {
			return fmt.Errorf("--daily-budget 为必填参数且必须大于 0（或使用 --file / --stdin）")
		}
		reqBody := map[string]interface{}{
			"daily_budget": o.DailyBudget,
		}
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	params := url.Values{"tenant_id": {o.TenantID}}
	client := o.f.AdsClient()
	resp, err := client.PostWithParams(
		fmt.Sprintf("/api/v1/projects/%s/adjust-budget-or-date", id),
		params,
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return err
	}

	fmt.Fprintln(internal.Stderr, "预算调整成功")
	return o.f.Print(resp.Data)
}

// --- bill-info ---

type projectBillInfoOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjectBillInfo(f *internal.Factory) *cobra.Command {
	o := &projectBillInfoOpts{f: f}
	cmd := &cobra.Command{
		Use:     "bill-info <project-id>",
		Short:   "查看项目账单信息",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project bill-info proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			client := o.f.AdsClient()
			params := url.Values{"tenant_id": {o.TenantID}}
			resp, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/bill-info", args[0]), params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// NewCmdProject 创建项目管理子命令组。
func NewCmdProject(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "项目管理（AdsCore）",
		Long: `项目管理命令组，覆盖广告项目的完整生命周期。

子命令:
  list           列出项目
  get            获取项目详情
  create         创建项目（JSON 文件或 flags）
  update         更新项目
  delete         删除项目
  activate       激活项目
  pause          暂停项目
  archive        归档项目
  diagnose       激活诊断
  metrics        查看指标时序数据
  adjust-budget  调整预算
  bill-info      查看账单信息
  channel        渠道配置管理
  attachment     项目附件管理`,
	}

	cmd.AddCommand(newCmdProjectList(f))
	cmd.AddCommand(newCmdProjectGet(f))
	cmd.AddCommand(newCmdProjectCreate(f))
	cmd.AddCommand(newCmdProjectUpdate(f))
	cmd.AddCommand(newCmdProjectDelete(f))
	cmd.AddCommand(newCmdProjectActivate(f))
	cmd.AddCommand(newCmdProjectPause(f))
	cmd.AddCommand(newCmdProjectArchive(f))
	cmd.AddCommand(newCmdProjectDiagnose(f))
	cmd.AddCommand(newCmdProjectMetrics(f))
	cmd.AddCommand(newCmdProjectAdjustBudget(f))
	cmd.AddCommand(newCmdProjectBillInfo(f))
	cmd.AddCommand(NewCmdProjectChannel(f))
	cmd.AddCommand(NewCmdProjectAttachment(f))

	return cmd
}

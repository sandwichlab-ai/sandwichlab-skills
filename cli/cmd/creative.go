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

type creativeListOpts struct {
	f         *internal.Factory
	ProjectID string
	TenantID  string
	Status    string
	Limit     int
	Offset    int
}

func newCmdCreativeList(f *internal.Factory) *cobra.Command {
	o := &creativeListOpts{f: f}
	cmd := &cobra.Command{
		Use:     "list [project-id]",
		Short:   "列出创意",
		Args:    cobra.MaximumNArgs(1),
		Example: `  ahcli ads creative list proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				o.ProjectID = args[0]
			}
			if o.ProjectID == "" {
				return fmt.Errorf("project-id 为必填参数")
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)
			return creativeListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVarP(&o.Status, "status", "s", "", "状态过滤 (active|draft|archived)")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func creativeListRun(o *creativeListOpts) error {
	client := o.f.AdsClient()
	params := url.Values{"project_id": {o.ProjectID}}
	if o.TenantID != "" {
		params.Set("tenant_id", o.TenantID)
	}
	if o.Status != "" {
		params.Set("status", o.Status)
	}
	params.Set("limit", fmt.Sprintf("%d", o.Limit))
	params.Set("offset", fmt.Sprintf("%d", o.Offset))

	resp, err := client.Get("/api/v1/creatives", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- get ---

type creativeGetOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdCreativeGet(f *internal.Factory) *cobra.Command {
	o := &creativeGetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get <creative-id>",
		Short:   "获取创意详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads creative get crtv-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Get(fmt.Sprintf("/api/v1/creatives/%s", args[0]), params)
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

type creativeCreateOpts struct {
	f         *internal.Factory
	ProjectID string
	AssetIDs  string
	CopyIDs   string
	File      string
	Stdin     bool
}

func newCmdCreativeCreate(f *internal.Factory) *cobra.Command {
	o := &creativeCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建创意",
		Example: `  ahcli ads creative create --project-id proj-xxx --asset-ids "masset-xxx" --copy-ids "copy-xxx"
  ahcli ads creative create -f creative.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return creativeCreateRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.AssetIDs, "asset-ids", "", "素材 ID 列表（逗号分隔）")
	cmd.Flags().StringVar(&o.CopyIDs, "copy-ids", "", "文案 ID 列表（逗号分隔）")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func creativeCreateRun(o *creativeCreateOpts) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}

	if jsonInput != nil {
		arguments := map[string]interface{}{"body": jsonInput}
		result, callErr := internal.ActionHubCall(o.f.ActionHubClient(), "POST_api_v1_creatives", arguments)
		if callErr != nil {
			return callErr
		}
		return o.f.Print(result)
	}

	if o.ProjectID == "" {
		return fmt.Errorf("--project-id 为必填参数")
	}

	var assetIDs, copyIDs []string
	if o.AssetIDs != "" {
		assetIDs = splitAndTrim(o.AssetIDs)
	}
	if o.CopyIDs != "" {
		copyIDs = splitAndTrim(o.CopyIDs)
	}
	if len(assetIDs) == 0 && len(copyIDs) == 0 {
		return fmt.Errorf("至少需要 --asset-ids 或 --copy-ids 之一")
	}

	body := map[string]interface{}{
		"project_id": o.ProjectID,
		"status":     "active",
	}
	if len(assetIDs) > 0 {
		body["asset_ids"] = assetIDs
	}
	if len(copyIDs) > 0 {
		body["copy_ids"] = copyIDs
	}

	arguments := map[string]interface{}{"body": body}
	result, err := internal.ActionHubCall(o.f.ActionHubClient(), "POST_api_v1_creatives", arguments)
	if err != nil {
		return err
	}
	return o.f.Print(result)
}

// --- update ---

type creativeUpdateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdCreativeUpdate(f *internal.Factory) *cobra.Command {
	o := &creativeUpdateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "update <creative-id>",
		Short:   "更新创意",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads creative update crtv-xxx -f updates.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 -f 或 --stdin 提供更新内容")
			}
			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}
			client := o.f.AdsClient()
			resp, err := client.Put(fmt.Sprintf("/api/v1/creatives/%s", args[0]), bytes.NewReader(bodyBytes))
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

type creativeDeleteOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdCreativeDelete(f *internal.Factory) *cobra.Command {
	o := &creativeDeleteOpts{f: f}
	cmd := &cobra.Command{
		Use:     "delete <creative-id>",
		Short:   "删除创意",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads creative delete crtv-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Delete(fmt.Sprintf("/api/v1/creatives/%s", args[0]), params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- activate ---

func newCmdCreativeActivate(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "activate <creative-id>",
		Short:   "激活创意",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads creative activate crtv-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/creatives/%s/activate", args[0]),
				nil, bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- archive ---

func newCmdCreativeArchive(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "archive <creative-id>",
		Short:   "归档创意",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads creative archive crtv-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/creatives/%s/archive", args[0]),
				nil, bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- bind ---

type creativeBindOpts struct {
	f          *internal.Factory
	CreativeID string
	ProjectID  string
	Action     string
}

func newCmdCreativeBind(f *internal.Factory) *cobra.Command {
	o := &creativeBindOpts{f: f}
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "绑定/解绑创意到项目",
		Example: `  ahcli ads creative bind --creative-id crtv-xxx --project-id proj-xxx --action add
  ahcli ads creative bind --creative-id crtv-xxx --project-id proj-xxx --action remove`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.CreativeID == "" {
				return fmt.Errorf("--creative-id 为必填参数")
			}
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			if o.Action == "" {
				return fmt.Errorf("--action 为必填参数")
			}
			if err := validateBindAction(o.Action); err != nil {
				return err
			}

			reqBody := map[string]interface{}{
				"creative_id": o.CreativeID,
				"project_id":  o.ProjectID,
				"action":      o.Action,
			}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/creatives/bind-project", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.CreativeID, "creative-id", "", "创意 ID")
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.Action, "action", "", "操作 (add|remove)")
	return cmd
}

// NewCmdCreative 创建创意管理子命令组。
func NewCmdCreative(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "creative",
		Short: "创意管理",
		Long: `管理广告创意（Creative = 素材 + 文案的组合）。

示例:
  ahcli ads creative list proj-xxx
  ahcli ads creative create -f creative.json
  ahcli ads creative bind --creative-id crtv-xxx --project-id proj-xxx --action add`,
	}

	cmd.AddCommand(newCmdCreativeList(f))
	cmd.AddCommand(newCmdCreativeGet(f))
	cmd.AddCommand(newCmdCreativeCreate(f))
	cmd.AddCommand(newCmdCreativeUpdate(f))
	cmd.AddCommand(newCmdCreativeDelete(f))
	cmd.AddCommand(newCmdCreativeActivate(f))
	cmd.AddCommand(newCmdCreativeArchive(f))
	cmd.AddCommand(newCmdCreativeBind(f))

	return cmd
}

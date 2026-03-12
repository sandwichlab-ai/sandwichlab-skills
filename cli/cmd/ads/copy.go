package ads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// --- list ---

type copyListOpts struct {
	f         *internal.Factory
	ProjectID string
	TenantID  string
	Limit     int
	Offset    int
}

func newCmdCopyList(f *internal.Factory) *cobra.Command {
	o := &copyListOpts{f: f}
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "列出文案",
		Example: `  ahcli ads copy list --project-id proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)
			return copyListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func copyListRun(o *copyListOpts) error {
	client := o.f.AdsClient()
	params := url.Values{"project_id": {o.ProjectID}}
	if o.TenantID != "" {
		params.Set("tenant_id", o.TenantID)
	}
	params.Set("limit", fmt.Sprintf("%d", o.Limit))
	params.Set("offset", fmt.Sprintf("%d", o.Offset))

	resp, err := client.Get("/api/v1/copies", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- get ---

type copyGetOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdCopyGet(f *internal.Factory) *cobra.Command {
	o := &copyGetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get <copy-id>",
		Short:   "获取文案详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads copy get copy-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Get(fmt.Sprintf("/api/v1/copies/%s", args[0]), params)
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

type copyCreateOpts struct {
	f         *internal.Factory
	ProjectID string
	Headline  string
	CopyText  string
	CTA       string
	File      string
	Stdin     bool
}

func newCmdCopyCreate(f *internal.Factory) *cobra.Command {
	o := &copyCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建文案",
		Example: `  ahcli ads copy create --project-id proj-xxx --headline "Title" --copy-text "Body" --cta LEARN_MORE
  ahcli ads copy create -f copy.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return copyCreateRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.Headline, "headline", "", "广告标题")
	cmd.Flags().StringVar(&o.CopyText, "copy-text", "", "广告正文")
	cmd.Flags().StringVar(&o.CTA, "cta", "LEARN_MORE", "Call-to-Action")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func copyCreateRun(o *copyCreateOpts) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}

	if jsonInput != nil {
		arguments := map[string]interface{}{"body": jsonInput}
		result, callErr := internal.ActionHubCall(o.f.ActionHubClient(), "POST_api_v1_copies", arguments)
		if callErr != nil {
			return callErr
		}
		return o.f.Print(result)
	}

	if o.ProjectID == "" {
		return fmt.Errorf("--project-id 为必填参数")
	}
	if o.Headline == "" || o.CopyText == "" {
		return fmt.Errorf("--headline, --copy-text 为必填参数")
	}
	if o.CTA == "" {
		o.CTA = "LEARN_MORE"
	}

	arguments := map[string]interface{}{
		"body": map[string]interface{}{
			"type":       "meta",
			"project_id": o.ProjectID,
			"content": map[string]interface{}{
				"headline":       []string{o.Headline},
				"copy_text":      []string{o.CopyText},
				"call_to_action": []string{o.CTA},
			},
		},
	}

	result, err := internal.ActionHubCall(o.f.ActionHubClient(), "POST_api_v1_copies", arguments)
	if err != nil {
		return err
	}
	return o.f.Print(result)
}

// --- delete ---

type copyDeleteOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdCopyDelete(f *internal.Factory) *cobra.Command {
	o := &copyDeleteOpts{f: f}
	cmd := &cobra.Command{
		Use:     "delete <copy-id>",
		Short:   "删除文案",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads copy delete copy-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Delete(fmt.Sprintf("/api/v1/copies/%s", args[0]), params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- batch-update ---

type copyBatchUpdateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdCopyBatchUpdate(f *internal.Factory) *cobra.Command {
	o := &copyBatchUpdateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "batch-update",
		Short:   "批量更新文案",
		Example: `  ahcli ads copy batch-update --file copies.json`,
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
			resp, err := client.Post("/api/v1/copies/batch-update", bytes.NewReader(bodyBytes))
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

// --- bind ---

type copyBindOpts struct {
	f         *internal.Factory
	CopyID    string
	ProjectID string
	Action    string
}

func newCmdCopyBind(f *internal.Factory) *cobra.Command {
	o := &copyBindOpts{f: f}
	cmd := &cobra.Command{
		Use:     "bind",
		Short:   "绑定/解绑文案到项目",
		Example: `  ahcli ads copy bind --copy-id copy-xxx --project-id proj-xxx --action add`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.CopyID == "" {
				return fmt.Errorf("--copy-id 为必填参数")
			}
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			if o.Action == "" {
				return fmt.Errorf("--action 为必填参数")
			}
			if err := internal.ValidateBindAction(o.Action); err != nil {
				return err
			}

			reqBody := map[string]interface{}{
				"copy_id":    o.CopyID,
				"project_id": o.ProjectID,
				"action":     o.Action,
			}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/copies/bind-project", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.CopyID, "copy-id", "", "文案 ID")
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.Action, "action", "", "操作 (add|remove)")
	return cmd
}

// NewCmdCopy 创建文案管理子命令组。
func NewCmdCopy(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "文案管理",
		Long: `管理广告文案（Copy = 标题 + 正文 + CTA）。

示例:
  ahcli ads copy list --project-id proj-xxx
  ahcli ads copy create --file copy.json
  ahcli ads copy bind --copy-id copy-xxx --project-id proj-xxx --action add`,
	}

	cmd.AddCommand(newCmdCopyList(f))
	cmd.AddCommand(newCmdCopyGet(f))
	cmd.AddCommand(newCmdCopyCreate(f))
	cmd.AddCommand(newCmdCopyDelete(f))
	cmd.AddCommand(newCmdCopyBatchUpdate(f))
	cmd.AddCommand(newCmdCopyBind(f))

	return cmd
}

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// --- list ---

type projAttachmentListOpts struct {
	f          *internal.Factory
	ProjectID  string
	Type       string
	OnlyLatest bool
	Limit      int
	Offset     int
}

func newCmdProjAttachmentList(f *internal.Factory) *cobra.Command {
	o := &projAttachmentListOpts{f: f}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出项目附件",
		Example: `  ahcli ads project attachment list --project-id proj-xxx
  ahcli ads project attachment list --project-id proj-xxx --type url_analysis`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			return projAttachmentListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.Type, "type", "",
		"附件类型（url_analysis|audience_insights|proposal|market_research|pre_upload_assets）")
	cmd.Flags().BoolVar(&o.OnlyLatest, "only-latest", false, "每种类型仅返回最新一条")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func projAttachmentListRun(o *projAttachmentListOpts) error {
	client := o.f.AdsClient()
	params := url.Values{
		"project_id": {o.ProjectID},
		"limit":      {fmt.Sprintf("%d", o.Limit)},
		"offset":     {fmt.Sprintf("%d", o.Offset)},
	}
	if o.Type != "" {
		params.Set("attachment_type", o.Type)
	}
	if o.OnlyLatest {
		params.Set("only_latest", "true")
	}

	resp, err := client.Get("/api/v1/project-attachments", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- get ---

func newCmdProjAttachmentGet(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <attachment-id>",
		Short:   "获取附件详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project attachment get att-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.Get(
				fmt.Sprintf("/api/v1/project-attachments/%s", args[0]),
				nil,
			)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- create ---

type projAttachmentCreateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdProjAttachmentCreate(f *internal.Factory) *cobra.Command {
	o := &projAttachmentCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "创建项目附件",
		Example: `  ahcli ads project attachment create --file attachment.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供附件 JSON")
			}
			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}
			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/project-attachments", bytes.NewReader(bodyBytes))
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

type projAttachmentUpdateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdProjAttachmentUpdate(f *internal.Factory) *cobra.Command {
	o := &projAttachmentUpdateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "update <attachment-id>",
		Short:   "更新附件",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project attachment update att-xxx --file updates.json`,
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
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/project-attachments/%s", args[0]),
				bytes.NewReader(bodyBytes),
			)
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

// --- upsert ---

type projAttachmentUpsertOpts struct {
	f         *internal.Factory
	ProjectID string
	Type      string
	Name      string
	File      string
	Stdin     bool
}

func newCmdProjAttachmentUpsert(f *internal.Factory) *cobra.Command {
	o := &projAttachmentUpsertOpts{f: f}
	cmd := &cobra.Command{
		Use:   "upsert",
		Short: "创建或更新项目附件",
		Long: `创建或更新项目附件。content 通过 --file 或 --stdin 传入 JSON。
type 枚举：url_analysis | audience_insights | proposal | market_research | pre_upload_assets`,
		Example: `  echo '{"url":"https://..."}' | \
    ahcli ads project attachment upsert --project-id proj-xxx --type url_analysis --name product_info --stdin

  ahcli ads project attachment upsert --project-id proj-xxx \
    --type proposal --name campaign_proposal --file proposal.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			if o.Type == "" || o.Name == "" {
				return fmt.Errorf("--type, --name 为必填参数")
			}
			return projAttachmentUpsertRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.Type, "type", "", "附件类型（必填）")
	cmd.Flags().StringVar(&o.Name, "name", "", "附件名称（必填）")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func projAttachmentUpsertRun(o *projAttachmentUpsertOpts) error {
	content, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}
	if content == nil {
		return fmt.Errorf("必须通过 --file 或 --stdin 提供 content JSON")
	}

	reqBody := map[string]interface{}{
		"attachment_type": o.Type,
		"name":            o.Name,
		"source_type":     "hui",
		"content":         content,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := o.f.AdsClient()
	params := url.Values{"project_id": {o.ProjectID}}
	resp, err := client.PostWithParams("/api/v1/project-attachments/upsert", params, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- latest ---

type projAttachmentLatestOpts struct {
	f         *internal.Factory
	ProjectID string
	Type      string
}

func newCmdProjAttachmentLatest(f *internal.Factory) *cobra.Command {
	o := &projAttachmentLatestOpts{f: f}
	cmd := &cobra.Command{
		Use:     "latest",
		Short:   "获取指定类型的最新附件",
		Example: `  ahcli ads project attachment latest --project-id proj-xxx --type url_analysis`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			if o.Type == "" {
				return fmt.Errorf("--type 为必填参数")
			}
			client := o.f.AdsClient()
			params := url.Values{
				"project_id":      {o.ProjectID},
				"attachment_type": {o.Type},
				"only_latest":     {"true"},
				"limit":           {"1"},
			}
			resp, err := client.Get("/api/v1/project-attachments", params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.Type, "type", "", "附件类型（必填）")
	return cmd
}

// --- delete ---

type projAttachmentDeleteOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjAttachmentDelete(f *internal.Factory) *cobra.Command {
	o := &projAttachmentDeleteOpts{f: f}
	cmd := &cobra.Command{
		Use:     "delete <attachment-id>",
		Short:   "删除附件",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project attachment delete att-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Delete(
				fmt.Sprintf("/api/v1/project-attachments/%s", args[0]),
				params,
			)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// NewCmdProjectAttachment 创建项目附件管理子命令组。
func NewCmdProjectAttachment(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attachment",
		Short: "项目附件管理（/project-attachments）",
		Long: `管理项目附件（品牌素材、投放提案等）。

注意：attachment API 是 flat 的（/project-attachments），用 query param project_id 过滤。

示例:
  ahcli ads project attachment list --project-id proj-xxx
  ahcli ads project attachment get att-xxx
  ahcli ads project attachment upsert --project-id proj-xxx --type url_analysis --name info --file content.json`,
	}

	cmd.AddCommand(newCmdProjAttachmentList(f))
	cmd.AddCommand(newCmdProjAttachmentGet(f))
	cmd.AddCommand(newCmdProjAttachmentCreate(f))
	cmd.AddCommand(newCmdProjAttachmentUpdate(f))
	cmd.AddCommand(newCmdProjAttachmentUpsert(f))
	cmd.AddCommand(newCmdProjAttachmentLatest(f))
	cmd.AddCommand(newCmdProjAttachmentDelete(f))

	return cmd
}

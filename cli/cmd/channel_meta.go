package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// ==================== Campaign ====================

// --- campaign list ---

type metaCampaignListOpts struct {
	f          *internal.Factory
	AccountIDs string
	TenantID   string
	Limit      int
	Offset     int
}

func newCmdMetaCampaignList(f *internal.Factory) *cobra.Command {
	o := &metaCampaignListOpts{f: f}
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "列出 Meta Campaigns",
		Example: `  ahcli channel meta campaign list --account-ids act_123,act_456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			return metaCampaignListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.AccountIDs, "account-ids", "", "Meta 广告账户 ID（逗号分隔）")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().IntVar(&o.Limit, "limit", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func metaCampaignListRun(o *metaCampaignListOpts) error {
	client := o.f.AdsClient()
	params := url.Values{}
	if o.TenantID != "" {
		params.Set("tenant_id", o.TenantID)
	}
	if o.AccountIDs != "" {
		params.Set("account_ids", o.AccountIDs)
	}
	params.Set("limit", fmt.Sprintf("%d", o.Limit))
	params.Set("offset", fmt.Sprintf("%d", o.Offset))

	resp, err := client.Get("/api/v1/channel/meta/campaigns", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- campaign by-projects ---

type metaCampaignByProjectsOpts struct {
	f          *internal.Factory
	ProjectIDs string
}

func newCmdMetaCampaignByProjects(f *internal.Factory) *cobra.Command {
	o := &metaCampaignByProjectsOpts{f: f}
	cmd := &cobra.Command{
		Use:     "by-projects",
		Short:   "按项目 ID 批量查询 Campaigns",
		Example: `  ahcli channel meta campaign by-projects --project-ids proj-xxx,proj-yyy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectIDs == "" {
				return fmt.Errorf("--project-ids 为必填参数")
			}
			return metaCampaignByProjectsRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectIDs, "project-ids", "", "项目 ID 列表（逗号分隔）")
	return cmd
}

func metaCampaignByProjectsRun(o *metaCampaignByProjectsOpts) error {
	projectIDs := splitAndTrim(o.ProjectIDs)
	reqBody := map[string]interface{}{"project_ids": projectIDs}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := o.f.AdsClient()
	resp, err := client.Post("/api/v1/channel/meta/campaigns/by-project-ids", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// ==================== Entity ====================

// --- entity get ---

type metaEntityGetOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdMetaEntityGet(f *internal.Factory) *cobra.Command {
	o := &metaEntityGetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get <type> <id>",
		Short:   "获取 Meta 实体详情",
		Long:    `type: campaign | adset | ad | creative`,
		Args:    cobra.ExactArgs(2),
		Example: `  ahcli channel meta entity get campaign 123456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entityType := args[0]
			entityID := args[1]
			o.TenantID = f.ResolveTenantID(o.TenantID)

			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}

			resp, err := client.Get(
				fmt.Sprintf("/api/v1/channel/meta/action/entity/%s/%s", entityType, entityID),
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

// --- entity list ---

type metaEntityListOpts struct {
	f        *internal.Factory
	ParentID string
	TenantID string
	Limit    int
}

func newCmdMetaEntityList(f *internal.Factory) *cobra.Command {
	o := &metaEntityListOpts{f: f}
	cmd := &cobra.Command{
		Use:     "list <type>",
		Short:   "列出 Meta 子实体",
		Long:    `type: campaign | adset | ad`,
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli channel meta entity list adset --parent-id 123456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			entityType := args[0]
			o.TenantID = f.ResolveTenantID(o.TenantID)

			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			if o.ParentID != "" {
				params.Set("parent_id", o.ParentID)
			}
			params.Set("limit", fmt.Sprintf("%d", o.Limit))

			resp, err := client.Get(
				fmt.Sprintf("/api/v1/channel/meta/action/entity_list/%s", entityType),
				params,
			)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.ParentID, "parent-id", "", "父实体 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().IntVar(&o.Limit, "limit", 20, "每页数量")
	return cmd
}

// ==================== Audience ====================

// --- audience create ---

type metaAudienceCreateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdMetaAudienceCreate(f *internal.Factory) *cobra.Command {
	o := &metaAudienceCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "创建自定义受众",
		Example: `  ahcli channel meta audience create --file audience.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供受众 JSON")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/meta/custom-audiences", bytes.NewReader(bodyBytes))
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

// --- audience get ---

func newCmdMetaAudienceGet(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <audience-id>",
		Short:   "获取受众详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli channel meta audience get aud-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			client := f.AdsClient()
			resp, err := client.Get(fmt.Sprintf("/api/v1/meta/custom-audiences/%s", id), nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- audience add-users ---

type metaAudienceAddUsersOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdMetaAudienceAddUsers(f *internal.Factory) *cobra.Command {
	o := &metaAudienceAddUsersOpts{f: f}
	cmd := &cobra.Command{
		Use:     "add-users <audience-id>",
		Short:   "向受众添加用户",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli channel meta audience add-users aud-xxx --file users.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供用户列表 JSON")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post(
				fmt.Sprintf("/api/v1/meta/custom-audiences/%s/users", id),
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

// --- audience remove-users ---

type metaAudienceRemoveUsersOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdMetaAudienceRemoveUsers(f *internal.Factory) *cobra.Command {
	o := &metaAudienceRemoveUsersOpts{f: f}
	cmd := &cobra.Command{
		Use:     "remove-users <audience-id>",
		Short:   "从受众移除用户",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli channel meta audience remove-users aud-xxx --file users.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供用户列表 JSON")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.DeleteWithBody(
				fmt.Sprintf("/api/v1/meta/custom-audiences/%s/users", id),
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

// ==================== Tools ====================

// --- account-info ---

type metaAccountInfoOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdMetaAccountInfo(f *internal.Factory) *cobra.Command {
	o := &metaAccountInfoOpts{f: f}
	cmd := &cobra.Command{
		Use:     "account-info",
		Short:   "获取 Meta 账户信息",
		Example: `  ahcli channel meta account-info --file params.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供参数 JSON")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/channel/meta/action/get_account_info", bytes.NewReader(bodyBytes))
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

// --- convert-to-usd ---

type metaConvertToUSDOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdMetaConvertToUSD(f *internal.Factory) *cobra.Command {
	o := &metaConvertToUSDOpts{f: f}
	cmd := &cobra.Command{
		Use:     "convert-to-usd",
		Short:   "货币转换为 USD",
		Example: `  ahcli channel meta convert-to-usd --file params.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供参数 JSON")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/channel/meta/action/convert_to_usd", bytes.NewReader(bodyBytes))
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

// NewCmdChannelMeta 创建渠道查询子命令组。
func NewCmdChannelMeta(f *internal.Factory) *cobra.Command {
	channelCmd := &cobra.Command{
		Use:   "channel",
		Short: "渠道查询",
		Long: `渠道相关的查询命令。

目前支持 Meta 渠道，未来扩展 Google / TikTok。

示例:
  ahcli channel meta campaign list
  ahcli channel meta entity get campaign 123456
  ahcli channel meta audience create --file audience.json`,
	}

	metaCmd := &cobra.Command{
		Use:   "meta",
		Short: "Meta 渠道",
	}

	// campaign 子命令组
	campaignCmd := &cobra.Command{
		Use:   "campaign",
		Short: "Meta Campaign 查询",
	}
	campaignCmd.AddCommand(newCmdMetaCampaignList(f))
	campaignCmd.AddCommand(newCmdMetaCampaignByProjects(f))

	// entity 子命令组
	entityCmd := &cobra.Command{
		Use:   "entity",
		Short: "Meta 实体查询",
	}
	entityCmd.AddCommand(newCmdMetaEntityGet(f))
	entityCmd.AddCommand(newCmdMetaEntityList(f))

	// audience 子命令组
	audienceCmd := &cobra.Command{
		Use:   "audience",
		Short: "Meta 自定义受众管理",
	}
	audienceCmd.AddCommand(newCmdMetaAudienceCreate(f))
	audienceCmd.AddCommand(newCmdMetaAudienceGet(f))
	audienceCmd.AddCommand(newCmdMetaAudienceAddUsers(f))
	audienceCmd.AddCommand(newCmdMetaAudienceRemoveUsers(f))

	metaCmd.AddCommand(campaignCmd)
	metaCmd.AddCommand(entityCmd)
	metaCmd.AddCommand(audienceCmd)
	metaCmd.AddCommand(newCmdMetaAccountInfo(f))
	metaCmd.AddCommand(newCmdMetaConvertToUSD(f))

	channelCmd.AddCommand(metaCmd)

	return channelCmd
}

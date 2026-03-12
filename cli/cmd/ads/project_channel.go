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

type projChannelListOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjChannelList(f *internal.Factory) *cobra.Command {
	o := &projChannelListOpts{f: f}
	cmd := &cobra.Command{
		Use:     "list <project-id>",
		Short:   "列出渠道配置",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project channel list proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Get(fmt.Sprintf("/api/v1/projects/%s/channel-configs", args[0]), params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- get ---

type projChannelGetOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjChannelGet(f *internal.Factory) *cobra.Command {
	o := &projChannelGetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get <project-id> <config-id>",
		Short:   "获取渠道配置详情",
		Args:    cobra.ExactArgs(2),
		Example: `  ahcli ads project channel get proj-xxx cfg-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Get(
				fmt.Sprintf("/api/v1/projects/%s/channel-configs/%s", args[0], args[1]),
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

// --- create ---

type projChannelCreateOpts struct {
	f        *internal.Factory
	TenantID string
	UserID   string
	Channel  string
	File     string
	Stdin    bool
}

func newCmdProjChannelCreate(f *internal.Factory) *cobra.Command {
	o := &projChannelCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "create <project-id>",
		Short: "创建渠道配置",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads project channel create proj-xxx --file config.json
  ahcli ads project channel create proj-xxx --channel Meta`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return projChannelCreateRun(o, args[0])
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.UserID, "user-id", "", "用户 ID")
	cmd.Flags().StringVar(&o.Channel, "channel", "Meta", "渠道名称（默认 Meta）")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func projChannelCreateRun(o *projChannelCreateOpts, projectID string) error {
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
		if requireErr := o.f.RequireTenantID(&o.TenantID); requireErr != nil {
			return requireErr
		}
		if o.UserID == "" {
			o.UserID = o.f.UserID()
		}
		if o.UserID == "" {
			return fmt.Errorf("--user-id 为必填参数（或通过登录设置）")
		}

		reqBody := map[string]interface{}{
			"project_id":        projectID,
			"name":              o.Channel,
			"enabled":           true,
			"budget_allocation": 1.0,
			"extra_config":      map[string]interface{}{"is_fully_managed": true},
		}

		// 只在非空时才添加（空值由 client.Post() 自动注入）
		if o.TenantID != "" {
			reqBody["tenant_id"] = o.TenantID
		}
		if o.UserID != "" {
			reqBody["user_id"] = o.UserID
		}
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	client := o.f.AdsClient()
	resp, err := client.Post(
		fmt.Sprintf("/api/v1/projects/%s/channel-configs", projectID),
		bytes.NewReader(bodyBytes),
	)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- update ---

type projChannelUpdateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdProjChannelUpdate(f *internal.Factory) *cobra.Command {
	o := &projChannelUpdateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "update <project-id> <config-id>",
		Short:   "更新渠道配置",
		Args:    cobra.ExactArgs(2),
		Example: `  ahcli ads project channel update proj-xxx cfg-xxx --file update.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供 JSON")
			}
			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}
			client := o.f.AdsClient()
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/projects/%s/channel-configs/%s", args[0], args[1]),
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

// --- delete ---

type projChannelDeleteOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjChannelDelete(f *internal.Factory) *cobra.Command {
	o := &projChannelDeleteOpts{f: f}
	cmd := &cobra.Command{
		Use:     "delete <project-id> <config-id>",
		Short:   "删除渠道配置",
		Args:    cobra.ExactArgs(2),
		Example: `  ahcli ads project channel delete proj-xxx cfg-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Delete(
				fmt.Sprintf("/api/v1/projects/%s/channel-configs/%s", args[0], args[1]),
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

// --- set-pixel ---

type projChannelSetPixelOpts struct {
	f               *internal.Factory
	PixelID         string
	ConversionEvent string
	PixelSource     string
}

func newCmdProjChannelSetPixel(f *internal.Factory) *cobra.Command {
	o := &projChannelSetPixelOpts{f: f}
	cmd := &cobra.Command{
		Use:   "set-pixel <project-id>",
		Short: "配置转化追踪（Pixel）",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads project channel set-pixel proj-xxx --pixel-id 654321
  ahcli ads project channel set-pixel proj-xxx --pixel-id 654321 --conversion-event Purchase --pixel-source transit_bm`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.PixelID == "" {
				return fmt.Errorf("--pixel-id 为必填参数")
			}
			reqBody := map[string]string{
				"pixel_id":         o.PixelID,
				"conversion_event": o.ConversionEvent,
				"pixel_source":     o.PixelSource,
			}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}
			client := o.f.AdsClient()
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/projects/%s/pixel-config", args[0]),
				bytes.NewReader(bodyBytes),
			)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.PixelID, "pixel-id", "", "Pixel ID（必填）")
	cmd.Flags().StringVar(&o.ConversionEvent, "conversion-event", "Purchase", "转化事件（默认 Purchase）")
	cmd.Flags().StringVar(&o.PixelSource, "pixel-source", "transit_bm", "Pixel 来源（transit_bm|user_owned）")
	return cmd
}

// --- set-google-tag ---

type projChannelSetGoogleTagOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdProjChannelSetGoogleTag(f *internal.Factory) *cobra.Command {
	o := &projChannelSetGoogleTagOpts{f: f}
	cmd := &cobra.Command{
		Use:     "set-google-tag <project-id>",
		Short:   "配置 Google Tag",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project channel set-google-tag proj-xxx --file google-tag.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供 JSON")
			}
			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}
			client := o.f.AdsClient()
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/projects/%s/google-tag-config", args[0]),
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

// --- auths ---

type projChannelAuthsOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdProjChannelAuths(f *internal.Factory) *cobra.Command {
	o := &projChannelAuthsOpts{f: f}
	cmd := &cobra.Command{
		Use:     "auths <project-id>",
		Short:   "查看渠道授权信息",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads project channel auths proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Get(
				fmt.Sprintf("/api/v1/projects/%s/channel-auths", args[0]),
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

// NewCmdProjectChannel 创建渠道配置管理子命令组。
func NewCmdProjectChannel(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "渠道配置管理（/projects/:id/channel-configs）",
		Long: `管理项目的渠道配置。

示例:
  ahcli ads project channel list <project-id>
  ahcli ads project channel get <project-id> <config-id>
  ahcli ads project channel create <project-id> --file config.json
  ahcli ads project channel set-pixel <project-id> --pixel-id px-xxx`,
	}

	cmd.AddCommand(newCmdProjChannelList(f))
	cmd.AddCommand(newCmdProjChannelGet(f))
	cmd.AddCommand(newCmdProjChannelCreate(f))
	cmd.AddCommand(newCmdProjChannelUpdate(f))
	cmd.AddCommand(newCmdProjChannelDelete(f))
	cmd.AddCommand(newCmdProjChannelSetPixel(f))
	cmd.AddCommand(newCmdProjChannelSetGoogleTag(f))
	cmd.AddCommand(newCmdProjChannelAuths(f))

	return cmd
}

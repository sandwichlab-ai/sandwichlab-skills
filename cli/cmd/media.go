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

// --- search ---

type mediaSearchOpts struct {
	f         *internal.Factory
	ProjectID string
	TenantID  string
	Query     string
	Type      string
	Limit     int
	Offset    int
}

func newCmdMediaSearch(f *internal.Factory) *cobra.Command {
	o := &mediaSearchOpts{f: f}
	cmd := &cobra.Command{
		Use:     "search",
		Short:   "搜索素材",
		Example: `  ahcli ads media search --project-id proj-xxx --type image`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)
			return mediaSearchRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.Query, "query", "", "搜索关键词")
	cmd.Flags().StringVar(&o.Type, "type", "", "素材类型 (image|video)")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 20, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func mediaSearchRun(o *mediaSearchOpts) error {
	client := o.f.AdsClient()
	params := url.Values{"project_id": {o.ProjectID}}
	if o.TenantID != "" {
		params.Set("tenant_id", o.TenantID)
	}
	if o.Query != "" {
		params.Set("query", o.Query)
	}
	if o.Type != "" {
		params.Set("type", o.Type)
	}
	params.Set("limit", fmt.Sprintf("%d", o.Limit))
	params.Set("offset", fmt.Sprintf("%d", o.Offset))

	resp, err := client.Get("/api/v1/media-library/search", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- get ---

type mediaGetOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdMediaGet(f *internal.Factory) *cobra.Command {
	o := &mediaGetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get <asset-id>",
		Short:   "获取素材详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads media get masset-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Get(fmt.Sprintf("/api/v1/media-library/assets/%s", args[0]), params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- batch-get ---

type mediaBatchGetOpts struct {
	f   *internal.Factory
	IDs string
}

func newCmdMediaBatchGet(f *internal.Factory) *cobra.Command {
	o := &mediaBatchGetOpts{f: f}
	cmd := &cobra.Command{
		Use:     "batch-get",
		Short:   "批量获取素材",
		Example: `  ahcli ads media batch-get --ids masset-xxx,masset-yyy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.IDs == "" {
				return fmt.Errorf("--ids 为必填参数")
			}

			ids := splitAndTrim(o.IDs)
			reqBody := map[string]interface{}{"ids": ids}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/media-library/assets/batch", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.IDs, "ids", "", "素材 ID 列表（逗号分隔）")
	return cmd
}

// --- quota ---

type mediaQuotaOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdMediaQuota(f *internal.Factory) *cobra.Command {
	o := &mediaQuotaOpts{f: f}
	cmd := &cobra.Command{
		Use:     "quota",
		Short:   "查看租户素材配额",
		Example: `  ahcli ads media quota`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)

			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Get("/api/v1/media-library/quota", params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- generate ---

type mediaGenerateOpts struct {
	f            *internal.Factory
	ProjectID    string
	TenantID     string
	ProductName  string
	SellingPoint string
	Preference   string
	Count        int
	File         string
	Stdin        bool
}

func newCmdMediaGenerate(f *internal.Factory) *cobra.Command {
	o := &mediaGenerateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "AI 生成素材",
		Example: `  ahcli ads media generate --file params.json
  ahcli ads media generate --project-id proj-xxx --product-name "Watch" --selling-point "AI Health" --count 2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mediaGenerateRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.ProductName, "product-name", "", "产品名称")
	cmd.Flags().StringVar(&o.SellingPoint, "selling-point", "", "核心卖点")
	cmd.Flags().StringVar(&o.Preference, "preference", "", "风格偏好")
	cmd.Flags().IntVar(&o.Count, "count", 2, "生成数量")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func mediaGenerateRun(o *mediaGenerateOpts) error {
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
		if o.ProjectID == "" {
			return fmt.Errorf("--project-id 为必填参数")
		}
		o.TenantID = o.f.ResolveTenantID(o.TenantID)

		count := o.Count
		if count <= 0 {
			count = 2
		}

		reqBody := map[string]interface{}{
			"project_id":    o.ProjectID,
			"project_name":  o.ProductName,
			"selling_point": o.SellingPoint,
			"language":      "en_US",
			"geo_locations": []map[string]string{{"country": "US"}},
			"count":         count,
			"type":          "asset_manual",
		}

		// 只在非空时才添加（空值由 client.Post() 自动注入）
		if o.TenantID != "" {
			reqBody["tenant_id"] = o.TenantID
		}
		if o.Preference != "" {
			reqBody["preference"] = o.Preference
		}
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	client := o.f.AdsClient()
	resp, err := client.Post("/api/v1/media-library/generate-asset", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- poll ---

type mediaPollOpts struct {
	f         *internal.Factory
	ProjectID string
	TenantID  string
	Wait      bool
}

func newCmdMediaPoll(f *internal.Factory) *cobra.Command {
	o := &mediaPollOpts{f: f}
	cmd := &cobra.Command{
		Use:   "poll <execution-arn>",
		Short: "轮询素材生成结果",
		Args:  cobra.ExactArgs(1),
		Example: `  ahcli ads media poll "arn:aws:states:..."
  ahcli ads media poll "arn:aws:states:..." --wait`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)
			return mediaPollRun(o, args[0])
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().BoolVar(&o.Wait, "wait", false, "自动轮询直到完成")
	return cmd
}

func mediaPollRun(o *mediaPollOpts, arn string) error {
	client := o.f.AdsClient()

	maxAttempts := 1
	if o.Wait {
		maxAttempts = 10
	}

	params := url.Values{
		"execution_arn": {arn},
		"project_id":    {o.ProjectID},
		"tenant_id":     {o.TenantID},
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := client.Get("/api/v1/media-library/execution-result", params)
		if err != nil {
			return err
		}

		if !o.Wait {
			return o.f.Print(resp.Data)
		}

		var result struct {
			Status string `json:"status"`
		}
		if json.Unmarshal(resp.Data, &result) == nil {
			fmt.Fprintf(internal.Stderr, "[polling] attempt %d/%d: %s\n", attempt, maxAttempts, result.Status)
			if result.Status == "SUCCEEDED" || result.Status == "FAILED" {
				return o.f.Print(resp.Data)
			}
		}

		if attempt < maxAttempts {
			time.Sleep(30 * time.Second)
		}
	}

	return fmt.Errorf("轮询超时：%d 次尝试后仍未完成", maxAttempts)
}

// --- import-s3 ---

type mediaImportS3Opts struct {
	f         *internal.Factory
	ProjectID string
	S3URLs    string
	File      string
	Stdin     bool
}

func newCmdMediaImportS3(f *internal.Factory) *cobra.Command {
	o := &mediaImportS3Opts{f: f}
	cmd := &cobra.Command{
		Use:   "import-s3",
		Short: "从 S3 导入素材",
		Example: `  ahcli ads media import-s3 --project-id proj-xxx --s3-urls "s3://bucket/img.png,s3://bucket/img2.png"
  ahcli ads media import-s3 --file import.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return mediaImportS3Run(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.S3URLs, "s3-urls", "", "S3 URL 列表（逗号分隔）")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func mediaImportS3Run(o *mediaImportS3Opts) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}

	var arguments map[string]interface{}
	if jsonInput != nil {
		arguments = map[string]interface{}{"body": jsonInput}
	} else {
		if o.ProjectID == "" {
			return fmt.Errorf("--project-id 为必填参数")
		}
		if o.S3URLs == "" {
			return fmt.Errorf("--s3-urls 为必填参数")
		}

		s3URLs := splitAndTrim(o.S3URLs)
		assets := make([]map[string]string, len(s3URLs))
		for i, u := range s3URLs {
			assets[i] = map[string]string{"s3_url": u}
		}

		arguments = map[string]interface{}{
			"body": map[string]interface{}{
				"project_id": o.ProjectID,
				"assets":     assets,
			},
		}
	}

	result, err := internal.ActionHubCall(o.f.ActionHubClient(), "POST_api_v1_media-library_import-from-s3", arguments)
	if err != nil {
		return err
	}
	return o.f.Print(result)
}

// --- edit ---

type mediaEditOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdMediaEdit(f *internal.Factory) *cobra.Command {
	o := &mediaEditOpts{f: f}
	cmd := &cobra.Command{
		Use:     "edit",
		Short:   "编辑素材",
		Example: `  ahcli ads media edit --file edit-params.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供编辑参数")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/media-library/edit-asset", bytes.NewReader(bodyBytes))
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

type mediaDeleteOpts struct {
	f   *internal.Factory
	IDs string
}

func newCmdMediaDelete(f *internal.Factory) *cobra.Command {
	o := &mediaDeleteOpts{f: f}
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "批量删除素材",
		Example: `  ahcli ads media delete --ids masset-xxx,masset-yyy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.IDs == "" {
				return fmt.Errorf("--ids 为必填参数")
			}

			ids := splitAndTrim(o.IDs)
			reqBody := map[string]interface{}{"ids": ids}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/media-library/assets/delete", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.IDs, "ids", "", "素材 ID 列表（逗号分隔）")
	return cmd
}

// --- favorite ---

type mediaFavoriteOpts struct {
	f   *internal.Factory
	IDs string
	Set bool
}

func newCmdMediaFavorite(f *internal.Factory) *cobra.Command {
	o := &mediaFavoriteOpts{f: f}
	cmd := &cobra.Command{
		Use:     "favorite",
		Short:   "批量收藏/取消收藏素材",
		Example: `  ahcli ads media favorite --ids masset-xxx,masset-yyy --set true`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.IDs == "" {
				return fmt.Errorf("--ids 为必填参数")
			}

			ids := splitAndTrim(o.IDs)
			reqBody := map[string]interface{}{
				"ids":      ids,
				"favorite": o.Set,
			}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Put("/api/v1/media-library/assets/favorite", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.IDs, "ids", "", "素材 ID 列表（逗号分隔）")
	cmd.Flags().BoolVar(&o.Set, "set", true, "设置收藏状态")
	return cmd
}

// --- trash ---

type mediaTrashOpts struct {
	f   *internal.Factory
	IDs string
	Set bool
}

func newCmdMediaTrash(f *internal.Factory) *cobra.Command {
	o := &mediaTrashOpts{f: f}
	cmd := &cobra.Command{
		Use:     "trash",
		Short:   "批量移入/移出回收站",
		Example: `  ahcli ads media trash --ids masset-xxx,masset-yyy --set true`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.IDs == "" {
				return fmt.Errorf("--ids 为必填参数")
			}

			ids := splitAndTrim(o.IDs)
			reqBody := map[string]interface{}{
				"ids":     ids,
				"trashed": o.Set,
			}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Put("/api/v1/media-library/assets/trash", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.IDs, "ids", "", "素材 ID 列表（逗号分隔）")
	cmd.Flags().BoolVar(&o.Set, "set", true, "设置回收站状态")
	return cmd
}

// --- batch-update ---

type mediaBatchUpdateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdMediaBatchUpdate(f *internal.Factory) *cobra.Command {
	o := &mediaBatchUpdateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "batch-update",
		Short:   "批量更新素材",
		Example: `  ahcli ads media batch-update --file updates.json`,
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
			resp, err := client.Post("/api/v1/media-library/assets/batch-update", bytes.NewReader(bodyBytes))
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

type mediaBindOpts struct {
	f         *internal.Factory
	AssetID   string
	ProjectID string
	Action    string
}

func newCmdMediaBind(f *internal.Factory) *cobra.Command {
	o := &mediaBindOpts{f: f}
	cmd := &cobra.Command{
		Use:     "bind",
		Short:   "绑定/解绑素材到项目",
		Example: `  ahcli ads media bind --asset-id masset-xxx --project-id proj-xxx --action add`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.AssetID == "" {
				return fmt.Errorf("--asset-id 为必填参数")
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
				"asset_ids":  []string{o.AssetID},
				"project_id": o.ProjectID,
				"action":     o.Action,
			}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/media-library/assets/bind-project", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.AssetID, "asset-id", "", "素材 ID")
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.Action, "action", "", "操作 (add|remove)")
	return cmd
}

// --- tag list ---

type mediaTagListOpts struct {
	f        *internal.Factory
	TenantID string
}

func newCmdMediaTagList(f *internal.Factory) *cobra.Command {
	o := &mediaTagListOpts{f: f}
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "列出标签",
		Example: `  ahcli ads media tag list`,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.TenantID = f.ResolveTenantID(o.TenantID)
			client := o.f.AdsClient()
			params := url.Values{}
			if o.TenantID != "" {
				params.Set("tenant_id", o.TenantID)
			}
			resp, err := client.Get("/api/v1/media-library/tags", params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- tag create ---

type mediaTagCreateOpts struct {
	f        *internal.Factory
	Name     string
	TenantID string
}

func newCmdMediaTagCreate(f *internal.Factory) *cobra.Command {
	o := &mediaTagCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "创建标签",
		Example: `  ahcli ads media tag create --name "Hero"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.Name == "" {
				return fmt.Errorf("--name 为必填参数")
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)

			reqBody := map[string]interface{}{"name": o.Name}
			if o.TenantID != "" {
				reqBody["tenant_id"] = o.TenantID
			}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/media-library/tags", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.Name, "name", "", "标签名称")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	return cmd
}

// --- tag update ---

type mediaTagUpdateOpts struct {
	f    *internal.Factory
	Name string
}

func newCmdMediaTagUpdate(f *internal.Factory) *cobra.Command {
	o := &mediaTagUpdateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "update <tag-id>",
		Short:   "更新标签",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads media tag update tag-xxx --name "New Name"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.Name == "" {
				return fmt.Errorf("--name 为必填参数")
			}

			reqBody := map[string]string{"name": o.Name}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Put(fmt.Sprintf("/api/v1/media-library/tags/%s", args[0]), bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.Name, "name", "", "标签名称")
	return cmd
}

// --- tag delete ---

func newCmdMediaTagDelete(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <tag-id>",
		Short:   "删除标签",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads media tag delete tag-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.Delete(fmt.Sprintf("/api/v1/media-library/tags/%s", args[0]), nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- tag copy-assets ---

type mediaTagCopyAssetsOpts struct {
	f           *internal.Factory
	SourceTagID string
	TargetTagID string
}

func newCmdMediaTagCopyAssets(f *internal.Factory) *cobra.Command {
	o := &mediaTagCopyAssetsOpts{f: f}
	cmd := &cobra.Command{
		Use:     "copy-assets",
		Short:   "复制标签下的素材到另一标签",
		Example: `  ahcli ads media tag copy-assets --source-tag-id tag-xxx --target-tag-id tag-yyy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.SourceTagID == "" || o.TargetTagID == "" {
				return fmt.Errorf("--source-tag-id 和 --target-tag-id 为必填参数")
			}

			reqBody := map[string]string{
				"source_tag_id": o.SourceTagID,
				"target_tag_id": o.TargetTagID,
			}
			bodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return fmt.Errorf("failed to marshal request body: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/media-library/tags/copy-assets", bytes.NewReader(bodyBytes))
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.SourceTagID, "source-tag-id", "", "源标签 ID")
	cmd.Flags().StringVar(&o.TargetTagID, "target-tag-id", "", "目标标签 ID")
	return cmd
}

// NewCmdMediaTag 创建素材标签管理子命令组。
func NewCmdMediaTag(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "素材标签管理",
	}

	cmd.AddCommand(newCmdMediaTagList(f))
	cmd.AddCommand(newCmdMediaTagCreate(f))
	cmd.AddCommand(newCmdMediaTagUpdate(f))
	cmd.AddCommand(newCmdMediaTagDelete(f))
	cmd.AddCommand(newCmdMediaTagCopyAssets(f))

	return cmd
}

// NewCmdMedia 创建素材库管理子命令组。
func NewCmdMedia(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "media",
		Short: "素材库管理",
		Long: `管理媒体素材（图片、视频）的搜索、生成、导入和标签。

示例:
  ahcli ads media search --project-id proj-xxx
  ahcli ads media get masset-xxx
  ahcli ads media generate --file params.json
  ahcli ads media tag list`,
	}

	cmd.AddCommand(newCmdMediaSearch(f))
	cmd.AddCommand(newCmdMediaGet(f))
	cmd.AddCommand(newCmdMediaBatchGet(f))
	cmd.AddCommand(newCmdMediaQuota(f))
	cmd.AddCommand(newCmdMediaGenerate(f))
	cmd.AddCommand(newCmdMediaPoll(f))
	cmd.AddCommand(newCmdMediaImportS3(f))
	cmd.AddCommand(newCmdMediaEdit(f))
	cmd.AddCommand(newCmdMediaDelete(f))
	cmd.AddCommand(newCmdMediaFavorite(f))
	cmd.AddCommand(newCmdMediaTrash(f))
	cmd.AddCommand(newCmdMediaBatchUpdate(f))
	cmd.AddCommand(newCmdMediaBind(f))
	cmd.AddCommand(NewCmdMediaTag(f))

	return cmd
}

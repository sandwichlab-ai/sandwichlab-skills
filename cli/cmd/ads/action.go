package ads

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// --- create ---

type actionCreateOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdActionCreate(f *internal.Factory) *cobra.Command {
	o := &actionCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "异步创建投放动作",
		Example: `  ahcli ads action create --file action.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供 action JSON")
			}

			arguments := map[string]interface{}{"body": jsonInput}
			result, err := internal.ActionHubCall(o.f.ActionHubClient(), "POST_api_v1_actions", arguments)
			if err != nil {
				return err
			}
			return o.f.Print(result)
		},
	}
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

// --- create-sync ---

type actionCreateSyncOpts struct {
	f              *internal.Factory
	File           string
	Stdin          bool
	TenantID       string
	ProjectID      string
	AccessToken    string
	AdAccountID    string
	CampaignName   string
	Objective      string
	DailyBudget    int
	CreativeID     string
	Countries      string
	StartTime      string
	EndTime        string
	AgeMin         int
	AgeMax         int
	TimeoutSeconds int
}

func newCmdActionCreateSync(f *internal.Factory) *cobra.Command {
	o := &actionCreateSyncOpts{f: f}
	cmd := &cobra.Command{
		Use:   "create-sync",
		Short: "同步创建投放动作（等待完成）",
		Long: `同步创建广告（Campaign + AdSet + Ads）。支持两种模式：
  1. JSON 文件/stdin（完全控制 payload）
  2. Flag 快捷方式（常见场景）`,
		Example: `  ahcli ads action create-sync --file campaign.json
  ahcli ads action create-sync \
    --tenant-id tnt-xxx --project-id proj-xxx \
    --access-token "EAAxxxx" --ad-account-id "act_123" \
    --campaign-name "Smart Watch Q1" --daily-budget 50 --creative-id crtv-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}

			var actionBody map[string]interface{}
			if jsonInput != nil {
				actionBody = jsonInput
			} else {
				actionBody, err = buildCampaignFromOpts(o)
				if err != nil {
					return err
				}
			}

			arguments := map[string]interface{}{"body": actionBody}
			result, err := internal.ActionHubCall(o.f.ActionHubClient(), "POST_api_v1_actions_sync", arguments)
			if err != nil {
				return err
			}
			return o.f.Print(result)
		},
	}
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.AccessToken, "access-token", "", "Meta Access Token")
	cmd.Flags().StringVar(&o.AdAccountID, "ad-account-id", "", "Meta Ad Account ID")
	cmd.Flags().StringVar(&o.CampaignName, "campaign-name", "", "广告系列名称")
	cmd.Flags().StringVar(&o.Objective, "objective", "OUTCOME_TRAFFIC", "目标")
	cmd.Flags().IntVar(&o.DailyBudget, "daily-budget", 0, "每日预算（美元）")
	cmd.Flags().StringVar(&o.CreativeID, "creative-id", "", "Creative ID")
	cmd.Flags().StringVar(&o.Countries, "countries", "US", "目标国家（逗号分隔）")
	cmd.Flags().StringVar(&o.StartTime, "start-time", "", "开始时间（ISO 8601）")
	cmd.Flags().StringVar(&o.EndTime, "end-time", "", "结束时间（ISO 8601）")
	cmd.Flags().IntVar(&o.AgeMin, "age-min", 18, "最低年龄")
	cmd.Flags().IntVar(&o.AgeMax, "age-max", 65, "最高年龄")
	cmd.Flags().IntVar(&o.TimeoutSeconds, "timeout-seconds", 120, "超时秒数")
	return cmd
}

// campaignObjectiveToOptimizationGoal 映射 campaign objective 到 optimization_goal
var campaignObjectiveToOptimizationGoal = map[string]string{
	"OUTCOME_TRAFFIC":   "LINK_CLICKS",
	"OUTCOME_SALES":     "OFFSITE_CONVERSIONS",
	"OUTCOME_AWARENESS": "REACH",
}

// buildCampaignFromOpts 从 opts 构建 campaign action body（供 action create-sync 使用）
func buildCampaignFromOpts(o *actionCreateSyncOpts) (map[string]interface{}, error) {
	o.TenantID = o.f.ResolveTenantID(o.TenantID)

	if o.TenantID == "" || o.ProjectID == "" || o.AccessToken == "" || o.AdAccountID == "" || o.CampaignName == "" {
		return nil, fmt.Errorf("--tenant-id, --project-id, --access-token, --ad-account-id, --campaign-name 为必填参数（tenant-id 可通过 ahcli auth login 自动获取）")
	}
	if o.DailyBudget <= 0 {
		return nil, fmt.Errorf("--daily-budget 必须大于 0")
	}

	optimizationGoal, ok := campaignObjectiveToOptimizationGoal[o.Objective]
	if !ok {
		return nil, fmt.Errorf("未知的 objective: %s", o.Objective)
	}

	countries := []string{"US"}
	if o.Countries != "" {
		countries = internal.SplitAndTrim(o.Countries)
	}

	targeting := map[string]interface{}{
		"age_min":       o.AgeMin,
		"age_max":       o.AgeMax,
		"genders":       []int{0},
		"geo_locations": map[string]interface{}{"countries": countries},
	}

	creative := map[string]interface{}{"creative_id": o.CreativeID}
	adsetName := o.CampaignName + "_adset_1"
	adName := o.CampaignName + "_ad_1"

	return map[string]interface{}{
		"tenant_id":    o.TenantID,
		"project_id":   o.ProjectID,
		"channel_type": "meta",
		"action_type":  "create_campaign_composite",
		"payload": map[string]interface{}{
			"create_campaign_composite": map[string]interface{}{
				"access_token":  o.AccessToken,
				"ad_account_id": o.AdAccountID,
				"campaign_name": o.CampaignName,
				"objective":     o.Objective,
				"status":        "PAUSED",
				"bid_strategy":  "LOWEST_COST_WITHOUT_CAP",
				"adsets": []map[string]interface{}{
					{
						"adset_name":         adsetName,
						"status":             "PAUSED",
						"daily_budget":       fmt.Sprintf("%d", o.DailyBudget),
						"optimization_goal":  optimizationGoal,
						"billing_event":      "IMPRESSIONS",
						"bid_strategy":       "LOWEST_COST_WITHOUT_CAP",
						"start_time":         o.StartTime,
						"end_time":           o.EndTime,
						"audience_targeting": targeting,
						"ads": []map[string]interface{}{
							{
								"ad_name":  adName,
								"status":   "PAUSED",
								"creative": creative,
							},
						},
					},
				},
			},
		},
		"timeout_seconds": o.TimeoutSeconds,
	}, nil
}

// --- batch ---

type actionBatchOpts struct {
	f     *internal.Factory
	File  string
	Stdin bool
}

func newCmdActionBatch(f *internal.Factory) *cobra.Command {
	o := &actionBatchOpts{f: f}
	cmd := &cobra.Command{
		Use:     "batch",
		Short:   "批量创建投放动作",
		Example: `  ahcli ads action batch --file actions.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
			if err != nil {
				return err
			}
			if jsonInput == nil {
				return fmt.Errorf("必须通过 --file 或 --stdin 提供 actions JSON")
			}

			bodyBytes, err := json.Marshal(jsonInput)
			if err != nil {
				return fmt.Errorf("failed to marshal JSON input: %w", err)
			}

			client := o.f.AdsClient()
			resp, err := client.Post("/api/v1/actions/batch", bytes.NewReader(bodyBytes))
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

// --- list ---

type actionListOpts struct {
	f         *internal.Factory
	ProjectID string
	TenantID  string
	Status    string
	Limit     int
	Offset    int
}

func newCmdActionList(f *internal.Factory) *cobra.Command {
	o := &actionListOpts{f: f}
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "列出投放动作",
		Example: `  ahcli ads action list --project-id proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ProjectID == "" {
				return fmt.Errorf("--project-id 为必填参数")
			}
			o.TenantID = f.ResolveTenantID(o.TenantID)
			return actionListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVarP(&o.Status, "status", "s", "", "状态过滤 (pending|processing|completed|failed|cancelled)")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func actionListRun(o *actionListOpts) error {
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

	resp, err := client.Get("/api/v1/actions", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- get ---

func newCmdActionGet(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <action-id>",
		Short:   "获取投放动作详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads action get act-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.Get(fmt.Sprintf("/api/v1/actions/%s", args[0]), nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- cancel ---

func newCmdActionCancel(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "cancel <action-id>",
		Short:   "取消投放动作",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ads action cancel act-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.PostWithParams(
				fmt.Sprintf("/api/v1/actions/%s/cancel", args[0]),
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

// --- stats ---

func newCmdActionStats(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "stats",
		Short:   "查看 Worker Pool 状态",
		Example: `  ahcli ads action stats`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.AdsClient()
			resp, err := client.Get("/api/v1/actions/stats", nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// --- query-by-biz-id ---

type actionQueryByBizIDOpts struct {
	f          *internal.Factory
	OuterBizID string
}

func newCmdActionQueryByBizID(f *internal.Factory) *cobra.Command {
	o := &actionQueryByBizIDOpts{f: f}
	cmd := &cobra.Command{
		Use:     "query-by-biz-id",
		Short:   "按外部业务 ID 查询动作",
		Example: `  ahcli ads action query-by-biz-id --outer-biz-id biz-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.OuterBizID == "" {
				return fmt.Errorf("--outer-biz-id 为必填参数")
			}

			client := o.f.AdsClient()
			params := url.Values{"outer_biz_id": {o.OuterBizID}}

			resp, err := client.Get("/api/v1/actions/query/by-outer-biz-id", params)
			if err != nil {
				return err
			}
			return o.f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.OuterBizID, "outer-biz-id", "", "外部业务 ID")
	return cmd
}

// NewCmdAction 创建投放动作子命令组。
func NewCmdAction(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "action",
		Short: "投放动作管理（Campaign 创建/管理）",
		Long: `管理投放动作（通过 ActionHub 创建 Campaign）。

示例:
  ahcli ads action create --file action.json
  ahcli ads action create-sync --file action.json
  ahcli ads action batch --file actions.json
  ahcli ads action list --project-id proj-xxx
  ahcli ads action get act-xxx`,
	}

	cmd.AddCommand(newCmdActionCreate(f))
	cmd.AddCommand(newCmdActionCreateSync(f))
	cmd.AddCommand(newCmdActionBatch(f))
	cmd.AddCommand(newCmdActionList(f))
	cmd.AddCommand(newCmdActionGet(f))
	cmd.AddCommand(newCmdActionCancel(f))
	cmd.AddCommand(newCmdActionStats(f))
	cmd.AddCommand(newCmdActionQueryByBizID(f))

	return cmd
}

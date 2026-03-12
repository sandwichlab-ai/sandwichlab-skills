package rules

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

// NewCmdRules 创建规则管理命令组。
func NewCmdRules(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "规则管理（HUI 后端）",
		Long: `规则管理命令组，支持列表查询、自然语言创建、审批等。

子命令:
  list      列出规则（支持 project_id 过滤）
  create    创建规则（自然语言或 JSON）
  get       查看规则详情
  submit    提交规则审批
  approve   审批通过规则

示例:
  ahcli rules list --project-id proj-xxx
  ahcli rules create "ROI低于0.01持续2小时暂停广告"
  ahcli rules create --file rule.json
  ahcli rules get 42
  ahcli rules submit 42
  ahcli rules approve 42`,
	}

	cmd.AddCommand(newCmdRulesList(f))
	cmd.AddCommand(newCmdRulesCreate(f))
	cmd.AddCommand(newCmdRulesGet(f))
	cmd.AddCommand(newCmdRulesSubmit(f))
	cmd.AddCommand(newCmdRulesApprove(f))

	return cmd
}

// ============================================================================
// list
// ============================================================================

type rulesListOpts struct {
	f         *internal.Factory
	TenantID  string
	ProjectID string
	Status    string
	Name      string
	Limit     int
	Offset    int
}

func newCmdRulesList(f *internal.Factory) *cobra.Command {
	o := &rulesListOpts{f: f}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出规则（支持 project_id 过滤）",
		Example: `  ahcli rules list
  ahcli rules list --project-id proj-xxx --status active
  ahcli -c rules list | jq '.items[].name'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return rulesListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.ProjectID, "project-id", "", "项目 ID（筛选 scope 含此项目或 scope.type=all 的规则）")
	cmd.Flags().StringVarP(&o.Status, "status", "s", "", "状态过滤 (draft|pending|active|inactive|rejected)")
	cmd.Flags().StringVar(&o.Name, "name", "", "规则名称模糊搜索")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func rulesListRun(o *rulesListOpts) error {
	client := o.f.HUIClient()
	params := url.Values{
		"tenant_id": {o.TenantID},
		"limit":     {strconv.Itoa(o.Limit)},
		"offset":    {strconv.Itoa(o.Offset)},
	}
	if o.ProjectID != "" {
		params.Set("project_id", o.ProjectID)
	}
	if o.Status != "" {
		params.Set("status", o.Status)
	}
	if o.Name != "" {
		params.Set("name", o.Name)
	}

	resp, err := client.Get("/api/v1/rule", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// ============================================================================
// get
// ============================================================================

func newCmdRulesGet(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <rule-id>",
		Short:   "查看规则详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli rules get 42`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.HUIClient()
			resp, err := client.Get(fmt.Sprintf("/api/v1/rule/%s", args[0]), nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// ============================================================================
// submit
// ============================================================================

func newCmdRulesSubmit(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "submit <rule-id>",
		Short:   "提交规则审批",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli rules submit 42`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.HUIClient()
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/rule/%s/submit", args[0]),
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "规则已提交审批")
			return f.Print(resp.Data)
		},
	}
}

// ============================================================================
// approve
// ============================================================================

type rulesApproveOpts struct {
	f          *internal.Factory
	Resolution string
}

func newCmdRulesApprove(f *internal.Factory) *cobra.Command {
	o := &rulesApproveOpts{f: f}
	cmd := &cobra.Command{
		Use:     "approve <rule-id>",
		Short:   "审批通过规则",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli rules approve 42 --resolution "已审核，符合安全要求"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]string{}
			if o.Resolution != "" {
				body["resolution"] = o.Resolution
			}
			bodyBytes, _ := json.Marshal(body)
			client := f.HUIClient()
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/rule/%s/approve", args[0]),
				bytes.NewReader(bodyBytes),
			)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "规则已审批通过")
			return f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.Resolution, "resolution", "", "审批说明")
	return cmd
}

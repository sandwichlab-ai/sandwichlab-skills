package ticket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

// NewCmdTicket 创建工单管理命令组。
func NewCmdTicket(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ticket",
		Short: "工单管理（HUI 后端）",
		Long: `工单管理命令组，支持创建、查询、审批等。

子命令:
  create    创建工单（flags 或 stdin JSON）
  list      列出工单
  get       查看工单详情
  approve   审批通过
  reject    审批拒绝
  cancel    取消工单

示例:
  ahcli ticket create --title "调整预算" --type adjust_budget --source-ref-id proj-xxx
  ahcli ticket list --status pending
  ahcli ticket get 42
  ahcli ticket approve 42`,
	}

	cmd.AddCommand(newCmdTicketCreate(f))
	cmd.AddCommand(newCmdTicketList(f))
	cmd.AddCommand(newCmdTicketGet(f))
	cmd.AddCommand(newCmdTicketApprove(f))
	cmd.AddCommand(newCmdTicketReject(f))
	cmd.AddCommand(newCmdTicketCancel(f))

	return cmd
}

// ============================================================================
// create
// ============================================================================

type ticketCreateOpts struct {
	f               *internal.Factory
	TenantID        string
	Title           string
	Description     string
	Type            string
	AssigneeRoleKey string
	SourceService   string
	SourceRefID     string
	SourceRefType   string
	Payload         string
	File            string
	Stdin           bool
}

func newCmdTicketCreate(f *internal.Factory) *cobra.Command {
	o := &ticketCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "创建工单",
		Long: `创建工单，支持 flags 模式或 JSON 模式。

Flags 模式:
  ahcli ticket create --title "调整预算至$30" --type adjust_budget \
    --source-ref-id proj-xxx --payload '{"daily_budget":30}'

JSON 模式:
  echo '{"title":"...","type":"adjust_budget",...}' | ahcli ticket create --stdin
  ahcli ticket create --file ticket.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}

			// JSON 模式
			if o.File != "" || o.Stdin {
				return ticketCreateJSONRun(o)
			}

			// Flags 模式
			if o.Title == "" {
				return fmt.Errorf("请提供 --title 或使用 --file/--stdin 提供 JSON")
			}
			if o.Type == "" {
				return fmt.Errorf("请提供 --type (pause_campaign|adjust_budget|enable_campaign|other)")
			}
			return ticketCreateFlagsRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.Title, "title", "", "工单标题")
	cmd.Flags().StringVar(&o.Description, "description", "", "工单描述")
	cmd.Flags().StringVar(&o.Type, "type", "", "工单类型 (pause_campaign|adjust_budget|enable_campaign|other)")
	cmd.Flags().StringVar(&o.AssigneeRoleKey, "assignee-role-key", "", "指派角色 key")
	cmd.Flags().StringVar(&o.SourceService, "source-service", "", "来源服务")
	cmd.Flags().StringVar(&o.SourceRefID, "source-ref-id", "", "关联资源 ID")
	cmd.Flags().StringVar(&o.SourceRefType, "source-ref-type", "", "关联资源类型")
	cmd.Flags().StringVar(&o.Payload, "payload", "", "附加数据 (JSON 字符串)")
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	return cmd
}

func ticketCreateFlagsRun(o *ticketCreateOpts) error {
	body := map[string]interface{}{
		"title": o.Title,
		"type":  o.Type,
	}
	if o.Description != "" {
		body["description"] = o.Description
	}
	if o.AssigneeRoleKey != "" {
		body["assignee_role_key"] = o.AssigneeRoleKey
	}
	if o.SourceService != "" {
		body["source_service"] = o.SourceService
	}
	if o.SourceRefID != "" {
		body["source_ref_id"] = o.SourceRefID
	}
	if o.SourceRefType != "" {
		body["source_ref_type"] = o.SourceRefType
	}
	if o.Payload != "" {
		var payload interface{}
		if err := json.Unmarshal([]byte(o.Payload), &payload); err != nil {
			return fmt.Errorf("--payload 不是有效的 JSON: %w", err)
		}
		body["payload"] = payload
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	client := o.f.HUIClient()
	resp, err := client.Post("/api/v1/ticket", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	fmt.Fprintln(internal.Stderr, "工单创建成功")
	return o.f.Print(resp.Data)
}

func ticketCreateJSONRun(o *ticketCreateOpts) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}
	if jsonInput == nil {
		return fmt.Errorf("请提供 --file 或 --stdin 输入 JSON")
	}

	bodyBytes, err := json.Marshal(jsonInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	client := o.f.HUIClient()
	resp, err := client.Post("/api/v1/ticket", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	fmt.Fprintln(internal.Stderr, "工单创建成功")
	return o.f.Print(resp.Data)
}

// ============================================================================
// list
// ============================================================================

type ticketListOpts struct {
	f        *internal.Factory
	TenantID string
	Status   string
	Type     string
	Limit    int
	Offset   int
}

func newCmdTicketList(f *internal.Factory) *cobra.Command {
	o := &ticketListOpts{f: f}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "列出工单",
		Example: `  ahcli ticket list
  ahcli ticket list --status pending --type adjust_budget
  ahcli -c ticket list | jq '.items[].title'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}
			return ticketListRun(o)
		},
	}
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVarP(&o.Status, "status", "s", "", "状态过滤 (pending|approved|rejected|cancelled|closed)")
	cmd.Flags().StringVarP(&o.Type, "type", "t", "", "类型过滤 (pause_campaign|adjust_budget|enable_campaign|other)")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 10, "每页数量")
	cmd.Flags().IntVar(&o.Offset, "offset", 0, "偏移量")
	return cmd
}

func ticketListRun(o *ticketListOpts) error {
	client := o.f.HUIClient()
	params := url.Values{
		"tenant_id": {o.TenantID},
		"limit":     {strconv.Itoa(o.Limit)},
		"offset":    {strconv.Itoa(o.Offset)},
	}
	if o.Status != "" {
		params.Set("status", o.Status)
	}
	if o.Type != "" {
		params.Set("type", o.Type)
	}

	resp, err := client.Get("/api/v1/ticket", params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// ============================================================================
// get
// ============================================================================

func newCmdTicketGet(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "get <ticket-id>",
		Short:   "查看工单详情",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ticket get 42`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.HUIClient()
			resp, err := client.Get(fmt.Sprintf("/api/v1/ticket/%s", args[0]), nil)
			if err != nil {
				return err
			}
			return f.Print(resp.Data)
		},
	}
}

// ============================================================================
// approve
// ============================================================================

type ticketApproveOpts struct {
	f          *internal.Factory
	Resolution string
}

func newCmdTicketApprove(f *internal.Factory) *cobra.Command {
	o := &ticketApproveOpts{f: f}
	cmd := &cobra.Command{
		Use:     "approve <ticket-id>",
		Short:   "审批通过工单",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ticket approve 42 --resolution "已审核，同意执行"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]string{}
			if o.Resolution != "" {
				body["resolution"] = o.Resolution
			}
			bodyBytes, _ := json.Marshal(body)
			client := f.HUIClient()
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/ticket/%s/approve", args[0]),
				bytes.NewReader(bodyBytes),
			)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "工单已审批通过")
			return f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.Resolution, "resolution", "", "审批说明")
	return cmd
}

// ============================================================================
// reject
// ============================================================================

type ticketRejectOpts struct {
	f          *internal.Factory
	Resolution string
}

func newCmdTicketReject(f *internal.Factory) *cobra.Command {
	o := &ticketRejectOpts{f: f}
	cmd := &cobra.Command{
		Use:     "reject <ticket-id>",
		Short:   "审批拒绝工单",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ticket reject 42 --resolution "预算调整幅度过大"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]string{}
			if o.Resolution != "" {
				body["resolution"] = o.Resolution
			}
			bodyBytes, _ := json.Marshal(body)
			client := f.HUIClient()
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/ticket/%s/reject", args[0]),
				bytes.NewReader(bodyBytes),
			)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "工单已拒绝")
			return f.Print(resp.Data)
		},
	}
	cmd.Flags().StringVar(&o.Resolution, "resolution", "", "拒绝原因")
	return cmd
}

// ============================================================================
// cancel
// ============================================================================

func newCmdTicketCancel(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "cancel <ticket-id>",
		Short:   "取消工单",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli ticket cancel 42`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := f.HUIClient()
			resp, err := client.Put(
				fmt.Sprintf("/api/v1/ticket/%s/cancel", args[0]),
				bytes.NewReader([]byte("{}")),
			)
			if err != nil {
				return err
			}
			fmt.Fprintln(internal.Stderr, "工单已取消")
			return f.Print(resp.Data)
		},
	}
}

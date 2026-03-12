package data

import (
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// --- get-shop ---

type getShopOpts struct {
	f      *internal.Factory
	ShopID string
}

func newCmdGetShop(f *internal.Factory) *cobra.Command {
	o := &getShopOpts{f: f}
	cmd := &cobra.Command{
		Use:     "get-shop",
		Short:   "获取店铺信息",
		Example: `  ahcli data get-shop --shop-id 12345`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ShopID == "" {
				return fmt.Errorf("--shop-id 为必填参数")
			}
			return getShopRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ShopID, "shop-id", "", "店铺 ID")
	return cmd
}

func getShopRun(o *getShopOpts) error {
	client := o.f.DataSyncerClient()
	resp, err := client.Get(fmt.Sprintf("/api/v1/shops/%s", o.ShopID), nil)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

// --- list-orders ---

type listOrdersOpts struct {
	f                 *internal.Factory
	ShopID            string
	Limit             int
	Cursor            string
	FinancialStatus   string
	FulfillmentStatus string
}

func newCmdListOrders(f *internal.Factory) *cobra.Command {
	o := &listOrdersOpts{f: f}
	cmd := &cobra.Command{
		Use:   "list-orders",
		Short: "列出店铺订单",
		Example: `  ahcli data list-orders --shop-id 12345
  ahcli data list-orders --shop-id 12345 --financial-status paid --limit 20
  ahcli data list-orders --shop-id 12345 --cursor eyJsYX...  # 翻页`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.ShopID == "" {
				return fmt.Errorf("--shop-id 为必填参数")
			}
			return listOrdersRun(o)
		},
	}
	cmd.Flags().StringVar(&o.ShopID, "shop-id", "", "店铺 ID")
	cmd.Flags().IntVar(&o.Limit, "limit", 10, "每页数量（最大 100）")
	cmd.Flags().StringVar(&o.Cursor, "cursor", "", "分页游标（从上次响应的 next_cursor 获取）")
	cmd.Flags().StringVar(&o.FinancialStatus, "financial-status", "", "按支付状态过滤（如 paid）")
	cmd.Flags().StringVar(&o.FulfillmentStatus, "fulfillment-status", "", "按履约状态过滤")
	return cmd
}

func listOrdersRun(o *listOrdersOpts) error {
	client := o.f.DataSyncerClient()
	params := url.Values{
		"limit": {fmt.Sprintf("%d", o.Limit)},
	}
	if o.Cursor != "" {
		params.Set("cursor", o.Cursor)
	}
	if o.FinancialStatus != "" {
		params.Set("financial_status", o.FinancialStatus)
	}
	if o.FulfillmentStatus != "" {
		params.Set("fulfillment_status", o.FulfillmentStatus)
	}

	resp, err := client.Get(fmt.Sprintf("/api/v1/shops/%s/orders", o.ShopID), params)
	if err != nil {
		return err
	}
	return o.f.Print(resp.Data)
}

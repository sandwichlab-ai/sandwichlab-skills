package data

import (
	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// NewCmdData 创建 DataSyncer 服务子命令组。
func NewCmdData(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "data",
		Short: "DataSyncer 数据同步服务命令",
		Long: `调用 DataSyncer 数据同步服务的 API，查询 Shoplazza 店铺、订单等数据。

可用命令:
  get-shop      根据 ID 获取店铺信息
  list-orders   列出店铺订单（游标分页，支持状态过滤）

注意: DataSyncer 使用游标分页（cursor），而非 offset/limit 分页。
首次请求不传 cursor，后续使用响应中的 next_cursor 翻页。

示例:
  ahcli data get-shop --shop-id 12345
  ahcli data list-orders --shop-id 12345 --financial-status paid
  ahcli data list-orders --shop-id 12345 --cursor eyJsYX...`,
	}

	cmd.AddCommand(newCmdGetShop(f))
	cmd.AddCommand(newCmdListOrders(f))

	return cmd
}

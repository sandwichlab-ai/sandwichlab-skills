package cmd

import (
	"sandwichlab_core/tools/ahcli/internal"

	"github.com/spf13/cobra"
)

// NewCmdOps 创建 OpsCore 服务子命令组。
func NewCmdOps(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ops",
		Short: "OpsCore 运营管理命令",
		Long: `调用 OpsCore 运营管理服务的 API，包括资产分配、凭证获取等。

可用命令:
  get-allocation    获取用户的资产分配
  get-asset         获取 Asset 详情
  get-access-config 获取 Asset 的 AccessToken
  get-credentials   一次性获取投放所需的全部凭证（组合命令）

示例:
  ahcli ops get-credentials --user-id usr-xxx --tenant-id tnt-xxx
  ahcli ops get-allocation --user-id usr-xxx --tenant-id tnt-xxx
  ahcli ops get-asset --asset-code asset-001`,
	}

	cmd.AddCommand(newCmdGetAllocation(f))
	cmd.AddCommand(newCmdGetAsset(f))
	cmd.AddCommand(newCmdGetAccessConfig(f))
	cmd.AddCommand(newCmdGetCredentials(f))

	return cmd
}

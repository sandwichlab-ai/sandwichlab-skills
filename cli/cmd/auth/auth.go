package auth

import (
	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// NewCmdAuth 创建认证管理子命令组。
func NewCmdAuth(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "认证管理（登录、登出、租户切换）",
		Long: `认证管理，凭证持久化到 ~/.config/ahcli/hosts.yml。

登录后 tenant_id 和 user_id 自动注入后续命令，无需重复指定。

示例:
  ahcli auth login --env dev --hui-user 1
  ahcli auth status
  ahcli auth switch tnt-xxx
  ahcli auth tenants
  ahcli auth logout`,
	}

	cmd.AddCommand(newCmdLogin(f))
	cmd.AddCommand(newCmdLogout(f))
	cmd.AddCommand(newCmdAuthStatus(f))
	cmd.AddCommand(NewCmdTenant(f))
	cmd.AddCommand(newCmdAuthSwitch(f))
	cmd.AddCommand(newCmdAuthTenants(f))
	cmd.AddCommand(newCmdOpenDashboard(f))

	return cmd
}

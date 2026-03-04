package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// --- tenant list ---

func newCmdTenantList(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "列出所有租户",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := internal.LoadTenantConfig()
			if err != nil {
				return err
			}

			if len(config.Tenants) == 0 {
				fmt.Fprintln(internal.Stderr, "没有已保存的租户，使用 'ahcli auth tenant add' 添加")
			}

			// JSON 输出
			result := map[string]interface{}{
				"current": config.Current,
				"tenants": config.Tenants,
			}
			data, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("failed to marshal result: %w", err)
			}
			return f.Print(data)
		},
	}
}

// --- tenant add ---

type tenantAddOpts struct {
	f        *internal.Factory
	Name     string
	TenantID string
	UserID   string
}

func newCmdTenantAdd(f *internal.Factory) *cobra.Command {
	o := &tenantAddOpts{f: f}
	cmd := &cobra.Command{
		Use:     "add",
		Short:   "添加租户",
		Example: `  ahcli auth tenant add --name "大客户1" --tenant-id tnt-123 --user-id usr-456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.Name == "" || o.TenantID == "" || o.UserID == "" {
				return fmt.Errorf("--name, --tenant-id, --user-id 均为必填参数")
			}

			tenant := internal.Tenant{
				Name:     o.Name,
				TenantID: o.TenantID,
				UserID:   o.UserID,
			}

			if err := internal.AddTenant(tenant); err != nil {
				return err
			}

			fmt.Fprintf(internal.Stderr, "✓ 已添加租户: %s\n", o.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&o.Name, "name", "", "租户名称（必填）")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID（必填）")
	cmd.Flags().StringVar(&o.UserID, "user-id", "", "用户 ID（必填）")
	return cmd
}

// --- tenant remove ---

func newCmdTenantRemove(_ *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "删除租户",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := internal.RemoveTenant(name); err != nil {
				return err
			}
			fmt.Fprintf(internal.Stderr, "✓ 已删除租户: %s\n", name)
			return nil
		},
	}
}

// --- tenant use ---

func newCmdTenantUse(_ *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "切换当前租户",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := internal.UseTenant(name); err != nil {
				return err
			}

			// 获取租户信息显示
			tenant, err := internal.GetCurrentTenant()
			if err != nil {
				return fmt.Errorf("failed to get tenant info: %w", err)
			}
			fmt.Fprintf(internal.Stderr, "✓ 已切换到租户: %s\n", name)
			if tenant != nil {
				fmt.Fprintf(internal.Stderr, "  tenant_id: %s\n", tenant.TenantID)
				fmt.Fprintf(internal.Stderr, "  user_id: %s\n", tenant.UserID)
			}
			return nil
		},
	}
}

// --- tenant current ---

func newCmdTenantCurrent(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "显示当前租户",
		RunE: func(cmd *cobra.Command, args []string) error {
			tenant, err := internal.GetCurrentTenant()
			if err != nil {
				return err
			}

			if tenant == nil {
				fmt.Fprintln(internal.Stderr, "未选择租户，使用 'ahcli auth tenant use <name>' 切换")
				result := map[string]interface{}{"current": nil}
				data, marshalErr := json.Marshal(result)
				if marshalErr != nil {
					return fmt.Errorf("failed to marshal result: %w", marshalErr)
				}
				return f.Print(data)
			}

			data, err := json.Marshal(tenant)
			if err != nil {
				return fmt.Errorf("failed to marshal tenant: %w", err)
			}
			return f.Print(data)
		},
	}
}

// NewCmdTenant 创建租户管理子命令组。
func NewCmdTenant(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "租户管理（添加、删除、切换）",
		Long: `管理多个租户配置，支持快速切换。

切换租户后，所有 API 调用会自动携带 tenant_id 和 user_id 参数。

示例:
  ahcli auth tenant list
  ahcli auth tenant add --name "大客户1" --tenant-id tnt-123 --user-id usr-456
  ahcli auth tenant use "大客户1"
  ahcli auth tenant current
  ahcli auth tenant remove "大客户1"`,
	}

	cmd.AddCommand(newCmdTenantList(f))
	cmd.AddCommand(newCmdTenantAdd(f))
	cmd.AddCommand(newCmdTenantRemove(f))
	cmd.AddCommand(newCmdTenantUse(f))
	cmd.AddCommand(newCmdTenantCurrent(f))

	return cmd
}

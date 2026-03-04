package cmd

import (
	"fmt"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

func newCmdAuthSwitch(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "switch <tenant-id>",
		Short:   "切换活跃租户",
		Args:    cobra.ExactArgs(1),
		Example: `  ahcli auth switch tnt-yyy`,
		RunE: func(cmd *cobra.Command, args []string) error {
			newTenantID := args[0]

			sess, err := internal.LoadSession(f.Env)
			if err != nil {
				return fmt.Errorf("failed to load session: %w", err)
			}
			if sess == nil {
				return fmt.Errorf("未登录（%s 环境），请先执行 'ahcli auth login'", f.Env)
			}

			oldTenantID := sess.TenantID
			sess.TenantID = newTenantID
			if err := internal.SaveSession(f.Env, sess); err != nil {
				return fmt.Errorf("failed to save session: %w", err)
			}

			fmt.Fprintf(internal.Stderr, "✓ 租户已切换: %s → %s\n", oldTenantID, newTenantID)
			return nil
		},
	}
}

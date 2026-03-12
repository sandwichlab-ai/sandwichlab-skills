package auth

import (
	"encoding/json"
	"fmt"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

func newCmdAuthTenants(f *internal.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "tenants",
		Short:   "列出所有已保存的登录环境",
		Example: `  ahcli auth tenants`,
		RunE: func(cmd *cobra.Command, args []string) error {
			sessions, err := internal.ListSessions()
			if err != nil {
				return fmt.Errorf("failed to list sessions: %w", err)
			}

			if len(sessions) == 0 {
				fmt.Fprintln(internal.Stderr, "没有已保存的登录信息")
				emptyData, marshalErr := json.Marshal(map[string]interface{}{"sessions": []interface{}{}})
				if marshalErr != nil {
					return fmt.Errorf("failed to marshal result: %w", marshalErr)
				}
				return f.Print(emptyData)
			}

			result := make([]map[string]string, 0, len(sessions))
			for envName, sess := range sessions {
				entry := map[string]string{
					"env":       envName,
					"tenant_id": sess.TenantID,
					"user_id":   sess.UserID,
					"email":     sess.Email,
				}
				if envName == f.Env {
					entry["active"] = "true"
				}
				result = append(result, entry)
			}

			data, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("failed to marshal result: %w", err)
			}
			return f.Print(data)
		},
	}
}

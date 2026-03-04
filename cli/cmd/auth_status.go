package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

func newCmdAuthStatus(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "查看当前登录状态",
		Long: `显示当前保存的 Cognito 凭证信息，包括用户、环境、Token 过期时间等。

示例:
  ahcli auth status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			creds, err := internal.LoadCredentials()
			if err != nil {
				return fmt.Errorf("failed to load credentials: %w", err)
			}

			if creds == nil {
				fmt.Fprintf(internal.Stderr, "未登录\n")
				fmt.Fprintf(internal.Stderr, "使用 'ahcli auth login --env dev' 登录\n")
				result := map[string]interface{}{
					"logged_in": false,
				}
				data, marshalErr := json.Marshal(result)
				if marshalErr != nil {
					return fmt.Errorf("failed to marshal result: %w", marshalErr)
				}
				return f.Print(data)
			}

			// 检查是否过期
			expired := creds.IsExpired()
			timeUntilExpiry := time.Until(creds.ExpiresAt)

			result := map[string]interface{}{
				"logged_in":      true,
				"environment":    creds.Environment,
				"user_id":        creds.UserID,
				"email":          creds.Email,
				"email_verified": creds.EmailVerified,
				"expires_at":     creds.ExpiresAt.Format(time.RFC3339),
				"expired":        expired,
			}

			if !expired {
				result["time_until_expiry"] = timeUntilExpiry.String()
			}

			data, err := json.Marshal(result)
			if err != nil {
				return fmt.Errorf("failed to marshal result: %w", err)
			}

			fmt.Fprintf(internal.Stderr, "✓ 已登录\n")
			fmt.Fprintf(internal.Stderr, "环境: %s\n", creds.Environment)
			fmt.Fprintf(internal.Stderr, "用户: %s (%s)\n", creds.Email, creds.UserID)
			fmt.Fprintf(internal.Stderr, "Token 过期时间: %s\n", creds.ExpiresAt.Format(time.RFC3339))
			if expired {
				fmt.Fprintf(internal.Stderr, "⚠️  Token 已过期，请重新登录\n")
			} else {
				fmt.Fprintf(internal.Stderr, "剩余有效期: %s\n", timeUntilExpiry.Round(time.Second))
			}

			return f.Print(data)
		},
	}
	return cmd
}

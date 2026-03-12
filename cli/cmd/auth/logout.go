package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

func newCmdLogout(_ *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "登出并清除凭证",
		Long: `清除本地保存的 Cognito 凭证，并在浏览器中登出 Cognito 会话。

示例:
  ahcli auth logout`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 先加载凭证，获取环境信息
			creds, err := internal.LoadCredentials()
			if err != nil {
				return fmt.Errorf("failed to load credentials: %w", err)
			}

			if creds == nil {
				fmt.Fprintf(internal.Stderr, "未登录，无需登出\n")
				return nil
			}

			// 获取 Cognito 配置
			cognitoConfig, err := internal.LoadCognitoConfig(creds.Environment)
			if err != nil {
				fmt.Fprintf(internal.Stderr, "警告：无法获取 Cognito 配置: %v\n", err)
			} else {
				// 启动本地服务器接收 logout 回调
				logoutURI := fmt.Sprintf("http://localhost:%d/logout-complete", cognitoConfig.CallbackPort)

				doneChan := make(chan struct{})
				mux := http.NewServeMux()
				mux.HandleFunc("/logout-complete", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					if err := writeLogoutPage(w); err != nil {
						fmt.Fprintf(internal.Stderr, "警告：写入登出页面失败: %v\n", err)
					}
					close(doneChan)
				})

				server := &http.Server{
					Addr:    fmt.Sprintf(":%d", cognitoConfig.CallbackPort),
					Handler: mux,
				}

				go func() {
					_ = server.ListenAndServe() //nolint:errcheck // fire-and-forget server
				}()

				time.Sleep(200 * time.Millisecond)

				params := url.Values{
					"client_id":  {cognitoConfig.ClientID},
					"logout_uri": {logoutURI},
				}
				logoutURL := fmt.Sprintf("https://%s/logout?%s", cognitoConfig.Domain, params.Encode())

				fmt.Fprintf(internal.Stderr, "正在浏览器中登出 Cognito 会话...\n")
				if err := internal.OpenBrowser(logoutURL); err != nil {
					fmt.Fprintf(internal.Stderr, "警告：无法打开浏览器: %v\n", err)
				}

				select {
				case <-doneChan:
					fmt.Fprintf(internal.Stderr, "✓ Cognito 会话已清除\n")
				case <-time.After(30 * time.Second):
					fmt.Fprintf(internal.Stderr, "警告：等待 Cognito 回调超时\n")
				}

				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				_ = server.Shutdown(ctx) //nolint:errcheck // best-effort shutdown
			}

			if err := internal.ClearCredentials(); err != nil {
				return fmt.Errorf("failed to clear credentials: %w", err)
			}

			// 删除以邮箱命名的默认租户
			if creds.Email != "" {
				_ = internal.RemoveTenant(creds.Email) //nolint:errcheck // best-effort cleanup
			}

			fmt.Fprintf(internal.Stderr, "✓ 已登出，本地凭证已清除\n")
			return nil
		},
	}
	return cmd
}

const logoutPageHTML = `<!DOCTYPE html>
<html>
<head><title>ahcli - Logout</title>
<style>
body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
.success { color: #28a745; font-size: 24px; margin-bottom: 20px; }
.message { color: #666; font-size: 16px; }
</style>
</head>
<body>
<div class="success">✓ Logout Successful</div>
<div class="message">You can close this window.</div>
</body>
</html>`

func writeLogoutPage(w http.ResponseWriter) error {
	_, err := w.Write([]byte(logoutPageHTML))
	return err
}

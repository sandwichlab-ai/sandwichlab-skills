package auth

import (
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

func newCmdOpenDashboard(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open-dashboard",
		Short: "在浏览器中打开 HUI Dashboard（免登录）",
		Long: `使用已保存的 HUI JWT Token 打开浏览器，自动登录 Dashboard。

需要先通过 'ahcli auth login' 登录。如果 HUI Token 已过期，
会自动使用 Cognito ID Token 重新交换。

示例:
  ahcli auth open-dashboard`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return openDashboardRun(f)
		},
	}
	return cmd
}

func openDashboardRun(f *internal.Factory) error {
	creds, err := internal.LoadCredentials()
	if err != nil {
		return fmt.Errorf("无法加载凭证: %w", err)
	}
	if creds == nil {
		return fmt.Errorf("未登录，请先运行 'ahcli auth login'")
	}

	// 如果 HUI Token 过期，尝试重新 exchange
	if creds.IsHUITokenExpired() {
		if creds.IsExpired() {
			return fmt.Errorf("Cognito Token 已过期，请先运行 'ahcli auth login' 重新登录")
		}
		fmt.Fprintf(internal.Stderr, "HUI Token 已过期，正在重新获取...\n")
		internal.ExchangeHUIToken(f, creds)
		if creds.IsHUITokenExpired() {
			return fmt.Errorf("无法获取 HUI Token，请检查 HUI 服务是否可用")
		}
	}

	// 获取前端 URL
	feURL, ok := internal.FrontendURLs[f.Env]
	if !ok {
		return fmt.Errorf("未知环境: %s", f.Env)
	}

	// 构建带 token 的 URL
	dashboardURL := fmt.Sprintf("%s/auth/token?t=%s", feURL, url.QueryEscape(creds.HUIToken))

	fmt.Fprintf(internal.Stderr, "正在打开 HUI Dashboard...\n")
	if err := internal.OpenBrowser(dashboardURL); err != nil {
		// 打开失败时打印 URL 让用户手动访问
		fmt.Fprintf(internal.Stderr, "无法自动打开浏览器，请手动访问：\n%s\n", dashboardURL)
		return nil
	}

	fmt.Fprintf(internal.Stderr, "✓ 已打开 HUI Dashboard\n")
	return nil
}

package ads

import (
	"fmt"
	"net/url"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

// openMonitorPage 打开外部盯盘页面，使用 Cognito ID Token 认证。
func openMonitorPage(f *internal.Factory, projectID string) error {
	creds, err := internal.LoadCredentials()
	if err != nil {
		return fmt.Errorf("无法加载凭证: %w", err)
	}
	if creds == nil || creds.IDToken == "" {
		return fmt.Errorf("未登录，请先运行 'ahcli auth login'")
	}
	if creds.IsExpired() {
		return fmt.Errorf("Cognito Token 已过期，请先运行 'ahcli auth login' 重新登录")
	}

	feURL, ok := internal.FrontendURLs[f.Env]
	if !ok {
		return fmt.Errorf("未知环境: %s", f.Env)
	}

	target := fmt.Sprintf("%s/open/dashboard?token=%s", feURL, url.QueryEscape(creds.IDToken))
	if f.TenantID() != "" {
		target += "&tenant_id=" + url.QueryEscape(f.TenantID())
	}
	if projectID != "" {
		target += "&project=" + url.QueryEscape(projectID)
	}

	fmt.Fprintf(internal.Stderr, "正在打开盯盘页面...\n")
	if err := internal.OpenBrowser(target); err != nil {
		fmt.Fprintf(internal.Stderr, "无法自动打开浏览器，请手动访问：\n%s\n", target)
		return nil
	}

	fmt.Fprintf(internal.Stderr, "✓ 已打开盯盘页面\n")
	return nil
}

func newCmdMonitorOpen(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [project-id]",
		Short: "在浏览器中打开盯盘页面",
		Example: `  ahcli ads monitor open
  ahcli ads monitor open proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := ""
			if len(args) > 0 {
				projectID = args[0]
			}
			return openMonitorPage(f, projectID)
		},
	}
	return cmd
}

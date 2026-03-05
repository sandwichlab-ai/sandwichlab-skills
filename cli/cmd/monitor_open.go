package cmd

import (
	"fmt"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

// openMonitorPage 打开前端盯盘页面，可选深链接到特定项目。
func openMonitorPage(f *internal.Factory, projectID string) {
	feURL, ok := frontendURLs[f.Env]
	if !ok {
		return
	}

	target := feURL + "/monitor"
	if projectID != "" {
		target += "?project=" + projectID
	}

	fmt.Fprintf(internal.Stderr, "正在打开: %s\n", target)
	if err := openBrowser(target); err != nil {
		fmt.Fprintf(internal.Stderr, "无法自动打开浏览器，请手动访问：\n%s\n", target)
	}
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
			openMonitorPage(f, projectID)
			return nil
		},
	}
	return cmd
}

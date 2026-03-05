package cmd

import (
	"fmt"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
)

func newCmdMonitorOpen(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "open [project-id]",
		Short: "在浏览器中打开盯盘页面",
		Example: `  ahcli ads monitor open
  ahcli ads monitor open proj-xxx`,
		RunE: func(cmd *cobra.Command, args []string) error {
			feURL, ok := frontendURLs[f.Env]
			if !ok {
				return fmt.Errorf("未知环境: %s", f.Env)
			}

			target := feURL + "/monitor"
			if len(args) > 0 {
				target += "?project=" + args[0]
			}

			fmt.Fprintf(internal.Stderr, "正在打开: %s\n", target)
			if err := openBrowser(target); err != nil {
				fmt.Fprintf(internal.Stderr, "无法自动打开浏览器，请手动访问：\n%s\n", target)
				return nil
			}
			return nil
		},
	}
	return cmd
}

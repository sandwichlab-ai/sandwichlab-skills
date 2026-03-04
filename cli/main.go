// ahcli 是 ActionHub 下游服务的直调 CLI 工具。
// 绕过 MCP 协议，直接调用 AdsCore、DataSyncer、BrowserService 的 HTTP API，
// 用于调试和脚本化场景。
//
// 使用示例:
//
//	go build -o ahcli ./tools/ahcli
//	./ahcli --env dev ads project list
//	./ahcli browser screenshot --url https://example.com --save out.png
package main

import (
	"os"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

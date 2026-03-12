// ahcli 是广告服务的命令行工具。
// 直接调用 AdsCore、DataSyncer、BrowserService、OpsCore 的 HTTP API，
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

// Version is set at build time via -ldflags "-X main.Version=vX.Y.Z"
var Version = "dev"

func main() {
	if err := cmd.Execute(Version); err != nil {
		os.Exit(1)
	}
}

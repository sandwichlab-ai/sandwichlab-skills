package cmd

import (
	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

// NewCmdAds 创建广告服务子命令组。
func NewCmdAds(f *internal.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ads",
		Short: "广告服务命令（项目、素材、创意、投放等）",
		Long: `广告服务管理，按实体组织子命令，覆盖广告投放完整生命周期。

实体命令:
  project    项目管理（含 channel / attachment 子命令）
  creative   创意管理
  copy       文案管理
  media      素材库管理
  action     投放动作（Campaign 创建/管理）
  plan       投放计划
  channel    渠道查询（Meta 等）

示例:
  ahcli ads project list
  ahcli ads creative create --file creative.json
  ahcli ads action create-sync --file campaign.json
  ahcli ads media search --project-id proj-xxx
  ahcli ads channel meta campaign list`,
	}

	cmd.AddCommand(NewCmdProject(f))
	cmd.AddCommand(NewCmdCreative(f))
	cmd.AddCommand(NewCmdCopy(f))
	cmd.AddCommand(NewCmdMedia(f))
	cmd.AddCommand(NewCmdAction(f))
	cmd.AddCommand(NewCmdPlan(f))
	cmd.AddCommand(NewCmdChannelMeta(f))
	cmd.AddCommand(NewCmdMonitor(f))

	return cmd
}

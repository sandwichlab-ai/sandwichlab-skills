// Package cmd 定义所有 CLI 命令。
// 使用 Cobra 框架，采用 Options struct + Factory 依赖注入模式。
package cmd

import (
	"fmt"
	"os"
	"strings"

	"sandwichlab_core/tools/ahcli/internal"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewRootCmd 创建 CLI 根命令，所有子命令挂载在其下。
// Factory 在 PersistentPreRunE 中初始化，通过 NewCmd* 函数传递给每个子命令。
func NewRootCmd() *cobra.Command {
	f := &internal.Factory{}
	var cfgFile string

	rootCmd := &cobra.Command{
		Use:   "ahcli",
		Short: "ActionHub CLI — 广告服务管理工具",
		Long: `ahcli 是 ActionHub 生态的命令行工具，按实体组织子命令，覆盖广告投放完整生命周期。

支持四个环境：local / dev / preprod / prod，通过 --env 切换。
登录后 tenant_id、user_id 自动注入，无需重复输入。

示例:
  ahcli auth login --env dev
  ahcli ads project list
  ahcli ads creative create --file creative.json
  ahcli ads action create-sync --file campaign.json`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rootPreRun(f, cmd)
		},
	}

	// 全局持久化 flags，绑定到 Factory
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径（默认 ./.ahcli.yaml）")
	rootCmd.PersistentFlags().StringVarP(&f.Env, "env", "e", "dev", "目标环境 (local|dev|preprod|prod)")
	rootCmd.PersistentFlags().BoolVarP(&f.Verbose, "verbose", "v", false, "输出调试信息")
	rootCmd.PersistentFlags().BoolVarP(&f.Compact, "compact", "c", false, "紧凑 JSON 输出（方便 pipe 给 jq）")

	cobra.OnInitialize(func() {
		initConfig(cfgFile, f.Verbose)
	})

	// 注册命令树，传入 Factory
	rootCmd.AddCommand(NewCmdAds(f))
	rootCmd.AddCommand(NewCmdData(f))
	rootCmd.AddCommand(NewCmdBrowser(f))
	rootCmd.AddCommand(NewCmdOps(f))
	rootCmd.AddCommand(NewCmdAuth(f))

	return rootCmd
}

// rootPreRun 初始化 Factory：解析 URL、加载凭证和租户。
func rootPreRun(f *internal.Factory, cmd *cobra.Command) error {
	overrides := map[string]string{
		"adscore_url":     viper.GetString("adscore_url"),
		"data_syncer_url": viper.GetString("data_syncer_url"),
		"browser_url":     viper.GetString("browser_url"),
		"opscore_url":     viper.GetString("opscore_url"),
		"authcenter_url":  viper.GetString("authcenter_url"),
		"actionhub_url":   viper.GetString("actionhub_url"),
	}

	var err error
	f.URLs, err = internal.ResolveURLs(f.Env, overrides)
	if err != nil {
		return err
	}

	loadCredentials(f)
	logVerboseState(f)
	loadTenant(f)

	// 检查登录状态（排除 auth 命令）
	cmdPath := cmd.CommandPath()
	if !strings.HasPrefix(cmdPath, "ahcli auth") && f.Credentials == nil {
		return fmt.Errorf("未登录，请先运行: ahcli auth login --env %s", f.Env)
	}

	return nil
}

// loadCredentials 加载 Cognito 凭证（静默，不影响无需认证的命令）。
func loadCredentials(f *internal.Factory) {
	cognitoConfig, err := loadCognitoConfig(f.Env)
	if err != nil {
		if f.Verbose {
			fmt.Fprintf(os.Stderr, "[verbose] failed to load Cognito config: %v\n", err)
		}
		return
	}

	creds, refreshed, loadErr := internal.LoadAndRefreshCredentials(f.Env, cognitoConfig)
	if loadErr != nil {
		if f.Verbose {
			fmt.Fprintf(os.Stderr, "[verbose] failed to load credentials: %v\n", loadErr)
		}
		return
	}
	if creds != nil {
		f.Credentials = creds
		if refreshed && f.Verbose {
			fmt.Fprintf(os.Stderr, "[verbose] Token was expired and has been automatically refreshed\n")
		}
	}
}

// logVerboseState 在 verbose 模式下输出当前状态。
func logVerboseState(f *internal.Factory) {
	if !f.Verbose {
		return
	}
	fmt.Fprintf(os.Stderr, "[verbose] config=%s env=%s\n", viper.ConfigFileUsed(), f.Env)
	fmt.Fprintf(os.Stderr, "[verbose] adscore=%s data_syncer=%s browser=%s\n",
		f.URLs.AdsCore, f.URLs.DataSyncer, f.URLs.Browser)
	fmt.Fprintf(os.Stderr, "[verbose] opscore=%s authcenter=%s actionhub=%s\n",
		f.URLs.OpsCore, f.URLs.AuthCenter, f.URLs.ActionHub)
	if f.Credentials != nil {
		fmt.Fprintf(os.Stderr, "[verbose] Cognito: logged in as %s (%s)\n",
			f.Credentials.Email, f.Credentials.UserID)
	} else {
		fmt.Fprintf(os.Stderr, "[verbose] auth: not logged in\n")
	}
}

// loadTenant 加载当前租户。
func loadTenant(f *internal.Factory) {
	tenant, err := internal.GetCurrentTenant()
	if err != nil {
		if f.Verbose {
			fmt.Fprintf(os.Stderr, "[verbose] failed to load tenant: %v\n", err)
		}
		return
	}
	f.CurrentTenant = tenant
	if f.Verbose && tenant != nil {
		fmt.Fprintf(os.Stderr, "[verbose] tenant: %s (tenant_id=%s, user_id=%s)\n",
			tenant.Name, tenant.TenantID, tenant.UserID)
	}
}

// initConfig 初始化 Viper 配置。
// 配置优先级：命令行 flag > 环境变量 (AHCLI_*) > 配置文件
func initConfig(cfgFile string, verbose bool) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".ahcli")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
	}

	viper.SetEnvPrefix("AHCLI")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "[verbose] 已加载配置文件: %s\n", viper.ConfigFileUsed())
		}
	}
}

// Execute 是 CLI 的入口函数，由 main.go 调用。
func Execute() error {
	return NewRootCmd().Execute()
}

// splitAndTrim 将逗号分隔的字符串拆分为切片，去除空白。
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// validBindActions 是 bind 命令 --action flag 的合法值。
var validBindActions = map[string]bool{"add": true, "remove": true}

// validateBindAction 校验 bind 命令的 --action 参数值。
func validateBindAction(action string) error {
	if !validBindActions[action] {
		return fmt.Errorf("--action 值无效: %q（可选: add, remove）", action)
	}
	return nil
}

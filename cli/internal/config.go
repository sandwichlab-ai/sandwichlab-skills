// Package internal 包含 ahcli 的内部实现：HTTP 客户端、配置解析、输出格式化。
// 不依赖 pkg/infra，仅使用标准库，避免引入 otel/logger/tracing 初始化。
package internal

import "fmt"

// ServiceURLs 保存各下游服务的基础 URL
type ServiceURLs struct {
	AdsCore    string // 广告核心服务
	DataSyncer string // 数据同步服务（Shoplazza 店铺/订单等）
	Browser    string // 浏览器截图服务
	OpsCore    string // 运营管理服务（资产分配、Pixel 管理）
	AuthCenter string // 认证中心（Session 管理）
	ActionHub  string // ActionHub JSON-RPC 网关
	HUI        string // HUI 管理后台服务
}

// envDefaults 各环境的默认服务 URL。
// local: 本地开发端口（参考 CLAUDE.md 端口分配）
// dev/preprod/prod: K8s 集群内部 DNS 地址（需要 VPN 或集群网络访问）
var envDefaults = map[string]ServiceURLs{
	"local": {
		AdsCore:    "http://localhost:8083",
		DataSyncer: "http://localhost:8086",
		Browser:    "http://localhost:8090",
		OpsCore:    "http://localhost:8088",
		AuthCenter: "http://localhost:8080",
		ActionHub:  "http://localhost:8085",
		HUI:        "http://localhost:8089",
	},
	"dev": {
		AdsCore:    "https://api.dev-us.sandwichlab.ai/adscore",
		DataSyncer: "https://api.dev-us.sandwichlab.ai/data-syncer",
		Browser:    "https://api.dev-us.sandwichlab.ai/browser",
		OpsCore:    "https://api.dev-us.sandwichlab.ai/opscore",
		AuthCenter: "https://api.dev-us.sandwichlab.ai/authcenter",
		ActionHub:  "https://api.dev-us.sandwichlab.ai/actionhub",
		HUI:        "https://api.dev-us.sandwichlab.ai/hui",
	},
	"preprod": {
		AdsCore:    "https://api.preprod-us.sandwichlab.ai/adscore",
		DataSyncer: "https://api.preprod-us.sandwichlab.ai/data-syncer",
		Browser:    "https://api.preprod-us.sandwichlab.ai/browser",
		OpsCore:    "https://api.preprod-us.sandwichlab.ai/opscore",
		AuthCenter: "https://api.preprod-us.sandwichlab.ai/authcenter",
		ActionHub:  "https://api.preprod-us.sandwichlab.ai/actionhub",
		HUI:        "https://api.preprod-us.sandwichlab.ai/hui",
	},
	"prod": {
		AdsCore:    "https://api.sandwichlab.ai/adscore",
		DataSyncer: "https://api.sandwichlab.ai/data-syncer",
		Browser:    "https://api.sandwichlab.ai/browser",
		OpsCore:    "https://api.sandwichlab.ai/opscore",
		AuthCenter: "https://api.sandwichlab.ai/authcenter",
		ActionHub:  "https://api.sandwichlab.ai/actionhub",
		HUI:        "https://api.sandwichlab.ai/hui",
	},
}

// FrontendURLs 各环境前端 Dashboard URL
var FrontendURLs = map[string]string{
	"local":   "http://localhost:5175",
	"dev":     "https://hui.lanbow.ai",
	"preprod": "https://hui.lanbow.ai",
	"prod":    "https://hui.lanbow.ai",
}

// ResolveURLs 根据环境名解析服务 URL。
// 优先使用 overrides 中的值（来自 Viper 配置或环境变量），否则使用环境默认值。
func ResolveURLs(env string, overrides map[string]string) (*ServiceURLs, error) {
	defaults, ok := envDefaults[env]
	if !ok {
		return nil, fmt.Errorf("unknown environment: %s (valid: local, dev, preprod, prod)", env)
	}

	urls := defaults

	// 逐个检查覆盖值，非空时替换默认 URL
	if v, ok := overrides["adscore_url"]; ok && v != "" {
		urls.AdsCore = v
	}
	if v, ok := overrides["data_syncer_url"]; ok && v != "" {
		urls.DataSyncer = v
	}
	if v, ok := overrides["browser_url"]; ok && v != "" {
		urls.Browser = v
	}
	if v, ok := overrides["opscore_url"]; ok && v != "" {
		urls.OpsCore = v
	}
	if v, ok := overrides["authcenter_url"]; ok && v != "" {
		urls.AuthCenter = v
	}
	if v, ok := overrides["actionhub_url"]; ok && v != "" {
		urls.ActionHub = v
	}
	if v, ok := overrides["hui_url"]; ok && v != "" {
		urls.HUI = v
	}

	return &urls, nil
}

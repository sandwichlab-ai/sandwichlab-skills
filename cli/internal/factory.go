package internal

import (
	"encoding/json"
	"fmt"
)

// Factory 集中管理所有共享依赖，替代全局变量。
// 由 NewRootCmd 创建，通过 NewCmd* 函数传递给每个子命令。
type Factory struct {
	URLs          *ServiceURLs // 解析后的各服务 URL
	Credentials   *Credentials // Cognito 凭证（可为 nil）
	CurrentTenant *Tenant      // 当前选中的租户（可为 nil）
	Env           string       // 目标环境：local / dev / preprod / prod
	Verbose       bool         // 是否输出调试信息
	Compact       bool         // 紧凑 JSON 输出
}

// idToken 从 Credentials 派生 ID Token。
func (f *Factory) idToken() string {
	if f.Credentials != nil {
		return f.Credentials.IDToken
	}
	return ""
}

// NewClient 创建指向指定服务的 HTTP 客户端，自动注入 token 和租户信息。
func (f *Factory) NewClient(baseURL string) *Client {
	client := NewClient(baseURL, f.Verbose)
	if token := f.idToken(); token != "" {
		client.SetIDToken(token)
	}
	if f.CurrentTenant != nil {
		client.SetTenant(f.CurrentTenant.TenantID, f.CurrentTenant.UserID)
	}
	return client
}

// AdsClient 返回 AdsCore 服务客户端。
func (f *Factory) AdsClient() *Client { return f.NewClient(f.URLs.AdsCore) }

// ActionHubClient 返回 ActionHub 服务客户端。
func (f *Factory) ActionHubClient() *Client { return f.NewClient(f.URLs.ActionHub) }

// OpsClient 返回 OpsCore 服务客户端。
func (f *Factory) OpsClient() *Client { return f.NewClient(f.URLs.OpsCore) }

// BrowserClient 返回 BrowserService 客户端。
func (f *Factory) BrowserClient() *Client { return f.NewClient(f.URLs.Browser) }

// DataSyncerClient 返回 DataSyncer 服务客户端。
func (f *Factory) DataSyncerClient() *Client { return f.NewClient(f.URLs.DataSyncer) }

// HUIClient 返回 HUI 后端服务客户端。
func (f *Factory) HUIClient() *Client { return f.NewClient(f.URLs.HUI) }

// AuthClient 返回 AuthCenter 服务客户端（不注入 token/tenant，用于登录流程）。
func (f *Factory) AuthClient() *Client { return NewClient(f.URLs.AuthCenter, f.Verbose) }

// TenantID 从当前租户获取 tenant_id，无租户时返回空。
func (f *Factory) TenantID() string {
	if f.CurrentTenant != nil {
		return f.CurrentTenant.TenantID
	}
	return ""
}

// UserID 从当前租户获取 user_id，无租户时返回空。
func (f *Factory) UserID() string {
	if f.CurrentTenant != nil {
		return f.CurrentTenant.UserID
	}
	return ""
}

// ResolveTenantID 返回 override（非空时）或当前租户的 tenant_id。
func (f *Factory) ResolveTenantID(override string) string {
	if override != "" {
		return override
	}
	return f.TenantID()
}

// RequireTenantID 解析 tenant_id 并写回 id 指针，为空则返回错误。
func (f *Factory) RequireTenantID(id *string) error {
	*id = f.ResolveTenantID(*id)
	if *id == "" {
		return fmt.Errorf("--tenant-id 为必填参数（或通过登录设置）")
	}
	return nil
}

// Print 统一输出 JSON 数据。
func (f *Factory) Print(data json.RawMessage) error {
	return PrintJSON(data, f.Compact)
}

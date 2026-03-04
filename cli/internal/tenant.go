package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Tenant 租户信息
type Tenant struct {
	Name     string `json:"name"`
	TenantID string `json:"tenant_id"`
	UserID   string `json:"user_id"`
}

// TenantConfig 租户配置文件结构
type TenantConfig struct {
	Current string   `json:"current"` // 当前选中的租户名称
	Tenants []Tenant `json:"tenants"`
}

// tenantConfigPath 返回租户配置文件路径：~/.ahcli/tenants.json
func tenantConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".ahcli", "tenants.json"), nil
}

// LoadTenantConfig 加载租户配置
func LoadTenantConfig() (*TenantConfig, error) {
	path, err := tenantConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TenantConfig{Tenants: []Tenant{}}, nil
		}
		return nil, fmt.Errorf("failed to read tenant config: %w", err)
	}

	var config TenantConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse tenant config: %w", err)
	}

	return &config, nil
}

// SaveTenantConfig 保存租户配置
func SaveTenantConfig(config *TenantConfig) error {
	path, err := tenantConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tenant config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write tenant config: %w", err)
	}

	return nil
}

// GetCurrentTenant 获取当前选中的租户
func GetCurrentTenant() (*Tenant, error) {
	config, err := LoadTenantConfig()
	if err != nil {
		return nil, err
	}

	if config.Current == "" {
		return nil, nil
	}

	for _, t := range config.Tenants {
		if t.Name == config.Current {
			return &t, nil
		}
	}

	return nil, nil
}

// AddTenant 添加租户
func AddTenant(tenant Tenant) error {
	config, err := LoadTenantConfig()
	if err != nil {
		return err
	}

	// 检查是否已存在
	for i, t := range config.Tenants {
		if t.Name == tenant.Name {
			// 更新已存在的租户
			config.Tenants[i] = tenant
			return SaveTenantConfig(config)
		}
	}

	config.Tenants = append(config.Tenants, tenant)
	return SaveTenantConfig(config)
}

// RemoveTenant 删除租户
func RemoveTenant(name string) error {
	config, err := LoadTenantConfig()
	if err != nil {
		return err
	}

	found := false
	newTenants := make([]Tenant, 0, len(config.Tenants))
	for _, t := range config.Tenants {
		if t.Name != name {
			newTenants = append(newTenants, t)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("租户 %q 不存在", name)
	}

	config.Tenants = newTenants

	// 如果删除的是当前租户，清空 current
	if config.Current == name {
		config.Current = ""
	}

	return SaveTenantConfig(config)
}

// UseTenant 切换当前租户
func UseTenant(name string) error {
	config, err := LoadTenantConfig()
	if err != nil {
		return err
	}

	// 检查租户是否存在
	found := false
	for _, t := range config.Tenants {
		if t.Name == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("租户 %q 不存在", name)
	}

	config.Current = name
	return SaveTenantConfig(config)
}

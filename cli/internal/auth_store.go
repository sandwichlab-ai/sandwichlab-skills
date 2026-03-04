package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Session 表示一个已登录的会话信息，存储在 hosts.yml 中。
type Session struct {
	TenantID string `yaml:"tenant_id" json:"tenant_id"`
	UserID   string `yaml:"user_id" json:"user_id"`
	Email    string `yaml:"email,omitempty" json:"email,omitempty"`
	Token    string `yaml:"token,omitempty" json:"token,omitempty"`
}

// HostsFile 是 hosts.yml 的顶层结构，按环境分隔。
type HostsFile struct {
	Hosts map[string]*Session `yaml:"hosts"`
}

// hostsFilePath 返回凭证文件路径：~/.config/ahcli/hosts.yml
func hostsFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".config", "ahcli", "hosts.yml"), nil
}

// loadHostsFile 读取并解析 hosts.yml。文件不存在时返回空结构。
func loadHostsFile() (*HostsFile, error) {
	path, err := hostsFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path) // #nosec G304 -- CLI tool, fixed path
	if err != nil {
		if os.IsNotExist(err) {
			return &HostsFile{Hosts: make(map[string]*Session)}, nil
		}
		return nil, fmt.Errorf("failed to read hosts file: %w", err)
	}

	var hosts HostsFile
	if err := yaml.Unmarshal(data, &hosts); err != nil {
		return nil, fmt.Errorf("failed to parse hosts file: %w", err)
	}
	if hosts.Hosts == nil {
		hosts.Hosts = make(map[string]*Session)
	}
	return &hosts, nil
}

// saveHostsFile 将 hosts 写回文件，确保目录存在且权限为 0600。
func saveHostsFile(hosts *HostsFile) error {
	path, err := hostsFilePath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if mkdirErr := os.MkdirAll(dir, 0700); mkdirErr != nil {
		return fmt.Errorf("failed to create config directory: %w", mkdirErr)
	}

	data, err := yaml.Marshal(hosts)
	if err != nil {
		return fmt.Errorf("failed to marshal hosts file: %w", err)
	}

	if writeErr := os.WriteFile(path, data, 0600); writeErr != nil {
		return fmt.Errorf("failed to write hosts file: %w", writeErr)
	}

	return nil
}

// LoadSession 加载指定环境的 session。不存在时返回 nil。
func LoadSession(env string) (*Session, error) {
	hosts, err := loadHostsFile()
	if err != nil {
		return nil, err
	}
	return hosts.Hosts[env], nil
}

// SaveSession 保存指定环境的 session。
func SaveSession(env string, session *Session) error {
	hosts, err := loadHostsFile()
	if err != nil {
		return err
	}
	hosts.Hosts[env] = session
	return saveHostsFile(hosts)
}

// ClearSession 清除指定环境的 session。
func ClearSession(env string) error {
	hosts, err := loadHostsFile()
	if err != nil {
		return err
	}
	delete(hosts.Hosts, env)
	return saveHostsFile(hosts)
}

// ListSessions 返回所有已保存的环境和 session。
func ListSessions() (map[string]*Session, error) {
	hosts, err := loadHostsFile()
	if err != nil {
		return nil, err
	}
	return hosts.Hosts, nil
}

// SessionToJSON 将 session 转为 json.RawMessage，方便 PrintJSON 输出。
func SessionToJSON(session *Session) (json.RawMessage, error) {
	data, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}
	return data, nil
}

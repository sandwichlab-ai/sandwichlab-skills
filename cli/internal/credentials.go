package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials 存储 Cognito OAuth 令牌和用户信息。
// 保存在 ~/.ahcli/credentials.json
type Credentials struct {
	Environment   string    `json:"environment"`    // 环境：dev / preprod / prod
	IDToken       string    `json:"id_token"`       // Cognito ID Token（用于 API 调用）
	AccessToken   string    `json:"access_token"`   // Cognito Access Token
	RefreshToken  string    `json:"refresh_token"`  // Refresh Token（用于刷新）
	TokenType     string    `json:"token_type"`     // 通常为 "Bearer"
	ExpiresAt     time.Time `json:"expires_at"`     // ID Token 过期时间
	UserID        string    `json:"user_id"`        // Cognito sub（用户唯一标识）
	Email         string    `json:"email"`          // 用户邮箱
	EmailVerified bool      `json:"email_verified"` // 邮箱是否验证
}

// IsExpired 检查 ID Token 是否已过期（提前 5 分钟判断）。
func (c *Credentials) IsExpired() bool {
	return time.Now().Add(5 * time.Minute).After(c.ExpiresAt)
}

// credentialsPath 返回凭证文件的完整路径：~/.ahcli/credentials.json
func credentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".ahcli", "credentials.json"), nil
}

// LoadCredentials 从 ~/.ahcli/credentials.json 加载凭证。
// 如果文件不存在，返回 nil（不报错）。
func LoadCredentials() (*Credentials, error) {
	path, err := credentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // 文件不存在，返回 nil
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// SaveCredentials 保存凭证到 ~/.ahcli/credentials.json。
// 自动创建 ~/.ahcli 目录（如果不存在）。
func SaveCredentials(creds *Credentials) error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}

	// 创建目录
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 序列化为 JSON
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// 写入文件（权限 0600，仅当前用户可读写）
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// ClearCredentials 删除凭证文件。
func ClearCredentials() error {
	path, err := credentialsPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	return nil
}

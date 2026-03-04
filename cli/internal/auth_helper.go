package internal

import (
	"fmt"
	"time"
)

// LoadAndRefreshCredentials 加载凭证，如果过期则自动刷新。
// 返回：凭证、是否已刷新、错误
func LoadAndRefreshCredentials(env string, cognitoConfig *CognitoConfig) (*Credentials, bool, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, false, fmt.Errorf("failed to load credentials: %w", err)
	}

	if creds == nil {
		return nil, false, nil // 未登录
	}

	// 检查环境是否匹配
	if creds.Environment != env {
		return nil, false, fmt.Errorf("credentials are for environment '%s', but current environment is '%s'. Please login again", creds.Environment, env)
	}

	// 检查是否过期
	if !creds.IsExpired() {
		return creds, false, nil // 未过期，直接返回
	}

	// Token 已过期，尝试刷新
	if creds.RefreshToken == "" {
		return nil, false, fmt.Errorf("token expired and no refresh token available. Please login again")
	}

	client := NewCognitoClient(cognitoConfig)
	tokenResp, err := client.RefreshTokens(creds.RefreshToken)
	if err != nil {
		return nil, false, fmt.Errorf("failed to refresh token: %w. Please login again", err)
	}

	// 解析新的 ID Token
	sub, email, emailVerified, exp, err := ParseIDToken(tokenResp.IDToken)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse refreshed ID token: %w", err)
	}

	// 更新凭证
	creds.IDToken = tokenResp.IDToken
	creds.AccessToken = tokenResp.AccessToken
	// 注意：refresh_token 可能不会返回新的，保留旧的
	if tokenResp.RefreshToken != "" {
		creds.RefreshToken = tokenResp.RefreshToken
	}
	creds.TokenType = tokenResp.TokenType
	creds.ExpiresAt = exp
	creds.UserID = sub
	creds.Email = email
	creds.EmailVerified = emailVerified

	// 保存更新后的凭证
	if err := SaveCredentials(creds); err != nil {
		return nil, false, fmt.Errorf("failed to save refreshed credentials: %w", err)
	}

	return creds, true, nil
}

// GetIDToken 获取当前有效的 ID Token（自动刷新）。
func GetIDToken(env string, cognitoConfig *CognitoConfig) (string, error) {
	creds, refreshed, err := LoadAndRefreshCredentials(env, cognitoConfig)
	if err != nil {
		return "", err
	}

	if creds == nil {
		return "", fmt.Errorf("not logged in. Please run 'ahcli auth login --env %s'", env)
	}

	if refreshed {
		fmt.Fprintf(Stderr, "[info] Token was expired and has been automatically refreshed\n")
	}

	return creds.IDToken, nil
}

// FormatExpiry 格式化过期时间为人类可读的字符串。
func FormatExpiry(t time.Time) string {
	now := time.Now()
	if t.Before(now) {
		return "已过期"
	}

	duration := t.Sub(now)
	if duration < time.Minute {
		return fmt.Sprintf("%d 秒后过期", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%d 分钟后过期", int(duration.Minutes()))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%d 小时后过期", int(duration.Hours()))
	}
	return fmt.Sprintf("%d 天后过期", int(duration.Hours()/24))
}

package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ExchangeHUIToken 使用 Cognito ID Token 换取 HUI JWT。
// 失败仅打印警告，不影响 CLI 正常使用。
func ExchangeHUIToken(f *Factory, creds *Credentials) {
	if f.URLs.HUI == "" {
		return
	}

	reqURL := f.URLs.HUI + "/public/api/v1/token-exchange"
	reqBody := fmt.Sprintf(`{"cognito_token":"%s"}`, creds.IDToken)

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(Stderr, "警告：无法创建 HUI token exchange 请求: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(Stderr, "警告：HUI token exchange 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(Stderr, "警告：无法读取 HUI token exchange 响应: %v\n", err)
		return
	}

	if f.Verbose {
		fmt.Fprintf(Stderr, "[verbose] HUI token exchange: HTTP %d, body: %s\n", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Success bool `json:"success"`
		Data    struct {
			Token     string    `json:"token"`
			ExpiresAt time.Time `json:"expires_at"`
		} `json:"data"`
		Message string `json:"message,omitempty"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Fprintf(Stderr, "警告：无法解析 HUI token exchange 响应: %v\n", err)
		return
	}

	if !apiResp.Success || apiResp.Data.Token == "" {
		fmt.Fprintf(Stderr, "警告：HUI token exchange 失败: %s\n", apiResp.Message)
		return
	}

	creds.HUIToken = apiResp.Data.Token
	creds.HUIExpiresAt = apiResp.Data.ExpiresAt
	if saveErr := SaveCredentials(creds); saveErr != nil {
		fmt.Fprintf(Stderr, "警告：保存 HUI Token 失败: %v\n", saveErr)
		return
	}

	fmt.Fprintf(Stderr, "✓ HUI Dashboard 已授权\n")
}

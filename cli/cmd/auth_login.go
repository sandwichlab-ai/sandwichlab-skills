package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"

	"github.com/spf13/cobra"
)

func newCmdLogin(f *internal.Factory) *cobra.Command {
	var tokenFlag string
	var tenantIDFlag string
	var userIDFlag string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "通过浏览器登录（Cognito OAuth）或 Token 无头登录",
		Long: `打开浏览器进行 Cognito OAuth 登录，或使用 --token 进行无头模式登录。

登录成功后，凭证会保存在 ~/.ahcli/credentials.json。

示例:
  ahcli auth login --env dev
  ahcli auth login --env dev --token <jwt> --tenant-id <tid> --user-id <uid>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenFlag != "" {
				return loginWithToken(f, tokenFlag, tenantIDFlag, userIDFlag)
			}
			return loginRun(f)
		},
	}

	cmd.Flags().StringVar(&tokenFlag, "token", "", "HUI JWT Token（无头模式，跳过浏览器 OAuth）")
	cmd.Flags().StringVar(&tenantIDFlag, "tenant-id", "", "租户 ID（无头模式下使用）")
	cmd.Flags().StringVar(&userIDFlag, "user-id", "", "用户 ID（无头模式下使用）")

	return cmd
}

// loginWithToken 使用 JWT Token 进行无头登录（Agent 容器中使用）。
func loginWithToken(f *internal.Factory, token, tenantID, userID string) error {
	creds := &internal.Credentials{
		Environment:   f.Env,
		IDToken:       token,
		AccessToken:   "",
		RefreshToken:  "",
		TokenType:     "Bearer",
		ExpiresAt:     time.Now().Add(24 * time.Hour),
		UserID:        userID,
		Email:         "",
		EmailVerified: true,
		HUIToken:      token,
		HUIExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	if err := internal.SaveCredentials(creds); err != nil {
		return fmt.Errorf("保存凭证失败: %w", err)
	}

	// 设置租户信息
	if tenantID != "" {
		tenant := internal.Tenant{
			Name:     tenantID,
			TenantID: tenantID,
			UserID:   userID,
		}
		if err := internal.AddTenant(tenant); err != nil {
			fmt.Fprintf(internal.Stderr, "警告：无法创建租户: %v\n", err)
		} else {
			_ = internal.UseTenant(tenantID)
		}
	}

	fmt.Fprintf(internal.Stderr, "✓ Token 登录成功（无头模式）\n")
	fmt.Fprintf(internal.Stderr, "环境: %s\n", f.Env)
	if tenantID != "" {
		fmt.Fprintf(internal.Stderr, "租户: %s\n", tenantID)
	}
	if userID != "" {
		fmt.Fprintf(internal.Stderr, "用户: %s\n", userID)
	}

	return nil
}

// loginRun 直接通过 Cognito OAuth 登录。
func loginRun(f *internal.Factory) error {
	cognitoConfig, err := loadCognitoConfig(f.Env)
	if err != nil {
		return fmt.Errorf("failed to load Cognito config: %w", err)
	}

	fmt.Fprintf(internal.Stderr, "正在登录到 %s 环境...\n", f.Env)
	fmt.Fprintf(internal.Stderr, "User Pool: %s\n", cognitoConfig.UserPoolID)

	client := internal.NewCognitoClient(cognitoConfig)

	authURL, codeVerifier, state, err := client.BuildAuthorizationURL()
	if err != nil {
		return fmt.Errorf("failed to build authorization URL: %w", err)
	}

	code, err := waitForAuthCode(client, authURL, state, cognitoConfig.CallbackPort)
	if err != nil {
		return err
	}

	fmt.Fprintf(internal.Stderr, "正在交换 Token...\n")
	tokenResp, err := client.ExchangeCodeForTokens(code, codeVerifier)
	if err != nil {
		return fmt.Errorf("failed to exchange tokens: %w", err)
	}

	sub, email, emailVerified, exp, provider, _, err := internal.ParseIDTokenFull(tokenResp.IDToken)
	if err != nil {
		return fmt.Errorf("failed to parse ID token: %w", err)
	}

	creds := &internal.Credentials{
		Environment:   f.Env,
		IDToken:       tokenResp.IDToken,
		AccessToken:   tokenResp.AccessToken,
		RefreshToken:  tokenResp.RefreshToken,
		TokenType:     tokenResp.TokenType,
		ExpiresAt:     exp,
		UserID:        sub,
		Email:         email,
		EmailVerified: emailVerified,
	}
	if saveErr := internal.SaveCredentials(creds); saveErr != nil {
		return fmt.Errorf("登录成功但保存凭证失败: %w", saveErr)
	}

	if f.Verbose {
		fmt.Fprintf(internal.Stderr, "[verbose] provider=%s, sub=%s\n", provider, sub)
	}
	opscoreUserID, err := fetchOpscoreUserID(f, tokenResp.IDToken)
	if err != nil {
		fmt.Fprintf(internal.Stderr, "警告：无法获取 opscore user_id: %v\n", err)
		opscoreUserID = sub
	}

	setupDefaultTenant(email, opscoreUserID)

	// 自动获取 HUI JWT（失败仅警告，不影响 CLI 正常使用）
	exchangeHUIToken(f, creds)

	fmt.Fprintf(internal.Stderr, "\n✓ 登录成功！\n")
	fmt.Fprintf(internal.Stderr, "环境: %s\n", f.Env)
	fmt.Fprintf(internal.Stderr, "用户: %s (%s)\n", email, sub)
	if opscoreUserID != sub {
		fmt.Fprintf(internal.Stderr, "Opscore User ID: %s\n", opscoreUserID)
	}
	fmt.Fprintf(internal.Stderr, "Token 过期时间: %s\n", exp.Format(time.RFC3339))

	return nil
}

// waitForAuthCode 启动回调服务器、打开浏览器并等待授权码。
func waitForAuthCode(client *internal.CognitoClient, authURL, state string, port int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Fprintf(internal.Stderr, "\n正在启动本地回调服务器（端口 %d）...\n", port)
	fmt.Fprintf(internal.Stderr, "请在浏览器中完成登录。如果浏览器未自动打开，请手动访问：\n%s\n\n", authURL)

	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)
	go func() {
		cbCode, cbErr := client.StartCallbackServer(ctx, state)
		if cbErr != nil {
			errChan <- cbErr
			return
		}
		codeChan <- cbCode
	}()

	time.Sleep(500 * time.Millisecond)

	if browserErr := openBrowser(authURL); browserErr != nil {
		fmt.Fprintf(internal.Stderr, "警告：无法自动打开浏览器: %v\n", browserErr)
	}

	select {
	case code := <-codeChan:
		fmt.Fprintf(internal.Stderr, "✓ 已收到授权码\n")
		return code, nil
	case cbErr := <-errChan:
		return "", fmt.Errorf("callback server error: %w", cbErr)
	case <-ctx.Done():
		return "", fmt.Errorf("登录超时（5 分钟）")
	}
}

// setupDefaultTenant 创建默认租户并自动切换。
func setupDefaultTenant(email, userID string) {
	tenant := internal.Tenant{
		Name:     email,
		TenantID: userID,
		UserID:   userID,
	}
	if addErr := internal.AddTenant(tenant); addErr != nil {
		fmt.Fprintf(internal.Stderr, "警告：无法创建默认租户: %v\n", addErr)
	} else {
		if useErr := internal.UseTenant(email); useErr != nil {
			fmt.Fprintf(internal.Stderr, "警告：无法切换到默认租户: %v\n", useErr)
		}
	}
}

// openBrowser 打开系统默认浏览器。
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// exchangeHUIToken 使用 Cognito ID Token 换取 HUI JWT。
// 失败仅打印警告，不影响 CLI 正常使用。
func exchangeHUIToken(f *internal.Factory, creds *internal.Credentials) {
	if f.URLs.HUI == "" {
		return
	}

	reqURL := f.URLs.HUI + "/public/api/v1/token-exchange"
	reqBody := fmt.Sprintf(`{"cognito_token":"%s"}`, creds.IDToken)

	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(internal.Stderr, "警告：无法创建 HUI token exchange 请求: %v\n", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(internal.Stderr, "警告：HUI token exchange 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(internal.Stderr, "警告：无法读取 HUI token exchange 响应: %v\n", err)
		return
	}

	if f.Verbose {
		fmt.Fprintf(internal.Stderr, "[verbose] HUI token exchange: HTTP %d, body: %s\n", resp.StatusCode, string(body))
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
		fmt.Fprintf(internal.Stderr, "警告：无法解析 HUI token exchange 响应: %v\n", err)
		return
	}

	if !apiResp.Success || apiResp.Data.Token == "" {
		fmt.Fprintf(internal.Stderr, "警告：HUI token exchange 失败: %s\n", apiResp.Message)
		return
	}

	creds.HUIToken = apiResp.Data.Token
	creds.HUIExpiresAt = apiResp.Data.ExpiresAt
	if saveErr := internal.SaveCredentials(creds); saveErr != nil {
		fmt.Fprintf(internal.Stderr, "警告：保存 HUI Token 失败: %v\n", saveErr)
		return
	}

	fmt.Fprintf(internal.Stderr, "✓ HUI Dashboard 已授权\n")
}

// opscoreUserInfoResponse opscore /api/v1/oauth/user/info 响应结构
// 注意：这是 data 字段的内容，不包含外层的 success/code/message
type opscoreUserInfoResponse struct {
	User *struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	} `json:"user"`
	Status string `json:"status"`
}

// fetchOpscoreUserID 从 opscore 获取真正的 user_id
// 使用 Cognito JWT token 通过 Kong Gateway 调用
func fetchOpscoreUserID(f *internal.Factory, idToken string) (string, error) {
	// 解析 ID Token 获取 provider 和 cognito sub
	sub, _, _, _, provider, cognitoUsername, err := internal.ParseIDTokenFull(idToken)
	if err != nil {
		return "", fmt.Errorf("failed to parse ID token: %w", err)
	}

	if f.Verbose {
		fmt.Fprintf(internal.Stderr, "[verbose] cognito:username=%s, provider=%s, sub=%s\n",
			cognitoUsername, provider, sub)
	}

	// 构建请求（手动设置 headers，因为 Kong 插件还没有实现这个逻辑）
	reqURL := f.URLs.OpsCore + "/api/v1/oauth/user/info"
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 设置 JWT token
	req.Header.Set("Authorization", "Bearer "+idToken)
	// 手动设置 provider headers（Kong 插件应该做这个，但目前还没实现）
	// 注意：X-App-Identity-Provider-User-ID 应该是 Cognito 的 sub，不是 Google 的 user ID
	req.Header.Set("X-App-Identity-Provider", provider)
	req.Header.Set("X-App-Identity-Provider-User-ID", sub)

	if f.Verbose {
		fmt.Fprintf(internal.Stderr, "[verbose] GET %s\n", reqURL)
		fmt.Fprintf(internal.Stderr, "[verbose] Headers: Authorization=Bearer {token}, X-App-Identity-Provider=%s, X-App-Identity-Provider-User-ID=%s\n",
			provider, sub)
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if f.Verbose {
		fmt.Fprintf(internal.Stderr, "[verbose] HTTP %d, body: %s\n", resp.StatusCode, string(body))
	}

	// 解析外层 APIResponse
	var apiResp struct {
		Success bool            `json:"success"`
		Message string          `json:"message,omitempty"`
		Data    json.RawMessage `json:"data,omitempty"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to decode API response: %w", err)
	}

	if !apiResp.Success {
		return "", fmt.Errorf("API error: %s", apiResp.Message)
	}

	// 解析 data 字段
	var result opscoreUserInfoResponse
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		return "", fmt.Errorf("failed to decode response data: %w", err)
	}

	if f.Verbose {
		fmt.Fprintf(internal.Stderr, "[verbose] Parsed result: status=%s, user=%v\n", result.Status, result.User)
	}

	// 检查用户是否存在
	if result.Status == "none" || result.User == nil {
		return "", fmt.Errorf("user not found in opscore (status: %s)", result.Status)
	}

	return result.User.UserID, nil
}

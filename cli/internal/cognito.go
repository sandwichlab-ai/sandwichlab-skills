package internal

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// CognitoConfig Cognito OAuth 配置。
type CognitoConfig struct {
	UserPoolID   string // User Pool ID，如 us-west-2_8D8Pp5UcN
	ClientID     string // App Client ID
	Domain       string // Cognito Domain，如 lexi2-dev.auth.us-west-2.amazoncognito.com
	Region       string // AWS Region，如 us-west-2
	CallbackPort int    // 本地回调服务器端口，默认 8888
}

// CognitoClient Cognito OAuth 客户端。
type CognitoClient struct {
	Config     *CognitoConfig
	HTTPClient *http.Client
}

// NewCognitoClient 创建 Cognito OAuth 客户端。
func NewCognitoClient(config *CognitoConfig) *CognitoClient {
	return &CognitoClient{
		Config: config,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// generatePKCE 生成 PKCE code_verifier 和 code_challenge。
// code_verifier: 43-128 字符的随机字符串
// code_challenge: BASE64URL(SHA256(code_verifier))
func generatePKCE() (verifier, challenge string, err error) {
	// 生成 32 字节随机数（base64 编码后约 43 字符）
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	verifier = base64.RawURLEncoding.EncodeToString(b)

	// 计算 SHA256 哈希
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])

	return verifier, challenge, nil
}

// BuildAuthorizationURL 构建 Cognito 授权 URL。
// 返回：授权 URL、code_verifier（用于后续 token 交换）、state（用于 CSRF 防护）
func (c *CognitoClient) BuildAuthorizationURL() (authURL, codeVerifier, state string, err error) {
	// 生成 PKCE
	codeVerifier, codeChallenge, err := generatePKCE()
	if err != nil {
		return "", "", "", err
	}

	// 生成 state（CSRF 防护）
	stateBytes := make([]byte, 16)
	if _, err := rand.Read(stateBytes); err != nil {
		return "", "", "", fmt.Errorf("failed to generate state: %w", err)
	}
	state = base64.RawURLEncoding.EncodeToString(stateBytes)

	// 构建授权 URL
	callbackURL := fmt.Sprintf("http://localhost:%d/callback", c.Config.CallbackPort)
	params := url.Values{
		"response_type":         {"code"},
		"client_id":             {c.Config.ClientID},
		"redirect_uri":          {callbackURL},
		"scope":                 {"openid email profile"},
		"state":                 {state},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"prompt":                {"select_account"}, // 强制显示账户选择页面
	}

	authURL = fmt.Sprintf("https://%s/oauth2/authorize?%s", c.Config.Domain, params.Encode())
	return authURL, codeVerifier, state, nil
}

// TokenResponse Cognito Token 端点响应。
type TokenResponse struct {
	IDToken      string `json:"id_token"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // 秒
}

// ExchangeCodeForTokens 使用授权码交换 Token。
func (c *CognitoClient) ExchangeCodeForTokens(code, codeVerifier string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("https://%s/oauth2/token", c.Config.Domain)
	callbackURL := fmt.Sprintf("http://localhost:%d/callback", c.Config.CallbackPort)

	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {c.Config.ClientID},
		"code":          {code},
		"redirect_uri":  {callbackURL},
		"code_verifier": {codeVerifier},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// RefreshTokens 使用 refresh_token 刷新 Token。
func (c *CognitoClient) RefreshTokens(refreshToken string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("https://%s/oauth2/token", c.Config.Domain)

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {c.Config.ClientID},
		"refresh_token": {refreshToken},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// ParseIDToken 解析 ID Token，提取用户信息（不验证签名）。
// 返回：sub（用户 ID）、email、email_verified、exp（过期时间）
func ParseIDToken(idToken string) (sub, email string, emailVerified bool, exp time.Time, err error) {
	// 使用 jwt 库解析（不验证签名，仅提取 claims）
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return "", "", false, time.Time{}, fmt.Errorf("failed to parse ID token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", false, time.Time{}, fmt.Errorf("invalid claims type")
	}

	// 提取 sub
	sub, _ = claims["sub"].(string)
	if sub == "" {
		return "", "", false, time.Time{}, fmt.Errorf("missing sub claim")
	}

	// 提取 email
	email, _ = claims["email"].(string)

	// 提取 email_verified
	emailVerified, _ = claims["email_verified"].(bool)

	// 提取 exp（Unix 时间戳）
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return "", "", false, time.Time{}, fmt.Errorf("missing or invalid exp claim")
	}
	exp = time.Unix(int64(expFloat), 0)

	return sub, email, emailVerified, exp, nil
}

// ParseIDTokenFull 解析 ID Token，提取完整用户信息（包括 provider）。
// 返回：sub、email、emailVerified、exp、provider、cognitoUsername
func ParseIDTokenFull(idToken string) (sub, email string, emailVerified bool, exp time.Time, provider, cognitoUsername string, err error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(idToken, jwt.MapClaims{})
	if err != nil {
		return "", "", false, time.Time{}, "", "", fmt.Errorf("failed to parse ID token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", false, time.Time{}, "", "", fmt.Errorf("invalid claims type")
	}

	// 提取 sub
	sub, _ = claims["sub"].(string)
	if sub == "" {
		return "", "", false, time.Time{}, "", "", fmt.Errorf("missing sub claim")
	}

	// 提取 email
	email, _ = claims["email"].(string)

	// 提取 email_verified
	emailVerified, _ = claims["email_verified"].(bool)

	// 提取 exp
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return "", "", false, time.Time{}, "", "", fmt.Errorf("missing or invalid exp claim")
	}
	exp = time.Unix(int64(expFloat), 0)

	// 提取 cognito:username（格式如 Google_xxx 或直接是 sub）
	cognitoUsername, _ = claims["cognito:username"].(string)
	if cognitoUsername == "" {
		cognitoUsername = sub
	}

	// 解析 provider：如果 username == sub，则是 email 登录；否则从 username 前缀解析
	if cognitoUsername == sub {
		provider = "email"
	} else {
		parts := strings.Split(cognitoUsername, "_")
		if len(parts) > 0 {
			provider = strings.ToLower(parts[0])
		}
	}

	return sub, email, emailVerified, exp, provider, cognitoUsername, nil
}

// StartCallbackServer 启动本地 HTTP 服务器，监听 Cognito 回调。
// 返回：授权码、错误
func (c *CognitoClient) StartCallbackServer(ctx context.Context, expectedState string) (string, error) {
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		// 检查 state（CSRF 防护）
		state := r.URL.Query().Get("state")
		if state != expectedState {
			errChan <- fmt.Errorf("invalid state parameter (CSRF attack?)")
			http.Error(w, "Invalid state", http.StatusBadRequest)
			return
		}

		// 检查是否有错误
		if errCode := r.URL.Query().Get("error"); errCode != "" {
			errDesc := r.URL.Query().Get("error_description")
			errChan <- fmt.Errorf("OAuth error: %s (%s)", errCode, errDesc)
			http.Error(w, "OAuth error: "+errDesc, http.StatusBadRequest)
			return
		}

		// 获取授权码
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("missing authorization code")
			http.Error(w, "Missing code", http.StatusBadRequest)
			return
		}

		// 返回成功页面
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>ahcli - Login Successful</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .success { color: #28a745; font-size: 24px; margin-bottom: 20px; }
        .message { color: #666; font-size: 16px; }
    </style>
</head>
<body>
    <div class="success">✓ Login Successful</div>
    <div class="message">You can close this window and return to the terminal.</div>
</body>
</html>
`)

		codeChan <- code
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", c.Config.CallbackPort),
		Handler: mux,
	}

	// 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to start callback server: %w", err)
		}
	}()

	// 等待授权码或错误
	var code string
	var err error
	select {
	case code = <-codeChan:
		// 成功获取授权码
	case err = <-errChan:
		// 发生错误
	case <-ctx.Done():
		err = fmt.Errorf("callback timeout")
	}

	// 关闭服务器
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)

	return code, err
}

// TokenCallbackResult 前端 OAuth 回调传回的 token 结果
type TokenCallbackResult struct {
	HUIToken       string
	CognitoIDToken string
}

// StartTokenCallbackServer 启动本地 HTTP 服务器，接收前端 OAuth 回调传回的 token。
// 前端完成 Cognito OAuth 后，会跳转到 http://localhost:{port}/callback?hui_token=X&cognito_id_token=Y
func StartTokenCallbackServer(ctx context.Context, port int) (*TokenCallbackResult, error) {
	resultChan := make(chan *TokenCallbackResult, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		huiToken := r.URL.Query().Get("hui_token")
		cognitoIDToken := r.URL.Query().Get("cognito_id_token")

		if huiToken == "" || cognitoIDToken == "" {
			errChan <- fmt.Errorf("missing hui_token or cognito_id_token")
			http.Error(w, "Missing tokens", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>ahcli - Login Successful</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; text-align: center; padding: 60px 20px; background: #fafafa; }
        .card { max-width: 400px; margin: 0 auto; background: #fff; border-radius: 12px; padding: 40px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .success { color: #22c55e; font-size: 48px; margin-bottom: 16px; }
        .title { font-size: 20px; font-weight: 600; color: #111; margin-bottom: 8px; }
        .message { color: #666; font-size: 14px; }
    </style>
</head>
<body>
    <div class="card">
        <div class="success">&#10003;</div>
        <div class="title">登录成功</div>
        <div class="message">可以关闭此窗口，返回终端继续使用。</div>
    </div>
</body>
</html>`)

		resultChan <- &TokenCallbackResult{
			HUIToken:       huiToken,
			CognitoIDToken: cognitoIDToken,
		}
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("failed to start callback server: %w", err)
		}
	}()

	var result *TokenCallbackResult
	var err error
	select {
	case result = <-resultChan:
	case err = <-errChan:
	case <-ctx.Done():
		err = fmt.Errorf("callback timeout")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = server.Shutdown(shutdownCtx)

	return result, err
}

package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// 全局认证信息（由 cmd/root.go 在启动时设置）
var (
	globalIDToken  string
	globalTenantID string
	globalUserID   string
	globalVerbose  bool
)

// SetGlobalAuth 设置全局认证信息，由 cmd/root.go 在启动时调用。
func SetGlobalAuth(idToken, tenantID, userID string, verbose bool) {
	globalIDToken = idToken
	globalTenantID = tenantID
	globalUserID = userID
	globalVerbose = verbose
}

// APIResponse 匹配 pkg/component/response.Response 的 JSON 结构。
// 使用 json.RawMessage 延迟解析 data 字段，便于直接输出或按需反序列化。
type APIResponse struct {
	Success bool            `json:"success"`
	Code    string          `json:"code,omitempty"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	TraceID string          `json:"trace_id,omitempty"`
}

// Client 是轻量级 HTTP 客户端，仅使用标准库 net/http，不依赖 pkg/infra。
type Client struct {
	BaseURL    string       // 下游服务的基础 URL（如 http://localhost:8083）
	HTTPClient *http.Client // 标准 HTTP 客户端
	Verbose    bool         // 是否输出调试信息到 stderr
	IDToken    string       // Cognito ID Token（自动注入到 Authorization header）
	TenantID   string       // 租户 ID（自动注入到请求参数）
	UserID     string       // 用户 ID（自动注入到请求参数）
}

// NewClient 创建一个新的 HTTP 客户端。默认超时 60 秒。
// 自动注入全局认证信息（ID Token 和租户信息）。
func NewClient(baseURL string, verbose bool) *Client {
	client := &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		Verbose: verbose,
	}

	// 自动注入全局认证信息
	if globalIDToken != "" {
		client.IDToken = globalIDToken
	}
	if globalTenantID != "" {
		client.TenantID = globalTenantID
	}
	if globalUserID != "" {
		client.UserID = globalUserID
	}

	return client
}

// SetIDToken 设置 Cognito ID Token，用于自动注入到请求头。
func (c *Client) SetIDToken(token string) {
	c.IDToken = token
}

// SetTenant 设置租户信息，用于自动注入到请求参数。
func (c *Client) SetTenant(tenantID, userID string) {
	c.TenantID = tenantID
	c.UserID = userID
}

// injectAuth 为请求注入 Authorization header（如果有 IDToken）。
func (c *Client) injectAuth(req *http.Request) {
	if c.IDToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.IDToken)
	}
}

// injectTenantToQuery 为 query 参数注入 tenant_id 和 user_id。
func (c *Client) injectTenantToQuery(params url.Values) url.Values {
	if params == nil {
		params = url.Values{}
	}
	if c.TenantID != "" && params.Get("tenant_id") == "" {
		params.Set("tenant_id", c.TenantID)
	}
	if c.UserID != "" && params.Get("user_id") == "" {
		params.Set("user_id", c.UserID)
	}
	return params
}

// injectTenantToBody 为 JSON body 注入 tenant_id 和 user_id。
func (c *Client) injectTenantToBody(body io.Reader) io.Reader {
	if body == nil || (c.TenantID == "" && c.UserID == "") {
		return body
	}

	// 读取原始 body
	data, err := io.ReadAll(body)
	if err != nil || len(data) == 0 {
		return body
	}

	// 解析为 map
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		// 不是 JSON 对象，返回原始 body
		return strings.NewReader(string(data))
	}

	// 注入 tenant_id 和 user_id（如果不存在）
	if c.TenantID != "" {
		if _, exists := m["tenant_id"]; !exists {
			m["tenant_id"] = c.TenantID
		}
	}
	if c.UserID != "" {
		if _, exists := m["user_id"]; !exists {
			m["user_id"] = c.UserID
		}
	}

	// 重新序列化
	newData, err := json.Marshal(m)
	if err != nil {
		return strings.NewReader(string(data))
	}

	return strings.NewReader(string(newData))
}

// Get 发起 GET 请求，params 为查询参数（可为 nil）。
func (c *Client) Get(path string, params url.Values) (*APIResponse, error) {
	params = c.injectTenantToQuery(params)
	reqURL := c.BaseURL + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	if c.Verbose {
		fmt.Fprintf(Stderr, "[verbose] GET %s\n", reqURL)
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.injectAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

// Post 发起 POST 请求，body 为 JSON 请求体。
func (c *Client) Post(path string, body io.Reader) (*APIResponse, error) {
	reqURL := c.BaseURL + path
	body = c.injectTenantToBody(body)

	// 在 verbose 模式下，读取并打印请求体
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		body = strings.NewReader(string(bodyBytes))
	}

	if c.Verbose {
		fmt.Fprintf(Stderr, "[verbose] POST %s\n", reqURL)
		if len(bodyBytes) > 0 {
			fmt.Fprintf(Stderr, "[verbose] Request body: %s\n", string(bodyBytes))
		}
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.injectAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

// Put 发起 PUT 请求，body 为 JSON 请求体。
func (c *Client) Put(path string, body io.Reader) (*APIResponse, error) {
	body = c.injectTenantToBody(body)
	return c.doJSON(http.MethodPut, c.BaseURL+path, body)
}

// Delete 发起 DELETE 请求，params 为查询参数（可为 nil）。
func (c *Client) Delete(path string, params url.Values) (*APIResponse, error) {
	params = c.injectTenantToQuery(params)
	reqURL := c.BaseURL + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}
	return c.doJSON(http.MethodDelete, reqURL, nil)
}

// DeleteWithBody 发起带 JSON body 的 DELETE 请求。
func (c *Client) DeleteWithBody(path string, body io.Reader) (*APIResponse, error) {
	body = c.injectTenantToBody(body)
	return c.doJSON(http.MethodDelete, c.BaseURL+path, body)
}

// Patch 发起 PATCH 请求，body 为 JSON 请求体。
func (c *Client) Patch(path string, body io.Reader) (*APIResponse, error) {
	body = c.injectTenantToBody(body)
	return c.doJSON(http.MethodPatch, c.BaseURL+path, body)
}

// doJSON 是通用的 JSON 请求方法，用于 PUT/DELETE/PATCH 等。
func (c *Client) doJSON(method, reqURL string, body io.Reader) (*APIResponse, error) {
	if c.Verbose {
		fmt.Fprintf(Stderr, "[verbose] %s %s\n", method, reqURL)
	}

	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.injectAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

// PostWithParams 发起 POST 请求，同时传 query params + JSON body。
func (c *Client) PostWithParams(path string, params url.Values, body io.Reader) (*APIResponse, error) {
	params = c.injectTenantToQuery(params)
	body = c.injectTenantToBody(body)
	reqURL := c.BaseURL + path
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	if c.Verbose {
		fmt.Fprintf(Stderr, "[verbose] POST %s\n", reqURL)
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	c.injectAuth(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return c.parseResponse(resp)
}

// parseResponse 解析 HTTP 响应。
// 处理逻辑：
//  1. HTTP 4xx/5xx → 尝试解析为 APIResponse 获取结构化错误信息，否则返回原始 body
//  2. 正常响应 → 解析为 APIResponse，检查 success 字段
//  3. 非标准 JSON → 将原始 body 包装为 data 字段返回
func (c *Client) parseResponse(resp *http.Response) (*APIResponse, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.Verbose {
		fmt.Fprintf(Stderr, "[verbose] HTTP %d, body length: %d\n", resp.StatusCode, len(body))
	}

	// HTTP 错误状态码处理
	if resp.StatusCode >= 400 {
		// 尝试解析为结构化错误
		var apiResp APIResponse
		if json.Unmarshal(body, &apiResp) == nil && apiResp.Message != "" {
			return &apiResp, fmt.Errorf("HTTP %d: %s", resp.StatusCode, apiResp.Message)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// 正常响应，尝试解析为标准 APIResponse
	var apiResp APIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// 非标准响应格式，包装为 data
		apiResp.Success = true
		apiResp.Data = body
		return &apiResp, nil
	}

	// 检查业务层错误（success=false）
	if !apiResp.Success {
		return &apiResp, fmt.Errorf("API error [%s]: %s", apiResp.Code, apiResp.Message)
	}

	// 非标准 APIResponse（如 BrowserService 直接返回扁平结构，有 success 字段但无 data 字段），
	// 将整个 body 作为 data 使用
	if len(apiResp.Data) == 0 {
		apiResp.Data = body
	}

	return &apiResp, nil
}

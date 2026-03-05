package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// jsonRPCRequest 是 ActionHub JSON-RPC 请求结构
type jsonRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
}

// toolCallParams 是 tools/call 的参数
type toolCallParams struct {
	Name      string      `json:"name"`
	Arguments interface{} `json:"arguments"`
}

// jsonRPCResponse 是 ActionHub JSON-RPC 响应结构
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// toolCallResult 是 tools/call 的返回结构
type toolCallResult struct {
	Content []toolCallContent `json:"content"`
}

type toolCallContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ActionHubCall 通过 ActionHub JSON-RPC 调用 tool，返回解析后的 JSON 数据。
// 封装 JSON-RPC 协议细节，对使用者透明。
func ActionHubCall(client *Client, toolName string, arguments interface{}) (json.RawMessage, error) {
	rpcReq := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: toolCallParams{
			Name:      toolName,
			Arguments: arguments,
		},
	}

	bodyBytes, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}

	// 调用 ActionHub 的公共 JSON-RPC 端点
	req, err := http.NewRequest(http.MethodPost, client.BaseURL+"/public/api/v1/mcp/jsonrpc", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client.injectAuth(req)

	httpResp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ActionHub request failed: %w", err)
	}
	defer httpResp.Body.Close()

	rawBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if httpResp.StatusCode >= 400 {
		return nil, fmt.Errorf("ActionHub HTTP error [%d]: %s", httpResp.StatusCode, string(rawBody))
	}

	var rpcResp jsonRPCResponse
	if err := json.Unmarshal(rawBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON-RPC response: %w\nraw body: %s", err, string(rawBody))
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("JSON-RPC error [%d]: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	// 解析 result.content[0].text 为 JSON
	var result toolCallResult
	if err := json.Unmarshal(rpcResp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tool result: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty tool result content")
	}

	// content[0].text 是 JSON 字符串，直接返回
	return json.RawMessage(result.Content[0].Text), nil
}

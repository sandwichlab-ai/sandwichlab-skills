package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

// Stderr 用于输出调试信息和错误信息，与 stdout 的数据输出分离，
// 使得 stdout 可以安全地 pipe 给 jq 等工具。
var Stderr = os.Stderr

// PrintJSON 将 JSON 数据输出到 stdout。
// compact=true 时输出紧凑格式（适合 pipe 给 jq）；
// compact=false 时输出带缩进的格式化 JSON（方便人工阅读）。
func PrintJSON(data json.RawMessage, compact bool) error {
	if compact {
		_, err := fmt.Fprintln(os.Stdout, string(data))
		return err
	}

	var indented []byte
	var err error

	// 尝试格式化为带缩进的 JSON
	var v interface{}
	if json.Unmarshal(data, &v) == nil {
		indented, err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format JSON: %w", err)
		}
	} else {
		// 非 JSON 内容，原样输出
		indented = data
	}

	_, err = fmt.Fprintln(os.Stdout, string(indented))
	return err
}

// PrintRawResponse 输出完整的 API 响应（包含 success、code、message、data 等元数据），
// 用于调试场景。
func PrintRawResponse(resp *APIResponse) error {
	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format response: %w", err)
	}
	_, err = fmt.Fprintln(os.Stdout, string(out))
	return err
}

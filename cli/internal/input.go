package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ReadJSONInputDirect 从文件路径或 stdin 读取 JSON 输入，不依赖 cobra.Command。
// filePath 和 useStdin 由调用方从 Options struct 传入。
func ReadJSONInputDirect(filePath string, useStdin bool) (map[string]interface{}, error) {
	var data []byte
	var err error

	if filePath != "" {
		data, err = os.ReadFile(filePath) // #nosec G304 -- CLI tool, user provides file path
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
	} else if useStdin {
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
	} else {
		return nil, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}

	return result, nil
}

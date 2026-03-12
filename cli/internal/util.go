package internal

import (
	"fmt"
	"strings"
)

// SplitAndTrim 将逗号分隔的字符串拆分为切片，去除空白。
func SplitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// validBindActions 是 bind 命令 --action flag 的合法值。
var validBindActions = map[string]bool{"add": true, "remove": true}

// ValidateBindAction 校验 bind 命令的 --action 参数值。
func ValidateBindAction(action string) error {
	if !validBindActions[action] {
		return fmt.Errorf("--action 值无效: %q（可选: add, remove）", action)
	}
	return nil
}

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sandwichlab-ai/sandwichlab-skills/cli/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type rulesCreateOpts struct {
	f              *internal.Factory
	File           string
	Stdin          bool
	TenantID       string
	ConstraintFile string // 本地约束文件路径覆盖
	Yes            bool   // 跳过确认提示
}

func newCmdRulesCreate(f *internal.Factory) *cobra.Command {
	o := &rulesCreateOpts{f: f}
	cmd := &cobra.Command{
		Use:   "create [natural-language-description]",
		Short: "创建规则（自然语言或 JSON）",
		Long: `创建广告自动化规则。支持两种模式：

NL 模式（默认）：传入自然语言描述，通过 Claude API 解析为结构化规则，确认后创建。
JSON 模式：通过 --file 或 --stdin 传入完整 JSON，直接创建。

NL 模式示例:
  ahcli rules create "ROI低于0.01持续2小时暂停广告"
  ahcli rules create "CPA大于50且花费超过200暂停Campaign" --yes

JSON 模式示例:
  ahcli rules create --file rule.json
  echo '{"name":"test","conditions":...}' | ahcli rules create --stdin`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireTenantID(&o.TenantID); err != nil {
				return err
			}

			// JSON 模式
			if o.File != "" || o.Stdin {
				return rulesCreateJSONRun(o)
			}

			// NL 模式
			if len(args) == 0 {
				return fmt.Errorf("请提供自然语言描述或使用 --file/--stdin 提供 JSON")
			}
			return rulesCreateNLRun(o, args[0])
		},
	}
	cmd.Flags().StringVarP(&o.File, "file", "f", "", "JSON 文件路径")
	cmd.Flags().BoolVar(&o.Stdin, "stdin", false, "从 stdin 读取 JSON")
	cmd.Flags().StringVar(&o.TenantID, "tenant-id", "", "租户 ID")
	cmd.Flags().StringVar(&o.ConstraintFile, "constraint-file", "", "本地约束文件路径（覆盖默认）")
	cmd.Flags().BoolVarP(&o.Yes, "yes", "y", false, "跳过确认提示")
	return cmd
}

// rulesCreateJSONRun JSON 模式：从文件或 stdin 读取 JSON 直接创建规则
func rulesCreateJSONRun(o *rulesCreateOpts) error {
	jsonInput, err := internal.ReadJSONInputDirect(o.File, o.Stdin)
	if err != nil {
		return err
	}
	if jsonInput == nil {
		return fmt.Errorf("请提供 --file 或 --stdin 输入 JSON")
	}

	bodyBytes, err := json.Marshal(jsonInput)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	client := o.f.HUIClient()
	resp, err := client.Post("/api/v1/rule", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	fmt.Fprintln(internal.Stderr, "规则创建成功")
	return o.f.Print(resp.Data)
}

// rulesCreateNLRun NL 模式：调用 Claude API 解析自然语言 → 结构化 JSON → 创建规则
func rulesCreateNLRun(o *rulesCreateOpts, description string) error {
	// 1. 加载约束 MD
	constraintMD, err := loadConstraintMD(o)
	if err != nil {
		return fmt.Errorf("failed to load constraint file: %w", err)
	}

	// 2. 解析 API Key
	apiKey := resolveAnthropicAPIKey()
	if apiKey == "" {
		return fmt.Errorf("Anthropic API Key 未配置。\n设置环境变量 ANTHROPIC_API_KEY 或在 .ahcli.yaml 中配置 anthropic_api_key")
	}

	fmt.Fprintf(internal.Stderr, "正在解析规则描述: %q\n", description)

	// 3. 调用 Claude API
	ruleJSON, err := callAnthropicForRule(apiKey, constraintMD, description)
	if err != nil {
		return fmt.Errorf("Claude API 调用失败: %w", err)
	}

	// 4. 展示解析结果
	formatted, _ := json.MarshalIndent(ruleJSON, "", "  ")
	fmt.Fprintf(internal.Stderr, "\n解析结果:\n%s\n\n", string(formatted))

	// 5. 确认
	if !o.Yes {
		fmt.Fprintf(internal.Stderr, "确认创建此规则？ (Y/n): ")
		var answer string
		fmt.Fscanln(os.Stdin, &answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(internal.Stderr, "已取消")
			return nil
		}
	}

	// 6. POST 到 HUI API
	bodyBytes, err := json.Marshal(ruleJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal rule JSON: %w", err)
	}

	client := o.f.HUIClient()
	resp, err := client.Post("/api/v1/rule", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}

	fmt.Fprintln(internal.Stderr, "规则创建成功")
	return o.f.Print(resp.Data)
}

// resolveAnthropicAPIKey 从环境变量或配置文件获取 Anthropic API Key
func resolveAnthropicAPIKey() string {
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key
	}
	return viper.GetString("anthropic_api_key")
}

// loadConstraintMD 加载规则约束 MD 文件
// 优先级：--constraint-file > 本地文件路径 > 内嵌默认值
func loadConstraintMD(o *rulesCreateOpts) (string, error) {
	// 1. 显式指定的路径
	if o.ConstraintFile != "" {
		data, err := os.ReadFile(o.ConstraintFile) // #nosec G304 -- CLI tool
		if err != nil {
			return "", fmt.Errorf("failed to read constraint file %s: %w", o.ConstraintFile, err)
		}
		return string(data), nil
	}

	// 2. 尝试本地已知路径
	candidates := []string{
		"configs/hui/rule_constraints.md",
		os.ExpandEnv("$HOME/sandwichlab_config/configs/hui/rule_constraints.md"),
	}
	for _, path := range candidates {
		data, err := os.ReadFile(path) // #nosec G304
		if err == nil {
			if o.f.Verbose {
				fmt.Fprintf(internal.Stderr, "[verbose] loaded constraint file from: %s\n", path)
			}
			return string(data), nil
		}
	}

	// 3. 使用内嵌默认值
	if o.f.Verbose {
		fmt.Fprintln(internal.Stderr, "[verbose] using embedded default constraint")
	}
	return defaultConstraintMD, nil
}

// callAnthropicForRule 调用 Claude API 将自然语言解析为结构化规则 JSON
func callAnthropicForRule(apiKey, constraintMD, description string) (map[string]any, error) {
	userPrompt := fmt.Sprintf(`请将以下自然语言规则描述解析为 JSON 格式的广告自动化规则。
只输出纯 JSON，不要包含 markdown 代码块或其他说明文字。

规则描述: %s`, description)

	responseText, err := internal.CallAnthropic(apiKey, constraintMD, userPrompt)
	if err != nil {
		return nil, err
	}

	// 去除可能的 markdown 代码块包裹
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	var result map[string]any
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("Claude 返回的内容无法解析为 JSON: %w\n原始响应:\n%s", err, responseText)
	}

	return result, nil
}

// defaultConstraintMD 内嵌的默认约束文档，当本地文件不可用时使用
const defaultConstraintMD = `# 广告自动化规则约束文档

你是一个广告自动化规则解析器。请将用户的自然语言描述转换为以下 JSON 结构。

## 输出 JSON 结构

{
  "name": "规则名称（从描述中提取，简洁描述规则目的）",
  "description": "规则详细描述",
  "priority": 0,
  "conditions": {
    "operator": "AND",
    "conditions": [
      {
        "metric": "<metric_name>",
        "operator": "<comparison_operator>",
        "value": <numeric_value>,
        "time_range": "<time_range>"
      }
    ]
  },
  "actions": [
    {
      "type": "<action_type>",
      "params": {},
      "priority": 0
    }
  ],
  "scope": {
    "type": "all",
    "entity_type": "<entity_type>",
    "project_ids": []
  }
}

## 可用指标 (metric)

| 指标名 | 说明 | 数据类型 |
|--------|------|---------|
| roi | 投资回报率 | float |
| cpa | 单次获客成本 | float (USD) |
| roas | 广告支出回报率 | float |
| spend | 花费 | float (USD) |
| impressions | 展示量 | integer |
| clicks | 点击量 | integer |
| ctr | 点击率 | float (0-1) |
| cvr | 转化率 | float (0-1) |

## 可用比较运算符 (operator)

| 运算符 | 说明 |
|--------|------|
| lt | 小于 |
| gt | 大于 |
| lte | 小于等于 |
| gte | 大于等于 |
| eq | 等于 |
| neq | 不等于 |

## 可用时间范围 (time_range)

| 值 | 说明 |
|-----|------|
| last_1h | 最近 1 小时 |
| last_2h | 最近 2 小时 |
| last_6h | 最近 6 小时 |
| last_12h | 最近 12 小时 |
| last_1d | 最近 1 天 |
| last_3d | 最近 3 天 |
| last_7d | 最近 7 天 |
| last_14d | 最近 14 天 |
| last_30d | 最近 30 天 |

## 可用动作类型 (action type)

| 动作 | 说明 | 参数示例 |
|------|------|---------|
| pause_campaign | 暂停广告系列 | {} |
| pause_ad | 暂停广告 | {} |
| close_ad | 关闭广告 | {} |
| adjust_budget | 调整预算 | {"adjustment_type": "decrease_percent", "value": 20} |
| notify | 发送通知 | {"channel": "feishu", "message": "..."} |

## 可用实体类型 (entity_type)

| 类型 | 说明 |
|------|------|
| campaign | 广告系列 |
| adset | 广告组 |
| ad | 广告 |

## 安全约束

1. 预算调整限制: adjust_budget 的 value 不得超过 50%
2. 禁止动作: 不支持直接删除广告（delete_ad/delete_campaign）
3. 条件要求: 每个规则至少有一个条件
4. 时间范围要求: 每个条件必须指定 time_range

## 解析规则

1. 如果描述中提到"持续 X 小时"，映射到 last_Xh
2. 如果描述中提到"持续 X 天"，映射到 last_Xd
3. 如果未明确提到时间，默认使用 last_1d
4. 如果未明确指定实体类型，默认 campaign
5. 如果有多个条件，默认使用 AND 组合
6. scope.type 默认为 "all"，除非用户明确指定了项目
7. name 字段从描述中提取关键词，简洁概括规则目的

## 示例

用户输入: "ROI低于0.01持续2小时暂停广告"
输出:
{
  "name": "低ROI自动暂停",
  "description": "当ROI低于0.01持续2小时，自动暂停Campaign",
  "priority": 0,
  "conditions": {
    "operator": "AND",
    "conditions": [
      {"metric": "roi", "operator": "lt", "value": 0.01, "time_range": "last_2h"}
    ]
  },
  "actions": [{"type": "pause_campaign", "params": {}, "priority": 0}],
  "scope": {"type": "all", "entity_type": "campaign", "project_ids": []}
}

用户输入: "CPA大于50且花费超过200时降低预算20%"
输出:
{
  "name": "高CPA高花费降预算",
  "description": "当CPA大于50且花费超过200时，降低预算20%",
  "priority": 0,
  "conditions": {
    "operator": "AND",
    "conditions": [
      {"metric": "cpa", "operator": "gt", "value": 50, "time_range": "last_1d"},
      {"metric": "spend", "operator": "gt", "value": 200, "time_range": "last_1d"}
    ]
  },
  "actions": [{"type": "adjust_budget", "params": {"adjustment_type": "decrease_percent", "value": 20}, "priority": 0}],
  "scope": {"type": "all", "entity_type": "campaign", "project_ids": []}
}
`

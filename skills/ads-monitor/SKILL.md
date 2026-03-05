---
name: ads-monitor
description: "广告投放盯盘，自动分析投放状态、识别异常、给出操作建议。通过 ahcli CLI 拉取数据并进行智能分析。使用场景：用户要求看投放情况、盯盘、看数据、有没有异常，如帮我盯一下盘、今天投放怎么样、看看有没有问题。"
---

# 广告投放盯盘

自动拉取投放数据 → 识别异常 → 分析趋势 → 给出操作建议。不是简单查数，而是帮用户做判断。

## 前提

- 已安装 `ahcli` CLI 工具（使用 `/ahcli-install` 安装）

---

## 执行流程

### Step 1: 检查登录状态

```bash
ahcli auth status --env $ENV
```

未登录则执行登录：

```bash
ahcli auth login --env $ENV
```

### Step 2: 拉取投放概览

```bash
ahcli -c ads monitor overview
```

使用 `-c`（compact）获取纯 JSON，解析以下关键字段：

```json
{
  "summary": { "total": 12, "active": 8, "paused": 3, "alerting": 2, "spend_today": 1234.56 },
  "projects": [
    {
      "project_id": "proj-xxx",
      "name": "Smart Watch",
      "status": "active",
      "spend_today": 450,
      "daily_budget": 500,
      "conversions": 12,
      "ctr": 0.85,
      "cpa": 37.5,
      "roas": 2.1,
      "budget_pct": 90,
      "alert_level": "critical",
      "alert_metrics": { "cpa": "warning", "budget_pct": "critical" }
    }
  ]
}
```

### Step 3: 向用户汇报全局状态

用自然语言总结，不要直接贴 JSON。格式参考：

> 当前共 **12** 个项目，**8** 个投放中，**3** 个已暂停。
> 今日总花费 **$1,234.56**。
> ⚠️ **2 个项目有告警**，需要关注。

### Step 4: 逐个分析告警项目

对每个 `alert_level != "normal"` 的项目，拉取趋势详情：

```bash
ahcli -c ads monitor detail $PROJECT_ID --days 7
```

分析 7 天趋势，关注以下模式：

| 模式 | 判断依据 | 严重程度 |
|------|---------|---------|
| CPA 飙升 | CPA 连续 3 天上涨且超阈值 | 🔴 高 |
| 预算即将耗尽 | budget_pct > 90% 且还未到晚间 | 🟡 中 |
| ROAS 持续下滑 | ROAS 连续 3 天下降 | 🔴 高 |
| CTR 异常偏低 | CTR < 0.5% | 🟡 中 |
| 花费停滞 | 活跃项目今日花费 ≈ 0 | 🔴 高（可能投放故障） |
| 效果突然变好 | CPA 大幅下降 + ROAS 上升 | ✅ 正面（值得追加预算） |

### Step 5: 给出具体操作建议

根据分析结果，给出**可执行的建议**，不要说"建议关注"这种空话：

**CPA 飙升：**
> 🔴 **Smart Watch** (proj-xxx) — CPA $180，超出阈值 $150
> 过去 7 天 CPA 趋势：$95 → $120 → $145 → $180（持续上涨）
> **建议：暂停项目止损** → 执行 `/ads-pause proj-xxx`

**预算即将耗尽但效果好：**
> 🟡 **Summer Sale** (proj-yyy) — 预算消耗 95%，ROAS 3.2x
> 效果良好但预算即将用完，今日还剩 $25
> **建议：追加日预算到 $800** → 执行 `/ads-adjust-budget proj-yyy`

**ROAS 下滑：**
> 🔴 **Brand Campaign** (proj-zzz) — ROAS 从 2.8x 跌至 1.3x
> 过去 7 天持续下滑，当前低于盈亏平衡线
> **建议：暂停项目，检查素材和受众定向** → 执行 `/ads-pause proj-zzz`

**花费停滞：**
> 🔴 **New Product Launch** (proj-aaa) — 今日花费 $0，状态仍为 active
> 可能是投放审核未通过或预算/出价设置问题
> **建议：检查项目状态** → 执行 `/ads-project-status proj-aaa`

**无异常：**
> ✅ 所有项目运行正常，无需操作。
> 花费最高：Smart Watch ($450)，效果最好：Summer Sale (ROAS 3.2x)

### Step 6: 询问用户是否执行

列出所有建议操作，让用户选择：

> 以上建议操作：
> 1. 暂停 Smart Watch (CPA 过高)
> 2. 追加 Summer Sale 预算到 $800
> 3. 暂停 Brand Campaign (ROAS 下滑)
>
> 要执行哪些？(输入编号，如 "1,3"，或 "全部"，或 "跳过")

用户确认后，调用对应 Skill 执行：
- 暂停 → `/ads-pause <project-id>`
- 调预算 → `/ads-adjust-budget <project-id>`
- 启动 → `/ads-enable <project-id>`

### Step 7: 打开前端页面（可选）

如果用户想在前端查看更多细节：

```bash
ahcli ads monitor open              # 打开盯盘总览
ahcli ads monitor open proj-xxx     # 打开特定项目
```

---

## 特殊场景处理

### 用户问"今天投放怎么样"

执行 Step 1-5 的完整流程。

### 用户问"有没有异常/告警"

跳到 Step 2，仅关注告警项目：

```bash
ahcli -c ads monitor overview --alert-only
```

无告警 → 回复"全部正常"；有告警 → 继续 Step 4-6。

### 用户问特定项目

直接进 detail：

```bash
ahcli -c ads monitor detail $PROJECT_ID --days 7
```

分析后给建议。

### 用户要求持续监控

```bash
ahcli ads monitor watch --interval 120 --alert-only --count 10
```

每 2 分钟检查一次，仅在有告警时输出，最多 10 次。

---

## 分析原则

1. **用数据说话** — 每个结论都附上具体数值和趋势
2. **给可执行建议** — 不说"建议关注"，直接说"暂停项目"或"追加预算到 $X"
3. **区分轻重缓急** — critical 先说，warning 后说，正常的简单带过
4. **对比说明** — "CPA $180，超出阈值 $150 的 20%"，而不是只说"CPA 偏高"
5. **趋势优先** — 单日数据可能有波动，看 7 天趋势做判断

---

## 错误处理

| 错误 | 原因 | 解决 |
|------|------|------|
| `tenant_id is required` | 未设置租户 | 重新登录 `ahcli auth login` |
| `unauthorized` | 登录过期 | 重新登录 |
| 趋势数据不足 | 新项目不足 2 天 | 仅基于当日数据分析，注明数据有限 |
| 全部项目都正常 | 无异常 | 汇报正常状态 + 花费 TOP3 + 效果 TOP3 |

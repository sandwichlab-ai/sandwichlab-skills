---
name: ads-monitor
description: "广告投放盯盘，查看项目概览、告警状态、持续监控。通过 ahcli CLI 查询盯盘数据。使用场景：用户要求查看投放概况、告警、盯盘，如查看投放概览、有哪些告警、monitor overview。"
---

# 广告投放盯盘

通过 `ahcli` CLI 工具查看广告投放概览、告警状态，支持持续监控模式。

## 使用方法

- `/ads-monitor` — 查看投放概览
- `/ads-monitor alerts` — 仅查看告警项目
- `/ads-monitor watch` — 持续监控模式

## 前提

- 已安装 `ahcli` CLI 工具（使用 `/ahcli-install` 安装）

## 用户需提供的信息

| 信息 | 必填 | 示例 |
|------|------|------|
| 环境 | 否 | `dev`（默认）/ `preprod` / `prod` |
| 是否仅看告警 | 否 | 默认显示全部项目 |
| 排序方式 | 否 | `spend`（默认）/ `cpa` / `ctr` / `roas` |
| 搜索关键词 | 否 | 项目名称或 ID |

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

### Step 2: 查看投放概览

```bash
ahcli ads monitor overview
```

可选参数：

| 参数 | 说明 | 示例 |
|------|------|------|
| `--status` | 状态过滤 | `--status active` |
| `--alert-only` | 仅显示告警 | `--alert-only` |
| `--sort-by` | 排序字段 | `--sort-by cpa` |
| `--search` | 搜索 | `--search "Smart Watch"` |
| `--limit` | 返回数量 | `--limit 20` |

命令会在 stderr 输出人性化摘要（项目数、花费、告警数），stdout 输出完整 JSON。

向用户展示关键信息：
- 总项目数 / 活跃 / 已暂停 / 告警中
- 今日总花费
- 告警项目列表（名称、告警级别、关键指标）

### Step 3: 仅查看告警（可选）

```bash
ahcli ads monitor alerts
```

等同于 `overview --alert-only`，快速查看需要关注的项目。

### Step 4: 持续监控（可选）

如果用户需要持续关注：

```bash
ahcli ads monitor watch --interval 60
```

| 参数 | 说明 | 默认 |
|------|------|------|
| `--interval` | 轮询间隔（秒） | 60 |
| `--alert-only` | 仅在有告警时输出 | false |
| `--count` | 最大轮询次数 | 0（无限） |

Ctrl+C 停止监控。

### Step 5: 打开前端盯盘页面（可选）

```bash
ahcli ads monitor open
ahcli ads monitor open proj-xxx  # 深链接到特定项目
```

### Step 6: 管理告警配置（可选）

查看当前告警阈值：

```bash
ahcli ads monitor config get
```

更新告警阈值：

```bash
ahcli ads monitor config set --file thresholds.json
```

重置为默认值：

```bash
ahcli ads monitor config reset
```

---

## 告警阈值 JSON 格式

```json
{
  "thresholds": [
    { "metric": "cpa", "label": "CPA", "direction": "above", "warning_value": 100, "critical_value": 150, "enabled": true },
    { "metric": "ctr", "label": "CTR", "direction": "below", "warning_value": 1.0, "critical_value": 0.5, "enabled": true },
    { "metric": "roas", "label": "ROAS", "direction": "below", "warning_value": 2.0, "critical_value": 1.5, "enabled": true },
    { "metric": "budget_pct", "label": "预算消耗%", "direction": "above", "warning_value": 80, "critical_value": 95, "enabled": true }
  ]
}
```

---

## 输出摘要

向用户返回以下信息：
- 投放概览：活跃/暂停/告警项目数 + 今日总花费
- 告警项目：名称、告警级别（warning/critical）、触发的指标
- 建议操作：对 critical 项目建议暂停或调整预算

---

## 错误处理

| 错误 | 原因 | 解决 |
|------|------|------|
| `tenant_id is required` | 未设置租户 | 重新登录 `ahcli auth login` |
| `unauthorized` | 登录过期 | 重新登录 |
| `查询失败` | 后端服务不可用 | 检查环境或稍后重试 |

---

## 与其他 Skill 联动

| 场景 | 推荐 Skill |
|------|------------|
| 发现 CPA 过高需要暂停 | `/ads-pause <project-id>` |
| 发现预算消耗过快需要调整 | `/ads-adjust-budget <project-id>` |
| 项目需要重新启动 | `/ads-enable <project-id>` |
| 需要查看详细项目信息 | `/ads-project-status <project-id>` |

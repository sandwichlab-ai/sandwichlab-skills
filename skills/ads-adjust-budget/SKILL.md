---
name: ads-adjust-budget
description: "调整指定广告项目的日预算。通过 ahcli CLI 查询当前预算并执行调整。使用场景：用户要求调整预算，如调整预算到50、修改日预算、把预算改成100。"
---

# 调整广告项目预算

调整指定广告项目的日预算。使用 `ahcli` CLI 工具执行。

## 使用方法

- `/ads-adjust-budget <project-id> --daily-budget <amount>` — 调整指定项目的日预算
- `/ads-adjust-budget` — 交互式引导

## 用户需提供的信息

| 信息 | 必填 | 示例 |
|------|------|------|
| 项目 ID | 是 | `proj-550e8400e29b41d4a716446655440000` |
| 日预算（美元） | 是 | `50` = $50/天 |
| 环境 | 否 | `dev`（默认）/ `preprod` / `prod` |

**预算说明：** `daily-budget` 单位是美元整数，`50` = $50/天。系统内部自动转换为美分。

---

## 执行流程

### Step 1: 检查登录状态

```bash
ahcli auth status --env $ENV
```

未安装 ahcli 先执行 `/ahcli-install`；未登录则执行登录：

```bash
ahcli auth login --env $ENV
```

### Step 2: 查询项目当前状态和预算

先确认项目存在并记录当前预算：

```bash
ahcli ads project get $PROJECT_ID
```

记录当前 `daily_budget` 值，供对比确认。

### Step 3: 执行预算调整

使用 `--daily-budget` flag：

```bash
ahcli ads project adjust-budget $PROJECT_ID --daily-budget $NEW_BUDGET
```

或使用 JSON 输入（支持更多参数）：

```bash
echo '{"daily_budget": 50}' | ahcli ads project adjust-budget $PROJECT_ID --stdin
```

### Step 4: 验证结果

确认预算已更新：

```bash
ahcli ads project get $PROJECT_ID
```

对比 `daily_budget` 字段，确认已变更为新值。

---

## 错误处理

| 错误 | 原因 | 解决 |
|------|------|------|
| `project not found` | 项目 ID 不存在 | 检查项目 ID 是否正确 |
| `daily-budget 为必填参数且必须大于 0` | 未指定预算或预算 ≤ 0 | 确保传入正整数 |
| `unauthorized` | 登录过期或权限不足 | 重新登录 `ahcli auth login` |

---
name: ads-pause
description: "暂停指定的广告投放项目。通过 ahcli CLI 查询项目当前状态并执行暂停操作。使用场景：用户要求暂停某个项目的投放，如暂停项目XXX、停止投放、pause project。"
---

# 暂停广告项目

暂停指定的广告投放项目。使用 `ahcli` CLI 工具执行。

## 使用方法

- `/ads-pause <project-id>` — 暂停指定项目
- `/ads-pause` — 交互式引导，输入项目 ID

## 前提

- 已安装 `ahcli` CLI 工具（使用 `/ahcli-install` 安装）

## 用户需提供的信息

| 信息 | 必填 | 示例 |
|------|------|------|
| 项目 ID | 是 | `proj-550e8400e29b41d4a716446655440000` |
| 环境 | 否 | `dev`（默认）/ `preprod` / `prod` |

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

### Step 2: 查询项目当前状态

先确认项目存在且处于可暂停状态（`active` 或 `processing`）：

```bash
ahcli ads project get $PROJECT_ID
```

检查返回的 `status` 字段：
- `active` / `processing` → 可以暂停，继续 Step 3
- `paused` → 已经暂停，无需操作，告知用户
- `draft` / `archived` → 不可暂停，告知用户当前状态

### Step 3: 执行暂停

```bash
ahcli ads project pause $PROJECT_ID
```

### Step 4: 验证结果

等待 2 秒后确认状态已变更：

```bash
ahcli ads project get $PROJECT_ID
```

确认 `status` 为 `paused`。

---

## 错误处理

| 错误 | 原因 | 解决 |
|------|------|------|
| `project not found` | 项目 ID 不存在 | 检查项目 ID 是否正确 |
| `invalid status transition` | 项目状态不支持暂停 | 查询当前状态，告知用户 |
| `unauthorized` | 登录过期或权限不足 | 重新登录 `ahcli auth login` |

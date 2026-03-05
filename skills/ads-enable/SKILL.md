---
name: ads-enable
description: "启用或激活已暂停的广告投放项目。通过 ahcli CLI 查询项目状态并执行激活操作。使用场景：用户要求启用/恢复/激活某个项目，如启用项目、恢复投放、activate project。"
---

# 启用广告项目

将已暂停的广告项目重新激活。使用 `ahcli` CLI 工具执行。

## 使用方法

- `/ads-enable <project-id>` — 启用指定项目
- `/ads-enable` — 交互式引导，输入项目 ID

## 前提

- 已安装 `ahcli` CLI 工具（使用 `/ahcli-install` 安装）

## 用户需提供的信息

| 信息 | 必填 | 示例 |
|------|------|------|
| 项目 ID | 是 | `proj-550e8400e29b41d4a716446655440000` |
| 跳过初始 Epoch | 否 | `true`（默认 false） |
| 环境 | 否 | `dev`（默认）/ `preprod` / `prod` |

**`skip-initial-epoch` 说明：** 当项目的广告已通过手动方式创建时，设置此标志跳过 Epoch 0 的自动广告组装。

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

先确认项目存在且处于可激活状态：

```bash
ahcli ads project get $PROJECT_ID
```

检查返回的 `status` 字段：
- `paused` / `draft` → 可以激活，继续 Step 3
- `active` / `processing` → 已经在运行，无需操作，告知用户
- `archived` → 已归档，不可激活，告知用户

### Step 3: 执行激活

不跳过初始 Epoch：

```bash
ahcli ads project activate $PROJECT_ID
```

跳过初始 Epoch（广告已手动创建时使用）：

```bash
ahcli ads project activate $PROJECT_ID --skip-initial-epoch
```

### Step 4: 验证结果

等待 3-5 秒后确认状态已变更：

```bash
ahcli ads project get $PROJECT_ID
```

确认 `status` 变为 `active` 或 `processing`。

---

## 错误处理

| 错误 | 原因 | 解决 |
|------|------|------|
| `project not found` | 项目 ID 不存在 | 检查项目 ID 是否正确 |
| `invalid status transition` | 项目状态不支持激活 | 查询当前状态，告知用户 |
| `channel auth not found` | 项目缺少渠道认证 | 需要先配置投放渠道和凭证 |
| `unauthorized` | 登录过期或权限不足 | 重新登录 `ahcli auth login` |

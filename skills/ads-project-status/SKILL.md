---
name: ads-project-status
description: "查询指定广告项目的详细信息和运行状态，支持诊断。通过 ahcli CLI 查询项目信息。使用场景：用户要求查看项目状态，如查看项目状态、项目诊断、project status。"
---

# 查询广告项目状态

查询指定广告项目的详细信息和运行状态。使用 `ahcli` CLI 工具执行。

## 使用方法

- `/ads-project-status <project-id>` — 查询指定项目状态
- `/ads-project-status` — 交互式引导，输入项目 ID

## 前提

- VPN 连接到对应环境的 K8s 集群
- `ahcli` 已编译（`cd tools/ahcli && go build -o ../../ahcli .`）

## 用户需提供的信息

| 信息 | 必填 | 示例 |
|------|------|------|
| 项目 ID | 是 | `proj-550e8400e29b41d4a716446655440000` |
| 环境 | 否 | `dev`（默认）/ `preprod` / `prod` |
| 是否需要诊断 | 否 | 默认仅查询基本状态 |

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

### Step 2: 查询项目详情

```bash
ahcli ads project get $PROJECT_ID
```

关注以下关键字段并向用户汇报：

| 字段 | 说明 |
|------|------|
| `status` | 项目状态：`draft` / `active` / `paused` / `processing` / `archived` |
| `stage` | 当前阶段 |
| `daily_budget` | 日预算（美元） |
| `name` | 项目名称 |
| `source_platform` | 来源平台 |
| `created_at` | 创建时间 |

### Step 3: 诊断（可选）

如果用户需要更详细的诊断信息（例如激活失败原因）：

```bash
ahcli ads project diagnose $PROJECT_ID
```

### Step 4: 查看指标（可选）

如果用户需要投放效果数据：

```bash
ahcli ads project metrics $PROJECT_ID --days 7
```

---

## 输出摘要

向用户返回以下信息：
- 项目名称和 ID
- 当前状态
- 日预算
- 创建时间
- 如有异常，附上诊断结果

---

## 错误处理

| 错误 | 原因 | 解决 |
|------|------|------|
| `project not found` | 项目 ID 不存在 | 检查项目 ID 是否正确 |
| `unauthorized` | 登录过期或权限不足 | 重新登录 `ahcli auth login` |

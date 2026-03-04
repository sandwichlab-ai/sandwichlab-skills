# ahcli - ActionHub CLI

ActionHub 生态的命令行工具，按实体组织子命令，覆盖广告投放完整生命周期。

## 快速开始

```bash
# 构建
go build -o ahcli ./tools/ahcli

# 登录（凭证自动持久化，后续命令无需重复指定 tenant_id/user_id）
./ahcli auth login --env dev --hui-user 1

# 查看帮助
./ahcli --help
./ahcli ads --help
./ahcli ads project --help
```

## 命令总览

```text
ahcli
├── auth                              # 认证管理
│   ├── login                         # 登录（创建 Session 并保存凭证）
│   ├── status                        # 查看当前登录状态
│   ├── tenants                       # 列出所有已保存的登录环境
│   ├── switch <tenant-id>            # 切换活跃租户
│   └── logout                        # 清除凭证
│
├── ads                               # 广告服务
│   ├── project                       # 项目管理
│   │   ├── list / get / create / update / delete
│   │   ├── activate / pause / archive
│   │   ├── diagnose / metrics / adjust-budget / bill-info
│   │   ├── channel                   # 渠道配置（/projects/:id/channel-configs）
│   │   │   ├── list / get / create / update / delete
│   │   │   ├── set-pixel / set-google-tag / auths
│   │   └── attachment                # 项目附件（/project-attachments）
│   │       ├── list / get / create / update / upsert / delete / latest
│   │
│   ├── creative                      # 创意管理
│   │   ├── list / get / create / update / delete
│   │   ├── activate / archive
│   │   └── bind                      # 绑定/解绑创意到项目
│   │
│   ├── copy                          # 文案管理
│   │   ├── list / get / create / delete
│   │   ├── batch-update
│   │   └── bind                      # 绑定/解绑文案到项目
│   │
│   ├── media                         # 素材库管理
│   │   ├── search / get / batch-get / quota
│   │   ├── generate / poll / import-s3 / edit
│   │   ├── delete / favorite / trash / batch-update
│   │   ├── bind                      # 绑定/解绑素材到项目
│   │   └── tag                       # 标签管理
│   │       ├── list / create / update / delete / copy-assets
│   │
│   ├── action                        # 投放动作
│   │   ├── create / create-sync / batch
│   │   ├── list / get / cancel / stats
│   │   └── query-by-biz-id
│   │
│   ├── plan                          # 投放计划
│   │   ├── list / get / create / update / delete
│   │   ├── activate / deactivate
│   │   └── epoch                     # Epoch 管理
│   │       ├── get / latest / retry
│   │
│   └── channel                       # 渠道查询
│       └── meta                      # Meta 渠道
│           ├── campaign              # Campaign 查询
│           │   ├── list / by-projects
│           ├── entity                # 实体查询
│           │   ├── get / list
│           ├── audience              # 自定义受众
│           │   ├── create / get / add-users / remove-users
│           ├── account-info          # 账户信息
│           └── convert-to-usd        # 货币转换
│
├── data                              # DataSyncer 数据同步
│   ├── get-shop / list-orders
├── browser                           # BrowserService 浏览器服务
│   └── screenshot
└── ops                               # OpsCore 运营管理
```

## 使用示例

### 认证

```bash
# HUI 用户快捷登录
ahcli auth login --env dev --hui-user 1

# 查看当前状态
ahcli auth status

# 切换租户
ahcli auth switch tnt-yyy

# 登出
ahcli auth logout
```

### 项目管理

```bash
# 列出项目
ahcli ads project list
ahcli ads project list --status active --limit 20

# 获取项目详情
ahcli ads project get proj-xxx

# 创建项目
ahcli ads project create --file project.json
ahcli ads project create --name "My Project" --description "..."

# 项目生命周期
ahcli ads project activate proj-xxx
ahcli ads project pause proj-xxx
ahcli ads project archive proj-xxx

# 项目诊断与指标
ahcli ads project diagnose proj-xxx
ahcli ads project metrics proj-xxx --days 7
ahcli ads project bill-info proj-xxx
```

### 渠道配置（project 子命令）

```bash
# 列出项目的渠道配置
ahcli ads project channel list proj-xxx

# 创建渠道配置
ahcli ads project channel create proj-xxx --file config.json

# 配置 Pixel
ahcli ads project channel set-pixel proj-xxx --pixel-id 654321
```

### 素材管理

```bash
# 搜索素材
ahcli ads media search --project-id proj-xxx --type image

# AI 生成素材
ahcli ads media generate --file params.json

# 轮询生成结果（自动等待）
ahcli ads media poll "arn:aws:states:..." --wait --project-id proj-xxx --tenant-id tnt-xxx

# 从 S3 导入
ahcli ads media import-s3 --project-id proj-xxx --s3-urls "s3://bucket/img.png"

# 批量操作
ahcli ads media delete --ids masset-xxx,masset-yyy
ahcli ads media favorite --ids masset-xxx --set true

# 标签
ahcli ads media tag list
ahcli ads media tag create --name "Hero"
```

### 创意与文案

```bash
# 创建创意（通过 ActionHub）
ahcli ads creative create --file creative.json
ahcli ads creative activate crtv-xxx
ahcli ads creative bind --creative-id crtv-xxx --project-id proj-xxx --action add

# 创建文案（通过 ActionHub）
ahcli ads copy create --file copy.json
ahcli ads copy bind --copy-id copy-xxx --project-id proj-xxx --action add
```

### 投放动作

```bash
# 同步创建 Campaign（等待完成）
ahcli ads action create-sync --file campaign.json

# 也可用 flag 快捷方式
ahcli ads action create-sync \
  --tenant-id tnt-xxx --project-id proj-xxx \
  --access-token "EAAxxxx" --ad-account-id "act_123" \
  --campaign-name "Smart Watch Q1" --daily-budget 50 --creative-id crtv-xxx

# 查看投放动作列表
ahcli ads action list --project-id proj-xxx --status completed

# 查看 Worker Pool 状态
ahcli ads action stats
```

### 投放计划

```bash
ahcli ads plan list --project-id proj-xxx
ahcli ads plan create --file plan.json
ahcli ads plan activate plan-xxx

# Epoch 查询
ahcli ads plan epoch get plan-xxx 3
ahcli ads plan epoch latest plan-xxx
ahcli ads plan epoch retry plan-xxx 3
```

### Meta 渠道查询

```bash
# Campaign 查询
ahcli ads channel meta campaign list --account-ids act_123,act_456
ahcli ads channel meta campaign by-projects --project-ids proj-xxx,proj-yyy

# 实体查询
ahcli ads channel meta entity get campaign 123456
ahcli ads channel meta entity list adset --parent-id 123456

# 自定义受众
ahcli ads channel meta audience create --file audience.json
ahcli ads channel meta audience get aud-xxx
ahcli ads channel meta audience add-users aud-xxx --file users.json
```

### 环境切换与调试

```bash
# 指定 prod 环境
ahcli --env prod ads project list

# verbose 模式查看请求细节
ahcli -v ads project channel list proj-xxx
# 输出: [verbose] GET http://localhost:8083/api/v1/projects/proj-xxx/channel-configs

# 紧凑输出 + jq 提取字段
ahcli -c ads project list | jq '.[].name'
```

## 全局参数

| 参数 | 缩写 | 默认值 | 说明 |
|------|------|--------|------|
| `--env` | `-e` | `dev` | 目标环境：`local` / `dev` / `preprod` / `prod` |
| `--verbose` | `-v` | `false` | 输出调试信息（请求 URL、HTTP 状态码等） |
| `--compact` | `-c` | `false` | 紧凑 JSON 输出，方便 pipe 给 `jq` |
| `--config` | - | `.ahcli.yaml` | 配置文件路径 |

## 配置文件

配置文件支持服务 URL 覆盖和常用默认参数，避免重复输入。

**查找顺序：** `--config` 指定路径 > `./.ahcli.yaml` > `~/.ahcli.yaml`

**优先级：** 命令行 flag > 环境变量 `AHCLI_*` > 配置文件 > auth session

```yaml
# .ahcli.yaml

# 服务 URL 覆盖（不设置时使用 --env 对应的默认 URL）
# adscore_url: "http://localhost:8083"
# data_syncer_url: "http://localhost:8086"
# browser_url: "http://localhost:8090"
# opscore_url: "http://localhost:8088"
# authcenter_url: "http://localhost:8080"
# actionhub_url: "http://localhost:8085"

# 常用默认参数（命令行未传时自动使用）
defaults:
  tenant_id: "tnt-xxxxxxxx"
  project_id: "proj-xxxxxxxx"
  user_id: "usr-xxxxxxxx"
```

也可通过环境变量覆盖：

```bash
export AHCLI_ADSCORE_URL="http://localhost:9999"
ahcli ads project list
```

## 认证机制

`ahcli auth login` 调用 AuthCenter 创建 Session，获取 `tenant_id` 和 `user_id`，持久化到 `~/.config/ahcli/hosts.yml`。

后续命令自动注入 session 中的 `tenant_id` / `user_id`，优先级低于 CLI flag 和配置文件。

```bash
# 登录后无需重复指定 tenant_id
ahcli auth login --env dev --hui-user 1
ahcli ads project list              # 自动使用 session 中的 tenant_id
ahcli ads project list --tenant-id tnt-other  # flag 优先
```

## 环境默认 URL

| 环境 | AdsCore | DataSyncer | BrowserService | OpsCore | AuthCenter | ActionHub |
|------|---------|------------|----------------|---------|------------|-----------|
| `local` | `:8083` | `:8086` | `:8090` | `:8088` | `:8080` | `:8085` |
| `dev` | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS |
| `preprod` | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS |
| `prod` | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS | K8s 内部 DNS |

> **注意：** `dev` / `preprod` / `prod` 环境使用 K8s 集群内部 DNS，需要 VPN 或集群网络访问。

## 目录结构

```text
tools/ahcli/
├── main.go                    # 入口
├── .ahcli.yaml                # 示例配置文件
├── README.md                  # 本文档
├── cmd/
│   ├── root.go                # 根命令 + 全局 flags + 配置加载
│   ├── helpers.go             # 共用工具函数（resolve*, require*, validate* 等）
│   ├── ads.go                 # ads 命令组（注册所有广告实体子命令 + 登录校验）
│   ├── auth.go                # auth 命令组
│   ├── auth_login.go          # auth login
│   ├── auth_status.go         # auth status
│   ├── auth_tenants.go        # auth tenants
│   ├── auth_switch.go         # auth switch
│   ├── auth_logout.go         # auth logout
│   ├── project.go             # ads project（CRUD + 生命周期 + 运营）
│   ├── project_channel.go     # ads project channel（渠道配置）
│   ├── project_attachment.go  # ads project attachment（项目附件）
│   ├── creative.go            # ads creative
│   ├── copy.go                # ads copy
│   ├── media.go               # ads media（含 tag 子命令）
│   ├── action.go              # ads action
│   ├── plan.go                # ads plan（含 epoch 子命令）
│   ├── channel_meta.go        # ads channel meta（Meta 渠道查询）
│   ├── data.go                # data 命令组
│   ├── browser.go             # browser 命令组
│   └── ops.go                 # ops 命令组（+ 登录校验）
└── internal/
    ├── client.go              # 轻量 HTTP 客户端（标准库 net/http）
    ├── config.go              # 环境 URL 解析
    ├── output.go              # JSON 输出格式化
    ├── actionhub.go           # ActionHub JSON-RPC 调用封装
    ├── json_input.go          # --file / --stdin JSON 输入读取
    └── auth_store.go          # Session 持久化（~/.config/ahcli/hosts.yml）
```

## 设计说明

- **不依赖 pkg/infra** — 使用标准库 `net/http`，避免引入 otel/logger/tracing 初始化开销
- **响应格式兼容** — `APIResponse` 匹配 `pkg/component/response.Response` 结构
- **两种 API 调用模式** — 直调 AdsCore REST API（`client.Get/Post/Put/Delete`）或通过 ActionHub JSON-RPC（`ActionHubCall`）
- **Session 自动注入** — 登录后 `tenant_id`/`user_id` 自动作为 defaults，优先级低于 CLI flag

### 参数解析与登录校验

**登录校验：** `ads` 和 `ops` 命令组通过 `PersistentPreRunE` 统一校验登录状态，未登录时返回友好错误提示。`auth`、`data`、`browser` 不需要登录。

**参数解析两层设计：** `resolve*` 函数从 flag / 配置文件读取值（可为空），`require*` 函数在 `resolve*` 基础上校验非空，为空时返回包含解决方案的错误消息。

| 函数 | 返回 | 用途 |
|------|------|------|
| `resolveTenantID` | `string` | 可选场景（如 get 命令的额外过滤） |
| `requireTenantID` | `string, error` | 必填场景（如 project list） |
| `resolveProjectID` | `string` | 可选场景 |
| `requireProjectID` | `string, error` | 必填场景（如 creative list、action list） |
| `resolveUserID` | `string` | 可选场景 |
| `requireUserID` | `string, error` | 必填场景（如 project create） |

**统一规则：**

| 命令场景 | 参数要求 |
|---------|---------|
| 实体级 list（project list） | `requireTenantID` |
| 项目级 list（creative/copy/media/action/plan list） | `requireProjectID` |
| get / delete（按 ID 查单个） | 位置参数 `args[0]`，tenant_id 可选 |
| create | 必填参数用 `require*`，可选参数用 `resolve*` |
| bind | `requireProjectID` + entity ID flag |

# AWS 基础设施排查

统一的 AWS 排查技能，覆盖认证、ECS 服务状态、ECR 镜像、CloudWatch 日志等。

## 使用方法

- `/aws ecs kong` - 查看 kong 的 ECS 状态（默认 dev）
- `/aws ecs kong prod` - 查看生产环境
- `/aws ecr kong` - 查看 kong 的 ECR 镜像
- `/aws logs kong` - 查看 kong 的 CloudWatch 日志
- `/aws who` - 查看当前 AWS 身份

## 1. AWS 认证

### SSO 架构

项目使用 AWS SSO（IAM Identity Center），通过 `~/.aws/config` 配置 profile。

### 账号与 Profile 映射

| Profile | 用途 | Account ID | Region | SSO Session |
|---------|------|-----------|--------|-------------|
| `dev-us` | **开发环境（主力）** | `856969325542` | `us-west-2` | `non-prod-sso` |
| `prod-us` | 生产环境 | `879794963971` | `us-west-2` | `prod-sso` |
| `dev` | 开发（新加坡） | `222262261015` | `us-west-2` | `non-prod-sso` |
| `prod` | 生产（新加坡） | `314146296901` | `ap-southeast-1` | `prod-sso` |
| `management` | 管理账号 | - | - | - |
| `security` | 安全账号 | - | - | - |
| `data-platform` | 数据平台 | - | - | - |
| `shared-services` | 共享服务 | - | - | - |

### 认证方式

用户在终端中使用 `aws_switch` 函数（定义在 `~/.zshrc`）：

```bash
aws_switch dev-us          # 切换到 dev（自动检查/刷新 SSO）
aws_switch prod-us         # 切换到 prod（需要输入 YES 确认）
aws_switch dev-us --force  # 强制重新登录
aws_who                    # 查看当前身份
```

**`aws_switch` 做了什么：**
1. 检查 profile 的 SSO 凭证是否有效
2. 过期则调用 `aws sso login --sso-session {session}`
3. 清除旧环境变量（`AWS_ACCESS_KEY_ID` 等）
4. `export AWS_PROFILE={profile}`
5. prod 环境需要二次确认（输入 YES）

### Claude Code 中的使用限制

**关键**：`aws_switch` 设置的 `AWS_PROFILE` 环境变量在 Claude Code 的 Bash shell 中**不可见**（不同进程）。

所以 Claude Code 必须：
1. 先尝试 `aws sts get-caller-identity --profile {profile}` 检查凭证
2. 如果失败，**提示用户**在终端中运行 `aws_switch {profile}`
3. 所有 AWS 命令都加 `--profile {profile}` 参数

### 环境到 Profile 的映射规则

```text
dev   → --profile dev-us
prod  → --profile prod-us
```

## 2. ECS 服务排查

### 基础信息

| 环境 | ECS Cluster |
|------|-------------|
| dev | `lexi2-ecs-cluster-dev-usw2` |
| preprod | `lexi2-ecs-cluster-preprod-usw2` |
| prod | `lexi2-ecs-cluster-prod-usw2` |

### 查看服务状态

```bash
aws ecs describe-services \
  --cluster lexi2-ecs-cluster-{env}-usw2 \
  --services {service} \
  --region us-west-2 \
  --profile {profile} \
  --query 'services[0].{status:status, runningCount:runningCount, desiredCount:desiredCount, taskDefinition:taskDefinition, deployments:deployments[*].{status:status, runningCount:runningCount, desiredCount:desiredCount, rolloutState:rolloutState}}' \
  --output json
```

### 查看运行镜像

```bash
# 先从上一步获取 taskDefinition ARN
aws ecs describe-task-definition \
  --task-definition {task-definition-arn} \
  --region us-west-2 \
  --profile {profile} \
  --query 'taskDefinition.containerDefinitions[*].{name:name, image:image}' \
  --output json
```

### 查看服务事件（排查启动失败）

```bash
aws ecs describe-services \
  --cluster lexi2-ecs-cluster-{env}-usw2 \
  --services {service} \
  --region us-west-2 \
  --profile {profile} \
  --query 'services[0].events[:10].{createdAt:createdAt, message:message}' \
  --output json
```

### 查看 Task 详情（排查容器崩溃）

```bash
# 列出 tasks
TASK_ARN=$(aws ecs list-tasks \
  --cluster lexi2-ecs-cluster-{env}-usw2 \
  --service-name {service} \
  --region us-west-2 --profile {profile} \
  --query 'taskArns[0]' --output text)

# 查看容器状态
aws ecs describe-tasks \
  --cluster lexi2-ecs-cluster-{env}-usw2 \
  --tasks "$TASK_ARN" \
  --region us-west-2 --profile {profile} \
  --query 'tasks[0].{lastStatus:lastStatus, startedAt:startedAt, stoppedReason:stoppedReason, containers:containers[*].{name:name, lastStatus:lastStatus, exitCode:exitCode, reason:reason}}' \
  --output json
```

### 查看已停止的 Tasks（排查反复重启）

```bash
aws ecs list-tasks \
  --cluster lexi2-ecs-cluster-{env}-usw2 \
  --service-name {service} \
  --desired-status STOPPED \
  --region us-west-2 --profile {profile} \
  --query 'taskArns[:5]' --output json
```

## 3. ECR 镜像排查

### ECR Registry

```text
009661764016.dkr.ecr.us-west-2.amazonaws.com/sandwichlab/lexi2/{service}
```

### 查看最近的镜像

```bash
aws ecr describe-images \
  --repository-name sandwichlab/lexi2/{service} \
  --region us-west-2 --profile {profile} \
  --query 'sort_by(imageDetails, &imagePushedAt)[-5:].{tags:imageTags, pushed:imagePushedAt, size:imageSizeInBytes}' \
  --output json
```

### 镜像标签格式

| 标签 | 说明 |
|------|------|
| `dev-{service}-2026.03.11.1` | dev 环境构建 |
| `hash-dev-{hash}` | 基于代码 hash 的缓存镜像 |
| `dev-{service}-latest` | 最新 dev 镜像 |

## 4. CloudWatch 日志

### 日志组命名

```text
/ecs/lexi2-{env}/{service}
```

### 查看最近日志

```bash
aws logs tail /ecs/lexi2-{env}/{service} \
  --since 30m \
  --region us-west-2 --profile {profile} \
  --format short
```

### 搜索日志

```bash
aws logs filter-log-events \
  --log-group-name /ecs/lexi2-{env}/{service} \
  --start-time $(date -v-1H +%s000) \
  --filter-pattern "ERROR" \
  --region us-west-2 --profile {profile} \
  --query 'events[:20].{time:timestamp, msg:message}' \
  --output json
```

## 5. 常见排查流程

### 部署后服务不正常

```text
1. /aws ecs {service} {env}     → 确认服务状态和 running count
2. 查看 task definition         → 确认镜像版本是否正确
3. 查看 events                  → 是否有健康检查失败
4. 查看 stopped tasks           → 容器是否崩溃重启
5. /aws logs {service} {env}    → 查看应用日志
```

### 服务 0/1 无法启动

```text
1. 查看 events → "unable to place a task" = 资源不足
2. 查看 stopped tasks → exitCode != 0 = 应用崩溃
3. 查看 logs → 启动报错信息
4. 检查 task definition → 端口、环境变量、镜像是否正确
```

### 镜像没更新

```text
1. /aws ecr {service}           → 确认最新镜像 tag 和推送时间
2. /aws ecs {service}           → 对比 task definition 中的镜像
3. deploy-selective 的 hash 机制会跳过代码未变的服务
4. 用 force_rebuild=true 强制重建
```

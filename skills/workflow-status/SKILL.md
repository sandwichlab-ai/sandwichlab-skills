# GitHub Workflow 状态查询

查询 GitHub Actions workflow 运行状态，排查部署失败原因。

## 使用方法

- `/workflow-status` - 查看最近的 workflow 运行
- `/workflow-status deploy` - 查看部署相关的 workflow
- `/workflow-status {run_id}` - 查看指定 run ID 的详情
- `/workflow-status pr {pr_number}` - 查看 PR 触发的部署

## 执行步骤

### 1. 列出最近的运行

```bash
# 部署 workflow
gh run list --workflow=deploy-selective.yml --limit 5 \
  --json databaseId,status,conclusion,createdAt,displayTitle \
  --jq '.[] | "\(.databaseId)\t\(.status)\t\(.conclusion)\t\(.createdAt)\t\(.displayTitle)"'

# PR staging workflow
gh run list --workflow=pr-staging.yml --limit 5 \
  --json databaseId,status,conclusion,createdAt,displayTitle \
  --jq '.[] | "\(.databaseId)\t\(.status)\t\(.conclusion)\t\(.createdAt)\t\(.displayTitle)"'

# Kong 配置部署
gh run list --workflow=kong-deploy.yml --limit 5 \
  --json databaseId,status,conclusion,createdAt,displayTitle \
  --jq '.[] | "\(.databaseId)\t\(.status)\t\(.conclusion)\t\(.createdAt)\t\(.displayTitle)"'

# CI 检查
gh run list --workflow=ci.yml --limit 5 \
  --json databaseId,status,conclusion,createdAt,displayTitle \
  --jq '.[] | "\(.databaseId)\t\(.status)\t\(.conclusion)\t\(.createdAt)\t\(.displayTitle)"'
```

### 2. 查看运行详情

```bash
# 查看所有 jobs 状态
gh run view {run_id} --json status,conclusion,jobs \
  --jq '{status, conclusion, jobs: [.jobs[] | {name, status, conclusion}]}'

# 只看失败的 jobs
gh run view {run_id} --json jobs \
  --jq '.jobs[] | select(.conclusion=="failure") | {name, steps: [.steps[] | select(.conclusion=="failure") | .name]}'
```

### 3. 查看失败日志

```bash
# 获取失败 job 的 ID
gh run view {run_id} --json jobs \
  --jq '.jobs[] | select(.conclusion=="failure") | .id'

# 查看日志（搜索关键错误信息）
gh run view {run_id} --log 2>/dev/null | grep "{job_name}" | grep -i "error\|fail\|cannot\|not found\|denied\|exit code" | head -20

# 查看特定 step 的完整日志
gh run view {run_id} --log 2>/dev/null | grep "{job_name}.*{step_name}" -A 30 | head -50
```

### 4. 查看 PR 触发的部署

```bash
# 查看 PR 的检查状态
gh pr checks {pr_number}

# 查看 PR 触发的 workflow runs
gh run list --branch {branch_name} --limit 5 \
  --json databaseId,status,conclusion,workflowName \
  --jq '.[] | "\(.databaseId)\t\(.status)\t\(.conclusion)\t\(.workflowName)"'
```

## 常用 Workflow

| Workflow | 文件 | 触发方式 | 说明 |
|---------|------|---------|------|
| deploy-selective | `deploy-selective.yml` | 手动 / pr-staging 触发 | 服务部署（构建镜像 + ECS 更新） |
| pr-staging | `pr-staging.yml` | PR 创建/更新时自动触发 | 合并 open PRs → 检测服务 → 触发部署 |
| Deploy Kong Configuration | `kong-deploy.yml` | 手动触发 | Kong YAML 配置同步（deck sync） |
| CI | `ci.yml` | PR / push | 代码检查、测试、Kong 配置验证 |
| db-migrate | `db-migrate.yml` | 手动触发 | 数据库迁移 |

## 部署流程排查

### PR 提交后没有触发部署？

1. 检查 pr-staging 是否运行：`gh run list --workflow=pr-staging.yml --limit 3`
2. 检查服务检测结果：在 run 日志里搜索 `Detected service`
3. **已知问题**：`gh pr view --json files` 最多返回 100 个文件，大 PR 可能漏检服务（已修复用 `gh api --paginate`）

### 部署触发了但失败？

1. 查看哪个 job 失败：`gh run view {run_id} --json jobs --jq '.jobs[] | select(.conclusion=="failure") | .name'`
2. 常见失败原因：
   - **Validate Kong Config**: Kong YAML 验证不过 → `/kong-validate`
   - **Build Docker Image**: 编译错误 → 查看 build 日志
   - **Deploy to ECS**: 服务不稳定 → `/ecs-status {service}`
   - **Deploy to ECS 超时（600s）**: 健康检查失败、CDK 未部署、端口不对

### 部署成功但镜像没更新？

- CI 用 content hash 做缓存，代码没变则复用已有镜像
- 日志里会显示 `♻️ Reused existing image` 或 `✅ Target image is already running`
- 强制重建：`gh workflow run deploy-selective.yml -f services={svc} -f environment={env} -f force_rebuild=true`

## 状态说明

| 状态 | 说明 |
|------|------|
| queued | 排队中 |
| in_progress | 运行中 |
| completed/success | 成功 |
| completed/failure | 失败 |
| cancelled | 被新 run 取消（正常，pr-staging 会取消旧 run） |

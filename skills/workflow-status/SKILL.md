# GitHub Workflow 状态查询

查询 GitHub Actions workflow 运行状态。

## 使用方法

- `/workflow-status` - 查看最近的 workflow 运行
- `/workflow-status deploy` - 查看部署相关的 workflow
- `/workflow-status 12345678` - 查看指定 run ID 的详情

## 执行步骤

```bash
# 列出最近的运行
gh run list --limit 10

# 查看特定 workflow
gh run list --workflow deploy-selective.yml --limit 5

# 查看运行详情
gh run view {run_id}

# 查看运行日志
gh run view {run_id} --log
```

## 常用 Workflow

| Workflow | 说明 |
|---------|------|
| deploy-selective.yml | 服务部署 |
| db-migrate.yml | 数据库迁移 |
| kong-deploy.yml | Kong 配置部署 |
| ci.yml | CI 检查 |

## 状态说明

| 状态 | 说明 |
|------|------|
| queued | 排队中 |
| in_progress | 运行中 |
| completed/success | 成功 |
| completed/failure | 失败 |
| cancelled | 已取消 |

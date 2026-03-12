# 服务部署

部署服务到指定环境。支持 Go 微服务、Kong 网关、Python Agent 等。

## 使用方法

- `/deploy adscore dev` - 部署 adscore 到 dev 环境
- `/deploy kong dev` - 部署 Kong 网关（Lua 插件变更）
- `/deploy chatcore,opscore preprod` - 部署多个服务到 preprod
- `/deploy authcenter prod` - 部署到生产环境（需确认）

## 可用服务

```text
actionhub adscore algocore authcenter browser_service capi_gateway
chatcore config_center creatives_grow data_gather data_syncer
devops_portal hello hui kong llm_gateway meta_api_gateway opscore
playground_apps sgtm_preview_server sgtm_tag_server superset workstation
```

## 两种部署机制

### 1. 服务部署（deploy-selective.yml）

Go 微服务、Kong、前端等应用代码部署：

```bash
gh workflow run deploy-selective.yml \
  -f environment=dev \
  -f services=kong \
  -f source_ref=dev_prs
```

**Kong 特殊说明**：Kong Lua 插件（`apps/kong/plugins/`）的变更需要**重建 Kong Docker 镜像**，走 `deploy-selective.yml`，而不是 Kong YAML 配置部署。

### 2. Kong YAML 配置部署（Deploy Kong Configuration）

仅用于 `configs/kong/*.yaml` 的路由/插件配置变更：

```bash
# 通过 GitHub Actions UI 手动触发 "Deploy Kong Configuration" workflow
# 或在 PR 合并后自动触发
```

| 变更类型 | 部署方式 |
|---------|---------|
| `apps/kong/plugins/*.lua` | `deploy-selective.yml`（重建镜像） |
| `configs/kong/*.yaml` | `Deploy Kong Configuration`（deck sync） |
| 两者都改了 | 两个都要部署 |

## 环境与分支映射

| 环境 | 默认分支 | source_ref |
|------|---------|------------|
| dev | dev | `dev_prs`（PR staging）或 `dev` |
| preprod | main | `main_prs` 或 `main` |
| prod | main | `main`（需确认） |

## 执行步骤

1. 解析用户输入的服务名和环境
2. 验证服务名是否在可用列表中
3. 触发部署
4. 使用 `/workflow-status {run_id}` 监控部署状态
5. 部署完成后可用 `/ecs-status {service} {env}` 确认

## 命令示例

```bash
# 单服务部署
gh workflow run deploy-selective.yml -f environment=dev -f services=adscore

# 多服务部署
gh workflow run deploy-selective.yml -f environment=dev -f services=adscore,chatcore

# 强制重建（跳过镜像缓存）
gh workflow run deploy-selective.yml -f environment=dev -f services=kong -f force_rebuild=true

# 生产部署
gh workflow run deploy-selective.yml -f environment=prod -f services=authcenter -f source_ref=main
```

## 部署失败排查

| 失败步骤 | 常见原因 | 解决方法 |
|---------|---------|---------|
| Validate Kong Config | Kong YAML 配置验证不过 | 运行 `/kong-validate` 查看详情 |
| Build Docker Image | 代码编译错误 | 查看 build 日志 |
| Deploy to ECS (超时) | 健康检查失败、端口不对、CDK 未部署 | `/ecs-status` 查看 task 状态 |
| Image hash reuse | 代码没变，复用缓存镜像 | 用 `force_rebuild=true` 强制重建 |

## 注意事项

- prod 部署**必须**让用户确认后再触发
- PR staging 部署（`dev_prs`/`main_prs`）会合并所有 open PR 的代码一起部署
- 部署后建议用 `/ecs-status` 确认服务状态
- Kong 的 Lua 插件和 YAML 配置是**两套独立的部署流程**

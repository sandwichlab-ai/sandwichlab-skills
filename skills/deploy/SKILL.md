# 服务部署

部署 Go 微服务到指定环境。

## 使用方法

- `/deploy adscore dev` - 部署 adscore 到 dev 环境
- `/deploy chatcore,opscore preprod` - 部署多个服务到 preprod
- `/deploy authcenter prod` - 部署到生产环境

## 可用服务

adscore, actionhub, authcenter, chatcore, data_syncer, hello, opscore, llm-gateway, workstation

## 环境与分支映射

| 环境 | 分支 | 说明 |
|------|------|------|
| dev | dev | 开发环境 |
| preprod | main | 预生产环境 |
| prod | main | 生产环境（需确认） |

## 执行步骤

1. 解析用户输入的服务名和环境
2. 验证服务名是否有效
3. 使用 `gh workflow run deploy-selective.yml` 触发部署
4. 监控部署状态

## 命令示例

```bash
# 单服务部署
gh workflow run deploy-selective.yml -f environment=dev -f services=adscore

# 多服务部署
gh workflow run deploy-selective.yml -f environment=dev -f services=adscore,chatcore

# 生产部署（需要 source_ref）
gh workflow run deploy-selective.yml -f environment=prod -f services=authcenter -f source_ref=main
```

## 注意事项

- prod 部署需要用户确认
- 部署后应检查服务健康状态
- 如果部署失败，查看 GitHub Actions 日志

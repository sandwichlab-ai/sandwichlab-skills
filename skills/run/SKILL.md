# 本地运行服务

在本地开发环境运行 Go 微服务。

## 使用方法

- `/run adscore` - 运行 adscore 服务
- `/run chatcore` - 运行 chatcore 服务

## 可用服务及端口

| 服务 | HTTP 端口 | gRPC 端口 |
|-----|---------|---------|
| authcenter | 8080 | - |
| config_center | 8081 | 9081 |
| hello | 8082 | 9082 |
| adscore | 8083 | 9083 |
| chatcore | 8084 | - |
| actionhub | 8085 | 9085 |
| data_syncer | 8086 | - |
| opscore | 8088 | 9088 |

## 执行步骤

1. 确保本地开发环境已启动（PostgreSQL, Redis 等）
2. 设置 ENV=local 环境变量
3. 运行服务

```bash
# 启动依赖服务
make dev

# 运行服务
ENV=local go run apps/{service}/main.go

# 或使用 make
make run service={service}
```

## 前置条件

1. **本地依赖服务**（通过 `make dev` 启动）：
   - PostgreSQL: localhost:5432
   - Redis: localhost:6379
   - LocalStack: localhost:4566

2. **配置文件**：
   - 确保 `apps/{service}/config/bootstrap.local.yaml` 存在

3. **依赖注入**（如果修改了依赖）：

   ```bash
   cd apps/{service} && wire
   ```

## 调试

- 查看日志输出
- 使用 `make dev-logs` 查看依赖服务日志
- 访问 Jaeger UI (<http://localhost:16686>) 查看链路追踪

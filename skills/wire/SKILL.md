# Wire 依赖注入生成

为 Go 服务生成 Wire 依赖注入代码。

## 使用方法

- `/wire adscore` - 为 adscore 生成依赖注入代码
- `/wire all` - 为所有服务生成

## 执行步骤

```bash
# 单个服务
cd apps/{service} && wire

# 所有服务
make generate cmd=wire

# 指定服务
make generate cmd=wire service=adscore
```

## 何时需要运行

- 添加新的 Provider
- 修改依赖关系
- 新增 Service/Repository/Handler

## 常见问题

### 循环依赖

```text
wire: ... has a cycle
```

解决：检查依赖关系，打破循环

### 缺少 Provider

```text
wire: no provider found for ...
```

解决：在 `wire.go` 中添加缺失的 Provider

## 项目结构

```text
apps/{service}/
├── wire.go           # Wire 配置
├── wire_gen.go       # 生成的代码（不要手动修改）
└── providers/        # Provider 定义
    ├── repository.go
    ├── service.go
    └── handler.go
```

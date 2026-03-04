# 代码检查

运行 Go 代码静态检查。

## 使用方法

- `/lint` - 检查所有代码
- `/lint adscore` - 检查指定服务
- `/lint fix` - 自动修复可修复的问题

## 执行步骤

```bash
# 检查所有
golangci-lint run ./...

# 检查指定目录
golangci-lint run ./apps/adscore/...

# 自动修复
golangci-lint run --fix ./...

# 使用 make
make lint
```

## 配置文件

`.golangci.yml` - 位于项目根目录

## 常见问题及修复

### unused variable

```go
// 删除未使用的变量或使用 _
_ = unusedVar
```

### error not checked

```go
// 错误必须处理
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed: %w", err)
}
```

### context as first parameter

```go
// Context 必须作为第一个参数
func DoSomething(ctx context.Context, param string) error
```

## 忽略规则

```go
//nolint:errcheck // reason for ignoring
```

**注意：** 尽量不要使用 nolint，除非有充分理由。

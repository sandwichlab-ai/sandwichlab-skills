# Proto 代码生成

生成 Protobuf 和 gRPC 代码。

## 使用方法

- `/proto` - 生成所有 Proto 代码
- `/proto lint` - 检查 Proto 格式
- `/proto breaking` - 检查 breaking changes

## 执行步骤

```bash
# 生成代码
make generate cmd=proto

# 格式检查
make generate cmd=proto-lint

# Breaking changes 检查
make generate cmd=proto-breaking
```

## Proto 文件位置

```text
proto/
├── adscore/
│   └── v1/
│       └── adscore.proto
├── chatcore/
│   └── v1/
│       └── chatcore.proto
└── ...
```

## 生成的代码位置

```text
pkg/proto/
├── adscore/
│   └── v1/
│       ├── adscore.pb.go
│       └── adscore_grpc.pb.go
└── ...
```

## 规范

- 使用 `buf` 进行格式化和检查
- 版本号使用 `v1`, `v2` 等目录
- 保持向后兼容，避免 breaking changes

## 常见问题

### buf 未安装

```bash
brew install bufbuild/buf/buf
```

### 格式不正确

```bash
buf format -w
```

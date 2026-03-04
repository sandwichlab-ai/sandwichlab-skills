# 数据库迁移

执行数据库 Schema 迁移操作。

## 使用方法

- `/db-migrate diff adscore dev` - 检测 adscore 在 dev 环境的 schema 差异
- `/db-migrate exec adscore dev 20260119171412` - 执行迁移
- `/db-migrate status adscore dev` - 查看迁移状态

## 可用服务

adscore, actionhub, authcenter, chatcore, data_syncer, opscore, workstation

## 可用环境

dev, preprod, prod

## 执行步骤

### diff 操作

1. 触发 workflow 检测 schema 差异
2. 等待 workflow 完成
3. 如果有差异，会自动创建 PR

```bash
gh workflow run db-migrate.yml --ref dev \
  -f environment=dev \
  -f app=adscore \
  -f action=diff
```

### exec 操作

1. 从 diff 生成的 PR 中获取 timestamp
2. 执行迁移

```bash
gh workflow run db-migrate.yml --ref dev \
  -f environment=dev \
  -f app=adscore \
  -f action=exec \
  -f timestamp=20260119171412
```

## 注意事项

- 执行 exec 前必须先运行 diff
- timestamp 参数从 diff 生成的 PR 标题或内容中获取
- prod 环境迁移需要特别谨慎
- 迁移后检查服务是否正常运行

## Model 规范

所有数据库 Model 必须实现 `GetIndexes()` 方法，否则 diff 会失败：

```go
func (MyModel) GetIndexes() []dbgen.Index {
    return nil  // 或返回自定义索引
}
```

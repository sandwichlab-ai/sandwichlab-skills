# Kong 配置验证

验证 Kong 网关配置的完整性和正确性。

## 使用方法

- `/kong-validate` - 验证所有 Kong 配置
- `/kong-validate preprod` - 验证 preprod 环境配置

## 配置文件位置

```text
configs/kong/
├── preprod/
│   └── kong.yaml
└── prod/
    └── kong.yaml
```

## 验证规则

### 1. Service 与 Plugin 对应

每个需要认证的 service 必须有对应的 plugin 配置：

```yaml
# ✅ 正确
services:
  - name: my-api
    routes: [...]

plugins:
  - name: custom-jwt-auth
    service: my-api
  - name: cors
    service: my-api

# ❌ 错误（缺少 plugin）
services:
  - name: my-api
    routes: [...]
# 没有对应的 plugin 配置！
```

### 2. 必需的 Plugin 类型

| 服务类型 | 必需插件 |
|---------|---------|
| API 服务（需要 JWT） | `custom-jwt-auth` + `cors` |
| 使用 key-auth 的服务 | `key-auth` + `cors` |
| OAuth 端点 | 仅 `cors` |
| 健康检查端点 | 无需插件 |

### 3. 路径规范

| 路径模式 | 鉴权要求 |
|---------|---------|
| `/{service}/public/api/v1/*` | 无鉴权 |
| `/{service}/api/v1/*` | 需要鉴权 |

## 验证步骤

1. 读取 Kong 配置文件
2. 提取所有 service 定义
3. 检查每个 service 是否有对应的 plugin
4. 报告缺失的配置

## 修复建议

如果发现配置缺失，需要在 `plugins` 部分添加对应配置：

```yaml
plugins:
  - name: custom-jwt-auth
    service: {missing-service-name}
    config:
      header_name: authorization
  - name: cors
    service: {missing-service-name}
    config:
      origins: ["*"]
```

## 部署

修改配置后，手动触发 `Deploy Kong Configuration` workflow。

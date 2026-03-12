# Kong 配置验证

验证 Kong 配置文件中的服务是否都有对应的插件配置，防止遗漏认证插件导致的安全问题。

## 使用方法

- `/kong-validate` - 验证所有环境的 Kong 配置
- `/kong-validate preprod` - 验证指定环境

## 参数

- `$ARGUMENTS`: 可选，指定要检查的配置文件路径或环境（dev/preprod/prod），默认检查所有环境

## 执行步骤

### 1. 读取 Kong 配置文件

检查以下目录中的配置文件：
- `configs/kong/dev/*.yaml`
- `configs/kong/preprod/*.yaml`
- `configs/kong/prod/*.yaml`

### 2. 解析服务和插件

从每个配置文件中提取：
- `services:` 部分定义的所有服务名称
- `plugins:` 部分中 `service:` 字段引用的服务名称
- `plugins:` 部分中 `custom-jwt-auth` 和 `rate-limiting` 插件的 `config:` 字段

### 3. 验证规则

#### 3.1 插件存在性检查

对于每个需要认证的服务（排除以下例外），必须配置：

| 服务类型 | 必需插件 |
|---------|---------|
| API 服务（需要 JWT） | `custom-jwt-auth` + `cors` |
| 使用 key-auth 的服务 | `key-auth` + `cors` |
| OAuth 端点 | 仅 `cors`（无需认证） |
| 公开接口（名称含 `public`） | 仅 `cors`（无需认证） |
| 健康检查端点 | 无需插件 |

**例外情况**（不需要 JWT 认证）：
- 服务名包含 `health` 的健康检查服务
- 服务名包含 `oauth-token` 的 OAuth 端点
- 服务名包含 `pre-authorize` 的预授权端点（使用 key-auth）
- 服务名包含 `public` 的公开接口

#### 3.2 过度认证检查

例外服务不应配置 JWT 认证插件，否则会导致功能异常：
- `health` 服务配置了 `custom-jwt-auth` → 监控系统无法访问
- `oauth-token` 服务配置了 `custom-jwt-auth` → 客户端无法获取 token
- `public` 服务配置了 `custom-jwt-auth` → 公开接口失去意义

#### 3.3 Rate-Limiting 插件配置检查

使用 `policy: redis` 的 `rate-limiting` 插件必须显式配置 `redis.host` 和 `redis.port`，否则 Kong 会拒绝配置（HTTP 400）。

```yaml
# ❌ 错误：缺少 redis 配置
- name: rate-limiting
  config:
    policy: redis  # 需要 redis 但没配置 host → Kong 400

# ✅ 正确
- name: rate-limiting
  config:
    policy: redis
    redis:
      host: clustercfg.lexi2-prod-usw2-xxx.cache.amazonaws.com
      port: 6379
```

#### 3.4 custom-jwt-auth 字段完整性检查（关键）

对于每个 `custom-jwt-auth` 插件，检查 `config` 中是否包含以下**全部必需字段**：

| 字段 | 说明 | 缺失后果 |
|------|------|---------|
| `cognito_issuer` | Cognito User Pool issuer URL | RS256 token 验证失败 |
| `platform_auth_issuer` | Platform Auth issuer | RS256 platform auth token 验证失败 |
| `platform_auth_jwks_url` | JWKS 端点 URL | 无法获取公钥验证签名 |
| `platform_token_issuer` | Platform Token issuer | HS256 token issuer 验证失败 |
| `platform_token_secret` | HS256 签名密钥 | **500 Server Error** |
| `jwks_cache_ttl` | JWKS 缓存时间 | 性能问题 |
| `required_scopes` | 必需的 scope 列表 | 可以为空数组 `[]` |

**SM 引用格式要求**：

`platform_token_secret` 在 prod/preprod **必须**使用 Secrets Manager 引用格式，匹配正则：

```text
^\$\{sm:@security:sandwichlab/infra/(prod|preprod)/kong/[^:]+:platform_key\}$
```

即必须包含：`${sm:@security:` + `sandwichlab/infra/{env}/kong/` + `{secret-name}:platform_key}`

示例：`${sm:@security:sandwichlab/infra/prod/kong/shoplazza:platform_key}`

dev 环境可以使用硬编码测试值。

#### 3.5 环境一致性检查

验证 dev/preprod/prod 三个环境的同名配置文件：
- 服务列表应一致（服务名和数量）
- 插件列表应一致（每个服务配置的插件类型）
- 仅配置值可以不同（如 SM 路径、超时时间等）

### 4. 输出报告

输出格式：

```text
## Kong 配置验证报告

### configs/kong/prod/shoplazza-external.yaml

✅ shoplazza-pre-authorize-external
   - key-auth: ✓
   - cors: ✓

✅ shoplazza-api-external
   - custom-jwt-auth: ✓ (7/7 字段完整)
   - cors: ✓

❌ ads-management-external
   - custom-jwt-auth: ✗ 缺少字段: platform_token_secret, platform_token_issuer
   - cors: ✓

### 环境一致性检查
✅ shoplazza-external.yaml: dev/preprod/prod 服务和插件结构一致
❌ chatcore.yaml: prod 缺少 platform_token_secret（dev/preprod 已配置）

### 总结
- 检查文件数: 46
- 检查服务数: 113
- Errors: 0
- Warnings: 3
```

### 5. 修复建议

对于缺失配置的服务，提供修复模板：

```yaml
# 需要添加到 plugins: 部分
- name: custom-jwt-auth
  service: {service-name}
  config:
    cognito_issuer: "https://cognito-idp.us-west-2.amazonaws.com/us-west-2_8D8Pp5UcN"
    platform_auth_issuer: "lexi-platform"
    platform_auth_jwks_url: "http://authcenter-dns.lexi2-{env}.svc:8080/.well-known/jwks.json"
    platform_token_issuer: "lexi-platform"
    platform_token_secret: "${sm:@security:sandwichlab/infra/{env}/kong/shoplazza:platform_key}"
    jwks_cache_ttl: 60      # dev: 60, preprod/prod: 300
    required_scopes: []

- name: cors
  service: {service-name}
  config:
    origins: ["*"]
    methods: [GET, POST, PUT, DELETE, OPTIONS]
    headers: [Content-Type, Authorization, X-Request-ID, Platform]
    credentials: false
    max_age: 86400
```

**dev 环境特殊值**：

```yaml
platform_auth_issuer: "http://authcenter-dns.lexi2-dev.svc:8080"
platform_token_secret: "local-dev-jwt-secret-key-for-testing-only"  # pragma: allowlist secret
```

## 验证范围说明

**本 Skill 检查的内容：**
- 服务与插件的对应关系（缺少认证 / 过度认证）
- `custom-jwt-auth` 插件字段完整性（7 个必需字段）
- `platform_token_secret` 的 SM 引用格式（prod/preprod）
- `rate-limiting` 插件的 Redis 配置完整性
- 跨环境服务列表一致性

**本 Skill 不检查的内容：**
- Kong 配置 YAML 语法正确性（由 Kong 自身和 CI 的 `yaml.safe_load` 验证）
- 路由路径冲突或 `regex_priority` 合理性
- Upstream 服务可用性或健康检查配置
- CORS 配置的具体值是否合理
- 插件配置值的业务正确性（如 `jwks_cache_ttl` 具体数值）

## 相关文件

- Kong 配置目录: `configs/kong/`
- Kong 插件代码: `apps/kong/plugins/`
- Kong 插件 handler: `apps/kong/plugins/custom-jwt-auth/handler.lua`
- 验证脚本: `scripts/validate-kong-config.py`
- 配置文档: `configs/kong/README.md`

## 注意事项

- 新增服务时，**必须**同时添加对应的插件配置
- 不同环境（dev/preprod/prod）可能有不同的配置值，但结构应该一致
- `platform_token_secret` 缺失会导致 500 错误，不是 401（这是已知的 handler.lua 行为）
- `rate-limiting` 使用 `policy: redis` 时缺少 `redis.host` 会导致 Kong 拒绝配置（HTTP 400）
- 修复配置问题时，使用全局搜索确认没有其他遗漏：`grep -r "关键词" configs/kong/prod/ configs/kong/preprod/`
- dev 环境硬编码的测试 secret 必须添加 `# pragma: allowlist secret` 注释，避免 secret 扫描工具误报
- 修改后记得触发 `Deploy Kong Configuration` workflow

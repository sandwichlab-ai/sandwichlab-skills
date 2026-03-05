---
name: hui-ads-launch
description: "Meta 广告端到端发布完整流程。从零创建并发布 Meta 广告，使用 ahcli CLI 逐步执行。使用场景：用户要求创建广告、发布新广告，如创建广告、发布新广告、launch ads。"
---

# Meta 广告端到端发布

从零创建并发布 Meta 广告的完整流程。使用 `ahcli` CLI 工具逐步执行。

## 使用方法

- `/hui-ads-launch` — 交互式引导，逐步收集信息并执行
- `/hui-ads-launch <产品描述或需求>` — 根据描述自动填充参数
- `/hui-ads-launch --env preprod` — 指定环境（默认 dev）

## 前提

- 已安装 `ahcli` CLI 工具（使用 `/ahcli-install` 安装）

## 用户需提供的信息

| 信息 | 必填 | 示例 |
|------|------|------|
| HUI 用户编号 | 是 | `1`（默认） |
| 产品/服务名称 | 是 | `"智能手表 Pro"` |
| Landing page URL | 是 | `"https://example.com/product"` |
| 广告目标 | 是 | OUTCOME_SALES / OUTCOME_TRAFFIC / OUTCOME_AWARENESS |
| 日预算（美元） | 是 | `50` = $50/天 |
| 货币 | 是 | `"USD"` |
| 素材图片 | 否 | S3 URL 列表 / AI 生成 / 跳过 |
| 广告文案 | 否 | 标题 + 正文（无则 Claude 自动生成） |
| Meta Pixel ID | 否 | `"123456789"` |
| 受众定向 | 否 | 年龄、性别、地域 |

**预算说明：** `daily-budget` 单位是美元，`50` = $50/天。系统内部自动转换为美分。

---

## 执行流程

**推荐执行顺序：** Steps 0→1→2→3→4→5(可选)→6→8a(skip)→7→8b

使用 `creative_id` 格式时，**必须先 activate 项目（Step 8a）再创建广告（Step 7）**。

### Step 0: 前置 — 登录 & 凭证获取

#### 0a. 登录并获取 Tenant ID

先检查登录状态，未登录则执行登录：

```bash
ahcli auth status --env $ENV
```

```bash
ahcli auth login --env $ENV
```

登录后 `tenant_id` 和 `user_id` 自动注入后续命令，无需重复指定。

#### 0b. 获取 Meta 凭证

```bash
ahcli ops get-credentials
```

一条命令获取投放所需的全部凭证（自动使用当前登录用户），输出：

```json
{
  "asset_code": "asset-001",
  "ad_account_id": "act_123456789",
  "page_id": "109876543",
  "pixel_id": "654321",
  "access_token": "EAAxxxx...",
  "allocation_code": "alloc-xxx"
}
```

**保存以下值供后续使用：** `ACCESS_TOKEN`, `AD_ACCOUNT_ID`, `PAGE_ID`, `PIXEL_ID`（可选）

---

### Step 1: 创建项目

```bash
ahcli ads project create \
  --name "项目名称" --description "项目描述"
```

`source_platform` 自动设为 `hui`。

---

### Step 2: 设置产品信息

```bash
echo '{
  "url": "https://example.com/product",
  "landing_page_url": "https://example.com/product",
  "product_info": "产品描述",
  "business_name": "商家名称",
  "locale": "en_US",
  "call_to_action": "SHOP_NOW",
  "selling_points": ["卖点1", "卖点2"]
}' | ahcli ads project attachment upsert \
  --project-id $PROJECT_ID --type url_analysis --name product_info --stdin
```

**重要：** `url` 和 `landing_page_url` 必须同时设置且值相同。缺少 `url` 会导致落地页 fallback 到 `example.com`。

---

### Step 3: 配置投放提案

```bash
echo '{
  "currency": "USD",
  "channels": ["Meta"],
  "daily_budget": "50",
  "total_budget": 0,
  "ads_goal": "OUTCOME_SALES",
  "campaign_objective": "OUTCOME_SALES",
  "bid_strategy": "LOWEST_COST_WITHOUT_CAP",
  "campaign_name": "广告系列名称",
  "audience_targeting": {
    "age_min": 18, "age_max": 65,
    "genders": [0],
    "geo_locations": {"countries": ["US"]}
  }
}' | ahcli ads project attachment upsert \
  --project-id $PROJECT_ID --type proposal --name campaign_proposal --stdin
```

| 枚举字段 | 可选值 |
|---------|--------|
| campaign_objective | `OUTCOME_SALES`, `OUTCOME_TRAFFIC`, `OUTCOME_AWARENESS` |
| bid_strategy | `LOWEST_COST_WITHOUT_CAP`（推荐）, `LOWEST_COST_WITH_BID_CAP`, `COST_CAP` |

---

### Step 4: 配置投放渠道

```bash
ahcli ads project channel create $PROJECT_ID --channel Meta
```

单渠道 `budget_allocation` 自动设为 `1.0`。

---

### Step 5: 配置转化追踪（可选）

**跳过条件：** 无 pixel_id 时跳过。

```bash
ahcli ads project channel set-pixel $PROJECT_ID \
  --pixel-id $PIXEL_ID --conversion-event Purchase
```

---

### Step 6: 创建素材（图片 + 文案 + 组合）

#### 6a. 获取图片素材

**方式一：AI 生成**

```bash
ahcli ads media generate \
  --project-id $PROJECT_ID \
  --product-name "产品名称" --selling-point "核心卖点" --count 2
```

响应提取 `execution_arn`，然后轮询结果（自动 30s 间隔）：

```bash
ahcli ads media poll "$EXECUTION_ARN" \
  --project-id $PROJECT_ID --wait
```

状态：`SUCCEEDED` → 提取 `asset_output[].asset_id`（`masset-xxx`），`FAILED` → 查看 error。

**注意：** AI 生成的图片不会自动绑定到项目素材库，需要手动绑定（见下方 6a-bind）。如果 `asset_output` 为空，用 `ahcli ads media search --project-id $PROJECT_ID --type image` 在租户素材库中查找最近生成的素材。

**方式二：从 S3 导入**

```bash
ahcli ads media import-s3 --project-id $PROJECT_ID \
  --s3-urls "s3://bucket/img1.png,s3://bucket/img2.png"
```

#### 6a-bind. 绑定素材到项目（必须）

AI 生成或 S3 导入的素材需要手动绑定到项目，否则项目素材库中看不到。**每个素材调用一次：**

```bash
ahcli ads media bind --asset-id masset-xxx --project-id $PROJECT_ID --action add
```

#### 6b. 创建 Copy（广告文案）

每组文案调用一次。**可并行创建多组 Copy。**

文案来源：用户提供 → 直接使用；未提供 → Claude 根据 Step 2 产品信息生成 2-3 组（headline ≤ 25 chars, copy_text ≤ 125 chars）。

```bash
ahcli ads copy create --project-id $PROJECT_ID \
  --headline "Your Data, Clear Reports" \
  --copy-text "Transform messy data into polished reports in minutes." \
  --cta LEARN_MORE
```

响应提取 `copy_id`。收集所有 copy_id。

#### 6c. 创建 Creative（素材组合）

```bash
ahcli ads creative create --project-id $PROJECT_ID \
  --asset-ids "masset-xxx,masset-yyy" \
  --copy-ids "copy-xxx,copy-yyy"
```

响应提取 `creative_id`。

---

### Step 7: 创建广告（create_campaign_composite）

使用 ABO 模式一次性创建 Campaign + AdSet + Ads。

#### ABO vs CBO（重要）

| 模式 | Campaign `budget` | AdSet `daily_budget` |
|------|-------------------|---------------------|
| **ABO（推荐）** | **不设置** | 设置 |
| CBO | 设置 | 不设置 |
| **同时设置** | — | **Meta 会拒绝！** |

#### 使用 Flag 快捷方式

```bash
ahcli ads action create-sync \
  --project-id $PROJECT_ID \
  --access-token "$ACCESS_TOKEN" --ad-account-id "$AD_ACCOUNT_ID" \
  --campaign-name "Smart Watch Q1" \
  --objective OUTCOME_TRAFFIC \
  --daily-budget 50 \
  --creative-id $CREATIVE_ID \
  --countries US \
  --start-time "2026-03-01T00:00:00Z" --end-time "2026-03-31T23:59:59Z"
```

内置业务知识：
- `objective` → `optimization_goal` 自动映射（OUTCOME_TRAFFIC→LINK_CLICKS, OUTCOME_SALES→OFFSITE_CONVERSIONS, OUTCOME_AWARENESS→REACH）
- ABO 模式：budget 只设在 AdSet 级别
- status 默认 `PAUSED`（安全第一）
- bid_strategy 默认 `LOWEST_COST_WITHOUT_CAP`

#### 使用 JSON 文件（完全控制 payload）

```bash
ahcli ads action create-sync --file campaign.json
cat campaign.json | ahcli ads action create-sync --stdin
```

**响应检查：** 确认 `success: true`，提取 `metadata` 中的 campaign_id、adset_ids、ad_ids。

---

### Step 8: 激活 & 验证

#### 8a. 激活项目

```bash
ahcli ads project activate $PROJECT_ID --skip-initial-epoch
```

`--skip-initial-epoch`：广告已通过 Step 7 手动创建时使用，跳过 Epoch 0 的自动广告组装。

#### 8b. 验证状态

等待 3-5 秒后查询：

```bash
ahcli ads project get $PROJECT_ID
```

验证：`status` → `"active"`，`stage` 变化表示 workflow 已启动。

可进一步诊断：

```bash
ahcli ads project diagnose $PROJECT_ID
```

**完成后输出摘要：** Project ID、Campaign 名称、预算配置、投放渠道、广告数量、当前状态。

---

## 错误处理

每步执行后统一检查错误输出。

| 错误 | 原因 | 解决 |
|------|------|------|
| `channel auth not found` | 使用 `creative_id` 但项目未 activate | **必须先执行 Step 8a activate，再执行 Step 7** |
| Budget conflict | 同时设置 Campaign + AdSet budget | 只设 AdSet `daily_budget`（ABO 模式） |
| 落地页显示 example.com | url_analysis 中缺少 `url` 字段 | Step 2 中 `url` 和 `landing_page_url` 都必须设置 |

## 贯穿全流程的关键规则

- `tenant_id` 必须一致贯穿所有请求（登录后自动注入）
- 创建时所有实体状态设为 `PAUSED`
- `access_token`、`ad_account_id`、`page_id` 从 Step 0b 获取
- **推荐使用 `creative_id` 格式**：通过 Step 6 创建素材组合后引用，系统自动处理图片上传和 Meta Creative 创建
- **使用 `creative_id` 时必须先 activate 项目**（Step 8a 在 Step 7 之前），否则报 `channel auth not found`

## 已验证的端到端测试

### 测试 1：纯文本广告（object_story_spec）

| 参数 | 值 |
|------|-----|
| objective | OUTCOME_TRAFFIC |
| optimization_goal | LINK_CLICKS |
| daily_budget (adset) | "50" ($50/天) |
| creative format | object_story_spec → link_data (无图片) |
| billing_event | IMPRESSIONS |
| bid_strategy | LOWEST_COST_WITHOUT_CAP |

结果：Campaign + AdSet + Ad 全部成功创建，项目 activate 后状态变为 `processing`。

### 测试 2：带图片广告（creative_id）

| 参数 | 值 |
|------|-----|
| objective | OUTCOME_TRAFFIC |
| optimization_goal | LINK_CLICKS |
| daily_budget (adset) | "20" ($20/天) |
| creative format | creative_id（引用 lexi creative，单张图片） |

结果：Campaign + AdSet + Ad 全部成功创建，Meta 上落地页 URL 正确。

### 测试 3：skip_initial_epoch + creative_id + AI 生成图片（完整流程）

| 参数 | 值 |
|------|-----|
| objective | OUTCOME_TRAFFIC |
| daily_budget (adset) | "20" ($20/天) |
| creative format | creative_id（AI 生成图片，2 张） |
| skip_initial_epoch | true |
| 执行顺序 | Steps 0→1→2→3→4→6(AI生成)→8a(skip)→7→8b |

结果：AI 生成 2 张图片（耗时 ~1.5min），创建 Creative 组合，activate 跳过 Epoch 0，Campaign + AdSet + Ad 全部成功创建。

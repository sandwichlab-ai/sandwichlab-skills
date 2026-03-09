---
name: hui-ads-creative
description: "为项目批量生成并绑定广告素材（图片+文案+Creative 组合）。使用 ahcli CLI 和 AI 生成工具。使用场景：用户要求生成素材、创建广告图片，如生成素材、创建广告图片、generate creatives。"
---

# 项目素材供给

为已有的 adscore 项目批量生成并绑定广告素材（图片 + 文案 + Creative 组合）。

## 使用方法

- `/hui-ads-creative <project_id>` — 为指定项目生成素材
- `/hui-ads-creative <project_id> --images-only` — 仅生成图片
- `/hui-ads-creative <project_id> --copies-only` — 仅创建文案
- `/hui-ads-creative <project_id> --env preprod` — 指定环境（默认 dev）

## 用户需提供的信息

| 信息 | 必填 | 示例 |
|------|------|------|
| project_id | 是 | `"hui-xxx"` |
| 产品描述/卖点 | 否 | 从项目 url_analysis 自动提取 |
| 图片数量 | 否 | 默认 2 张 |
| 图片风格偏好 | 否 | `"Modern minimalist product photography"` |
| 广告文案 | 否 | Claude 根据产品信息自动生成 |
| 文案组数 | 否 | 默认 2 组 |

---

## 执行流程

### Step 0: 检查登录状态

先确认已登录到目标环境，登录后 `tenant_id` 和 `user_id` 自动注入后续命令。

```bash
ahcli auth status --env $ENV
```

未安装 ahcli 先执行 `/ahcli-install`；未登录时执行登录：

```bash
ahcli auth login --env $ENV
```

---

### Step 1: 获取项目信息

查询项目详情，提取产品信息用于后续素材生成。

```bash
ahcli ads project get $PROJECT_ID
```

从响应的附件 `url_analysis.content` 中提取：
- `product_info` → 用于图片生成的 `selling_point`
- `business_name` → 用于图片生成的 `product_name`
- `landing_page_url` → 验证落地页
- `selling_points` → 文案生成参考
- `call_to_action` → 文案 CTA

如果需要附件详情：

```bash
ahcli ads project attachment list --project-id $PROJECT_ID --type url_analysis --only-latest
```

**如果项目无 url_analysis**，需要用户提供产品描述。

---

### Step 2: AI 生成图片

```bash
ahcli ads media generate \
  --project-id $PROJECT_ID \
  --product-name "产品名称" --selling-point "核心卖点" --count 2
```

响应提取 `execution_arn`，然后自动轮询结果：

```bash
ahcli ads media poll "$EXECUTION_ARN" \
  --project-id $PROJECT_ID --wait
```

`--wait` 模式自动轮询（30s 间隔，最多 10 次），状态变化输出到 stderr。

状态：`SUCCEEDED` → 提取 `asset_output[].asset_id`（`masset-xxx`），`FAILED` → 查看 error。

---

### Step 3: 创建 Copy（广告文案）

每组文案调用一次。**可并行创建多组 Copy。**

文案来源：用户提供 → 直接使用；未提供 → Claude 根据 Step 1 产品信息生成 2-3 组。

**文案规范：** headline ≤ 25 chars, copy_text ≤ 125 chars。

```bash
ahcli ads copy create --project-id $PROJECT_ID \
  --headline "Your Catchy Headline" \
  --copy-text "Compelling ad copy that drives action." \
  --cta SHOP_NOW
```

响应提取 `copy_id`。收集所有 copy_id。

**常用 CTA 选项：** `SHOP_NOW`, `LEARN_MORE`, `SIGN_UP`, `BUY_NOW`, `GET_OFFER`, `SUBSCRIBE`

---

### Step 4: 创建 Creative（素材组合）

将图片和文案组合成 Creative，供广告创建时引用。

```bash
ahcli ads creative create --project-id $PROJECT_ID \
  --asset-ids "masset-xxx,masset-yyy" \
  --copy-ids "copy-xxx,copy-yyy"
```

响应提取 `creative_id`（`crtv-xxx` 格式）。

**注意：** creative 绑定多张图片时，当前只使用第一张（多图轮播待支持 `child_attachments` 格式）。如需多个广告使用不同图片，创建多个 Creative 各绑一张图片。

---

### Step 5: 验证（可选）

确认素材已出现在项目素材库中：

```bash
ahcli ads media search --project-id $PROJECT_ID
```

---

## 从 S3 导入图片（替代 AI 生成）

如果用户提供了 S3 图片 URL，使用导入方式替代 Step 2：

```bash
ahcli ads media import-s3 --project-id $PROJECT_ID \
  --s3-urls "s3://bucket/img1.png,s3://bucket/img2.png"
```

自动完成：下载 → 上传 → 缩略图 → 绑定到项目。响应提取 `media_asset_id`。

---

## 错误处理

| 错误 | 原因 | 解决 |
|------|------|------|
| 图片生成超时（>5min） | Step Functions 执行缓慢 | 继续 `media poll "ARN" --wait` 或重新触发 |
| 素材库看不到图片 | 未执行 generate 或导入失败 | 确认 asset_id 存在，检查 `media search` 输出 |

## 完成后输出摘要

执行完成后输出：
- Project ID
- 生成的图片数量和 asset_id 列表
- 创建的文案数量和 copy_id 列表
- Creative ID（供 `/hui-ads-launch` Step 7 使用）
- 下一步操作建议（如：使用 creative_id 创建广告）

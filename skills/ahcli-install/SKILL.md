---
name: ahcli-install
description: "安装 SandwichLab ahcli CLI 工具。检查是否已安装，未安装则执行安装。使用场景：首次使用 ahcli 相关功能前，或 ahcli 命令不可用时。"
---

# 安装 ahcli CLI 工具

安装 SandwichLab 的命令行工具 `ahcli`，用于广告投放、项目管理等操作。

## 使用方法

- `/ahcli-install` — 检查并安装 ahcli

## 执行流程

### Step 1: 检查是否已安装

```bash
ahcli --help
```

如果命令执行成功并显示帮助信息，说明已安装，无需继续。

如果提示 `command not found` 或 `ahcli: command not found`，继续 Step 2。

---

### Step 2: 执行安装

根据操作系统选择对应的安装命令：

**macOS / Linux:**

```bash
# 先下载脚本到临时文件
curl -fsSL https://raw.githubusercontent.com/sandwichlab-ai/sandwichlab-skills/main/cli/scripts/install-ahcli.sh -o /tmp/install-ahcli.sh

# 审查脚本内容（安全起见）
cat /tmp/install-ahcli.sh

# 确认无问题后执行
bash /tmp/install-ahcli.sh
```

**Windows (PowerShell):**

```powershell
# 先下载脚本到临时文件
Invoke-WebRequest -Uri https://raw.githubusercontent.com/sandwichlab-ai/sandwichlab-skills/main/cli/scripts/install-ahcli.ps1 -OutFile $env:TEMP\install-ahcli.ps1

# 审查脚本内容
Get-Content $env:TEMP\install-ahcli.ps1

# 确认无问题后执行
& $env:TEMP\install-ahcli.ps1
```

> **安全提示：** 请勿使用 `curl ... | bash` 管道方式直接执行远程脚本。始终先下载、审查、再执行，以防供应链攻击或中间人篡改。

安装脚本会自动：
1. 检测操作系统和架构
2. 下载对应的 ahcli 二进制文件
3. 安装到系统路径（`/usr/local/bin` 或 `C:\Program Files\ahcli`）
4. 验证安装

---

### Step 3: 验证安装

安装完成后，验证 ahcli 是否可用：

```bash
ahcli --help
```

应该显示 ahcli 的帮助信息和可用命令列表。

---

## 安装位置

| 操作系统 | 安装路径 |
|---------|---------|
| macOS / Linux | `/usr/local/bin/ahcli` |
| Windows | `C:\Program Files\ahcli\ahcli.exe` |

---

## 常见问题

### 权限不足

**macOS / Linux:**

如果提示权限错误，可能需要 sudo：

```bash
sudo bash /tmp/install-ahcli.sh
```

**Windows:**

以管理员身份运行 PowerShell。

### 网络问题

如果下载失败，检查网络连接或使用代理：

```bash
# 使用代理（macOS/Linux）
export https_proxy=http://proxy.example.com:8080
curl -fsSL https://raw.githubusercontent.com/sandwichlab-ai/sandwichlab-skills/main/cli/scripts/install-ahcli.sh -o /tmp/install-ahcli.sh
bash /tmp/install-ahcli.sh
```

### 手动安装

如果自动安装失败，可以手动下载：

1. 访问 [GitHub Releases](https://github.com/sandwichlab-ai/sandwichlab-skills/releases)
2. 下载对应平台的二进制文件
3. 解压并移动到系统路径

---

## 下一步

安装完成后，使用以下命令登录：

```bash
ahcli auth login
```

然后就可以使用其他 ahcli 相关的 skills 了。

# SandwichLab Skills for Claude

Official Agent Skills and CLI tool for SandwichLab advertising platform.

## 🚀 Quick Start

### Install ahcli

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/sandwichlab-ai/sandwichlab-skills/main/cli/scripts/install-ahcli.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/sandwichlab-ai/sandwichlab-skills/main/cli/scripts/install-ahcli.ps1 | iex
```

### Login

```bash
ahcli auth login
```

### Use Skills in Claude

**Claude Code:**
```
/plugin marketplace add sandwichlab-ai/sandwichlab-skills
/plugin install ads-skills@sandwichlab-skills
```

**Claude.ai:**
1. Click "Skills" button
2. Select "Upload custom skill"
3. Upload any skill folder from `skills/` directory

## 📦 Available Skills

| Skill | Description |
|-------|-------------|
| **ads-adjust-budget** | 调整广告项目预算 |
| **ads-enable** | 启用广告项目 |
| **ads-pause** | 暂停广告项目 |
| **ads-project-status** | 查询广告项目状态 |
| **hui-ads-creative** | 项目素材供给 |
| **hui-ads-launch** | Meta 广告端到端发布 |
| **workflow-status** | 查询工作流状态 |

## 📖 Documentation

- [Installation Guide](docs/INSTALLATION.md)
- [Usage Guide](docs/USAGE.md)
- [Contributing](docs/CONTRIBUTING.md)
- [CLI Development](cli/README.md)

## 🛠️ Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/sandwichlab-ai/sandwichlab-skills.git
cd sandwichlab-skills/cli

# Install dependencies
go mod download

# Build
go build -o ahcli .

# Run
./ahcli --help
```

## 📝 License

Apache 2.0 - see [LICENSE](LICENSE) for details.

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## 🔗 Links

- [SandwichLab Platform](https://sandwichlab.ai)
- [API Documentation](https://api.sandwichlab.ai/docs)
- [Report Issues](https://github.com/sandwichlab-ai/sandwichlab-skills/issues)

---

Built with ❤️ by SandwichLab Team

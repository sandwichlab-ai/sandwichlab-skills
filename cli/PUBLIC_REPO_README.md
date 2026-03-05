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

### Setup Skills

**For Claude Desktop:**

1. Clone this repository:
```bash
git clone https://github.com/sandwichlab-ai/sandwichlab-skills.git
cd sandwichlab-skills
```

2. Copy skills to Claude's directory:
```bash
cp -r skills ~/.claude/skills/
```

3. Restart Claude Desktop to load the skills.

**For development (using symbolic link):**
```bash
ln -s "$(pwd)/skills" ~/.claude/skills
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
| **ads-adjust-budget** | Adjust advertising project budget |
| **ads-enable** | Enable advertising projects |
| **ads-pause** | Pause advertising projects |
| **ads-project-status** | Query advertising project status |
| **hui-ads-creative** | Project creative supply |
| **hui-ads-launch** | Meta ads end-to-end launch |
| **workflow-status** | Query workflow status |

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

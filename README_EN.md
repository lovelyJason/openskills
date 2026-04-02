# 🛠️ openskills

English | [中文](README.md)

Cross-editor extension manager for AI coding tools. Manage **plugins**, **skills**, and **marketplaces** across Codex, Claude, Cursor, and more.

## ✨ Features

- 🎯 **Multi-editor support** — install once, deploy to Codex, Claude, Cursor (or all at once)
- 🏪 **Marketplace system** — add any Git-hosted marketplace, discover and install resources
- 🔗 **Dual install modes** — symlink (live updates) or native copy, locked per marketplace
- 📌 **Version pinning** — pin marketplaces or individual resources to specific tags/commits
- ⚡ **Conflict detection** — disambiguate when multiple marketplaces offer the same resource
- 🛡️ **Atomic operations** — transactional install/uninstall with automatic rollback on failure
- 🐚 **Shell completions** — bash, zsh, fish, powershell

## 📦 Install

### Homebrew (recommended)

```bash
brew tap lovelyJason/openskills https://github.com/lovelyJason/openskills
brew install openskills
```

### Build from source

```bash
git clone https://github.com/lovelyJason/openskills.git
cd openskills
make build && make install
```

## 🚀 Quick Start

```bash
# 1️⃣ Add a marketplace or skill repo
openskills marketplace add https://github.com/example/my-marketplace.git
openskills skill add https://github.com/someone/my-skills.git

# 2️⃣ List available plugins and skills
openskills plugin list
openskills skill list

# 3️⃣ Install (interactive target selection)
openskills plugin install jira-to-code
openskills skill install git-commit

# 4️⃣ Check what's installed
openskills list

# 5️⃣ Check system status
openskills status
```

## 📖 Documentation

- [📘 Usage Guide](docs/USAGE.md) — full command reference and configuration
- [🏗️ Architecture](docs/ARCHITECTURE.md) — internal design, adapter system, directory layout

## 📄 License

[MIT](LICENSE)

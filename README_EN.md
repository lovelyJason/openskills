# 🛠️ openskills

English | [中文](README.md)

![osk-demo](https://github.com/user-attachments/assets/7eecd22e-5dcc-407b-b806-1a18f83d82de)

Cross-editor extension manager for AI coding tools. Manage **plugins**, **skills**, and **marketplaces** across Codex, Claude, Cursor, and more. The motivation behind this project: Codex recently launched its plugin feature with a marketplace design, but its conventions differ from Claude's. Beyond cross-platform skill installation, I urgently needed a tool that lets me manage plugins on both Codex and Claude Code at the same time.

## ✨ Features

- 🎯 **Multi-editor support** — install once, deploy to Codex, Claude, Cursor (or all at once)
- 🏪 **Marketplace repos** — add any Git-hosted marketplace repo, auto-discover its plugins and skills
- 📚 **Skills repos** — add standalone skills repos (e.g. [awesome-copilot](https://github.com/github/awesome-copilot)), install community skills on demand
- 🔒 **Namespace isolation** — all resources install as `osk--<source>--<name>` (e.g. `osk--awesome-copilot--prd`), preventing directory conflicts even when different sources have identically named resources. Note: the slash command recognized by AI editors is determined by the `name` field in SKILL.md (e.g. `name: prd` → `/prd`), not by the directory name
- 🔗 **Dual install modes** — symlink (live updates) or native copy, chosen at install time
- 📌 **Version pinning** — pin marketplaces or individual resources to specific tags/commits
- ⚡ **Conflict detection** — disambiguate when multiple sources offer the same resource name, use `name@source` to specify
- 🛡️ **Atomic operations** — transactional install/uninstall with automatic rollback on failure
- 🐚 **Shell completions** — bash, zsh, fish, powershell

## 📦 Install

### Homebrew (recommended)

```bash
brew install lovelyJason/openskills/openskills
```

### Build from source

```bash
git clone https://github.com/lovelyJason/openskills.git
cd openskills
make build && make install
```

## 🚀 Quick Start

Most commands accept a `--target` or `-t` flag to specify the AI editor (e.g. codex, claude). If omitted, an interactive prompt will ask.

```bash
# 1️⃣ Register a marketplace or skills repo (source only, no install). Supports -t flag
osk marketplace add https://github.com/example/my-marketplace.git
osk skill add https://github.com/github/awesome-copilot

# 2️⃣ Overview of all platform resources and registered sources
osk list

# 3️⃣ List available plugins and skills
osk plugin list
osk skill list

# 4️⃣ Install plugins/skills to specific editors
# When -t or -m is omitted, an interactive prompt will ask
osk plugin install plugin@marketplace-source -t codex,claude
osk skill install prd@awesome-copilot -t claude -m symlink
# Short names work when unambiguous (omit @source)
osk skill install prd

# 5️⃣ Namespace isolation on install:
#   ~/.agents/skills/osk--awesome-copilot--prd        (Codex)
#   ~/.claude/skills/osk--awesome-copilot--prd        (Claude)
#   ~/.cursor/skills-cursor/osk--awesome-copilot--prd (Cursor)

# 6️⃣ Uninstall plugin/skill
osk skill uninstall prd@awesome-copilot
# Remove a marketplace. Supports -t flag to specify editors
osk marketplace remove my-marketplace

# 7️⃣ More management commands
osk marketplace list          # list registered marketplaces
osk marketplace update        # update all marketplaces
osk marketplace pin my-mp v1  # pin version
osk status                    # check system status
osk doctor                    # health check
```

## 🔌 Editor Plugin Systems

Each AI editor has a different plugin architecture. OpenSkills automatically detects repo format and matches compatible editors:

### Codex

Codex's plugin system is relatively new. It requires a specific directory structure:

```
plugins/
  └── <plugin-name>/
        └── .codex-plugin/
              └── plugin.json    ← plugin manifest (required)
```

Refer to `~/.codex/skills/.system/plugin-creator` for how to write plugin.json.

Only repos following this convention can be installed as Codex plugins. When adding a marketplace, if the repo lacks a `plugins/` directory, the Codex option is greyed out and unselectable. (😭 Desperately need a "First Emperor" of the AI world to unify plugin and skill standards)

### Claude

Claude uses its own marketplace mechanism. Plugins are declared at the repo root:

```
.claude-plugin/
  └── plugin.json    ← plugin manifest
```

When adding a marketplace, OpenSkills calls `claude plugin marketplace add` via the Claude CLI, letting Claude manage it internally.

### Cursor

Cursor does not currently support marketplace-level plugin operations. Only skill installation is supported.

> When adding a marketplace, OpenSkills scans the repo format — **incompatible editors are automatically greyed out** (interactive mode) or skipped with a notice (`--target` mode).

## 📖 Documentation

- [📘 Usage Guide](docs/USAGE.md) — full command reference and configuration
- [🏗️ Architecture](docs/ARCHITECTURE.md) — internal design, adapter system, directory layout

## 📄 License

[MIT](LICENSE)

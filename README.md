# 🛠️ openskills

[English](README_EN.md) | 中文

![osk-demo](https://github.com/user-attachments/assets/7eecd22e-5dcc-407b-b806-1a18f83d82de)

跨编辑器的 AI 扩展管理器。统一管理 Codex、Claude、Cursor 等 AI 编辑器的 **插件（plugins）**、**技能（skills）** 和 **市场（marketplaces）**。写这个项目的初衷是近期codex开启了plugin的功能， 也有marketplace的设计，但是同claude的规则不太一样。所以除了在skill的跨平台安装外，我迫切需要一个工具，让我能同时安装在codex和claude code上管理plugin

## ✨ 特性

- 🎯 **多编辑器支持** — 一次安装，同时部署到 Codex、Claude、Cursor（或全选）
- 🏪 **市场仓库（Marketplace Repo）** — 添加任意 Git 托管的 marketplace 仓库，自动发现其中的 plugins 和 skills
- 📚 **技能仓库（Skills Repo）** — 独立添加纯 skills 仓库（如 [awesome-copilot](https://github.com/github/awesome-copilot)），按需安装社区技能
- 🔒 **命名空间隔离** — 所有资源以 `osk--<source>--<name>` 格式安装（如 `osk--awesome-copilot--prd`），即使不同来源存在同名资源也不会产生目录冲突。注意：AI 编辑器识别的 slash 命令名由 SKILL.md 中的 `name` 字段决定（如 `name: prd` → `/prd`），与目录名无关
- 🔗 **双安装模式** — 软链接（实时同步）或物理拷贝，安装时选择
- 📌 **版本锁定** — 可将市场或单个资源锁定到指定 tag/commit
- ⚡ **冲突检测** — 多市场同名资源时自动提示，支持 `name@source` 消歧
- 🛡️ **原子操作** — 事务性安装/卸载，失败自动回滚
- 🐚 **Shell 补全** — 支持 bash、zsh、fish、powershell

## 📦 安装

### Homebrew（推荐）

```bash
brew install lovelyJason/openskills/openskills
```

### 从源码构建

```bash
git clone https://github.com/lovelyJason/openskills.git
cd openskills
make build && make install
```

## 🚀 快速开始

以下命令多数有--target或者-t参数，指定ai编辑器，如codex,claude,如果不指定，会交互式询问

```bash
# 1️⃣ 添加市场仓库或技能仓库（仅注册源，不安装资源）,支持-t参数
osk marketplace add https://github.com/example/my-marketplace.git
osk skill add https://github.com/github/awesome-copilot

# 2️⃣ 查看所有平台资源和已注册的源
osk list

# 3️⃣ 查看可用的插件和技能
osk plugin list
osk skill list

# 4️⃣ 安装插件/技能到指定编辑器
# -t,-m不传的时候就会交互式询问
osk plugin install 插件@marketplace源 -t codex,claude
osk skill install prd@awesome-copilot -t claude -m symlink
# 短名称在无歧义时也可以省略 @source
osk skill install prd

# 5️⃣ 安装后的命名空间隔离：
#   ~/.agents/skills/osk--awesome-copilot--prd     (Codex)
#   ~/.claude/skills/osk--awesome-copilot--prd     (Claude)
#   ~/.cursor/skills-cursor/osk--awesome-copilot--prd (Cursor)

# 6️⃣ 卸载插件
osk skill uninstall prd@awesome-copilot
# 卸载marketplace
osk marketplace remove my-marketplace # 支持-t参数指定ai编辑器

# 7️⃣ 更多管理命令
osk marketplace list          # 查看已注册的市场
osk marketplace update        # 更新所有市场
osk marketplace pin my-mp v1  # 锁定版本
osk status                    # 检查系统状态
osk doctor                    # 健康检查
```

## 🔌 各编辑器插件机制

不同 AI 编辑器的插件体系差异较大，OpenSkills 会自动检测仓库格式并匹配兼容的编辑器：

### Codex

Codex 的插件规范刚推出不久，要求仓库中存在如下目录结构：

```
plugins/
  └── <plugin-name>/
        └── .codex-plugin/
              └── plugin.json    ← 插件清单（必须）
```

plugin.json怎么书写参考~/.codex/skills/.system/plugin-creator中的描述即可

只有符合此规范的仓库才能作为 Codex 插件安装。添加 marketplace 时，如果仓库中没有 `plugins/` 目录，Codex 选项会自动置灰不可选。

### Claude

Claude 使用独立的 marketplace 机制，插件在仓库根目录声明：

```
.claude-plugin/
  └── plugin.json    ← 插件清单
```

添加 marketplace 时，OpenSkills 通过 `claude` CLI 调用 `claude plugin marketplace add` 完成注册，由 Claude 自行管理。

### Cursor

Cursor 目前不支持 marketplace 级别的插件操作，仅支持 skill 安装。

> 添加 marketplace 时，OpenSkills 会扫描仓库格式，**不兼容的编辑器会自动置灰不可选**（交互模式）或输出跳过提示（`--target` 模式）。

## 📖 文档

- [📘 使用指南](docs/USAGE.md) — 完整命令参考和配置说明
- [🏗️ 架构设计](docs/ARCHITECTURE.md) — 内部设计、适配器系统、目录结构

## 📄 License

[MIT](LICENSE)

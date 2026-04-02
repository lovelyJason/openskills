# 🛠️ openskills

[English](README_EN.md) | 中文

跨编辑器的 AI 扩展管理器。统一管理 Codex、Claude、Cursor 等 AI 编辑器的 **插件（plugins）**、**技能（skills）** 和 **市场（marketplaces）**。

## ✨ 特性

- 🎯 **多编辑器支持** — 一次安装，同时部署到 Codex、Claude、Cursor（或全选）
- 🏪 **市场系统** — 添加任意 Git 托管的市场仓库，发现并安装资源
- 🔗 **双安装模式** — 软链接（实时同步）或物理拷贝，安装时选择
- 📌 **版本锁定** — 可将市场或单个资源锁定到指定 tag/commit
- ⚡ **冲突检测** — 多市场同名资源时自动提示并交互式消歧
- 🛡️ **原子操作** — 事务性安装/卸载，失败自动回滚
- 🐚 **Shell 补全** — 支持 bash、zsh、fish、powershell

## 📦 安装

### Homebrew（推荐）

```bash
brew tap lovelyJason/openskills https://github.com/lovelyJason/openskills
brew install openskills
```

### 从源码构建

```bash
git clone https://github.com/lovelyJason/openskills.git
cd openskills
make build && make install
```

## 🚀 快速开始

```bash
# 1️⃣ 添加市场或技能仓库（仅注册，不安装插件）
openskills marketplace add https://github.com/example/my-marketplace.git
openskills marketplace add https://example.com/repo.git --target codex,claude
openskills skill add https://github.com/someone/my-skills.git

# 2️⃣ 查看可用的插件和技能
openskills plugin list
openskills skill list

# 3️⃣ 安装插件/技能（交互式选择安装模式：symlink / native）
openskills plugin install jira-to-code@my-marketplace
openskills plugin install jira-to-code@my-marketplace --mode symlink
openskills skill install git-commit@my-skills

# 短名称在无歧义时也可以省略 @marketplace
openskills plugin install jira-to-code

# 4️⃣ 查看已安装的资源
openskills list

# 5️⃣ 卸载 / 移除
openskills plugin uninstall jira-to-code@my-marketplace
openskills marketplace remove my-marketplace

# 6️⃣ 更多管理命令
openskills marketplace list          # 查看已注册的市场
openskills marketplace update        # 更新所有市场
openskills marketplace pin my-mp v1  # 锁定版本
openskills status                    # 检查系统状态
```

## 📖 文档

- [📘 使用指南](docs/USAGE.md) — 完整命令参考和配置说明
- [🏗️ 架构设计](docs/ARCHITECTURE.md) — 内部设计、适配器系统、目录结构

## 📄 License

[MIT](LICENSE)

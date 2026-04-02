# 📘 OpenSkills 使用指南

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
make build
make install  # 软链接到 /opt/homebrew/bin/，使用 'make uninstall' 可恢复 Homebrew 版本
```

## 🚀 快速开始

```bash
# 1️⃣ 添加一个市场
openskills marketplace add https://github.com/example/my-marketplace.git

# 2️⃣ 查看可用的插件和技能
openskills plugin list
openskills skill list

# 3️⃣ 安装插件（交互式选择目标编辑器）
openskills plugin install jira-to-code

# 4️⃣ 查看已安装的资源
openskills list

# 5️⃣ 检查系统健康状态
openskills status
```

## 🎮 命令参考

### 🏪 市场管理

```bash
# 添加市场（交互式选择安装模式：symlink 或 native）
openskills marketplace add <git-url>
openskills marketplace add <git-url> --name my-mp --mode symlink

# 列出已注册的市场
openskills marketplace list

# 更新市场（git pull）
openskills marketplace update
openskills marketplace update my-mp

# 移除市场
openskills marketplace remove my-mp

# 版本锁定
openskills marketplace pin my-mp v1.2.3
openskills marketplace unpin my-mp
```

### 🧩 插件管理

```bash
# 列出所有市场的可用插件
openskills plugin list
openskills plugin list my-mp  # 指定市场

# 安装插件（未指定 target 时交互式多选）
openskills plugin install jira-to-code
openskills plugin install jira-to-code --target codex,claude
openskills plugin install jira-to-code@my-mp  # 完全限定名

# 安装指定版本
openskills plugin install jira-to-code@1.2.3

# 卸载
openskills plugin uninstall jira-to-code
openskills plugin uninstall jira-to-code --target codex

# 查看状态
openskills plugin status
openskills plugin status jira-to-code
```

### 🎓 技能管理

```bash
# 从任意 Git 仓库添加技能源（不限于 marketplace）
openskills skill add https://github.com/someone/my-skills.git
openskills skill add https://github.com/someone/my-skills.git --name my-skills --mode symlink

# 列出所有来源的可用技能（marketplace + skill repo）
openskills skill list

# 安装技能（symlink 模式自动创建两层链路）
openskills skill install git-commit
openskills skill install git-commit --target codex,claude,cursor

# 卸载
openskills skill uninstall git-commit

# 移除技能源
openskills skill remove-source my-skills
```

### 🔧 通用命令

```bash
# 列出所有已安装的资源（按类型分组）
openskills list

# 系统状态（检测到的编辑器、市场数量、Codex 版本等）
openskills status

# 更新所有市场
openskills update

# 健康检查（检测孤立安装、缺失文件、版本兼容性等）
openskills doctor
```

### 🖥️ Codex 专属命令

```bash
# 查看 Codex CLI 版本及兼容性
openskills codex version

# 强制重新同步 Codex 聚合 marketplace
openskills codex sync

# 列出 Codex 内置插件（来自官方 marketplace）
openskills codex builtin-list

# 列出 Codex 已安装的插件（来自 config.toml）
openskills codex installed-list

# 清理所有 Codex 管理的状态（注册表、缓存、prepared 目录等）
openskills codex cleanup
```

> 💡 Codex 要求最低版本 **0.117.0**。低于此版本时安装命令会跳过 Codex 并继续其他编辑器。

## 🔗 安装模式

### Symlink 模式（软链接）

市场仓库被克隆到 `~/.osk/repos/<name>/`，资源以 **软链接** 方式链接到目标编辑器目录。

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| 仓库更新后资源立即生效 | 依赖克隆仓库保持完整 |
| 节省磁盘空间（无重复） | |
| 方便开发迭代 | |

### Native 模式（物理拷贝）

资源被 **物理拷贝** 到编辑器的原生目录结构中，OpenSkills 仅记录元数据。

| ✅ 优点 | ❌ 缺点 |
|---------|---------|
| 独立于源仓库 | 更新需重新拷贝 |
| 符合编辑器原生期望 | 更多磁盘占用 |

### 🔒 模式锁定

一旦市场以某种模式添加，**该市场下所有资源都遵循相同模式**，避免不一致性。

## ⚙️ 配置

全局配置位于 `~/.osk/config.toml`：

```toml
default_targets = ["codex", "claude"]
default_install_mode = "symlink"
max_backups = 10
```

## 🌍 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `OSK_HOME` | 配置目录 | `~/.osk` |
| `CODEX_HOME` | Codex 主目录 | `~/.codex` |

## ⚡ 冲突解决

当同名插件存在于多个市场时：

```bash
$ openskills plugin install auth-helper
# → 发现冲突！存在于：
#   1. auth-helper@team-marketplace
#   2. auth-helper@community-marketplace
# → 请使用完全限定名：

$ openskills plugin install auth-helper@team-marketplace
```

## 🛡️ 回滚机制

安装/卸载失败时，OpenSkills 会自动回滚所有变更。你也可以手动检查备份：`~/.osk/backups/`

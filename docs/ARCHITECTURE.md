# 🏗️ OpenSkills 架构设计

## 📐 概览

OpenSkills (`openskills`) 是一个跨编辑器的 AI 扩展管理器，统一管理 **插件（plugins）**、**技能（skills）** 和 **市场（marketplaces）**，支持 Codex、Claude、Cursor 及未来更多编辑器。

```
┌─────────────────────────────────────────────────┐
│               🎮 CLI Layer                       │
│           (cobra + huh interactive)              │
├─────────────────────────────────────────────────┤
│                ⚙️ Core Engine                    │
│  ┌───────────┐ ┌──────────┐ ┌──────────────┐   │
│  │🏪 Market  │ │📌Resolver│ │🛡️ Installer  │   │
│  │  Manager  │ │(version, │ │ (transaction  │   │
│  │           │ │ conflict)│ │  + rollback)  │   │
│  └───────────┘ └──────────┘ └──────────────┘   │
├─────────────────────────────────────────────────┤
│            💾 State Manager (~/.osk/)            │
├─────────────────────────────────────────────────┤
│            🔌 Target Adapter Layer               │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐         │
│  │  Codex  │  │ Claude  │  │ Cursor  │  ...     │
│  │ Adapter │  │ Adapter │  │ Adapter │          │
│  └─────────┘  └─────────┘  └─────────┘         │
└─────────────────────────────────────────────────┘
```

## 🎯 关键设计决策

### 1. 二进制名称 vs 配置目录

| 项目 | 值 |
|------|------|
| CLI 二进制 | `openskills` |
| 配置目录 | `~/.osk/` |
| 环境变量覆盖 | `OSK_HOME` |

### 2. 🔗 安装模式架构

市场注册时会指定 **安装模式**，决定该市场下所有资源的安装方式：

- **`symlink`**：OpenSkills 全程管理。市场仓库克隆到 `~/.osk/repos/<name>/`，资源以软链接方式安装到目标编辑器目录。
- **`native`**：OpenSkills 委托给编辑器原生机制。仅在 `~/.osk/state.json` 中记录元数据用于追踪。

**一旦市场以某种模式注册，其所有资源都遵循该模式。** 防止不一致。

### 3. 🎯 多目标安装

省略 `--target` 时，OpenSkills 会弹出交互式多选：

```
$ openskills plugin install jira-to-code
? 选择目标编辑器: (空格选择, 回车确认)
  ◉ codex
  ◉ claude
  ○ cursor
```

### 4. ⚡ 冲突检测

短名称在全局唯一时直接安装。同名冲突时触发交互式消歧：

```
$ openskills plugin install auth-helper
发现 "auth-helper" 存在于多个市场:
  1. auth-helper@tsai-marketplace (v1.2.0)
  2. auth-helper@community-tools (v2.0.1)
请指定: openskills plugin install auth-helper@tsai-marketplace
```

完全限定名 `name@marketplace` 始终无歧义。

### 5. 📦 资源来源（Source Types）

OpenSkills 支持两种 Git 仓库来源：

| 类型 | 命令 | 说明 |
|------|------|------|
| **marketplace** | `openskills marketplace add <url>` | 可包含 `plugins/` 和 `skills/` |
| **skill repo** | `openskills skill add <url>` | 只扫描 `skills/` 目录 |

两者在 `state.json` 中统一存储，通过 `sourceType` 字段区分。`update` 命令同时更新两种来源。

### 6. 🔗 Skill 两层安装（Symlink 模式）

Symlink 模式下，skill 安装采用两层链路：

```
仓库来源                          中间层                         编辑器目录
~/.osk/repos/my-skills/      ~/.osk/skills/               ~/.agents/skills/
  skills/git-commit/  ←────── git-commit (symlink) ←────── git-commit (symlink)
```

- **中间层** `~/.osk/skills/` 提供所有已安装 skill 的统一视图
- **编辑器目录** 只需指向中间层，不关心 skill 来自哪个仓库
- 仓库更新后，整条链路自动生效（无需重新安装）

Native 模式下直接拷贝到编辑器目录，不经过中间层。

### 7. 📌 版本锁定

- **市场级别**: `openskills marketplace pin <name> <tag/commit>`
- **资源级别**: `openskills plugin install foo@1.2.3`
- Lock 文件: `~/.osk/openskills.lock` 记录精确的 git commit SHA

### 8. 🛡️ 回滚机制

每次安装/卸载操作：
1. 在 `~/.osk/backups/<timestamp>/` 创建备份
2. 快照所有将被修改的文件
3. 失败时自动从快照恢复
4. 保留最近 N 份备份（可配置，默认 10）

## 📁 目录结构

### `~/.osk/`（OpenSkills 主目录）

```
~/.osk/
├── config.toml              # 全局配置
├── state.json               # 安装状态
├── openskills.lock          # 版本锁定文件
├── repos/                   # 克隆的仓库（marketplace + skill repo）
│   ├── tsai-marketplace/
│   ├── community-tools/
│   └── my-skills/           # 纯 skill 仓库
├── skills/                  # 已安装 skill 的中间层（symlink 到 repos）
│   ├── git-commit → ../repos/my-skills/skills/git-commit
│   └── code-review → ../repos/tsai-marketplace/skills/code-review
└── backups/                 # 回滚快照
    └── 20260401T100000/
        ├── manifest.json
        └── files/
```

### 💾 State 文件 (`state.json`)

```json
{
  "version": 1,
  "marketplaces": [
    {
      "name": "tsai-marketplace",
      "url": "git@...",
      "localPath": "/Users/jason/.osk/repos/tsai-marketplace",
      "branch": "main",
      "pinnedVersion": "",
      "mode": "symlink",
      "lastUpdated": "2026-04-01T10:00:00Z"
    }
  ],
  "installations": [
    {
      "id": "jira-to-code@tsai-marketplace",
      "resourceType": "plugin",
      "name": "jira-to-code",
      "marketplace": "tsai-marketplace",
      "version": "1.0.0",
      "gitCommitSha": "abc123def456",
      "mode": "symlink",
      "targets": {
        "codex": {
          "installedAt": "2026-04-01T10:00:00Z",
          "paths": ["~/.codex/plugins/cache/opencodex-local/tsai-marketplace--jira-to-code/local/"],
          "configEntries": ["[plugins.\"tsai-marketplace--jira-to-code@opencodex-local\"]"]
        },
        "claude": {
          "installedAt": "2026-04-01T10:05:00Z",
          "nativeRef": "jira-to-code@tsai-marketplace"
        }
      }
    }
  ]
}
```

## 📦 Go 包结构

```
openskills/
├── cmd/openskills/main.go          # 入口
├── internal/
│   ├── cli/                        # Cobra 命令 + 交互式提示
│   │   ├── root.go                 # App 结构体、目标解析
│   │   ├── marketplace.go          # marketplace add/list/update/remove/pin/unpin
│   │   ├── plugin.go               # plugin list/install/uninstall/status
│   │   ├── skill.go                # skill list/install/uninstall
│   │   ├── toplevel.go             # list/status/update/doctor
│   │   ├── codex.go                # openskills codex sync/cleanup/builtin-list/...
│   │   └── completion.go           # shell 补全
│   ├── config/config.go            # TOML 配置 (~/.osk/config.toml)
│   ├── state/state.go              # JSON 状态 (~/.osk/state.json)
│   ├── lockfile/lock.go            # Lock 文件管理
│   ├── marketplace/marketplace.go  # 市场 CRUD + git 操作
│   ├── scanner/scanner.go          # 扫描仓库发现插件/技能
│   ├── resolver/resolver.go        # 名称解析 + 冲突检测
│   ├── installer/installer.go      # 安装引擎（事务支持）
│   ├── target/                     # 目标适配器（上层）
│   │   ├── target.go               # Adapter 接口 + Registry
│   │   ├── codex.go                # Codex 适配器 → 调用 codexmgr
│   │   ├── claude.go               # Claude 适配器
│   │   └── cursor.go               # Cursor 适配器
│   ├── codexmgr/                   # Codex 专属底层驱动
│   │   ├── manager.go              # 对外入口（MarketplaceAdd/PluginInstall 等）
│   │   ├── sync.go                 # 插件扫描、聚合同步、marketplace.json 生成
│   │   ├── registry.go             # Codex 的 marketplaces.json 注册表
│   │   ├── pluginstate.go          # Codex 的 plugins.json 状态
│   │   ├── toml.go                 # Codex config.toml 读写
│   │   ├── rpc.go                  # Codex CLI RPC 调用
│   │   ├── paths.go                # Codex 专属路径常量
│   │   └── version.go              # Codex 版本检测
│   ├── claudecli/                  # Claude 专属底层驱动
│   │   └── cli.go                  # 执行 claude plugin ... CLI 命令
│   ├── backup/backup.go            # 备份/回滚事务系统
│   ├── gitutil/git.go              # Git 操作封装
│   ├── fsutil/fsutil.go            # 文件系统工具
│   ├── paths/paths.go              # 路径常量 + 辅助函数
│   └── ui/prompt.go                # 交互式提示 + 输出格式化
├── docs/
│   ├── USAGE.md                    # 使用指南
│   └── ARCHITECTURE.md             # 本文件
├── Formula/openskills.rb           # Homebrew formula（GoReleaser 自动更新）
├── Makefile
├── go.mod
└── go.sum
```

## 🏛️ 分层调用关系

整个系统分为三层，自上而下调用：

```
用户命令行 (openskills plugin install xxx --target codex,claude)
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  CLI 层 (internal/cli/)                                  │
│  解析命令、交互式提示、调用 installer                        │
└────────────────────────┬────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────┐
│  适配器层 (internal/target/)                              │
│  统一的 Adapter 接口，按目标编辑器分发                       │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐             │
│  │ codex.go │   │claude.go │   │cursor.go │             │
│  └────┬─────┘   └────┬─────┘   └──────────┘             │
└───────┼──────────────┼──────────────────────────────────┘
        │              │
        ▼              ▼
┌──────────────┐ ┌─────────────┐
│  codexmgr/   │ │  claudecli/ │
│  Codex 专属   │ │  Claude 专属 │
│  底层驱动     │ │  底层驱动    │
└──────────────┘ └─────────────┘
```

**为什么需要编辑器专属驱动？**

每个编辑器的插件系统差异很大，适配器层（`target/`）提供统一接口，但底层实现细节必须各自处理：

- **Codex** 最复杂：有自己的 `config.toml` 配置、RPC 接口、插件缓存目录结构（`~/.codex/plugins/cache/`），以及 marketplace.json 聚合机制。`codexmgr/` 包封装了这些 Codex 特有的操作。
- **Claude** 有自己的 `installed_plugins.json` 管理机制，由 `claudecli/` 封装。
- **Cursor** 相对简单，skill 直接软链接/拷贝到 `~/.cursor/skills-cursor/`，无需专属驱动包。

### 历史沿革

项目最初以单文件 shell 脚本 `opencodex.sh` 实现，其中内嵌了大段 Python 代码来处理 JSON 操作和插件管理逻辑（纯 Bash 难以胜任）。后来用 Go 重写为当前架构，`codexmgr/` 包的核心逻辑（`scanRepoPlugins`、`syncAggregate`、`normalizeLocalName` 等）即源自那段 Python 代码的直接移植。`opencodex.sh` 现仅作为历史参考保留在仓库中。

## 🔌 Target Adapter 系统

每个 AI 编辑器都有一个实现 `target.Adapter` 接口的适配器：

```go
type Adapter interface {
    Name() string
    Detect() bool
    SupportedResources() []resource.Type
    Install(ctx, resource, mode, sourcePath) (*InstallResult, error)
    Uninstall(ctx, resource) error
    IsInstalled(resource) (bool, error)
}
```

此外，适配器可以选择性实现扩展接口：

```go
// 市场生命周期钩子（marketplace add/remove/update 时自动触发）
type MarketplaceHook interface {
    OnMarketplaceAdd(ctx, url, name, repoDir) error
    OnMarketplaceRemove(ctx, name) error
    OnMarketplaceUpdate(ctx, name, repoDir) error
}

// 版本检查（安装前自动执行，低于最低版本则跳过该 target）
type VersionChecker interface {
    CheckVersion() error
}
```

### Codex Adapter（完整移植自 opencodex.sh）

Codex 的插件系统最为复杂，底层由 `codexmgr/` 包驱动：

- 🧩 **插件安装**：
  1. 插件源被拷贝到 `~/plugins/<marketplace>--<plugin>/`（prepared copy），manifest 中的 `name` 字段被改写为 `localName`
  2. 聚合所有 marketplace 的插件到 `~/.agents/plugins/marketplace.json`
  3. 尝试通过 JSON-RPC 调用 `codex app-server --listen stdio://` 的 `plugin/install` 方法
  4. RPC 失败时自动 fallback：手动拷贝到 `~/.codex/plugins/cache/opencodex-local/<localName>/local/` + 写入 `config.toml`
- 🎓 **技能**: 软链接/拷贝到 `~/.agents/skills/<name>`
- 📌 **版本要求**: codex-cli >= 0.117.0，低于此版本报错跳过
- 🔗 **MarketplaceHook**: marketplace 操作时同步更新 Codex 注册表 + 重新聚合

### Claude Adapter（原生 CLI 命令）

Claude 有完善的原生 CLI，插件操作直接调用 `claude plugin ...` 命令：

- 🧩 **插件安装**: 执行 `claude plugin install <name>@<marketplace>`
- 🧩 **插件卸载**: 执行 `claude plugin uninstall <name>`
- 🎓 **技能**: 保持文件级操作（Claude 无 skill CLI），拷贝/软链接到 `~/.claude/plugins/cache/osk/<name>/<version>/skills/<name>`
- 🔗 **MarketplaceHook**: 
  - add → `claude plugin marketplace add <url>`
  - remove → `claude plugin marketplace remove <name>`
  - update → `claude plugin marketplace update <name>`

### Cursor Adapter
- 🎓 技能: 拷贝/软链接到 `~/.cursor/skills-cursor/<name>`
- 🧩 插件: 暂不支持（Cursor 的插件系统不同）

### ⚡ 多目标容错

安装命令按 target **逐一执行**，单个 target 失败（如 codex 版本过低）不阻塞其他 target：

```
$ openskills plugin install jira-to-code --target codex,claude
  ✗ [codex] codex-cli 0.115.0 is below minimum required 0.117.0
  ℹ Installing jira-to-code to claude ...
  ✓ Installed jira-to-code (native) to claude
```

## 🎯 Codex 管理机制详解

### 文件路径（Codex 专属）

| 路径 | 用途 |
|------|------|
| `~/.codex/opencodex/marketplaces.json` | Codex 侧的市场注册表 |
| `~/.codex/opencodex/plugins.json` | Codex 侧的插件状态 |
| `~/.agents/plugins/marketplace.json` | 聚合后的本地 marketplace（Codex 读取） |
| `~/plugins/<localName>/` | 已准备的插件副本（manifest 已改写） |
| `~/.codex/plugins/cache/opencodex-local/<localName>/local/` | 插件缓存（Codex 运行时读取） |
| `~/.codex/config.toml` | 插件启用/禁用配置 |

### 命名规则

- `localName` = `<marketplace>--<plugin>`（双横线分隔）
- 本地 marketplace 名称 = `opencodex-local`
- config.toml 段 = `[plugins."<localName>@opencodex-local"]`

### JSON-RPC 协议

与 `codex app-server --listen stdio://` 通信，使用换行分隔的 JSON：

```
→ {"method":"initialize","id":"<uuid>","params":{"clientInfo":{"name":"openskills",...},"capabilities":{"experimentalApi":true}}}
← {"id":"<uuid>","result":{...}}
→ {"method":"initialized"}
→ {"method":"plugin/install","id":"<uuid>","params":{"marketplacePath":"...","pluginName":"..."}}
← {"id":"<uuid>","result":{...}}
```

RPC 失败时自动 fallback 到手动文件操作（拷贝 cache + 写 config.toml），用户可在输出中看到 `(fallback)` 标记。

### 版本要求

codex-cli 最低版本 **0.117.0**。低于此版本时：
- 所有 Codex 操作报错跳过
- 多 target 安装中不阻塞其他编辑器
- `openskills doctor` 和 `openskills codex version` 会检查并报告版本状态

## ➕ 添加新 Target

1. 创建 `internal/target/newtarget.go` 实现 `Adapter` 接口
2. 在 `internal/cli/root.go` 中注册: `reg.Register(target.NewMyEditor())`
3. 如需要，在 `internal/paths/paths.go` 中添加路径辅助函数

## 📦 分发

通过 Homebrew tap 分发（源码仓库同时作为 tap）：

```bash
brew tap lovelyJason/openskills https://github.com/lovelyJason/openskills
brew install openskills
```

二进制文件安装到 `/opt/homebrew/bin/openskills`，已在 PATH 中，无需 sudo。

使用 GoReleaser 构建跨平台二进制：
- 🍎 macOS: arm64, amd64 (`.tar.gz`)
- 🐧 Linux: arm64, amd64 (`.tar.gz`)
- 🪟 Windows: amd64 (`.zip`)

# 📝 OpenSkills 设计改动记录

## 2026-04-01 第二轮迭代

### 🔍 代码审查 & 问题修复

启动 3 个 subagent 对项目进行全面测试：
1. **Go 编译 & 单元测试** — `go vet`、`go test -race`、`go build`
2. **CLI 端到端冒烟测试** — 14 条命令全部验证
3. **项目配置 & CI/CD 审查** — 扫描所有配置文件一致性

#### Critical 修复

| 编号 | 问题 | 修复 |
|------|------|------|
| C1 | 项目声明 MIT 许可但缺少 LICENSE 文件 | 新增 `LICENSE` 文件 |
| C2 | Release Skill 中 `GITHUB_TOKEN` 说明有误 | 明确标注 Actions 自动提供，只需手动配置 `HOMEBREW_TAP_TOKEN` |

#### Important 修复

| 编号 | 问题 | 修复 |
|------|------|------|
| I2 | Makefile shell 变量未加引号，路径含空格时出错 | 所有 `$(BIN_PATH)` 等变量加双引号 |
| I3 | Release Skill 把 `GITHUB_TOKEN` 列为需手动配置 | 合并到 C2 |
| I4 | CI 用 `-race` 但 GoReleaser hooks 和 Makefile 没加 | 统一所有测试入口使用 `-race` |

#### Suggestions 修复

| 编号 | 问题 | 修复 |
|------|------|------|
| S1 | CI 生成 `coverage.out` 但没上传 | CI 移除 `-coverprofile`，本地 `make coverage` 保留 |
| S2 | `README.md` 是空的 | 写了完整中英文 README（`README.md` 中文主体，`README_EN.md` 英文版） |
| S3 | 项目内 `Formula/` 和 GoReleaser 自动生成的 tap 是两份 | Formula 添加头部注释说明是参考模板 |
| S4 | Makefile `git describe` 版本带 `v` 前缀，GoReleaser 不带 | `sed 's/^v//'` 统一格式 |

---

### 🏷️ Homebrew Tap 仓库名变更

**背景**：用户原始仓库名为 `homebrew-tap`，后改为 `openskills`。

**影响**：Homebrew 约定 `brew tap user/xxx` 自动找 `user/homebrew-xxx`。仓库名为 `openskills` 时，用户必须显式指定 URL：

```bash
brew tap lovelyJason/openskills https://github.com/lovelyJason/openskills
brew install openskills
```

**涉及文件**：
- `.goreleaser.yaml` — `brews.repository.name`
- `.claude/skills/release/SKILL.md` — 所有 tap 引用和安装命令
- `Formula/openskills.rb` — 头部注释
- `docs/USAGE.md`、`docs/ARCHITECTURE.md` — brew 安装命令
- `README.md`、`README_EN.md` — 安装说明

---

### 👤 GitHub 用户名修正

**变更**：`jasonhuang` → `lovelyJason`

`go.mod` 已经是 `github.com/lovelyJason/openskills`，但 Go 源码的 import 路径还是 `jasonhuang`。macOS 文件系统大小写不敏感所以本地能编译，但 Linux CI 会失败。

**全量替换**：30 个文件，包括：
- 所有 Go 源码 import 路径（26 个 `.go` 文件）
- `.goreleaser.yaml`、`Makefile` — ldflags 路径
- 所有文档、Formula、Release Skill — GitHub URL

---

### 📖 文档中文化 + Emoji

- `README.md` — 改为中文主文档，顶部 `English | 中文` 语言切换
- `README_EN.md` — 新建英文版
- `docs/USAGE.md` — 全面中文化，章节加 emoji 标注
- `docs/ARCHITECTURE.md` — 全面中文化，架构图加 emoji

---

### 🍺 Homebrew Caveats

在 `.goreleaser.yaml` 和 `Formula/openskills.rb` 中添加 `caveats`，`brew install` 后显示：

```
🛠️  openskills installed successfully!

Get started:
  openskills marketplace add <git-url>    # add a marketplace
  openskills plugin list                   # browse plugins
  openskills skill list                    # browse skills
  openskills status                        # check system

Docs: https://github.com/lovelyJason/openskills
```

---

### 🎓 独立 Skill 仓库支持 + 两层 Symlink 架构

**背景**：并非所有 Git 仓库都是 marketplace（包含 plugins + skills），有些只是纯 skill 仓库。之前只支持"先加 marketplace，再从中安装 skill"的单一路径。

#### 新增功能

**新命令**：
- `openskills skill add <git-url>` — 添加任意 Git 仓库作为技能来源
- `openskills skill remove-source <name>` — 移除技能来源

**两层 Symlink 架构**（symlink 模式）：

```
仓库来源                        中间层                       编辑器目录
~/.osk/repos/my-skills/    ~/.osk/skills/             ~/.agents/skills/
  skills/git-commit/ ←──── git-commit (symlink) ←──── git-commit (symlink)
```

- `~/.osk/skills/` 作为所有已安装 skill 的统一视图
- 编辑器目录只需指向中间层，不关心 skill 来自哪个仓库
- 仓库更新后，整条链路自动生效

**Source Types**：

`state.json` 中 `MarketplaceEntry` 新增 `sourceType` 字段：

| 类型 | 命令 | 扫描内容 |
|------|------|----------|
| `marketplace` | `marketplace add` | `plugins/` + `skills/` |
| `skills` | `skill add` | 仅 `skills/` |

#### 涉及文件

| 文件 | 改动 |
|------|------|
| `internal/paths/paths.go` | 新增 `SkillsDir()` → `~/.osk/skills` |
| `internal/state/state.go` | 新增 `SourceType`、`SourceMarketplace`、`SourceSkillRepo` |
| `internal/marketplace/marketplace.go` | 新增 `AddSkillRepo()`、`AddWithSource()` |
| `internal/cli/skill.go` | 新增 `skill add`、`skill remove-source`；`stageSkillToOsk()` 实现两层链路；`cleanupStagedSkill()` 卸载时清理 |
| `internal/cli/marketplace.go` | `marketplace list` 显示来源类型 |
| `internal/cli/toplevel.go` | `status`、`update`、`doctor` 区分 marketplace 和 skill repo |
| `docs/USAGE.md` | 技能管理章节新增 `skill add`、`skill remove-source` |
| `docs/ARCHITECTURE.md` | 新增 Source Types 和两层安装架构说明 |
| `README.md` / `README_EN.md` | Quick Start 新增 `skill add` |

---

### ✅ 最终验证

3 个 subagent 再次全面测试：
1. `go vet` — 0 警告
2. `go test -race` — 35 个测试全部通过
3. CLI 冒烟测试 — 14 条命令全部通过
4. GitHub 用户名一致性检查 — 0 个 `jasonhuang` 残留，所有路径和 URL 一致

---

## 2026-04-01 第三轮迭代：opencodex.sh 完整移植 + Claude CLI 对接

### 📋 背景

项目最初的 Codex 插件/marketplace 管理依赖 `opencodex.sh`（1477 行 Bash + 894 行内嵌 Python）。此轮迭代将其逻辑完整移植到 Go，同时将 Claude 的插件管理改为调用原生 `claude plugin ...` CLI 命令。

### 🏗️ 新增包：`internal/codexmgr/`（8 个文件）

完整移植 opencodex.sh 中 Python 管理器的全部逻辑：

| 文件 | 行数 | 对应 opencodex.sh 功能 |
|------|------|----------------------|
| `paths.go` | ~65 | 路径常量（`$STATE_DIR`、`$REGISTRY_FILE` 等） |
| `version.go` | ~85 | codex-cli 版本检测，最低要求 0.117.0 |
| `rpc.go` | ~160 | JSON-RPC 客户端（`call_codex_app_server()`） |
| `registry.go` | ~90 | `marketplaces.json` 注册表 CRUD（`register_marketplace()`） |
| `pluginstate.go` | ~70 | `plugins.json` 插件状态管理 |
| `toml.go` | ~110 | `config.toml` 操作（`set_plugin_enabled()`、`clear_plugin_config()`） |
| `sync.go` | ~240 | `sync_aggregate()` + `scan_repo_plugins()` + `prepare_local_plugin_copy()` |
| `manager.go` | ~180 | 高级编排（MarketplaceAdd/Remove/Update、PluginInstall/Uninstall） |

#### 关键移植细节

- **JSON-RPC 协议**：通过 `codex app-server --listen stdio://` 的 stdio 传输，newline-delimited JSON，initialize→initialized→method 三步握手
- **Fallback 机制**：RPC 失败时手动拷贝到 `~/.codex/plugins/cache/opencodex-local/<localName>/local/` + 写 config.toml
- **命名兼容**：保持 opencodex.sh 的 `<marketplace>--<plugin>` localName 格式和 `opencodex-local` marketplace 名称

### 🏗️ 新增包：`internal/claudecli/`（1 个文件）

| 文件 | 行数 | 说明 |
|------|------|------|
| `cli.go` | ~75 | 封装 `claude plugin marketplace add/remove/update` 和 `claude plugin install/uninstall` 命令 |

### 🏗️ 新增 CLI：`openskills codex` 子命令组

| 子命令 | 说明 |
|--------|------|
| `openskills codex version` | 显示 Codex CLI 版本 + 最低版本要求 |
| `openskills codex sync` | 强制重新同步 Codex 聚合 marketplace |
| `openskills codex cleanup` | 清理所有 Codex 管理的状态 |
| `openskills codex builtin-list` | 列出 Codex 内置插件 |
| `openskills codex installed-list` | 列出 config.toml 中已安装的插件 |

### 🔧 新增接口

在 `target/target.go` 中新增两个可选接口（适配器按需实现）：

```go
type MarketplaceHook interface {
    OnMarketplaceAdd(ctx, url, name, repoDir) error
    OnMarketplaceRemove(ctx, name) error
    OnMarketplaceUpdate(ctx, name, repoDir) error
}

type VersionChecker interface {
    CheckVersion() error
}
```

- **Codex** 实现 `MarketplaceHook`（同步注册表 + 聚合）+ `VersionChecker`（>= 0.117.0）
- **Claude** 实现 `MarketplaceHook`（调用 `claude plugin marketplace ...`）

### 🔄 重写文件

| 文件 | 改动 |
|------|------|
| `target/codex.go` | 完全重写：插件操作委托 `codexmgr.Manager`（RPC + fallback），技能保持不变 |
| `target/claude.go` | 完全重写：插件操作改为执行 `claude plugin ...` CLI 命令，技能保持文件级操作 |
| `cli/marketplace.go` | add/remove/update 后触发所有适配器的 `MarketplaceHook` |
| `cli/plugin.go` | 安装改为按 target 逐一执行 + 版本检查，单个 target 失败不阻塞其他 |
| `cli/root.go` | 注册 `openskills codex` 子命令组 |
| `cli/toplevel.go` | `doctor` 加入版本检查，`status` 显示 Codex 版本号 |

### ⚡ 多目标容错机制

安装/卸载命令从"全部一起"改为"按 target 逐一执行"：

```
// 之前：所有 target 一起，一个失败全部回滚
a.inst.Install(ctx, req{TargetNames: ["codex", "claude"]})

// 之后：逐一执行，codex 失败继续 claude
for _, tName := range targetNames {
    if vc, ok := adapter.(VersionChecker); ok {
        if err := vc.CheckVersion(); err != nil {
            ui.Error("[%s] %v", tName, err)
            continue  // 跳过此 target
        }
    }
    a.inst.Install(ctx, req{TargetNames: []string{tName}})
}
```

### ✅ 验证

- `go build ./...` — 编译通过
- `go vet ./...` — 0 警告
- `go test ./...` — 全部通过（resolver 3 + scanner 1 + state 31 = 35 个测试）
- 无 linter 错误

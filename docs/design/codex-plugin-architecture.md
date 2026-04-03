# Codex Plugin 架构

> 本文档描述 Codex 内置的 `plugin-creator` skill 的插件目录规范，以及 OpenSkills 如何与各编辑器的 marketplace 体系对接。

## 各编辑器 Marketplace 目录

不同 AI 编辑器的 marketplace 存放位置和管理方式不同：

| 编辑器 | Marketplace 位置 | 管理方式 |
|--------|------------------|----------|
| Codex | `~/.agents/plugins/marketplace.json` | 单个 JSON 文件，聚合所有 marketplace 的插件列表 |
| Claude | `~/.claude/plugins/marketplaces/<name>/` | 每个 marketplace 一个目录（git clone），由 `claude` CLI 管理 |
| Cursor | 暂不支持 marketplace 级别操作 | 仅支持 skill 安装 |

### OpenSkills 的 `marketplace add` 流程

```bash
# 带 --target 指定目标
openskills marketplace add https://example.com/marketplace.git --target codex,claude

# 不带 --target，交互式多选
openskills marketplace add https://example.com/marketplace.git
# ? Select target editors: (空格选择, 回车确认)
#   ◉ codex
#   ◉ claude
#   ○ cursor
```

选择目标后，OpenSkills 会：
1. Git clone 仓库到 `~/.osk/repos/<name>/`
2. 对每个选中的目标触发对应 hook：
   - **Codex** → `codexmgr` 仅将 marketplace 注册到 `marketplaces.json`（**懒加载**，不会立即拷贝插件到 `~/plugins/`；插件在 `plugin install` 时才按需准备）
   - **Claude** → 调用 `claude plugin marketplace add <url>`，Claude CLI 自行克隆到 `~/.claude/plugins/marketplaces/` 并更新 `known_marketplaces.json`

### 安装模式（plugin/skill install 时选择）

`marketplace add` 和 `skill add` **不再**询问安装模式。安装模式在 `plugin install` / `skill install` 时选择：

```bash
openskills plugin install jira-to-code@my-marketplace
# ? Install mode:
#   > symlink  — link to source repo, auto-updates on git pull
#     native   — copy files to editor directory, manual update

# 或直接指定
openskills plugin install jira-to-code@my-marketplace --mode symlink
```

- **symlink** — 软链接到源仓库，`git pull` 后自动生效
- **native** — 物理拷贝文件到编辑器目录，需要手动更新

---

## Plugin Creator 脚手架

`plugin-creator` 支持两种部署方式，涉及的路径不同：

## 场景 A：仓库插件（repo-local plugin）

```
<repo-root>/
├── .agents/
│   └── plugins/
│       └── marketplace.json          ← 创建或更新（--with-marketplace）
└── plugins/
    └── <plugin-name>/                ← 创建（必须）
        ├── .codex-plugin/
        │   └── plugin.json           ← 创建（必须，插件身份证）
        ├── skills/                   ← 可选（--with-skills）
        ├── hooks/                    ← 可选（--with-hooks）
        ├── scripts/                  ← 可选（--with-scripts）
        ├── assets/                   ← 可选（--with-assets）
        ├── .mcp.json                 ← 可选（--with-mcp）
        └── .app.json                 ← 可选（--with-apps）
```

## 场景 B：本地插件（home-local plugin）

```
~/
├── .agents/
│   └── plugins/
│       └── marketplace.json          ← 创建或更新（--with-marketplace --marketplace-path）
└── plugins/
    └── <plugin-name>/                ← 创建（必须）
        ├── .codex-plugin/
        │   └── plugin.json           ← 创建（必须）
        ├── skills/                   ← 可选
        ├── hooks/                    ← 可选
        ├── scripts/                  ← 可选
        ├── assets/                   ← 可选
        ├── .mcp.json                 ← 可选
        └── .app.json                 ← 可选
```

## 按操作类型分

| 操作 | 文件 | 触发条件 |
|------|------|----------|
| 必定创建 | `<path>/<plugin-name>/` 目录 | 始终 |
| 必定创建 | `<path>/<plugin-name>/.codex-plugin/plugin.json` | 始终 |
| 创建或更新 | `.agents/plugins/marketplace.json` | 加 `--with-marketplace` |
| 可选创建 | `<plugin>/skills/` 目录 | 加 `--with-skills` |
| 可选创建 | `<plugin>/hooks/` 目录 | 加 `--with-hooks` |
| 可选创建 | `<plugin>/scripts/` 目录 | 加 `--with-scripts` |
| 可选创建 | `<plugin>/assets/` 目录 | 加 `--with-assets` |
| 可选创建 | `<plugin>/.mcp.json` 文件 | 加 `--with-mcp` |
| 可选创建 | `<plugin>/.app.json` 文件 | 加 `--with-apps` |

## 插件生命周期全景

`plugin-creator` 只负责第 ① 步（脚手架），插件从定义到运行的完整流程如下：

```
① 插件定义（源码）            ② 插件注册（索引）             ③ 插件运行（加载）

.agents/plugins/              ~/.codex/.tmp/plugins/         ~/.codex/plugins/cache/
  marketplace.json              (Codex git clone 的官方仓库)    <marketplace>/
  + plugins/<name>/                                              <plugin>/
      .codex-plugin/                  │                            <version>/
          plugin.json                 │                            ├── .codex-plugin/
                                      │                            │   └── plugin.json
  "有哪些插件可装"                     │                            └── skills/...
                                      ▼
                                用户在 Codex 中                "Codex 从这里加载插件"
                                点击"安装" ──────────────────▶
```

| 阶段 | 位置 | 谁负责 |
|------|------|--------|
| ① 定义 | `<repo>/.agents/plugins/marketplace.json` + `<repo>/plugins/` | `plugin-creator`（脚手架工具） |
| ② 索引 | `~/.codex/.tmp/plugins/`（git clone） | Codex 自动拉取 |
| ③ 加载 | `~/.codex/plugins/cache/<marketplace>/<plugin>/<version>/` | Codex 安装时拷贝到此处 |
| ④ 启用 | `~/.codex/config.toml` 中 `[plugins."name@marketplace"]` | Codex 写入配置 |

这样做的好处是：

1. 版本隔离 — 官方插件用 git commit SHA 做目录名（如 `f78e3ad49297...`），本地插件用 `local`，互不干扰
2. 离线可用 — 装好之后即使源仓库删了，cache 里的插件照样能用
3. 快速加载 — Codex 启动时直接扫描 `cache/` 目录，不用再去 git 仓库里找

### 版本目录命名规则

- 官方插件：用 git commit SHA 做目录名（如 `f78e3ad49297...`）
- 本地/自定义插件：用 `local` 作为版本目录名

## Codex App-Server JSON-RPC

Codex 提供了 `codex app-server --listen stdio://` 作为 JSON-RPC 服务端，外部工具可通过 stdin/stdout 进行插件安装/卸载操作。

### 协议流程

```
openskills (osk)                                codex app-server
        │                                              │
        │──── initialize (clientInfo, capabilities) ──▶│
        │◀─── initialize response ─────────────────────│
        │──── initialized (notification) ──────────────▶│
        │                                              │
        │──── plugin/install (marketplacePath, name) ──▶│
        │◀─── result ─────────────────────────────────│
        │                                              │
        │──── plugin/uninstall (pluginId) ─────────────▶│
        │◀─── result ─────────────────────────────────│
```

### RPC 方法

| 方法 | 参数 | 作用 |
|------|------|------|
| `initialize` | `clientInfo`, `capabilities` | 握手，建立连接 |
| `initialized` | 无（notification） | 确认初始化完成 |
| `plugin/install` | `marketplacePath`, `pluginName` | 安装插件到 cache 并启用 |
| `plugin/uninstall` | `pluginId`（格式：`name@marketplace`） | 卸载插件并清理 |

### Fallback 机制

当 `codex app-server` 不可用时（Codex 未安装或进程异常），自动降级为手动操作：

1. 将插件文件拷贝到 `~/.codex/plugins/cache/<marketplace>/<plugin>/local/`
2. 在 `~/.codex/config.toml` 中写入启用配置

## 总结

`plugin-creator` 一共就管两个核心文件：

- **`plugin.json`** — 单个插件的清单（总是创建）
- **`marketplace.json`** — 插件目录注册表（按需创建/更新）

安装到 `~/.codex/plugins/cache/` 和写入 `config.toml` 是 Codex 自身（或通过 JSON-RPC）完成的，不属于 `plugin-creator` 的职责。

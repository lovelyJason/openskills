# osk marketplace update 内部流转

## 命令格式

```bash
osk marketplace update [name...] [-t codex,claude,cursor]
```

- `name` 可选，不传则更新所有已注册的 marketplace
- `-t` / `--target` 可选，指定只触发哪些编辑器的 adapter hook；不传则触发所有

## 整体流程

```
osk marketplace update -t codex
        │
        ▼
┌──────────────────────────┐
│  1. 加载 state.json      │  ~/.config/openskills/state.json
│     获取已注册的           │  包含所有 marketplace 的 URL、本地路径、pin 状态等
│     marketplace 列表      │
└──────────┬───────────────┘
           │
           ▼  （遍历每个 marketplace）
┌──────────────────────────┐
│  2. marketplace.Update() │  内部执行 git pull
│     拉取最新代码           │  更新 lastUpdated 时间戳
│                          │  如果 pinned 则跳过
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│  3. 触发 adapter hook    │  由 -t 参数过滤
│     fireMarketplaceHooks │  只调用匹配的 adapter
└──────────┬───────────────┘
           │
     ┌─────┴─────┐
     ▼           ▼
  Codex       Claude
  adapter     adapter
     │           │
     ▼           ▼
  SyncSingle  claude CLI
```

## 各 Adapter 详细流程

### Codex Adapter

`OnMarketplaceUpdate()` → `codexmgr.MarketplaceUpdate()` → `SyncSingle()`

```
SyncSingle(repoDir, name)
    │
    ├─ 1. scanRepoPlugins(repoDir, name)
    │     扫描仓库 plugins/ 目录
    │     查找每个子目录的 .codex-plugin/plugin.json
    │     提取 name、displayName、category
    │     生成 localName = "<marketplace>--<plugin>"
    │
    ├─ 2. prepareLocalPluginCopy()  （对每个扫描到的插件）
    │     拷贝插件到 ~/plugins/<localName>/
    │     改写 plugin.json 的 name 为 localName
    │
    ├─ 3. savePluginState()
    │     写入 ~/.codex/openskills/plugins.json
    │     记录所有插件的 marketplace、name、localName、路径等
    │
    └─ 4. rebuildMarketplaceJSON()
          重新生成 ~/.agents/plugins/marketplace.json
          格式：
          {
            "name": "openskills-local",
            "interface": { "displayName": "OpenSkills Local" },
            "plugins": [
              {
                "name": "<localName>",
                "source": { "source": "local", "path": "./plugins/<localName>" },
                "policy": { "installation": "AVAILABLE", "authentication": "ON_INSTALL" },
                "category": "..."
              }
            ]
          }
```

**涉及的文件路径：**

| 路径 | 用途 |
|------|------|
| `~/.config/openskills/repos/<name>/` | marketplace 本地 git clone |
| `~/plugins/<marketplace>--<plugin>/` | 已准备的插件副本（manifest 已改写） |
| `~/.codex/openskills/plugins.json` | openskills 管理的插件状态 |
| `~/.agents/plugins/marketplace.json` | Codex 读取的本地 marketplace 注册文件 |

### Claude Adapter

`OnMarketplaceUpdate()` → `claudecli.MarketplaceUpdate(name)`

直接调用 Claude CLI：

```bash
claude plugin marketplace update <name>
```

由 Claude 自行管理更新逻辑。如果该 marketplace 不是通过 `claude plugin marketplace add` 注册的（例如纯 skill 仓库），Claude 会报 "Marketplace not found"，此时 openskills 输出 warning 但不影响整体流程。

### Cursor Adapter

Cursor 目前不实现 `MarketplaceHook` 接口，不会被触发。

## 后续：plugin install 如何消费 update 的产物

`osk marketplace update` 完成后，`marketplace.json` 已包含最新的插件列表。执行 `osk plugin install` 时：

1. 优先通过 JSON-RPC 调用 `codex app-server` 的 `plugin/install`，传入 `marketplace.json` 路径
2. Codex 读取 `marketplace.json`，在 `~/.codex/plugins/cache/openskills-local/<localName>/local/` 下创建缓存
3. RPC 失败时 fallback：手动拷贝 + 写入 `~/.codex/config.toml` 的 `[plugins."<localName>@openskills-local"]` 段

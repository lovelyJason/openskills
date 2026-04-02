#!/usr/bin/env bash
# ============================================================================
# 这个是最原始的脚本，现在只作为考古了，不使用
# opencodex — Codex marketplace / plugin / skill manager
# 执行 opencodex marketplace add http://git.inner.truesightai.com/web-platform/tsai-claude-marketplace.git tsai-claude-marketplace 会按顺序触发以下操作：
# 1. 克隆仓库到~/.codex/opencodex/repos/tsai-claude-marketplace/
# 2. 创建或更新	~/.codex/opencodex/marketplaces.json
# 3. 聚合同步
#   3.1 创建目录 ~/plugins
#   3.2 ~/plugins/tsai-claude-marketplace--<plugin-name>/ ← 拷贝自仓库
#   3.3 改写 plugin.json 中的 name	 ~/plugins/tsai-claude-marketplace--<plugin-name>/.codex-plugin/plugin.json
#   3.4 创建或覆盖	 ~/.agents/plugins/marketplace.json

# ============================================================================
set -euo pipefail

REPO_URL="${OPENCODEX_REPO_URL:-http://git.inner.truesightai.com/web-platform/tsai-claude-marketplace.git}"
CODEX_HOME="${CODEX_HOME:-$HOME/.codex}"
CLONE_DIR="${OPENCODEX_CLONE_DIR:-$CODEX_HOME/tsai-claude-marketplace}"
STATE_DIR="${OPENCODEX_STATE_DIR:-$CODEX_HOME/opencodex}"
REGISTRY_FILE="$STATE_DIR/marketplaces.json"
PLUGIN_STATE_FILE="$STATE_DIR/plugins.json"
REPOS_DIR="$STATE_DIR/repos"
SKILLS_DIR="${OPENCODEX_SKILLS_DIR:-$HOME/.agents/skills}"
SKILL_LINK_NAME="${OPENCODEX_SKILL_LINK_NAME:-tsai-fe-toolkit}"
MARKETPLACE_DIR="${OPENCODEX_MARKETPLACE_DIR:-$HOME/.agents/plugins}"
MARKETPLACE_PATH="$MARKETPLACE_DIR/marketplace.json"
LOCAL_PLUGIN_DIR="${OPENCODEX_LOCAL_PLUGIN_DIR:-$HOME/plugins}"
BIN_DIR="${OPENCODEX_BIN_DIR:-/usr/local/bin}"
CMD_NAME="opencodex"
LOCAL_MARKETPLACE_NAME="${OPENCODEX_LOCAL_MARKETPLACE_NAME:-opencodex-local}"
LOCAL_MARKETPLACE_DISPLAY="${OPENCODEX_LOCAL_MARKETPLACE_DISPLAY:-OpenCodex Local}"

log() { printf '\033[1;36m[opencodex]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[opencodex]\033[0m %s\n' "$*" >&2; }
err() { printf '\033[1;31m[opencodex]\033[0m %s\n' "$*" >&2; }
dim() { printf '\033[2m%s\033[0m\n' "$*"; }

resolve_path() {
  local target="$1"
  python3 - "$target" <<'PY'
from pathlib import Path
import sys

print(Path(sys.argv[1]).expanduser().resolve())
PY
}

SCRIPT_SOURCE="${BASH_SOURCE[0]-}"
if [[ -n "$SCRIPT_SOURCE" && "$SCRIPT_SOURCE" != "bash" && "$SCRIPT_SOURCE" != "-bash" ]]; then
  SCRIPT_FILE="$(resolve_path "$SCRIPT_SOURCE")"
  SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_FILE")" && pwd)"
  CURRENT_SCRIPT_REPO="$(cd "$SCRIPT_DIR/.." && pwd)"
else
  SCRIPT_FILE=""
  SCRIPT_DIR=""
  CURRENT_SCRIPT_REPO=""
fi

require_cmd() {
  local cmd="$1"
  if ! command -v "$cmd" >/dev/null 2>&1; then
    err "需要命令: $cmd"
    exit 1
  fi
}

is_repo_root() {
  local candidate="${1:-}"
  [[ -n "$candidate" ]] \
    && [[ -d "$candidate/plugins" ]] \
    && [[ -d "$candidate/skills" ]] \
    && [[ -f "$candidate/VERSION" ]] \
    && [[ -f "$candidate/scripts/opencodex.sh" ]]
}

installed_command_target() {
  if [[ -L "$BIN_DIR/$CMD_NAME" ]] || [[ -f "$BIN_DIR/$CMD_NAME" ]]; then
    resolve_path "$BIN_DIR/$CMD_NAME"
  fi
}

installed_repo_root() {
  local target candidate
  target="$(installed_command_target || true)"
  [[ -n "$target" ]] || return 1
  candidate="$(cd "$(dirname "$target")/.." && pwd)"
  if is_repo_root "$candidate"; then
    printf '%s\n' "$candidate"
    return 0
  fi
  return 1
}

runtime_repo_root() {
  if is_repo_root "$CURRENT_SCRIPT_REPO"; then
    printf '%s\n' "$CURRENT_SCRIPT_REPO"
    return 0
  fi
  if installed_repo_root >/dev/null 2>&1; then
    installed_repo_root
    return 0
  fi
  if is_repo_root "$CLONE_DIR"; then
    printf '%s\n' "$CLONE_DIR"
    return 0
  fi
  return 1
}

repo_version() {
  local repo="${1:-}"
  if [[ -n "$repo" ]] && [[ -f "$repo/VERSION" ]]; then
    cat "$repo/VERSION"
  else
    echo "unknown"
  fi
}

detect_repo_marketplace_name() {
  local repo="$1"
  python3 - "$repo" <<'PY'
from pathlib import Path
import json
import re
import sys

repo = Path(sys.argv[1])
marketplace_json = repo / ".claude-plugin" / "marketplace.json"

def normalize(value: str) -> str:
    value = value.strip().lower()
    value = re.sub(r"[^a-z0-9]+", "-", value)
    value = re.sub(r"-{2,}", "-", value).strip("-")
    return value or "marketplace"

if marketplace_json.exists():
    try:
        payload = json.loads(marketplace_json.read_text())
        name = payload.get("name")
        if isinstance(name, str) and name.strip():
            print(normalize(name))
            raise SystemExit(0)
    except Exception:
        pass

print(normalize(repo.name))
PY
}

show_version() {
  local repo
  repo="$(runtime_repo_root || true)"
  echo "opencodex v$(repo_version "$repo")"
}

show_help() {
  show_version
  cat <<'EOF'

OpenCodex — Codex marketplace / plugin / skill manager

用法:
  opencodex [命令]

快捷命令:
  install                 安装或更新本地 opencodex 命令，并注册当前仓库
  update, upgrade         刷新当前仓库 skills，并更新所有已注册 marketplace
  uninstall, remove       卸载 opencodex 管理的 skills / plugins / 本地命令
  status                  查看命令、skills、marketplaces、plugins 状态
  list, ls                列出当前仓库 skills、已注册 marketplaces、已聚合 plugins

Marketplace:
  marketplace add <git-url> [name]
  marketplace list
  marketplace update [name]
  marketplace remove <name>

Plugin:
  plugin list [marketplace]
  plugin status [plugin@marketplace ...]
  plugin install <plugin@marketplace ...>
  plugin uninstall <plugin@marketplace ...>

兼容别名:
  plugins ...             等同于 plugin ...

Skills:
  skills link [name]      将仓库 skills 软链到 ~/.agents/skills/<name>
                          默认 name = tsai-fe-toolkit
  skills unlink [name]    移除指定软链（默认 tsai-fe-toolkit）
  skills list             列出 ~/.agents/skills/ 下所有由 opencodex 管理的软链

说明:
  1. 当前仓库的 top-level skills 仍通过 ~/.agents/skills/tsai-fe-toolkit 软链接入
  2. marketplace plugins 会被聚合到 ~/.agents/plugins/marketplace.json
  3. 聚合后的本地 marketplace 名称固定为 opencodex-local
  4. skills link 可多次调用以创建多个别名软链
EOF
}

register_global_command() {
  local repo="$1"
  local script_path="$repo/scripts/opencodex.sh"

  if [[ ! -f "$script_path" ]]; then
    err "未找到脚本: $script_path"
    exit 1
  fi

  if [[ -d "$BIN_DIR" ]] && [[ -w "$BIN_DIR" ]]; then
    ln -sf "$script_path" "$BIN_DIR/$CMD_NAME"
  else
    log "注册全局命令需要权限，尝试 sudo..."
    sudo ln -sf "$script_path" "$BIN_DIR/$CMD_NAME"
  fi
  log "已注册全局命令: $CMD_NAME -> $script_path"
}

sync_current_repo_skills() {
  local repo="$1"
  mkdir -p "$SKILLS_DIR"
  if [[ -L "$SKILLS_DIR/$SKILL_LINK_NAME" ]] || [[ -e "$SKILLS_DIR/$SKILL_LINK_NAME" ]]; then
    rm -f "$SKILLS_DIR/$SKILL_LINK_NAME"
  fi
  ln -s "$repo/skills" "$SKILLS_DIR/$SKILL_LINK_NAME"
  log "已创建 skills 符号链接: $SKILLS_DIR/$SKILL_LINK_NAME -> $repo/skills"
}

ensure_repo_for_install() {
  if is_repo_root "$CURRENT_SCRIPT_REPO"; then
    printf '%s\n' "$CURRENT_SCRIPT_REPO"
    return 0
  fi

  require_cmd git
  if [[ -d "$CLONE_DIR/.git" ]]; then
    log "仓库已存在，拉取最新代码..." >&2
    git -C "$CLONE_DIR" pull >&2 || return 1
  else
    log "克隆仓库到 $CLONE_DIR ..." >&2
    git clone "$REPO_URL" "$CLONE_DIR" >&2 || return 1
  fi
  printf '%s\n' "$CLONE_DIR"
}

refresh_current_repo_if_needed() {
  local repo="$1"
  require_cmd git
  if [[ "$repo" == "$CLONE_DIR" ]]; then
    log "更新当前仓库源..."
    git -C "$repo" pull
  else
    log "当前命令源是本地工作区，跳过对当前仓库执行 git pull"
  fi
}

run_manager() {
  local command="$1"
  shift || true

  require_cmd python3
  python3 - \
    "$REGISTRY_FILE" \
    "$REPOS_DIR" \
    "$MARKETPLACE_PATH" \
    "$LOCAL_PLUGIN_DIR" \
    "$CODEX_HOME/config.toml" \
    "$PLUGIN_STATE_FILE" \
    "$LOCAL_MARKETPLACE_NAME" \
    "$LOCAL_MARKETPLACE_DISPLAY" \
    "$command" \
    "$@" <<'PY'
from __future__ import annotations

import json
import os
import re
import shutil
import subprocess
import sys
import uuid
from pathlib import Path
from typing import Any


REGISTRY_FILE = Path(sys.argv[1]).expanduser()
REPOS_DIR = Path(sys.argv[2]).expanduser()
MARKETPLACE_PATH = Path(sys.argv[3]).expanduser()
LOCAL_PLUGIN_DIR = Path(sys.argv[4]).expanduser()
CONFIG_PATH = Path(sys.argv[5]).expanduser()
PLUGIN_STATE_FILE = Path(sys.argv[6]).expanduser()
LOCAL_MARKETPLACE_NAME = sys.argv[7]
LOCAL_MARKETPLACE_DISPLAY = sys.argv[8]
COMMAND = sys.argv[9]
ARGS = sys.argv[10:]
OFFICIAL_MARKETPLACE_PATH = CONFIG_PATH.parent / ".tmp" / "plugins" / ".agents" / "plugins" / "marketplace.json"


def normalize_name(value: str) -> str:
    value = value.strip().lower()
    value = re.sub(r"[^a-z0-9]+", "-", value)
    value = re.sub(r"-{2,}", "-", value).strip("-")
    return value or "marketplace"


def read_json(path: Path, default: dict[str, Any] | None = None) -> dict[str, Any]:
    if not path.exists():
        if default is None:
            raise FileNotFoundError(path)
        return default

    with path.open() as handle:
        payload = json.load(handle)
    if not isinstance(payload, dict):
        raise ValueError(f"{path} must contain a JSON object.")
    return payload


def write_json(path: Path, payload: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w") as handle:
        json.dump(payload, handle, indent=2, ensure_ascii=False)
        handle.write("\n")


def load_registry() -> dict[str, Any]:
    return read_json(REGISTRY_FILE, default={"version": 1, "marketplaces": []})


def save_registry(registry: dict[str, Any]) -> None:
    write_json(REGISTRY_FILE, registry)


def load_plugin_state() -> dict[str, Any]:
    return read_json(PLUGIN_STATE_FILE, default={"version": 1, "plugins": []})


def save_plugin_state(payload: dict[str, Any]) -> None:
    write_json(PLUGIN_STATE_FILE, payload)


def load_official_marketplace() -> dict[str, Any]:
    return read_json(
        OFFICIAL_MARKETPLACE_PATH,
        default={"name": "openai-curated", "interface": {"displayName": "Codex official"}, "plugins": []},
    )


def load_local_marketplace() -> dict[str, Any]:
    return read_json(MARKETPLACE_PATH, default=build_local_marketplace())


def resolve_symlink_target(path: Path) -> Path:
    target = path.readlink()
    if not target.is_absolute():
        target = (path.parent / target).resolve()
    return target.resolve()


def remove_path(path: Path) -> None:
    if path.is_symlink() or path.is_file():
        path.unlink()
    elif path.exists():
        shutil.rmtree(path)


def scan_repo_plugins(repo_root: Path, marketplace_name: str) -> list[dict[str, Any]]:
    plugins_root = repo_root / "plugins"
    if not plugins_root.exists():
        return []

    plugins: list[dict[str, Any]] = []
    for plugin_dir in sorted(plugins_root.iterdir()):
        if not plugin_dir.is_dir() or plugin_dir.name.startswith("."):
            continue

        manifest_path = plugin_dir / ".codex-plugin" / "plugin.json"
        if not manifest_path.exists():
            continue

        manifest = read_json(manifest_path)
        interface = manifest.get("interface") if isinstance(manifest.get("interface"), dict) else {}
        plugin_name = str(manifest.get("name") or plugin_dir.name).strip()
        local_name = f"{marketplace_name}--{plugin_name}"
        plugins.append(
            {
                "marketplace": marketplace_name,
                "name": plugin_name,
                "localName": local_name,
                "displayName": interface.get("displayName") or plugin_name,
                "category": interface.get("category") or "Developer Tools",
                "pluginDir": str(plugin_dir.resolve()),
            }
        )
    return plugins


def build_local_marketplace() -> dict[str, Any]:
    return {
        "name": LOCAL_MARKETPLACE_NAME,
        "interface": {
            "displayName": LOCAL_MARKETPLACE_DISPLAY,
        },
        "plugins": [],
    }


def prepare_local_plugin_copy(
    source: Path, dest: Path, local_name: str, previously_managed: set[str]
) -> tuple[bool, str]:
    if not source.is_dir():
        return False, f"{source} does not exist or is not a directory"

    if dest.exists() and dest.name not in previously_managed:
        return False, f"{dest} already exists and is not managed by opencodex"

    if dest.exists():
        remove_path(dest)

    shutil.copytree(source, dest, symlinks=False)
    manifest_path = dest / ".codex-plugin" / "plugin.json"
    if not manifest_path.exists():
        remove_path(dest)
        return False, f"missing plugin manifest in {source}"

    manifest = read_json(manifest_path)
    manifest["name"] = local_name
    write_json(manifest_path, manifest)
    return True, "prepared"


def ensure_prepared_plugin_source(plugin: dict[str, Any]) -> tuple[bool, str]:
    local_name = plugin["localName"]
    dest = LOCAL_PLUGIN_DIR / local_name
    source = Path(plugin["pluginDir"])

    manifest_path = dest / ".codex-plugin" / "plugin.json"
    if dest.is_dir() and not dest.is_symlink() and manifest_path.is_file():
        try:
            manifest = read_json(manifest_path)
            if manifest.get("name") == local_name:
                return True, "ready"
        except Exception:
            pass

    return prepare_local_plugin_copy(
        source=source,
        dest=dest,
        local_name=local_name,
        previously_managed={local_name},
    )


def sync_aggregate() -> tuple[list[dict[str, Any]], list[str]]:
    registry = load_registry()
    state = load_plugin_state()
    previous = {item["localName"] for item in state.get("plugins", []) if isinstance(item, dict) and item.get("localName")}
    warnings: list[str] = []

    all_plugins: list[dict[str, Any]] = []
    for marketplace in registry.get("marketplaces", []):
        if not isinstance(marketplace, dict):
            continue
        name = marketplace.get("name")
        repo_dir = marketplace.get("repoDir")
        if not isinstance(name, str) or not isinstance(repo_dir, str):
            continue
        repo_path = Path(repo_dir).expanduser()
        if not repo_path.exists():
            warnings.append(f"Marketplace {name} repo missing: {repo_path}")
            continue
        all_plugins.extend(scan_repo_plugins(repo_path, name))

    LOCAL_PLUGIN_DIR.mkdir(parents=True, exist_ok=True)
    managed_now: set[str] = set()
    written_plugins: list[dict[str, Any]] = []
    for plugin in all_plugins:
        local_name = plugin["localName"]
        dest = LOCAL_PLUGIN_DIR / local_name
        target = Path(plugin["pluginDir"])
        ok, status = prepare_local_plugin_copy(dest=dest, source=target, local_name=local_name, previously_managed=previous)
        if not ok:
            warnings.append(f"Skip plugin {plugin['name']}@{plugin['marketplace']}: {status}")
            continue
        managed_now.add(local_name)
        plugin["preparedDir"] = str(dest)
        written_plugins.append(plugin)

    for stale in sorted(previous - managed_now):
        dest = LOCAL_PLUGIN_DIR / stale
        if dest.exists():
            remove_path(dest)

    payload = build_local_marketplace()
    for plugin in written_plugins:
        payload["plugins"].append(
            {
                "name": plugin["localName"],
                "source": {
                    "source": "local",
                    "path": f"./plugins/{plugin['localName']}",
                },
                "policy": {
                    "installation": "AVAILABLE",
                    "authentication": "ON_INSTALL",
                },
                "category": plugin["category"],
            }
        )
    write_json(MARKETPLACE_PATH, payload)
    save_plugin_state({"version": 1, "plugins": written_plugins})
    return written_plugins, warnings


def registry_entry(registry: dict[str, Any], name: str) -> dict[str, Any] | None:
    for item in registry.get("marketplaces", []):
        if isinstance(item, dict) and item.get("name") == name:
            return item
    return None


def run_git(args: list[str], action: str) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(args, capture_output=True, text=True, check=False)
    if result.returncode == 0:
        return result

    detail = (result.stderr or result.stdout).strip() or f"exit code {result.returncode}"
    hint = ""
    lower = detail.lower()
    if "could not resolve host" in lower:
        hint = " Hint: check network / VPN / DNS access to the Git host."
    elif "could not read username" in lower or "authentication failed" in lower or "access denied" in lower:
        hint = " Hint: this Git URL likely needs credentials or a configured git credential helper."
    raise SystemExit(f"{action} failed: {detail}{hint}")


def ensure_git_repo(url: str, alias: str) -> Path:
    REPOS_DIR.mkdir(parents=True, exist_ok=True)
    repo_dir = REPOS_DIR / alias
    if (repo_dir / ".git").exists():
        existing_url = subprocess.run(
            ["git", "-C", str(repo_dir), "config", "--get", "remote.origin.url"],
            capture_output=True,
            text=True,
            check=False,
        ).stdout.strip()
        if existing_url and existing_url != url:
            raise SystemExit(f"Marketplace {alias} already exists with another URL: {existing_url}")
        run_git(["git", "-C", str(repo_dir), "pull"], f"git pull for marketplace {alias}")
    else:
        run_git(["git", "clone", url, str(repo_dir)], f"git clone for marketplace {alias}")
    return repo_dir.resolve()


def register_marketplace(name: str, source: str, repo_dir: str, kind: str) -> None:
    registry = load_registry()
    marketplaces = [item for item in registry.get("marketplaces", []) if isinstance(item, dict) and item.get("name") != name]
    marketplaces.append(
        {
            "name": name,
            "kind": kind,
            "source": source,
            "repoDir": repo_dir,
        }
    )
    marketplaces.sort(key=lambda item: item["name"])
    registry["marketplaces"] = marketplaces
    save_registry(registry)


def plugin_ref_map() -> dict[str, dict[str, Any]]:
    state = load_plugin_state()
    mapping: dict[str, dict[str, Any]] = {}
    for plugin in state.get("plugins", []):
        if not isinstance(plugin, dict):
            continue
        ref = f"{plugin['name']}@{plugin['marketplace']}"
        mapping[ref] = plugin
    return mapping


def local_name_map() -> dict[str, dict[str, Any]]:
    state = load_plugin_state()
    mapping: dict[str, dict[str, Any]] = {}
    for plugin in state.get("plugins", []):
        if isinstance(plugin, dict) and isinstance(plugin.get("localName"), str):
            mapping[plugin["localName"]] = plugin
    return mapping


def parse_installed_plugins() -> list[dict[str, Any]]:
    if not CONFIG_PATH.exists():
        return []

    text = CONFIG_PATH.read_text()
    section_pattern = re.compile(r'(?ms)^\[plugins\."([^"]+)"\]\n(.*?)(?=^\[|\Z)')
    local_plugins = local_name_map()
    official_marketplace = load_official_marketplace()
    official_name = str(official_marketplace.get("name") or "openai-curated")
    local_marketplace = load_local_marketplace()
    local_name = str(local_marketplace.get("name") or LOCAL_MARKETPLACE_NAME)
    local_entries = local_marketplace.get("plugins", [])
    local_plugin_names = {
        entry.get("name")
        for entry in local_entries
        if isinstance(entry, dict) and isinstance(entry.get("name"), str)
    }
    installed: list[dict[str, Any]] = []

    for match in section_pattern.finditer(text):
        ref = match.group(1)
        body = match.group(2)
        enabled_match = re.search(r"^\s*enabled\s*=\s*(true|false)\s*$", body, flags=re.MULTILINE)
        enabled = None if enabled_match is None else enabled_match.group(1) == "true"

        if "@" in ref:
            plugin_name, marketplace_name = ref.split("@", 1)
        else:
            plugin_name, marketplace_name = ref, ""

        source = "external"
        display_ref = ref
        orphaned = False
        mapped = local_plugins.get(plugin_name)
        if marketplace_name == LOCAL_MARKETPLACE_NAME and mapped is not None:
            source = "opencodex"
            display_ref = f"{mapped['name']}@{mapped['marketplace']}"
        elif marketplace_name == official_name:
            source = "builtin"
        else:
            orphaned = not (marketplace_name == local_name and plugin_name in local_plugin_names)

        installed.append(
            {
                "rawRef": ref,
                "displayRef": display_ref,
                "marketplace": marketplace_name,
                "enabled": enabled,
                "source": source,
                "orphaned": orphaned,
                "cachePresent": cache_path_for_plugin_ref(plugin_name, marketplace_name).is_dir(),
                "configSection": f'[plugins."{ref}"]',
                "cachePath": str(
                    cache_path_for_plugin_ref(plugin_name, marketplace_name).expanduser()
                ),
            }
        )

    installed.sort(key=lambda item: item["displayRef"])
    return installed


def split_plugin_refs(refs: list[str]) -> tuple[list[dict[str, Any]], list[str]]:
    mapping = plugin_ref_map()
    if not refs:
        return list(mapping.values()), []

    selected: list[dict[str, Any]] = []
    unknown: list[str] = []
    seen: set[str] = set()
    for ref in refs:
        key = ref.strip()
        if key in mapping:
            local_name = mapping[key]["localName"]
            if local_name not in seen:
                selected.append(mapping[key])
                seen.add(local_name)
        else:
            unknown.append(key)
    return selected, unknown


def plugin_cache_root(local_name: str) -> Path:
    return CONFIG_PATH.parent / "plugins" / "cache" / LOCAL_MARKETPLACE_NAME / local_name


def plugin_cache_version_root(local_name: str) -> Path:
    return plugin_cache_root(local_name) / "local"


def cache_path_for_plugin_ref(plugin_name: str, marketplace_name: str) -> Path:
    return CONFIG_PATH.parent / "plugins" / "cache" / marketplace_name / plugin_name


def plugin_enabled(local_name: str) -> str:
    section = f'[plugins."{local_name}@{LOCAL_MARKETPLACE_NAME}"]'
    if not CONFIG_PATH.exists():
        return "not-installed"
    text = CONFIG_PATH.read_text()
    start = text.find(section)
    if start == -1:
        return "not-installed"
    body_start = start + len(section)
    next_section = text.find("\n[", body_start)
    body = text[body_start:] if next_section == -1 else text[body_start:next_section]
    match = re.search(r"^\s*enabled\s*=\s*(true|false)\s*$", body, flags=re.MULTILINE)
    if not match:
        return "installed"
    return "enabled" if match.group(1) == "true" else "disabled"


def set_plugin_enabled(local_name: str, enabled: bool) -> None:
    section = f'[plugins."{local_name}@{LOCAL_MARKETPLACE_NAME}"]'
    value = "true" if enabled else "false"
    text = CONFIG_PATH.read_text() if CONFIG_PATH.exists() else ""
    pattern = re.compile(rf'(?ms)^\[plugins\."{re.escape(local_name)}@{re.escape(LOCAL_MARKETPLACE_NAME)}"\]\n.*?(?=^\[|\Z)')
    match = pattern.search(text)
    if match:
        block = match.group(0)
        if re.search(r"^\s*enabled\s*=", block, flags=re.MULTILINE):
            block = re.sub(r"^\s*enabled\s*=.*$", f"enabled = {value}", block, flags=re.MULTILINE)
        else:
            block = block.rstrip() + f"\nenabled = {value}\n"
        text = text[:match.start()] + block + text[match.end():]
    else:
        if text and not text.endswith("\n"):
            text += "\n"
        text += f'\n{section}\nenabled = {value}\n'
    CONFIG_PATH.parent.mkdir(parents=True, exist_ok=True)
    CONFIG_PATH.write_text(text)


def clear_plugin_config(local_name: str) -> None:
    if not CONFIG_PATH.exists():
        return
    text = CONFIG_PATH.read_text()
    pattern = re.compile(
        rf'(?ms)^\[plugins\."{re.escape(local_name)}@{re.escape(LOCAL_MARKETPLACE_NAME)}"\]\n.*?(?=^\[|\Z)'
    )
    text = pattern.sub("", text)
    text = re.sub(r"\n{3,}", "\n\n", text).lstrip("\n")
    CONFIG_PATH.write_text(text)


def call_codex_app_server(method: str, params: dict[str, Any]) -> tuple[bool, str]:
    codex_bin = shutil.which("codex")
    if codex_bin is None:
        return False, "codex binary not found in PATH"

    proc = subprocess.Popen(
        [codex_bin, "app-server", "--listen", "stdio://"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        env=os.environ.copy(),
    )

    def send(message: dict[str, Any]) -> None:
        assert proc.stdin is not None
        proc.stdin.write(json.dumps(message, ensure_ascii=False) + "\n")
        proc.stdin.flush()

    try:
        init_id = str(uuid.uuid4())
        send(
            {
                "method": "initialize",
                "id": init_id,
                "params": {
                    "clientInfo": {
                        "name": "opencodex",
                        "title": "OpenCodex",
                        "version": "1.0.0",
                    },
                    "capabilities": {
                        "experimentalApi": True,
                    },
                },
            }
        )

        assert proc.stdout is not None
        while True:
            line = proc.stdout.readline()
            if not line:
                stderr = proc.stderr.read() if proc.stderr else ""
                return False, f"initialize failed: {stderr.strip() or 'unexpected EOF'}"
            message = json.loads(line)
            if message.get("id") == init_id and "result" in message:
                break
            if message.get("id") == init_id and "error" in message:
                return False, f"initialize error: {message['error']}"

        send({"method": "initialized"})

        request_id = str(uuid.uuid4())
        send({"method": method, "id": request_id, "params": params})

        while True:
            line = proc.stdout.readline()
            if not line:
                stderr = proc.stderr.read() if proc.stderr else ""
                return False, f"{method} failed: {stderr.strip() or 'unexpected EOF'}"
            message = json.loads(line)
            if message.get("id") == request_id and "result" in message:
                return True, json.dumps(message["result"], ensure_ascii=False)
            if message.get("id") == request_id and "error" in message:
                return False, f"{method} error: {message['error']}"
    finally:
        if proc.stdin:
            proc.stdin.close()
        try:
            proc.terminate()
            proc.wait(timeout=2)
        except Exception:
            proc.kill()


def cmd_register_local() -> int:
    if len(ARGS) < 2:
        raise SystemExit("register-local requires <repo-path> <name>")
    repo = Path(ARGS[0]).expanduser().resolve()
    alias = normalize_name(ARGS[1])
    register_marketplace(alias, str(repo), str(repo), "local")
    plugins, warnings = sync_aggregate()
    print(f"Registered local marketplace: {alias}")
    print(f"Aggregated plugins: {len(plugins)}")
    if warnings:
      print("Warnings:")
      for warning in warnings:
        print(f"  - {warning}")
    return 0


def cmd_marketplace_add() -> int:
    if not ARGS:
        raise SystemExit("marketplace add requires <git-url> [name]")
    url = ARGS[0]
    alias = normalize_name(ARGS[1] if len(ARGS) > 1 else Path(url).name.removesuffix(".git"))
    repo_dir = ensure_git_repo(url, alias)
    contributed_plugins = scan_repo_plugins(repo_dir, alias)
    register_marketplace(alias, url, str(repo_dir), "git")
    plugins, warnings = sync_aggregate()
    if not contributed_plugins:
        warnings = [
            f"Marketplace {alias} contributed 0 Codex plugins. The cloned repo may not contain committed .codex-plugin/plugin.json files yet."
        ] + warnings
    print(f"Added marketplace: {alias}")
    print(f"Source: {url}")
    print(f"Repo: {repo_dir}")
    print(f"Aggregated plugins: {len(plugins)}")
    if warnings:
        print("Warnings:")
        for warning in warnings:
            print(f"  - {warning}")
    return 0


def cmd_marketplace_list() -> int:
    registry = load_registry()
    marketplaces = [item for item in registry.get("marketplaces", []) if isinstance(item, dict)]
    if not marketplaces:
        print("No marketplaces registered.")
        return 0
    print("Registered marketplaces:")
    for item in marketplaces:
        print(f"  - {item['name']}: {item.get('kind', 'unknown')} | {item.get('source', '')}")
    return 0


def cmd_marketplace_update() -> int:
    registry = load_registry()
    marketplaces = [item for item in registry.get("marketplaces", []) if isinstance(item, dict)]
    if ARGS:
        wanted = {normalize_name(name) for name in ARGS}
        marketplaces = [item for item in marketplaces if item.get("name") in wanted]
        missing = sorted(wanted - {item.get("name") for item in marketplaces})
    else:
        missing = []

    for item in marketplaces:
        if item.get("kind") == "git":
            repo_dir = Path(item["repoDir"]).expanduser()
            run_git(["git", "-C", str(repo_dir), "pull"], f"git pull for marketplace {item.get('name')}")

    plugins, warnings = sync_aggregate()
    print(f"Updated marketplaces: {len(marketplaces)}")
    print(f"Aggregated plugins: {len(plugins)}")
    for name in missing:
        print(f"  - Unknown marketplace: {name}")
    if warnings:
        print("Warnings:")
        for warning in warnings:
            print(f"  - {warning}")
    return 0 if not missing else 1


def cmd_marketplace_remove() -> int:
    if not ARGS:
        raise SystemExit("marketplace remove requires <name>")
    target = normalize_name(ARGS[0])
    registry = load_registry()
    entry = registry_entry(registry, target)
    if not entry:
        print(f"Marketplace not found: {target}")
        return 1

    state = load_plugin_state()
    removed_plugins = [
        item
        for item in state.get("plugins", [])
        if isinstance(item, dict) and item.get("marketplace") == target
    ]

    for plugin in removed_plugins:
        local_name = plugin.get("localName")
        if not isinstance(local_name, str):
            continue
        remove_path(plugin_cache_root(local_name))
        clear_plugin_config(local_name)
        prepared_dir = LOCAL_PLUGIN_DIR / local_name
        if prepared_dir.exists():
            remove_path(prepared_dir)

    registry["marketplaces"] = [item for item in registry.get("marketplaces", []) if isinstance(item, dict) and item.get("name") != target]
    save_registry(registry)

    if entry.get("kind") == "git":
        repo_dir = Path(entry.get("repoDir", "")).expanduser()
        if repo_dir.exists():
            shutil.rmtree(repo_dir)

    plugins, warnings = sync_aggregate()
    print(f"Removed marketplace: {target}")
    print(f"Uninstalled plugins from marketplace: {len(removed_plugins)}")
    print(f"Aggregated plugins: {len(plugins)}")
    if warnings:
        print("Warnings:")
        for warning in warnings:
            print(f"  - {warning}")
    return 0


def cmd_plugin_list() -> int:
    state = load_plugin_state()
    plugins = [item for item in state.get("plugins", []) if isinstance(item, dict)]
    if ARGS:
        wanted_marketplace = normalize_name(ARGS[0])
        plugins = [item for item in plugins if item.get("marketplace") == wanted_marketplace]
    if not plugins:
        print("No plugins available.")
        return 0
    print("Available plugins:")
    for plugin in plugins:
        print(f"  - {plugin['name']}@{plugin['marketplace']} -> {plugin['displayName']} [{plugin['category']}]")
    return 0


def cmd_builtin_plugin_list() -> int:
    payload = load_official_marketplace()
    plugins = payload.get("plugins", [])
    marketplace_name = str(payload.get("name") or "openai-curated")
    if not isinstance(plugins, list) or not plugins:
        print("No built-in plugins found.")
        return 0

    print("Built-in plugins:")
    for plugin in plugins:
        if not isinstance(plugin, dict):
            continue
        name = plugin.get("name")
        category = plugin.get("category", "Unknown")
        if isinstance(name, str):
            print(f"  - {name}@{marketplace_name} [{category}]")
    return 0


def cmd_installed_plugin_list() -> int:
    installed = parse_installed_plugins()
    if not installed:
        print("No plugins installed.")
        return 0

    print("Installed plugins:")
    orphaned_items: list[dict[str, Any]] = []
    cache_missing_items: list[dict[str, Any]] = []
    for item in installed:
        state = "enabled" if item["enabled"] is True else ("disabled" if item["enabled"] is False else "installed")
        tags = [item["source"], state]
        if item.get("orphaned"):
            tags.append("orphaned")
        if not item.get("cachePresent"):
            tags.append("not-cached")
        print(f"  - {item['displayRef']} [{', '.join(tags)}]")
        if item.get("orphaned"):
            orphaned_items.append(item)
        elif not item.get("cachePresent"):
            cache_missing_items.append(item)
    if orphaned_items:
        print("")
        print("Orphan plugin cleanup:")
        for item in orphaned_items:
            print(
                f"  - {item['displayRef']}: remove {item['configSection']} from {CONFIG_PATH}, "
                f"then optionally delete {item['cachePath']}"
            )
    if cache_missing_items:
        print("")
        print("Cache missing:")
        for item in cache_missing_items:
            if item["source"] == "opencodex":
                print(
                    f"  - {item['displayRef']}: config exists but cache is missing; rerun "
                    f"`opencodex plugin install {item['displayRef']}` or restart from /plugins"
                )
            else:
                print(
                    f"  - {item['displayRef']}: config exists but cache is missing at {item['cachePath']}"
                )
    return 0


def cmd_plugin_status() -> int:
    plugins, unknown = split_plugin_refs(ARGS)
    if not plugins:
        print("No plugins matched.")
        for item in unknown:
            print(f"Unknown plugin: {item}")
        return 1 if unknown else 0

    print(f"Local marketplace: {LOCAL_MARKETPLACE_NAME}")
    for plugin in plugins:
        local_path = LOCAL_PLUGIN_DIR / plugin["localName"]
        local_status = "prepared" if local_path.is_dir() else ("conflict" if local_path.exists() else "missing")
        installed = plugin_enabled(plugin["localName"])
        cache_status = "cached" if plugin_cache_version_root(plugin["localName"]).is_dir() else "not-cached"
        print(
            f"  - {plugin['name']}@{plugin['marketplace']}: "
            f"local={plugin['localName']}, source={local_status}, cache={cache_status}, installed={installed}"
        )
    for item in unknown:
        print(f"Unknown plugin: {item}")
    return 0 if not unknown else 1


def cmd_plugin_install() -> int:
    plugins, unknown = split_plugin_refs(ARGS)
    if not plugins:
        print("No plugins matched.")
        for item in unknown:
            print(f"Unknown plugin: {item}")
        return 1 if unknown else 0

    for plugin in plugins:
        prepared_ok, prepared_status = ensure_prepared_plugin_source(plugin)
        prepared_dir = LOCAL_PLUGIN_DIR / plugin["localName"]
        if not prepared_ok or not prepared_dir.is_dir():
            print(
                f"Plugin source missing for {plugin['name']}@{plugin['marketplace']}: "
                f"{prepared_dir} ({prepared_status})"
            )
            continue
        ok, detail = call_codex_app_server(
            "plugin/install",
            {
                "marketplacePath": str(MARKETPLACE_PATH),
                "pluginName": plugin["localName"],
            },
        )
        if ok:
            print(f"Installed plugin: {plugin['name']}@{plugin['marketplace']}")
            continue
        cache_root = plugin_cache_root(plugin["localName"])
        version_root = plugin_cache_version_root(plugin["localName"])
        if cache_root.exists():
            remove_path(cache_root)
        version_root.parent.mkdir(parents=True, exist_ok=True)
        shutil.copytree(prepared_dir, version_root, symlinks=False)
        set_plugin_enabled(plugin["localName"], True)
        print(
            f"Installed plugin: {plugin['name']}@{plugin['marketplace']} "
            f"(fallback, app-server unavailable: {detail})"
        )
    if unknown:
        for item in unknown:
            print(f"Unknown plugin: {item}")
        return 1
    return 0


def cmd_plugin_uninstall() -> int:
    plugins, unknown = split_plugin_refs(ARGS)
    if not plugins:
        print("No plugins matched.")
        for item in unknown:
            print(f"Unknown plugin: {item}")
        return 1 if unknown else 0

    for plugin in plugins:
        ensure_prepared_plugin_source(plugin)
        ok, detail = call_codex_app_server(
            "plugin/uninstall",
            {
                "pluginId": f"{plugin['localName']}@{LOCAL_MARKETPLACE_NAME}",
            },
        )
        if ok:
            print(f"Uninstalled plugin: {plugin['name']}@{plugin['marketplace']}")
            continue
        remove_path(plugin_cache_root(plugin["localName"]))
        clear_plugin_config(plugin["localName"])
        print(
            f"Uninstalled plugin: {plugin['name']}@{plugin['marketplace']} "
            f"(fallback, app-server unavailable: {detail})"
        )
    if unknown:
        for item in unknown:
            print(f"Unknown plugin: {item}")
        return 1
    return 0


def cmd_sync() -> int:
    plugins, warnings = sync_aggregate()
    print(f"Aggregated plugins: {len(plugins)}")
    if warnings:
        print("Warnings:")
        for warning in warnings:
            print(f"  - {warning}")
    return 0


def cmd_cleanup_all() -> int:
    state = load_plugin_state()
    for item in state.get("plugins", []):
        if not isinstance(item, dict):
            continue
        dest = LOCAL_PLUGIN_DIR / item.get("localName", "")
        if dest.is_symlink():
            dest.unlink()
    if MARKETPLACE_PATH.exists():
        payload = build_local_marketplace()
        write_json(MARKETPLACE_PATH, payload)
    if PLUGIN_STATE_FILE.exists():
        PLUGIN_STATE_FILE.unlink()
    return 0


COMMANDS = {
    "register-local": cmd_register_local,
    "marketplace-add": cmd_marketplace_add,
    "marketplace-list": cmd_marketplace_list,
    "marketplace-update": cmd_marketplace_update,
    "marketplace-remove": cmd_marketplace_remove,
    "plugin-list": cmd_plugin_list,
    "plugin-status": cmd_plugin_status,
    "plugin-install": cmd_plugin_install,
    "plugin-uninstall": cmd_plugin_uninstall,
    "builtin-plugin-list": cmd_builtin_plugin_list,
    "installed-plugin-list": cmd_installed_plugin_list,
    "sync": cmd_sync,
    "cleanup-all": cmd_cleanup_all,
}


def main() -> int:
    handler = COMMANDS.get(COMMAND)
    if handler is None:
        raise SystemExit(f"Unsupported manager command: {COMMAND}")
    return handler()


if __name__ == "__main__":
    sys.exit(main())
PY
}

sync_current_repo_marketplace() {
  local repo="$1"
  local alias
  alias="$(detect_repo_marketplace_name "$repo")"
  run_manager register-local "$repo" "$alias" >/dev/null
  printf '%s\n' "$alias"
}

do_install() {
  local repo alias
  if ! repo="$(ensure_repo_for_install)"; then
    err "安装失败：无法准备仓库源"
    exit 1
  fi
  alias="$(sync_current_repo_marketplace "$repo")"
  sync_current_repo_skills "$repo"
  register_global_command "$repo"
  log "已注册当前仓库 marketplace: $alias"
  log "已聚合 Codex plugin 本地源到 $MARKETPLACE_PATH"
  echo ""
  log "安装完成！v$(repo_version "$repo")"
  log "当前仓库 marketplace: $alias"
  log "重启 Codex 后生效。若要直接启用插件，可用:"
  log "  opencodex plugin install jira-to-code@$alias"
}

do_update() {
  local repo alias
  repo="$(runtime_repo_root || true)"
  if [[ -z "$repo" ]]; then
    err "未找到当前仓库源，请先执行 opencodex install"
    exit 1
  fi

  refresh_current_repo_if_needed "$repo"
  alias="$(sync_current_repo_marketplace "$repo")"
  sync_current_repo_skills "$repo"
  log "更新所有已注册 marketplaces..."
  run_manager marketplace-update || true
  register_global_command "$repo"
  log "更新完成。当前仓库 marketplace: $alias"
}

do_uninstall() {
  local installed_repo
  installed_repo="$(installed_repo_root || true)"

  log "清理 opencodex 管理的 plugins..."
  run_manager cleanup-all || true

  rm -f "$SKILLS_DIR/$SKILL_LINK_NAME" && log "已移除 skills 符号链接"
  rm -rf "$STATE_DIR" && log "已移除 opencodex 状态目录"

  if [[ -L "$BIN_DIR/$CMD_NAME" ]] || [[ -f "$BIN_DIR/$CMD_NAME" ]]; then
    rm -f "$BIN_DIR/$CMD_NAME" 2>/dev/null || sudo rm -f "$BIN_DIR/$CMD_NAME"
    log "已移除全局命令 $CMD_NAME"
  fi

  if [[ -n "$installed_repo" ]] && [[ "$installed_repo" == "$CLONE_DIR" ]] && [[ -d "$CLONE_DIR" ]]; then
    rm -rf "$CLONE_DIR"
    log "已删除由脚本克隆的仓库: $CLONE_DIR"
  else
    log "未删除本地源码仓库"
  fi
  log "卸载完成"
}

do_status() {
  local repo installed_target
  repo="$(runtime_repo_root || true)"
  installed_target="$(installed_command_target || true)"

  show_version
  echo ""
  if [[ -n "$repo" ]]; then
    log "当前仓库源: $repo"
    if [[ -d "$repo/.git" ]]; then
      local branch commit
      branch="$(git -C "$repo" branch --show-current 2>/dev/null || echo 'unknown')"
      commit="$(git -C "$repo" log -1 --format='%h %s' 2>/dev/null || echo 'unknown')"
      dim "  分支: $branch"
      dim "  最新提交: $commit"
    fi
  else
    err "当前仓库源未找到"
  fi

  echo ""
  if [[ -L "$SKILLS_DIR/$SKILL_LINK_NAME" ]]; then
    log "Skills 符号链接: $SKILLS_DIR/$SKILL_LINK_NAME -> $(readlink "$SKILLS_DIR/$SKILL_LINK_NAME")"
  else
    err "Skills 符号链接不存在"
  fi

  echo ""
  log "Registered marketplaces:"
  run_manager marketplace-list || true

  echo ""
  log "Aggregated plugins:"
  run_manager plugin-status || true

  echo ""
  log "Codex installed plugins:"
  run_manager installed-plugin-list || true

  echo ""
  if [[ -n "$installed_target" ]]; then
    log "全局命令: $BIN_DIR/$CMD_NAME -> $installed_target"
  else
    err "全局命令未注册"
  fi
}

do_list() {
  local repo
  repo="$(runtime_repo_root || true)"
  if [[ -z "$repo" ]] || [[ ! -d "$repo/skills" ]]; then
    err "未找到当前仓库源，请先执行 opencodex install"
    exit 1
  fi

  show_version
  echo ""
  log "当前仓库 Skills："
  echo ""
  for skill_dir in "$repo/skills"/*/; do
    [[ -d "$skill_dir" ]] || continue
    local name desc
    name="$(basename "$skill_dir")"
    desc="$(grep -m1 '^description:' "$skill_dir/SKILL.md" 2>/dev/null | sed 's/^description:[[:space:]]*//' | head -c 60 || echo '')"
    if [[ -z "$desc" ]] || [[ "$desc" == "|" ]]; then
      desc="$(sed -n 's/^name:[[:space:]]*//p' "$skill_dir/SKILL.md" 2>/dev/null || echo "$name")"
    fi
    printf "  \033[1;33m%-20s\033[0m %s\n" "$name" "$desc"
  done

  echo ""
  log "Registered marketplaces:"
  run_manager marketplace-list || true

  echo ""
  log "Available plugins:"
  run_manager plugin-list || true

  echo ""
  log "Codex built-in plugins:"
  run_manager builtin-plugin-list || true

  echo ""
  log "Codex installed plugins:"
  run_manager installed-plugin-list || true
}

do_marketplace() {
  local subcommand="${1:-list}"
  shift || true
  case "$subcommand" in
    add)
      if [[ $# -lt 1 ]]; then
        err "用法: opencodex marketplace add <git-url> [name]"
        exit 1
      fi
      run_manager marketplace-add "$@"
      ;;
    list)
      run_manager marketplace-list
      ;;
    update)
      run_manager marketplace-update "$@"
      ;;
    remove)
      if [[ $# -lt 1 ]]; then
        err "用法: opencodex marketplace remove <name>"
        exit 1
      fi
      run_manager marketplace-remove "$@"
      ;;
    *)
      err "未知 marketplace 子命令: $subcommand"
      exit 1
      ;;
  esac
}

do_plugin() {
  local subcommand="${1:-list}"
  shift || true
  case "$subcommand" in
    list)
      run_manager plugin-list "$@"
      ;;
    status)
      run_manager plugin-status "$@"
      ;;
    install)
      if [[ $# -lt 1 ]]; then
        err "用法: opencodex plugin install <plugin@marketplace ...>"
        exit 1
      fi
      run_manager plugin-install "$@"
      ;;
    uninstall|remove)
      if [[ $# -lt 1 ]]; then
        err "用法: opencodex plugin uninstall <plugin@marketplace ...>"
        exit 1
      fi
      run_manager plugin-uninstall "$@"
      ;;
    *)
      err "未知 plugin 子命令: $subcommand"
      exit 1
      ;;
  esac
}

do_skills() {
  local subcommand="${1:-list}"
  shift || true

  local repo
  repo="$(runtime_repo_root || true)"
  if [[ -z "$repo" ]]; then
    err "未找到仓库源，请先执行 opencodex install"
    exit 1
  fi

  local skills_source="$repo/skills"
  if [[ ! -d "$skills_source" ]]; then
    err "仓库中不存在 skills 目录: $skills_source"
    exit 1
  fi

  case "$subcommand" in
    link)
      local link_name="${1:-$SKILL_LINK_NAME}"
      mkdir -p "$SKILLS_DIR"
      if [[ -L "$SKILLS_DIR/$link_name" ]]; then
        local existing_target
        existing_target="$(readlink "$SKILLS_DIR/$link_name")"
        if [[ "$existing_target" == "$skills_source" ]]; then
          log "软链已存在且指向正确: $SKILLS_DIR/$link_name -> $skills_source"
          return 0
        fi
        warn "软链已存在但指向不同: $existing_target"
        warn "将更新为: $skills_source"
        rm -f "$SKILLS_DIR/$link_name"
      elif [[ -e "$SKILLS_DIR/$link_name" ]]; then
        err "$SKILLS_DIR/$link_name 已存在且不是软链，请手动处理"
        exit 1
      fi
      ln -s "$skills_source" "$SKILLS_DIR/$link_name"
      log "已创建软链: $SKILLS_DIR/$link_name -> $skills_source"
      ;;
    unlink)
      local link_name="${1:-$SKILL_LINK_NAME}"
      if [[ -L "$SKILLS_DIR/$link_name" ]]; then
        rm -f "$SKILLS_DIR/$link_name"
        log "已移除软链: $SKILLS_DIR/$link_name"
      elif [[ -e "$SKILLS_DIR/$link_name" ]]; then
        err "$SKILLS_DIR/$link_name 不是软链，拒绝删除"
        exit 1
      else
        warn "软链不存在: $SKILLS_DIR/$link_name"
      fi
      ;;
    list)
      log "Skills 目录: $SKILLS_DIR"
      echo ""
      local found=0
      if [[ -d "$SKILLS_DIR" ]]; then
        for entry in "$SKILLS_DIR"/*/; do
          [[ -L "${entry%/}" ]] || continue
          local name target
          name="$(basename "${entry%/}")"
          target="$(readlink "${entry%/}")"
          if [[ "$target" == "$skills_source" ]] || [[ "$target" == "$repo/skills" ]]; then
            printf "  \033[1;32m%-25s\033[0m -> %s\n" "$name" "$target"
            found=$((found + 1))
          fi
        done
      fi
      if [[ $found -eq 0 ]]; then
        dim "  (无 opencodex 管理的 skills 软链)"
      fi
      ;;
    *)
      err "未知 skills 子命令: $subcommand"
      echo ""
      echo "用法:"
      echo "  opencodex skills link [name]    创建软链（默认 $SKILL_LINK_NAME）"
      echo "  opencodex skills unlink [name]  移除软链（默认 $SKILL_LINK_NAME）"
      echo "  opencodex skills list           列出所有 opencodex 管理的软链"
      exit 1
      ;;
  esac
}

case "${1:-install}" in
  install)                 do_install ;;
  update|upgrade)          do_update ;;
  uninstall|remove)        do_uninstall ;;
  status)                  do_status ;;
  list|ls)                 do_list ;;
  marketplace)             shift; do_marketplace "$@" ;;
  plugin|plugins)          shift; do_plugin "$@" ;;
  skills|skill)            shift; do_skills "$@" ;;
  version|-v|-V|--version) show_version ;;
  help|-h|--help)          show_help ;;
  *)
    err "未知命令: $1"
    echo ""
    show_help
    exit 1
    ;;
esac

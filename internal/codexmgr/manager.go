package codexmgr

import (
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) MarketplaceAdd(repoDir, url, name string) error {
	reg, err := loadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	reg.register(name, url, repoDir, "git")
	return saveRegistry(reg)
}

func (m *Manager) MarketplaceRemove(name string) error {
	reg, err := loadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	entry := reg.find(name)
	if entry == nil {
		return fmt.Errorf("marketplace %q not found in codex registry", name)
	}

	ps, _ := loadPluginState()
	for _, plugin := range ps.Plugins {
		if plugin.Marketplace != name {
			continue
		}
		os.RemoveAll(pluginCacheRoot(plugin.LocalName))
		clearPluginConfig(plugin.LocalName)
		prepared := filepath.Join(localPluginDir(), plugin.LocalName)
		os.RemoveAll(prepared)
	}

	reg.remove(name)
	return saveRegistry(reg)
}

func (m *Manager) MarketplaceUpdate(name, repoDir string) error {
	reg, err := loadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}

	entry := reg.find(name)
	if entry != nil {
		entry.RepoDir = repoDir
		return saveRegistry(reg)
	}
	return nil
}

func (m *Manager) PluginInstall(desc PluginDescriptor) (fallback bool, err error) {
	ok, status := ensurePreparedPluginSource(&desc)
	if !ok {
		return false, fmt.Errorf("failed to prepare plugin source: %s", status)
	}

	rpcOK, detail := rpcPluginInstall(desc.LocalName)
	if rpcOK {
		return false, nil
	}

	cacheRoot := pluginCacheRoot(desc.LocalName)
	versionRoot := pluginCacheVersionRoot(desc.LocalName)

	os.RemoveAll(cacheRoot)
	if err := os.MkdirAll(filepath.Dir(versionRoot), 0755); err != nil {
		return true, fmt.Errorf("mkdir cache: %w", err)
	}

	preparedDir := filepath.Join(localPluginDir(), desc.LocalName)
	if err := copyDirRecursive(preparedDir, versionRoot); err != nil {
		return true, fmt.Errorf("copy to cache: %w", err)
	}

	if err := setPluginEnabled(desc.LocalName, true); err != nil {
		return true, fmt.Errorf("enable in config.toml: %w", err)
	}

	_ = detail
	return true, nil
}

func (m *Manager) PluginUninstall(desc PluginDescriptor) (fallback bool, err error) {
	rpcOK, detail := rpcPluginUninstall(desc.LocalName)
	if rpcOK {
		return false, nil
	}

	os.RemoveAll(pluginCacheRoot(desc.LocalName))
	if err := clearPluginConfig(desc.LocalName); err != nil {
		return true, fmt.Errorf("clear config: %w", err)
	}

	_ = detail
	return true, nil
}

func (m *Manager) Sync() error {
	reg, err := loadRegistry()
	if err != nil {
		return fmt.Errorf("load registry: %w", err)
	}
	_, warnings := syncAggregate(reg)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "  [codex] warning: %s\n", w)
	}
	return nil
}

func (m *Manager) CleanupAll() error {
	ps, _ := loadPluginState()
	for _, plugin := range ps.Plugins {
		os.RemoveAll(pluginCacheRoot(plugin.LocalName))
		clearPluginConfig(plugin.LocalName)
		prepared := filepath.Join(localPluginDir(), plugin.LocalName)
		os.RemoveAll(prepared)
	}

	os.Remove(registryFile())
	os.Remove(pluginStateFile())
	os.Remove(marketplacePath())

	return nil
}

func (m *Manager) BuiltinPluginList() ([]map[string]interface{}, error) {
	return readOfficialMarketplace()
}

func (m *Manager) InstalledPluginList() (map[string]bool, error) {
	return installedPluginsFromConfig()
}

func (m *Manager) FindPluginDescriptor(marketplace, pluginName string) (*PluginDescriptor, error) {
	ps, err := loadPluginState()
	if err != nil {
		return nil, err
	}
	localName := normalizeLocalName(marketplace, pluginName)
	desc := ps.findByLocalName(localName)
	if desc != nil {
		return desc, nil
	}

	for i := range ps.Plugins {
		if ps.Plugins[i].Name == pluginName && ps.Plugins[i].Marketplace == marketplace {
			return &ps.Plugins[i], nil
		}
	}
	return nil, fmt.Errorf("plugin %s@%s not found in codex plugin state", pluginName, marketplace)
}

func copyDirRecursive(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode())
	})
}

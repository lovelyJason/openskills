package codexmgr

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/lovelyJason/openskills/internal/fsutil"
)

func normalizeLocalName(marketplace, plugin string) string {
	return marketplace + "--" + plugin
}

func scanRepoPlugins(repoDir, marketplace string) []PluginDescriptor {
	pluginsDir := filepath.Join(repoDir, "plugins")
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil
	}

	var result []PluginDescriptor
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}

		manifestPath := filepath.Join(pluginsDir, e.Name(), ".codex-plugin", "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifest struct {
			Name      string                 `json:"name"`
			Interface map[string]interface{} `json:"interface,omitempty"`
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		name := manifest.Name
		if name == "" {
			name = e.Name()
		}

		displayName := name
		category := "Developer Tools"
		if iface := manifest.Interface; iface != nil {
			if dn, ok := iface["displayName"].(string); ok && dn != "" {
				displayName = dn
			}
			if cat, ok := iface["category"].(string); ok && cat != "" {
				category = cat
			}
		}

		result = append(result, PluginDescriptor{
			Marketplace: marketplace,
			Name:        name,
			LocalName:   normalizeLocalName(marketplace, name),
			DisplayName: displayName,
			Category:    category,
			PluginDir:   filepath.Join(pluginsDir, e.Name()),
		})
	}
	return result
}

func prepareLocalPluginCopy(dest, source, localName string) (bool, string) {
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		os.RemoveAll(dest)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return false, fmt.Sprintf("mkdir: %v", err)
	}

	if err := fsutil.CopyDir(source, dest); err != nil {
		return false, fmt.Sprintf("copy: %v", err)
	}

	manifestPath := filepath.Join(dest, ".codex-plugin", "plugin.json")
	if err := rewriteManifestName(manifestPath, localName); err != nil {
		return false, fmt.Sprintf("rewrite manifest: %v", err)
	}

	return true, ""
}

func rewriteManifestName(manifestPath, newName string) error {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}

	manifest["name"] = newName
	out, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, append(out, '\n'), 0644)
}

func buildLocalMarketplace() map[string]interface{} {
	return map[string]interface{}{
		"name": LocalMarketplaceName,
		"interface": map[string]interface{}{
			"displayName": LocalMarketplaceDisplay,
		},
		"plugins": []interface{}{},
	}
}

func syncAggregate(reg *Registry) ([]PluginDescriptor, []string) {
	ps, _ := loadPluginState()
	previous := ps.localNames()
	var warnings []string

	var allPlugins []PluginDescriptor
	for _, mp := range reg.Marketplaces {
		repoPath := mp.RepoDir
		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			warnings = append(warnings, fmt.Sprintf("Marketplace %s repo missing: %s", mp.Name, repoPath))
			continue
		}
		allPlugins = append(allPlugins, scanRepoPlugins(repoPath, mp.Name)...)
	}

	lpDir := localPluginDir()
	os.MkdirAll(lpDir, 0755)

	managedNow := make(map[string]struct{})
	var writtenPlugins []PluginDescriptor

	for _, plugin := range allPlugins {
		dest := filepath.Join(lpDir, plugin.LocalName)
		ok, status := prepareLocalPluginCopy(dest, plugin.PluginDir, plugin.LocalName)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("Skip plugin %s@%s: %s", plugin.Name, plugin.Marketplace, status))
			continue
		}
		managedNow[plugin.LocalName] = struct{}{}
		plugin.PreparedDir = dest
		writtenPlugins = append(writtenPlugins, plugin)
	}

	for stale := range previous {
		if _, ok := managedNow[stale]; !ok {
			stalePath := filepath.Join(lpDir, stale)
			os.RemoveAll(stalePath)
		}
	}

	payload := buildLocalMarketplace()
	pluginEntries := make([]interface{}, 0, len(writtenPlugins))
	for _, plugin := range writtenPlugins {
		pluginEntries = append(pluginEntries, map[string]interface{}{
			"name": plugin.LocalName,
			"source": map[string]interface{}{
				"source": "local",
				"path":   fmt.Sprintf("./plugins/%s", plugin.LocalName),
			},
			"policy": map[string]interface{}{
				"installation":   "AVAILABLE",
				"authentication": "ON_INSTALL",
			},
			"category": plugin.Category,
		})
	}
	payload["plugins"] = pluginEntries

	os.MkdirAll(marketplaceDir(), 0755)
	writeJSON(marketplacePath(), payload)

	savePluginState(&PluginState{
		Version: 1,
		Plugins: writtenPlugins,
	})

	return writtenPlugins, warnings
}

func writeJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func ensurePreparedPluginSource(desc *PluginDescriptor) (bool, string) {
	dest := filepath.Join(localPluginDir(), desc.LocalName)
	if info, err := os.Stat(dest); err == nil && info.IsDir() {
		desc.PreparedDir = dest
		return true, ""
	}

	if desc.PluginDir == "" {
		return false, "no plugin source directory"
	}

	ok, status := prepareLocalPluginCopy(dest, desc.PluginDir, desc.LocalName)
	if ok {
		desc.PreparedDir = dest
	}
	return ok, status
}

func readOfficialMarketplace() ([]map[string]interface{}, error) {
	data, err := os.ReadFile(officialMarketplacePath())
	if err != nil {
		return nil, err
	}
	var mp map[string]interface{}
	if err := json.Unmarshal(data, &mp); err != nil {
		return nil, err
	}
	plugins, ok := mp["plugins"].([]interface{})
	if !ok {
		return nil, nil
	}
	var result []map[string]interface{}
	for _, p := range plugins {
		if pm, ok := p.(map[string]interface{}); ok {
			result = append(result, pm)
		}
	}
	return result, nil
}

func listInstalledPluginDirs() ([]string, error) {
	cacheBase := filepath.Join(pluginCacheRoot(""), "..")
	cacheBase = filepath.Clean(cacheBase)
	if _, err := os.Stat(cacheBase); os.IsNotExist(err) {
		return nil, nil
	}

	var installed []string
	_ = filepath.WalkDir(cacheBase, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "local" {
			rel, _ := filepath.Rel(cacheBase, filepath.Dir(path))
			installed = append(installed, rel)
		}
		return nil
	})
	return installed, nil
}

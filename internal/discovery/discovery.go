package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/lovelyJason/openskills/internal/codexmgr"
	"github.com/lovelyJason/openskills/internal/paths"
)

type ResourceInfo struct {
	Name   string
	Source string // marketplace name, "system", "agents", etc.
	Tag    string // "official", "community", "enabled", "disabled", etc.
	IsOSK  bool   // managed by openskills (osk-- prefix)
}

type PlatformResources struct {
	Marketplaces []ResourceInfo
	Plugins      []ResourceInfo
	Skills       []ResourceInfo
}

func DiscoverClaude() *PlatformResources {
	res := &PlatformResources{}
	home := paths.ClaudeHome()

	res.Marketplaces = discoverClaudeMarketplaces(home)
	res.Plugins = DiscoverClaudePluginDetails(home)
	res.Skills = discoverClaudeSkills(home)

	return res
}

func DiscoverCodex() *PlatformResources {
	res := &PlatformResources{}

	res.Plugins = discoverCodexPlugins()
	res.Skills = discoverCodexSkills()

	return res
}

func DiscoverCursor() *PlatformResources {
	res := &PlatformResources{}
	res.Skills = discoverCursorSkills()
	return res
}

func discoverClaudeMarketplaces(home string) []ResourceInfo {
	mpDir := filepath.Join(home, "plugins", "marketplaces")
	entries, err := os.ReadDir(mpDir)
	if err != nil {
		return nil
	}

	var result []ResourceInfo
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") || e.Name() == "temp" {
			continue
		}
		name := e.Name()
		tag := "community"
		if strings.Contains(name, "anthropic") || name == "claude-code-plugins" || name == "claude-plugins-official" {
			tag = "official"
		}
		result = append(result, ResourceInfo{Name: name, Tag: tag})
	}
	return result
}


func discoverClaudeSkills(home string) []ResourceInfo {
	var result []ResourceInfo

	skillsDir := filepath.Join(home, "skills")
	if entries, err := os.ReadDir(skillsDir); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			entryPath := filepath.Join(skillsDir, e.Name())
			fi, err := os.Stat(entryPath)
			if err != nil || !fi.IsDir() {
				continue
			}
			result = append(result, ResourceInfo{
				Name:  e.Name(),
				IsOSK: strings.HasPrefix(e.Name(), "osk--"),
			})
		}
	}

	return result
}

func discoverCodexPlugins() []ResourceInfo {
	if result := discoverCodexPluginsViaRPC(); result != nil {
		return result
	}
	return discoverCodexPluginsFallback()
}

func discoverCodexPluginsViaRPC() []ResourceInfo {
	resp, err := codexmgr.RPCPluginList()
	if err != nil {
		return nil
	}

	var result []ResourceInfo
	for _, mp := range resp.Marketplaces {
		mpDisplay := mp.Name
		if iface := mp.Interface; iface != nil {
			if dn, ok := iface["displayName"].(string); ok && dn != "" {
				mpDisplay = dn
			}
		}

		for _, p := range mp.Plugins {
			displayName := p.Name
			if iface := p.Interface; iface != nil {
				if dn, ok := iface["displayName"].(string); ok && dn != "" {
					displayName = dn
				}
			}

			tag := "available"
			if p.Installed {
				tag = "installed"
			}

			source := mpDisplay
			if parts := strings.SplitN(p.Name, "--", 2); len(parts) == 2 {
				source = parts[0]
			}

			result = append(result, ResourceInfo{
				Name:   displayName,
				Source: source,
				Tag:    tag,
			})
		}
	}
	return result
}

func discoverCodexPluginsFallback() []ResourceInfo {
	installed, _ := codexmgr.InstalledPluginsFromConfig()
	if installed == nil {
		installed = make(map[string]bool)
	}

	registered, _ := codexmgr.RegisteredPlugins()

	seen := make(map[string]bool)
	var result []ResourceInfo

	for _, desc := range registered {
		seen[desc.LocalName] = true
		tag := "available"
		if _, isInstalled := installed[desc.LocalName]; isInstalled {
			tag = "installed"
		}
		result = append(result, ResourceInfo{
			Name:   desc.DisplayName,
			Source: codexmgr.LocalMarketplaceDisplay,
			Tag:    tag,
		})
	}

	for name := range installed {
		if seen[name] {
			continue
		}
		result = append(result, ResourceInfo{Name: name, Tag: "installed"})
	}

	return result
}

func discoverCodexSkills() []ResourceInfo {
	var result []ResourceInfo

	systemDir := filepath.Join(paths.CodexHome(), "skills", ".system")
	if entries, err := os.ReadDir(systemDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
				continue
			}
			result = append(result, ResourceInfo{
				Name:   e.Name(),
				Source: "system",
				Tag:    "System",
			})
		}
	}

	agentsDir := filepath.Join(paths.AgentsDir(), "skills")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
			entryPath := filepath.Join(agentsDir, e.Name())
			fi, err := os.Stat(entryPath)
			if err != nil || !fi.IsDir() {
				continue
			}

			subEntries, err := os.ReadDir(entryPath)
			if err != nil {
				continue
			}

			hasSkillMD := false
			for _, sub := range subEntries {
				if sub.Name() == "SKILL.md" {
					hasSkillMD = true
					break
				}
			}

			if hasSkillMD {
				result = append(result, ResourceInfo{
					Name:  e.Name(),
					Source: "agents",
					Tag:   "Agents",
					IsOSK: strings.HasPrefix(e.Name(), "osk--"),
				})
			} else {
				groupName := e.Name()
				for _, sub := range subEntries {
					if strings.HasPrefix(sub.Name(), ".") {
						continue
					}
					subPath := filepath.Join(entryPath, sub.Name())
					subFi, err := os.Stat(subPath)
					if err != nil || !subFi.IsDir() {
						continue
					}
					result = append(result, ResourceInfo{
						Name:   sub.Name(),
						Source:  groupName,
						Tag:    groupName,
						IsOSK:  strings.HasPrefix(sub.Name(), "osk--"),
					})
				}
			}
		}
	}

	return result
}

func discoverCursorSkills() []ResourceInfo {
	skillsDir := filepath.Join(paths.CursorHome(), "skills-cursor")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil
	}

	var result []ResourceInfo
	for _, e := range entries {
		if !e.IsDir() && !isSymlink(filepath.Join(skillsDir, e.Name())) {
			if strings.HasPrefix(e.Name(), ".") {
				continue
			}
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		result = append(result, ResourceInfo{
			Name:  e.Name(),
			IsOSK: strings.HasPrefix(e.Name(), "osk--"),
		})
	}
	return result
}

func isSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// DiscoverClaudePluginDetails reads plugin metadata from cache directories.
func DiscoverClaudePluginDetails(home string) []ResourceInfo {
	cacheDir := filepath.Join(home, "plugins", "cache")
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil
	}

	var result []ResourceInfo
	for _, marketplace := range entries {
		if !marketplace.IsDir() || strings.HasPrefix(marketplace.Name(), ".") {
			continue
		}
		mpName := marketplace.Name()
		if mpName == "osk" {
			continue
		}

		mpDir := filepath.Join(cacheDir, mpName)
		plugins, err := os.ReadDir(mpDir)
		if err != nil {
			continue
		}
		for _, plugin := range plugins {
			if !plugin.IsDir() || strings.HasPrefix(plugin.Name(), ".") {
				continue
			}
			name := plugin.Name()
			source := mpName
			displayName := readPluginDisplayName(filepath.Join(mpDir, name))
			if displayName != "" {
				name = displayName
			}
			result = append(result, ResourceInfo{Name: name, Source: source})
		}
	}
	return result
}

func readPluginDisplayName(pluginDir string) string {
	versions, err := os.ReadDir(pluginDir)
	if err != nil {
		return ""
	}
	for _, v := range versions {
		if !v.IsDir() {
			continue
		}
		manifestPath := filepath.Join(pluginDir, v.Name(), "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			manifestPath = filepath.Join(pluginDir, v.Name(), ".claude-plugin", "plugin.json")
			data, err = os.ReadFile(manifestPath)
			if err != nil {
				continue
			}
		}
		var m struct {
			Name string `json:"name"`
		}
		if json.Unmarshal(data, &m) == nil && m.Name != "" {
			return m.Name
		}
	}
	return ""
}

package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/lovelyJason/openskills/internal/resource"
)

type PluginManifest struct {
	Name      string                 `json:"name"`
	Interface map[string]interface{} `json:"interface,omitempty"`
}

type SkillMeta struct {
	Name        string
	Description string
	Dir         string
}

func ScanPlugins(repoRoot, marketplaceName string) ([]resource.Resource, error) {
	pluginsDir := filepath.Join(repoRoot, "plugins")
	if _, err := os.Stat(pluginsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return nil, err
	}

	var plugins []resource.Resource
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		manifestPath := filepath.Join(pluginsDir, entry.Name(), ".codex-plugin", "plugin.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifest PluginManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		name := manifest.Name
		if name == "" {
			name = entry.Name()
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

		plugins = append(plugins, resource.Resource{
			Name:        name,
			Type:        resource.TypePlugin,
			Marketplace: marketplaceName,
			LocalPath:   filepath.Join(pluginsDir, entry.Name()),
			Description: displayName,
			Category:    category,
		})
	}
	return plugins, nil
}

func ScanSkills(repoRoot, marketplaceName string) ([]resource.Resource, error) {
	skillsDir := filepath.Join(repoRoot, "skills")
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}

	var skills []resource.Resource
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		skillMD := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMD); os.IsNotExist(err) {
			continue
		}

		desc := extractSkillDescription(skillMD)

		skills = append(skills, resource.Resource{
			Name:        entry.Name(),
			Type:        resource.TypeSkill,
			Marketplace: marketplaceName,
			LocalPath:   filepath.Join(skillsDir, entry.Name()),
			Description: desc,
		})
	}
	return skills, nil
}

func ScanAll(repoRoot, marketplaceName string) ([]resource.Resource, error) {
	plugins, err := ScanPlugins(repoRoot, marketplaceName)
	if err != nil {
		return nil, err
	}
	skills, err := ScanSkills(repoRoot, marketplaceName)
	if err != nil {
		return nil, err
	}
	return append(plugins, skills...), nil
}

// HasCodexPlugins checks if the repo has plugins/<name>/.codex-plugin/plugin.json structure.
func HasCodexPlugins(repoRoot string) bool {
	pluginsDir := filepath.Join(repoRoot, "plugins")
	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		mp := filepath.Join(pluginsDir, e.Name(), ".codex-plugin", "plugin.json")
		if _, err := os.Stat(mp); err == nil {
			return true
		}
	}
	return false
}

// HasClaudePlugin checks if the repo has .claude-plugin/plugin.json or .claude-plugin/marketplace.json at the root.
func HasClaudePlugin(repoRoot string) bool {
	dir := filepath.Join(repoRoot, ".claude-plugin")
	for _, f := range []string{"plugin.json", "marketplace.json"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			return true
		}
	}
	return false
}

// ScanClaudePlugin scans .claude-plugin/marketplace.json or .claude-plugin/plugin.json and returns Resources.
func ScanClaudePlugin(repoRoot, marketplaceName string) ([]resource.Resource, error) {
	dir := filepath.Join(repoRoot, ".claude-plugin")

	// Try marketplace.json first (multi-plugin marketplace format)
	mpPath := filepath.Join(dir, "marketplace.json")
	if data, err := os.ReadFile(mpPath); err == nil {
		return parseClaudeMarketplace(data, repoRoot, marketplaceName)
	}

	// Fallback to plugin.json (single-plugin format)
	pluginPath := filepath.Join(dir, "plugin.json")
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		return nil, nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil
	}

	name, _ := raw["name"].(string)
	if name == "" {
		name = filepath.Base(repoRoot)
	}

	desc := name
	if d, _ := raw["description"].(string); d != "" {
		desc = d
	}

	return []resource.Resource{{
		Name:        name,
		Type:        resource.TypePlugin,
		Marketplace: marketplaceName,
		LocalPath:   repoRoot,
		Description: desc,
		Category:    "Developer Tools",
	}}, nil
}

type claudeMarketplaceJSON struct {
	Name    string `json:"name"`
	Plugins []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Source      string `json:"source"`
		Category    string `json:"category"`
	} `json:"plugins"`
}

func parseClaudeMarketplace(data []byte, repoRoot, marketplaceName string) ([]resource.Resource, error) {
	var mp claudeMarketplaceJSON
	if err := json.Unmarshal(data, &mp); err != nil {
		return nil, nil
	}

	var resources []resource.Resource
	for _, p := range mp.Plugins {
		localPath := repoRoot
		if p.Source != "" && p.Source != "./" {
			localPath = filepath.Join(repoRoot, p.Source)
		}
		cat := p.Category
		if cat == "" {
			cat = "Developer Tools"
		}
		resources = append(resources, resource.Resource{
			Name:        p.Name,
			Type:        resource.TypePlugin,
			Marketplace: marketplaceName,
			LocalPath:   localPath,
			Description: p.Description,
			Category:    cat,
		})
	}
	return resources, nil
}

func extractSkillDescription(skillMDPath string) string {
	data, err := os.ReadFile(skillMDPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	i := 0

	// Skip YAML frontmatter (--- ... ---)
	if i < len(lines) && strings.TrimSpace(lines[i]) == "---" {
		i++
		for i < len(lines) {
			if strings.TrimSpace(lines[i]) == "---" {
				i++
				break
			}
			i++
		}
	}

	for ; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
		if trimmed != "" && !strings.HasPrefix(trimmed, "```") {
			if len(trimmed) > 80 {
				return trimmed[:80] + "..."
			}
			return trimmed
		}
	}
	return ""
}

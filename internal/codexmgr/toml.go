package codexmgr

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func setPluginEnabled(localName string, enabled bool) error {
	return setPluginEnabledIn(configPath(), localName, enabled)
}

func clearPluginConfig(localName string) error {
	return clearPluginConfigIn(configPath(), localName)
}

func setPluginEnabledIn(cfgPath, localName string, enabled bool) error {
	section := fmt.Sprintf(`[plugins."%s@%s"]`, localName, LocalMarketplaceName)
	value := "true"
	if !enabled {
		value = "false"
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	text := string(data)

	headerRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(section))

	if loc := headerRe.FindStringIndex(text); loc != nil {
		blockStart := loc[0]
		blockEnd := len(text)
		if next := regexp.MustCompile(`(?m)^\[`).FindStringIndex(text[loc[1]:]); next != nil {
			blockEnd = loc[1] + next[0]
		}
		block := text[blockStart:blockEnd]
		enabledRe := regexp.MustCompile(`(?m)^\s*enabled\s*=.*$`)
		if enabledRe.MatchString(block) {
			block = enabledRe.ReplaceAllString(block, "enabled = "+value)
		} else {
			block = strings.TrimRight(block, "\n") + "\nenabled = " + value + "\n"
		}
		text = text[:blockStart] + block + text[blockEnd:]
	} else {
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		text += fmt.Sprintf("\n%s\nenabled = %s\n", section, value)
	}

	if err := os.MkdirAll(strings.TrimSuffix(cfgPath, "/config.toml"), 0755); err != nil {
		// best-effort directory creation
	}
	return os.WriteFile(cfgPath, []byte(text), 0644)
}

func clearPluginConfigIn(cfgPath, localName string) error {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	text := string(data)
	header := fmt.Sprintf(`[plugins."%s@%s"]`, localName, LocalMarketplaceName)
	headerRe := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(header))

	if loc := headerRe.FindStringIndex(text); loc != nil {
		blockEnd := len(text)
		if next := regexp.MustCompile(`(?m)^\[`).FindStringIndex(text[loc[1]:]); next != nil {
			blockEnd = loc[1] + next[0]
		}
		text = text[:loc[0]] + text[blockEnd:]
	}

	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	text = strings.TrimLeft(text, "\n")

	return os.WriteFile(cfgPath, []byte(text), 0644)
}

func InstalledPluginsFromConfig() (map[string]bool, error) {
	return installedPluginsFromConfig()
}

type InstalledPlugin struct {
	Name        string
	Marketplace string
	Enabled     bool
}

// AllInstalledPlugins reads every [plugins."name@marketplace"] section from config.toml.
func AllInstalledPlugins() ([]InstalledPlugin, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	text := string(data)
	pattern := regexp.MustCompile(`(?m)^\[plugins\."([^"]+)@([^"]+)"\]`)
	matches := pattern.FindAllStringSubmatchIndex(text, -1)

	var result []InstalledPlugin
	for _, loc := range matches {
		pluginName := text[loc[2]:loc[3]]
		marketplace := text[loc[4]:loc[5]]
		blockStart := loc[1]
		blockEnd := len(text)
		nextSection := regexp.MustCompile(`(?m)^\[`).FindStringIndex(text[blockStart:])
		if nextSection != nil {
			blockEnd = blockStart + nextSection[0]
		}
		block := text[blockStart:blockEnd]
		enabled := false
		enabledRe := regexp.MustCompile(`(?m)^\s*enabled\s*=\s*(\w+)`)
		if m := enabledRe.FindStringSubmatch(block); m != nil {
			enabled = m[1] == "true"
		}
		result = append(result, InstalledPlugin{
			Name:        pluginName,
			Marketplace: marketplace,
			Enabled:     enabled,
		})
	}
	return result, nil
}

func installedPluginsFromConfig() (map[string]bool, error) {
	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	result := make(map[string]bool)
	pattern := regexp.MustCompile(
		`(?m)^\[plugins\."([^"]+)@` + regexp.QuoteMeta(LocalMarketplaceName) + `"\]`,
	)

	text := string(data)
	matches := pattern.FindAllStringSubmatchIndex(text, -1)
	for _, loc := range matches {
		localName := text[loc[2]:loc[3]]
		blockStart := loc[1]
		blockEnd := len(text)
		nextSection := regexp.MustCompile(`(?m)^\[`).FindStringIndex(text[blockStart:])
		if nextSection != nil {
			blockEnd = blockStart + nextSection[0]
		}
		block := text[blockStart:blockEnd]
		enabledRe := regexp.MustCompile(`(?m)^\s*enabled\s*=\s*(\w+)`)
		if m := enabledRe.FindStringSubmatch(block); m != nil {
			result[localName] = m[1] == "true"
		} else {
			result[localName] = false
		}
	}
	return result, nil
}

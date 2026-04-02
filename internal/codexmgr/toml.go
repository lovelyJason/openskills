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

	pattern := regexp.MustCompile(fmt.Sprintf(
		`(?ms)^\[plugins\."%s@%s"\]\n.*?(?=^\[|\z)`,
		regexp.QuoteMeta(localName), regexp.QuoteMeta(LocalMarketplaceName),
	))

	if pattern.MatchString(text) {
		text = pattern.ReplaceAllStringFunc(text, func(block string) string {
			enabledRe := regexp.MustCompile(`(?m)^\s*enabled\s*=.*$`)
			if enabledRe.MatchString(block) {
				return enabledRe.ReplaceAllString(block, "enabled = "+value)
			}
			return strings.TrimRight(block, "\n") + "\nenabled = " + value + "\n"
		})
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

	pattern := regexp.MustCompile(fmt.Sprintf(
		`(?ms)^\[plugins\."%s@%s"\]\n.*?(?=^\[|\z)`,
		regexp.QuoteMeta(localName), regexp.QuoteMeta(LocalMarketplaceName),
	))

	text := pattern.ReplaceAllString(string(data), "")
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	text = strings.TrimLeft(text, "\n")

	return os.WriteFile(cfgPath, []byte(text), 0644)
}

func InstalledPluginsFromConfig() (map[string]bool, error) {
	return installedPluginsFromConfig()
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

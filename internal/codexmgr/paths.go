package codexmgr

import (
	"path/filepath"

	"github.com/lovelyJason/openskills/internal/paths"
)

const (
	LocalMarketplaceName    = "opencodex-local"
	LocalMarketplaceDisplay = "OpenCodex Local"
)

func stateDir() string {
	return filepath.Join(paths.CodexHome(), "opencodex")
}

func pluginStateFile() string {
	return filepath.Join(stateDir(), "plugins.json")
}

func reposDir() string {
	return paths.ReposDir()
}

func marketplaceDir() string {
	return filepath.Join(paths.AgentsDir(), "plugins")
}

func marketplacePath() string {
	return filepath.Join(marketplaceDir(), "marketplace.json")
}

func localPluginDir() string {
	return filepath.Join(paths.Home(), "plugins")
}

func configPath() string {
	return filepath.Join(paths.CodexHome(), "config.toml")
}

func pluginCacheRoot(localName string) string {
	return filepath.Join(paths.CodexHome(), "plugins", "cache", LocalMarketplaceName, localName)
}

func pluginCacheVersionRoot(localName string) string {
	return filepath.Join(pluginCacheRoot(localName), "local")
}

func officialMarketplacePath() string {
	return filepath.Join(paths.CodexHome(), ".tmp", "plugins", ".agents", "plugins", "marketplace.json")
}

func CacheVersionRootPath(localName string) string {
	return pluginCacheVersionRoot(localName)
}

func NormalizeLocalName(marketplace, plugin string) string {
	return normalizeLocalName(marketplace, plugin)
}

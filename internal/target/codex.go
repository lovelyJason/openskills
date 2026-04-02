package target

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lovelyJason/openskills/internal/codexmgr"
	"github.com/lovelyJason/openskills/internal/fsutil"
	"github.com/lovelyJason/openskills/internal/paths"
	"github.com/lovelyJason/openskills/internal/resource"
)

type Codex struct {
	mgr *codexmgr.Manager
}

func NewCodex() *Codex {
	return &Codex{mgr: codexmgr.NewManager()}
}

func (c *Codex) Name() string { return "codex" }

func (c *Codex) Detect() bool {
	_, err := os.Stat(paths.CodexHome())
	return err == nil
}

func (c *Codex) CheckVersion() error {
	return codexmgr.CheckVersion()
}

func (c *Codex) SupportedResources() []resource.Type {
	return []resource.Type{resource.TypePlugin, resource.TypeSkill}
}

func (c *Codex) OnMarketplaceAdd(_ context.Context, url, name, repoDir string) error {
	return c.mgr.MarketplaceAdd(repoDir, url, name)
}

func (c *Codex) OnMarketplaceRemove(_ context.Context, name string) error {
	return c.mgr.MarketplaceRemove(name)
}

func (c *Codex) OnMarketplaceUpdate(_ context.Context, name, repoDir string) error {
	return c.mgr.MarketplaceUpdate(name, repoDir)
}

func (c *Codex) Install(_ context.Context, res *resource.Resource, mode resource.InstallMode, sourcePath string) (*InstallResult, error) {
	switch res.Type {
	case resource.TypePlugin:
		return c.installPlugin(res, sourcePath)
	case resource.TypeSkill:
		return c.installSkill(res, mode, sourcePath)
	default:
		return nil, fmt.Errorf("unsupported resource type %s for codex", res.Type)
	}
}

func (c *Codex) installPlugin(res *resource.Resource, sourcePath string) (*InstallResult, error) {
	desc, err := c.mgr.FindPluginDescriptor(res.Marketplace, res.Name)
	if err != nil {
		localName := res.Marketplace + "--" + res.Name
		desc = &codexmgr.PluginDescriptor{
			Marketplace: res.Marketplace,
			Name:        res.Name,
			LocalName:   localName,
			DisplayName: res.Description,
			Category:    res.Category,
			PluginDir:   sourcePath,
		}
	}

	fallback, err := c.mgr.PluginInstall(*desc)
	if err != nil {
		return nil, err
	}

	section := fmt.Sprintf(`[plugins."%s@%s"]`, desc.LocalName, codexmgr.LocalMarketplaceName)
	result := &InstallResult{
		Paths:         []string{codexmgr.CacheVersionRootPath(desc.LocalName)},
		ConfigEntries: []string{section},
	}
	if fallback {
		result.NativeRef = "fallback"
	}
	return result, nil
}

func (c *Codex) installSkill(res *resource.Resource, mode resource.InstallMode, sourcePath string) (*InstallResult, error) {
	skillsDir := filepath.Join(paths.AgentsDir(), "skills")
	linkName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
	linkPath := filepath.Join(skillsDir, linkName)

	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return nil, err
	}
	os.Remove(linkPath)

	switch mode {
	case resource.ModeSymlink:
		if err := os.Symlink(sourcePath, linkPath); err != nil {
			return nil, fmt.Errorf("symlink skill: %w", err)
		}
	case resource.ModeNative:
		if err := fsutil.CopyDir(sourcePath, linkPath); err != nil {
			return nil, fmt.Errorf("copy skill: %w", err)
		}
	}

	return &InstallResult{Paths: []string{linkPath}}, nil
}

func (c *Codex) Uninstall(_ context.Context, res *resource.Resource) error {
	switch res.Type {
	case resource.TypePlugin:
		return c.uninstallPlugin(res)
	case resource.TypeSkill:
		linkName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
		linkPath := filepath.Join(paths.AgentsDir(), "skills", linkName)
		os.Remove(linkPath)
		return nil
	}
	return nil
}

func (c *Codex) uninstallPlugin(res *resource.Resource) error {
	desc, err := c.mgr.FindPluginDescriptor(res.Marketplace, res.Name)
	if err != nil {
		localName := res.Marketplace + "--" + res.Name
		desc = &codexmgr.PluginDescriptor{
			Marketplace: res.Marketplace,
			Name:        res.Name,
			LocalName:   localName,
		}
	}

	_, uninstallErr := c.mgr.PluginUninstall(*desc)
	return uninstallErr
}

func (c *Codex) IsInstalled(res *resource.Resource) (bool, error) {
	switch res.Type {
	case resource.TypePlugin:
		localName := res.Marketplace + "--" + res.Name
		_, err := os.Stat(codexmgr.CacheVersionRootPath(localName))
		return err == nil, nil
	case resource.TypeSkill:
		linkName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
		linkPath := filepath.Join(paths.AgentsDir(), "skills", linkName)
		_, err := os.Lstat(linkPath)
		return err == nil, nil
	}
	return false, nil
}

func (c *Codex) Manager() *codexmgr.Manager {
	return c.mgr
}

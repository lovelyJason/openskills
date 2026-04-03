package target

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lovelyJason/openskills/internal/claudecli"
	"github.com/lovelyJason/openskills/internal/fsutil"
	"github.com/lovelyJason/openskills/internal/paths"
	"github.com/lovelyJason/openskills/internal/resource"
)

type Claude struct{}

func NewClaude() *Claude { return &Claude{} }

func (c *Claude) Name() string { return "claude" }

func (c *Claude) Detect() bool {
	_, err := os.Stat(paths.ClaudeHome())
	return err == nil
}

func (c *Claude) SupportedResources() []resource.Type {
	return []resource.Type{resource.TypePlugin, resource.TypeSkill}
}

func (c *Claude) OnMarketplaceAdd(_ context.Context, url, name, repoDir string) error {
	cli, err := claudecli.New()
	if err != nil {
		return fmt.Errorf("claude CLI not available: %w", err)
	}
	output, err := cli.MarketplaceAdd(url)
	if err != nil {
		return err
	}
	if output != "" {
		fmt.Printf("  \033[2m[claude] %s\033[0m\n", output)
	}
	return nil
}

func (c *Claude) OnMarketplaceRemove(_ context.Context, name string) error {
	cli, err := claudecli.New()
	if err != nil {
		return fmt.Errorf("claude CLI not available: %w", err)
	}
	return cli.MarketplaceRemove(name)
}

func (c *Claude) OnMarketplaceUpdate(_ context.Context, name, repoDir string) error {
	cli, err := claudecli.New()
	if err != nil {
		return fmt.Errorf("claude CLI not available: %w", err)
	}
	return cli.MarketplaceUpdate(name)
}

func (c *Claude) Install(_ context.Context, res *resource.Resource, mode resource.InstallMode, sourcePath string) (*InstallResult, error) {
	switch res.Type {
	case resource.TypePlugin:
		return c.installPlugin(res)
	case resource.TypeSkill:
		return c.installSkill(res, mode, sourcePath)
	default:
		return nil, fmt.Errorf("unsupported resource type %s for claude", res.Type)
	}
}

func (c *Claude) installPlugin(res *resource.Resource) (*InstallResult, error) {
	cli, err := claudecli.New()
	if err != nil {
		return nil, fmt.Errorf("claude CLI not available: %w", err)
	}

	ref := fmt.Sprintf("%s@%s", res.Name, res.Marketplace)
	if err := cli.PluginInstall(ref); err != nil {
		return nil, err
	}

	return &InstallResult{
		NativeRef: ref,
	}, nil
}

func (c *Claude) installSkill(res *resource.Resource, mode resource.InstallMode, sourcePath string) (*InstallResult, error) {
	skillsDir := filepath.Join(paths.ClaudeHome(), "skills")
	skillName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
	skillDir := filepath.Join(skillsDir, skillName)

	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return nil, err
	}
	os.RemoveAll(skillDir)

	switch mode {
	case resource.ModeSymlink:
		if err := os.Symlink(sourcePath, skillDir); err != nil {
			return nil, fmt.Errorf("symlink skill: %w", err)
		}
	case resource.ModeNative:
		if err := fsutil.CopyDir(sourcePath, skillDir); err != nil {
			return nil, fmt.Errorf("copy skill: %w", err)
		}
	}

	return &InstallResult{Paths: []string{skillDir}}, nil
}

func (c *Claude) Uninstall(_ context.Context, res *resource.Resource) error {
	switch res.Type {
	case resource.TypePlugin:
		cli, err := claudecli.New()
		if err != nil {
			return fmt.Errorf("claude CLI not available: %w", err)
		}
		return cli.PluginUninstall(res.Name)
	case resource.TypeSkill:
		skillName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
		skillDir := filepath.Join(paths.ClaudeHome(), "skills", skillName)
		os.RemoveAll(skillDir)
		return nil
	}
	return nil
}

func (c *Claude) IsInstalled(res *resource.Resource) (bool, error) {
	switch res.Type {
	case resource.TypePlugin:
		cacheRoot := filepath.Join(paths.ClaudeHome(), "plugins", "cache", "osk", res.Name)
		_, err := os.Stat(cacheRoot)
		return err == nil, nil
	case resource.TypeSkill:
		skillName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
		skillDir := filepath.Join(paths.ClaudeHome(), "skills", skillName)
		_, err := os.Lstat(skillDir)
		return err == nil, nil
	}
	return false, nil
}

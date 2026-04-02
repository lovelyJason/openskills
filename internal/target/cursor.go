package target

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lovelyJason/openskills/internal/fsutil"
	"github.com/lovelyJason/openskills/internal/paths"
	"github.com/lovelyJason/openskills/internal/resource"
)

type Cursor struct{}

func NewCursor() *Cursor { return &Cursor{} }

func (c *Cursor) Name() string { return "cursor" }

func (c *Cursor) Detect() bool {
	_, err := os.Stat(paths.CursorHome())
	return err == nil
}

func (c *Cursor) SupportedResources() []resource.Type {
	return []resource.Type{resource.TypeSkill}
}

func (c *Cursor) Install(_ context.Context, res *resource.Resource, mode resource.InstallMode, sourcePath string) (*InstallResult, error) {
	if res.Type != resource.TypeSkill {
		return nil, fmt.Errorf("cursor only supports skills, not %s", res.Type)
	}

	skillsDir := filepath.Join(paths.CursorHome(), "skills-cursor")
	skillName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
	destPath := filepath.Join(skillsDir, skillName)

	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return nil, err
	}

	os.RemoveAll(destPath)

	switch mode {
	case resource.ModeSymlink:
		if err := os.Symlink(sourcePath, destPath); err != nil {
			return nil, fmt.Errorf("symlink skill: %w", err)
		}
	case resource.ModeNative:
		if err := fsutil.CopyDir(sourcePath, destPath); err != nil {
			return nil, fmt.Errorf("copy skill: %w", err)
		}
	}

	return &InstallResult{
		Paths: []string{destPath},
	}, nil
}

func (c *Cursor) Uninstall(_ context.Context, res *resource.Resource) error {
	skillName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
	destPath := filepath.Join(paths.CursorHome(), "skills-cursor", skillName)
	return os.RemoveAll(destPath)
}

func (c *Cursor) IsInstalled(res *resource.Resource) (bool, error) {
	skillName := fmt.Sprintf("osk--%s--%s", res.Marketplace, res.Name)
	destPath := filepath.Join(paths.CursorHome(), "skills-cursor", skillName)
	_, err := os.Lstat(destPath)
	return err == nil, nil
}

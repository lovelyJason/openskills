package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lovelyJason/openskills/internal/resource"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	pluginDir := filepath.Join(root, "plugins", "my-plugin", ".codex-plugin")
	os.MkdirAll(pluginDir, 0755)
	manifest := PluginManifest{
		Name: "my-plugin",
		Interface: map[string]interface{}{
			"displayName": "My Plugin",
			"category":    "Testing",
		},
	}
	data, _ := json.Marshal(manifest)
	os.WriteFile(filepath.Join(pluginDir, "plugin.json"), data, 0644)

	pluginDir2 := filepath.Join(root, "plugins", "bare-plugin", ".codex-plugin")
	os.MkdirAll(pluginDir2, 0755)
	bare := PluginManifest{}
	data2, _ := json.Marshal(bare)
	os.WriteFile(filepath.Join(pluginDir2, "plugin.json"), data2, 0644)

	skillDir := filepath.Join(root, "skills", "git-commit")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Git Commit Helper\nAutomatically generates commits."), 0644)

	skillDir2 := filepath.Join(root, "skills", "code-review")
	os.MkdirAll(skillDir2, 0755)
	os.WriteFile(filepath.Join(skillDir2, "SKILL.md"), []byte("---\nname: Code Review\n---\nReviews code changes."), 0644)

	os.MkdirAll(filepath.Join(root, "skills", "no-skill-md"), 0755)

	os.MkdirAll(filepath.Join(root, "plugins", ".hidden"), 0755)

	return root
}

func TestScanPlugins(t *testing.T) {
	root := setupTestRepo(t)
	plugins, err := ScanPlugins(root, "test-mp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	var found bool
	for _, p := range plugins {
		if p.Name == "my-plugin" {
			found = true
			if p.Type != resource.TypePlugin {
				t.Errorf("wrong type: %s", p.Type)
			}
			if p.Marketplace != "test-mp" {
				t.Errorf("wrong marketplace: %s", p.Marketplace)
			}
			if p.Description != "My Plugin" {
				t.Errorf("wrong description: %s", p.Description)
			}
			if p.Category != "Testing" {
				t.Errorf("wrong category: %s", p.Category)
			}
		}
	}
	if !found {
		t.Error("my-plugin not found in results")
	}
}

func TestScanPlugins_FallbackName(t *testing.T) {
	root := setupTestRepo(t)
	plugins, _ := ScanPlugins(root, "test-mp")
	for _, p := range plugins {
		if p.Name == "bare-plugin" {
			if p.Category != "Developer Tools" {
				t.Errorf("expected default category, got %s", p.Category)
			}
			return
		}
	}
	t.Error("bare-plugin not found")
}

func TestScanPlugins_NoPluginsDir(t *testing.T) {
	root := t.TempDir()
	plugins, err := ScanPlugins(root, "test-mp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plugins != nil {
		t.Errorf("expected nil, got %v", plugins)
	}
}

func TestScanPlugins_SkipsHidden(t *testing.T) {
	root := setupTestRepo(t)
	plugins, _ := ScanPlugins(root, "test-mp")
	for _, p := range plugins {
		if p.Name == ".hidden" {
			t.Error("should not include hidden directories")
		}
	}
}

func TestScanSkills(t *testing.T) {
	root := setupTestRepo(t)
	skills, err := ScanSkills(root, "test-mp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	names := map[string]bool{}
	for _, s := range skills {
		names[s.Name] = true
		if s.Type != resource.TypeSkill {
			t.Errorf("wrong type for %s: %s", s.Name, s.Type)
		}
		if s.Marketplace != "test-mp" {
			t.Errorf("wrong marketplace for %s: %s", s.Name, s.Marketplace)
		}
	}
	if !names["git-commit"] {
		t.Error("git-commit skill not found")
	}
	if !names["code-review"] {
		t.Error("code-review skill not found")
	}
	if names["no-skill-md"] {
		t.Error("no-skill-md should be excluded (no SKILL.md)")
	}
}

func TestScanSkills_NoSkillsDir(t *testing.T) {
	root := t.TempDir()
	skills, err := ScanSkills(root, "test-mp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skills != nil {
		t.Errorf("expected nil, got %v", skills)
	}
}

func TestScanAll(t *testing.T) {
	root := setupTestRepo(t)
	all, err := ScanAll(root, "test-mp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 4 {
		t.Errorf("expected 4 resources (2 plugins + 2 skills), got %d", len(all))
	}
}

func TestExtractSkillDescription_Heading(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	os.WriteFile(path, []byte("# My Great Skill\nSome content."), 0644)
	desc := extractSkillDescription(path)
	if desc != "My Great Skill" {
		t.Errorf("got %q, want 'My Great Skill'", desc)
	}
}

func TestExtractSkillDescription_NoHeading(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	os.WriteFile(path, []byte("This is a skill without headings."), 0644)
	desc := extractSkillDescription(path)
	if desc != "This is a skill without headings." {
		t.Errorf("got %q", desc)
	}
}

func TestExtractSkillDescription_LongLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	long := ""
	for i := 0; i < 100; i++ {
		long += "x"
	}
	os.WriteFile(path, []byte(long), 0644)
	desc := extractSkillDescription(path)
	if len(desc) > 84 {
		t.Errorf("expected truncated desc, got len %d", len(desc))
	}
}

func TestExtractSkillDescription_Frontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	os.WriteFile(path, []byte("---\nname: test\n---\nActual content here."), 0644)
	desc := extractSkillDescription(path)
	if desc != "Actual content here." {
		t.Errorf("got %q, want 'Actual content here.'", desc)
	}
}

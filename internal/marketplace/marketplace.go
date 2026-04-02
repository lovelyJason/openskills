package marketplace

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lovelyJason/openskills/internal/gitutil"
	"github.com/lovelyJason/openskills/internal/paths"
	"github.com/lovelyJason/openskills/internal/resource"
	"github.com/lovelyJason/openskills/internal/scanner"
	"github.com/lovelyJason/openskills/internal/state"
)

var nameRe = regexp.MustCompile(`[^a-z0-9]+`)

func NormalizeName(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = nameRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "marketplace"
	}
	return s
}

func NameFromURL(url string) string {
	base := filepath.Base(url)
	base = strings.TrimSuffix(base, ".git")
	return NormalizeName(base)
}

func RepoDir(name string) string {
	return filepath.Join(paths.ReposDir(), name)
}

func Add(st *state.Store, url, name string) (*state.MarketplaceEntry, error) {
	return AddWithSource(st, url, name, state.SourceMarketplace)
}

func AddSkillRepo(st *state.Store, url, name string) (*state.MarketplaceEntry, error) {
	return AddWithSource(st, url, name, state.SourceSkillRepo)
}

func AddWithSource(st *state.Store, url, name string, source state.SourceType) (*state.MarketplaceEntry, error) {
	if name == "" {
		name = NameFromURL(url)
	} else {
		name = NormalizeName(name)
	}

	if existing := st.FindMarketplace(name); existing != nil {
		if existing.URL != url {
			return nil, fmt.Errorf("source %q already exists with different URL: %s", name, existing.URL)
		}
		if err := gitutil.Pull(existing.LocalPath); err != nil {
			return nil, err
		}
		existing.LastUpdated = time.Now()
		existing.Source = source
		st.UpsertMarketplace(*existing)
		return existing, nil
	}

	repoDir := RepoDir(name)
	if err := gitutil.Clone(url, repoDir); err != nil {
		return nil, err
	}

	branch, _ := gitutil.CurrentBranch(repoDir)

	entry := state.MarketplaceEntry{
		Name:        name,
		URL:         url,
		LocalPath:   repoDir,
		Branch:      branch,
		Source:      source,
		LastUpdated: time.Now(),
	}
	st.UpsertMarketplace(entry)
	return &entry, nil
}

func Update(entry *state.MarketplaceEntry) error {
	if entry.PinnedVer != "" {
		return fmt.Errorf("marketplace %q is pinned to %s, skipping update", entry.Name, entry.PinnedVer)
	}
	if err := gitutil.Pull(entry.LocalPath); err != nil {
		return err
	}
	entry.LastUpdated = time.Now()
	return nil
}

func Pin(entry *state.MarketplaceEntry, version string) error {
	if err := gitutil.Checkout(entry.LocalPath, version); err != nil {
		return err
	}
	entry.PinnedVer = version
	return nil
}

func Unpin(entry *state.MarketplaceEntry) error {
	branch := entry.Branch
	if branch == "" {
		branch = "main"
	}
	if err := gitutil.Checkout(entry.LocalPath, branch); err != nil {
		return err
	}
	entry.PinnedVer = ""
	return nil
}

func ListResources(entry *state.MarketplaceEntry, resourceType resource.Type) ([]resource.Resource, error) {
	switch resourceType {
	case resource.TypePlugin:
		return scanner.ScanPlugins(entry.LocalPath, entry.Name)
	case resource.TypeSkill:
		return scanner.ScanSkills(entry.LocalPath, entry.Name)
	default:
		return scanner.ScanAll(entry.LocalPath, entry.Name)
	}
}

func ListAllResources(st *state.Store) ([]resource.Resource, error) {
	var all []resource.Resource
	for _, m := range st.Marketplaces {
		resources, err := scanner.ScanAll(m.LocalPath, m.Name)
		if err != nil {
			continue
		}
		all = append(all, resources...)
	}
	return all, nil
}

package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/lovelyJason/openskills/internal/paths"
	"github.com/lovelyJason/openskills/internal/resource"
)

type SourceType string

const (
	SourceMarketplace SourceType = "marketplace"
	SourceSkillRepo   SourceType = "skills"
)

type MarketplaceEntry struct {
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	LocalPath   string     `json:"localPath"`
	Branch      string     `json:"branch,omitempty"`
	PinnedVer   string     `json:"pinnedVersion,omitempty"`
	Source      SourceType `json:"sourceType,omitempty"`
	LastUpdated time.Time  `json:"lastUpdated"`
}

type InstallationTarget struct {
	InstalledAt   time.Time `json:"installedAt"`
	Paths         []string  `json:"paths,omitempty"`
	ConfigEntries []string  `json:"configEntries,omitempty"`
	NativeRef     string    `json:"nativeRef,omitempty"`
}

type Installation struct {
	ID           string                        `json:"id"`
	ResourceType resource.Type                 `json:"resourceType"`
	Name         string                        `json:"name"`
	Marketplace  string                        `json:"marketplace"`
	Version      string                        `json:"version,omitempty"`
	GitCommitSHA string                        `json:"gitCommitSha,omitempty"`
	Mode         resource.InstallMode          `json:"mode"`
	Targets      map[string]InstallationTarget `json:"targets"`
}

type Store struct {
	Version       int                `json:"version"`
	Marketplaces  []MarketplaceEntry `json:"marketplaces"`
	Installations []Installation     `json:"installations"`
}

type Manager struct {
	mu   sync.Mutex
	path string
}

func NewManager() *Manager {
	return &Manager{path: paths.StateFile()}
}

func (m *Manager) Load() (*Store, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.loadLocked()
}

func (m *Manager) loadLocked() (*Store, error) {
	data, err := os.ReadFile(m.path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Store{Version: 1}, nil
		}
		return nil, err
	}
	var s Store
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func (m *Manager) Save(s *Store) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveLocked(s)
}

func (m *Manager) saveLocked(s *Store) error {
	if err := paths.EnsureDir(paths.ConfigDir()); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := m.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, m.path)
}

func (s *Store) FindMarketplace(name string) *MarketplaceEntry {
	for i := range s.Marketplaces {
		if s.Marketplaces[i].Name == name {
			return &s.Marketplaces[i]
		}
	}
	return nil
}

func (s *Store) UpsertMarketplace(entry MarketplaceEntry) {
	for i := range s.Marketplaces {
		if s.Marketplaces[i].Name == entry.Name {
			s.Marketplaces[i] = entry
			return
		}
	}
	s.Marketplaces = append(s.Marketplaces, entry)
}

func (s *Store) RemoveMarketplace(name string) {
	filtered := s.Marketplaces[:0]
	for _, m := range s.Marketplaces {
		if m.Name != name {
			filtered = append(filtered, m)
		}
	}
	s.Marketplaces = filtered
}

func (s *Store) FindInstallation(qualifiedName string) *Installation {
	for i := range s.Installations {
		if s.Installations[i].ID == qualifiedName {
			return &s.Installations[i]
		}
	}
	return nil
}

func (s *Store) UpsertInstallation(inst Installation) {
	for i := range s.Installations {
		if s.Installations[i].ID == inst.ID {
			s.Installations[i] = inst
			return
		}
	}
	s.Installations = append(s.Installations, inst)
}

func (s *Store) RemoveInstallation(qualifiedName string) {
	filtered := s.Installations[:0]
	for _, inst := range s.Installations {
		if inst.ID != qualifiedName {
			filtered = append(filtered, inst)
		}
	}
	s.Installations = filtered
}

func (s *Store) InstallationsByMarketplace(marketplace string) []Installation {
	var result []Installation
	for _, inst := range s.Installations {
		if inst.Marketplace == marketplace {
			result = append(result, inst)
		}
	}
	return result
}

func (s *Store) InstallationsByTarget(targetName string) []Installation {
	var result []Installation
	for _, inst := range s.Installations {
		if _, ok := inst.Targets[targetName]; ok {
			result = append(result, inst)
		}
	}
	return result
}

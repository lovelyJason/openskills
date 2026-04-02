package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lovelyJason/openskills/internal/resource"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	dir := t.TempDir()
	return &Manager{path: filepath.Join(dir, "state.json")}
}

func TestManager_LoadEmpty(t *testing.T) {
	m := newTestManager(t)
	s, err := m.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Version != 1 {
		t.Errorf("expected version 1, got %d", s.Version)
	}
	if len(s.Marketplaces) != 0 {
		t.Errorf("expected empty marketplaces, got %d", len(s.Marketplaces))
	}
	if len(s.Installations) != 0 {
		t.Errorf("expected empty installations, got %d", len(s.Installations))
	}
}

func TestManager_SaveAndLoad(t *testing.T) {
	m := newTestManager(t)
	s := &Store{
		Version: 1,
		Marketplaces: []MarketplaceEntry{
			{Name: "test-mp", URL: "https://example.com/test.git"},
		},
		Installations: []Installation{
			{
				ID:           "plugin-a@test-mp",
				ResourceType: resource.TypePlugin,
				Name:         "plugin-a",
				Marketplace:  "test-mp",
				Version:      "1.0.0",
				Mode:         resource.ModeSymlink,
				Targets:      map[string]InstallationTarget{},
			},
		},
	}

	if err := m.Save(s); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := m.Load()
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Marketplaces) != 1 {
		t.Fatalf("expected 1 marketplace, got %d", len(loaded.Marketplaces))
	}
	if loaded.Marketplaces[0].Name != "test-mp" {
		t.Errorf("got name %s, want test-mp", loaded.Marketplaces[0].Name)
	}
	if len(loaded.Installations) != 1 {
		t.Fatalf("expected 1 installation, got %d", len(loaded.Installations))
	}
	if loaded.Installations[0].Version != "1.0.0" {
		t.Errorf("got version %s, want 1.0.0", loaded.Installations[0].Version)
	}
}

func TestManager_AtomicWrite(t *testing.T) {
	m := newTestManager(t)
	s := &Store{Version: 1}
	if err := m.Save(s); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	tmpPath := m.path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("temp file should not remain after save")
	}
}

func TestStore_FindMarketplace(t *testing.T) {
	s := &Store{
		Marketplaces: []MarketplaceEntry{
			{Name: "alpha"},
			{Name: "beta"},
		},
	}

	if m := s.FindMarketplace("alpha"); m == nil || m.Name != "alpha" {
		t.Error("expected to find alpha")
	}
	if m := s.FindMarketplace("gamma"); m != nil {
		t.Error("expected nil for gamma")
	}
}

func TestStore_UpsertMarketplace_Insert(t *testing.T) {
	s := &Store{}
	s.UpsertMarketplace(MarketplaceEntry{Name: "new", URL: "https://example.com"})
	if len(s.Marketplaces) != 1 {
		t.Fatalf("expected 1, got %d", len(s.Marketplaces))
	}
	if s.Marketplaces[0].URL != "https://example.com" {
		t.Errorf("wrong URL: %s", s.Marketplaces[0].URL)
	}
}

func TestStore_UpsertMarketplace_Update(t *testing.T) {
	s := &Store{
		Marketplaces: []MarketplaceEntry{{Name: "existing", URL: "old-url"}},
	}
	s.UpsertMarketplace(MarketplaceEntry{Name: "existing", URL: "new-url"})
	if len(s.Marketplaces) != 1 {
		t.Fatalf("expected 1, got %d", len(s.Marketplaces))
	}
	if s.Marketplaces[0].URL != "new-url" {
		t.Errorf("expected new-url, got %s", s.Marketplaces[0].URL)
	}
}

func TestStore_RemoveMarketplace(t *testing.T) {
	s := &Store{
		Marketplaces: []MarketplaceEntry{{Name: "a"}, {Name: "b"}, {Name: "c"}},
	}
	s.RemoveMarketplace("b")
	if len(s.Marketplaces) != 2 {
		t.Fatalf("expected 2, got %d", len(s.Marketplaces))
	}
	for _, m := range s.Marketplaces {
		if m.Name == "b" {
			t.Error("b should be removed")
		}
	}
}

func TestStore_FindInstallation(t *testing.T) {
	s := &Store{
		Installations: []Installation{
			{ID: "foo@bar", Name: "foo"},
			{ID: "baz@qux", Name: "baz"},
		},
	}
	if i := s.FindInstallation("foo@bar"); i == nil || i.Name != "foo" {
		t.Error("expected to find foo@bar")
	}
	if i := s.FindInstallation("nonexistent"); i != nil {
		t.Error("expected nil")
	}
}

func TestStore_UpsertInstallation(t *testing.T) {
	s := &Store{}
	s.UpsertInstallation(Installation{ID: "a@b", Version: "1.0"})
	if len(s.Installations) != 1 {
		t.Fatalf("expected 1, got %d", len(s.Installations))
	}
	s.UpsertInstallation(Installation{ID: "a@b", Version: "2.0"})
	if len(s.Installations) != 1 {
		t.Fatalf("expected 1 after upsert, got %d", len(s.Installations))
	}
	if s.Installations[0].Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", s.Installations[0].Version)
	}
}

func TestStore_RemoveInstallation(t *testing.T) {
	s := &Store{
		Installations: []Installation{{ID: "a@b"}, {ID: "c@d"}, {ID: "e@f"}},
	}
	s.RemoveInstallation("c@d")
	if len(s.Installations) != 2 {
		t.Fatalf("expected 2, got %d", len(s.Installations))
	}
}

func TestStore_InstallationsByMarketplace(t *testing.T) {
	s := &Store{
		Installations: []Installation{
			{ID: "a@mp1", Marketplace: "mp1"},
			{ID: "b@mp2", Marketplace: "mp2"},
			{ID: "c@mp1", Marketplace: "mp1"},
		},
	}
	results := s.InstallationsByMarketplace("mp1")
	if len(results) != 2 {
		t.Errorf("expected 2, got %d", len(results))
	}
}

func TestStore_InstallationsByTarget(t *testing.T) {
	s := &Store{
		Installations: []Installation{
			{ID: "a", Targets: map[string]InstallationTarget{
				"codex": {InstalledAt: time.Now()},
			}},
			{ID: "b", Targets: map[string]InstallationTarget{
				"claude": {InstalledAt: time.Now()},
			}},
			{ID: "c", Targets: map[string]InstallationTarget{
				"codex":  {InstalledAt: time.Now()},
				"claude": {InstalledAt: time.Now()},
			}},
		},
	}
	codex := s.InstallationsByTarget("codex")
	if len(codex) != 2 {
		t.Errorf("expected 2 codex, got %d", len(codex))
	}
	claude := s.InstallationsByTarget("claude")
	if len(claude) != 2 {
		t.Errorf("expected 2 claude, got %d", len(claude))
	}
}

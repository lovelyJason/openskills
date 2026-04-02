package codexmgr

import (
	"encoding/json"
	"os"
	"sort"
)

type RegistryEntry struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"`
	Source  string `json:"source"`
	RepoDir string `json:"repoDir"`
}

type Registry struct {
	Version      int             `json:"version"`
	Marketplaces []RegistryEntry `json:"marketplaces"`
}

func loadRegistry() (*Registry, error) {
	data, err := os.ReadFile(registryFile())
	if err != nil {
		if os.IsNotExist(err) {
			return &Registry{Version: 1}, nil
		}
		return nil, err
	}
	var r Registry
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if r.Marketplaces == nil {
		r.Marketplaces = []RegistryEntry{}
	}
	return &r, nil
}

func saveRegistry(r *Registry) error {
	if err := os.MkdirAll(stateDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := registryFile() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, registryFile())
}

func (r *Registry) register(name, source, repoDir, kind string) {
	filtered := make([]RegistryEntry, 0, len(r.Marketplaces))
	for _, e := range r.Marketplaces {
		if e.Name != name {
			filtered = append(filtered, e)
		}
	}
	filtered = append(filtered, RegistryEntry{
		Name:    name,
		Kind:    kind,
		Source:  source,
		RepoDir: repoDir,
	})
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Name < filtered[j].Name
	})
	r.Marketplaces = filtered
}

func (r *Registry) remove(name string) {
	filtered := make([]RegistryEntry, 0, len(r.Marketplaces))
	for _, e := range r.Marketplaces {
		if e.Name != name {
			filtered = append(filtered, e)
		}
	}
	r.Marketplaces = filtered
}

func (r *Registry) find(name string) *RegistryEntry {
	for i := range r.Marketplaces {
		if r.Marketplaces[i].Name == name {
			return &r.Marketplaces[i]
		}
	}
	return nil
}

package codexmgr

import (
	"encoding/json"
	"os"
)

type PluginDescriptor struct {
	Marketplace string `json:"marketplace"`
	Name        string `json:"name"`
	LocalName   string `json:"localName"`
	DisplayName string `json:"displayName"`
	Category    string `json:"category"`
	PluginDir   string `json:"pluginDir"`
	PreparedDir string `json:"preparedDir,omitempty"`
}

type PluginState struct {
	Version int                `json:"version"`
	Plugins []PluginDescriptor `json:"plugins"`
}

func loadPluginState() (*PluginState, error) {
	data, err := os.ReadFile(pluginStateFile())
	if err != nil {
		if os.IsNotExist(err) {
			return &PluginState{Version: 1}, nil
		}
		return nil, err
	}
	var ps PluginState
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil, err
	}
	return &ps, nil
}

func savePluginState(ps *PluginState) error {
	if err := os.MkdirAll(stateDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := pluginStateFile() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, pluginStateFile())
}

func (ps *PluginState) localNames() map[string]struct{} {
	m := make(map[string]struct{}, len(ps.Plugins))
	for _, p := range ps.Plugins {
		m[p.LocalName] = struct{}{}
	}
	return m
}

func (ps *PluginState) findByLocalName(localName string) *PluginDescriptor {
	for i := range ps.Plugins {
		if ps.Plugins[i].LocalName == localName {
			return &ps.Plugins[i]
		}
	}
	return nil
}

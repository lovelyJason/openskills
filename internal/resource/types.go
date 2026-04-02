package resource

import "time"

type Type string

const (
	TypePlugin Type = "plugin"
	TypeSkill  Type = "skill"
)

type InstallMode string

const (
	ModeSymlink InstallMode = "symlink"
	ModeNative  InstallMode = "native"
)

type Resource struct {
	Name        string `json:"name"`
	Type        Type   `json:"type"`
	Marketplace string `json:"marketplace"`
	Version     string `json:"version,omitempty"`
	LocalPath   string `json:"localPath,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
}

func (r *Resource) QualifiedName() string {
	return r.Name + "@" + r.Marketplace
}

type InstalledResource struct {
	Resource
	Mode         InstallMode       `json:"mode"`
	GitCommitSHA string            `json:"gitCommitSha,omitempty"`
	Targets      map[string]Target `json:"targets"`
}

type Target struct {
	InstalledAt   time.Time `json:"installedAt"`
	Paths         []string  `json:"paths,omitempty"`
	ConfigEntries []string  `json:"configEntries,omitempty"`
	NativeRef     string    `json:"nativeRef,omitempty"`
}

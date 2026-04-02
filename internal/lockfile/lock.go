package lockfile

import (
	"encoding/json"
	"os"

	"github.com/lovelyJason/openskills/internal/paths"
	"github.com/lovelyJason/openskills/internal/resource"
)

type LockedResource struct {
	QualifiedName string        `json:"qualifiedName"`
	Type          resource.Type `json:"type"`
	Version       string        `json:"version"`
	GitCommitSHA  string        `json:"gitCommitSha"`
}

type LockData struct {
	Version   int              `json:"version"`
	Resources []LockedResource `json:"resources"`
}

func Load() (*LockData, error) {
	data, err := os.ReadFile(paths.LockFile())
	if err != nil {
		if os.IsNotExist(err) {
			return &LockData{Version: 1}, nil
		}
		return nil, err
	}
	var ld LockData
	if err := json.Unmarshal(data, &ld); err != nil {
		return nil, err
	}
	return &ld, nil
}

func (ld *LockData) Save() error {
	if err := paths.EnsureDir(paths.ConfigDir()); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ld, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(paths.LockFile(), data, 0644)
}

func (ld *LockData) Find(qualifiedName string) *LockedResource {
	for i := range ld.Resources {
		if ld.Resources[i].QualifiedName == qualifiedName {
			return &ld.Resources[i]
		}
	}
	return nil
}

func (ld *LockData) Upsert(lr LockedResource) {
	for i := range ld.Resources {
		if ld.Resources[i].QualifiedName == lr.QualifiedName {
			ld.Resources[i] = lr
			return
		}
	}
	ld.Resources = append(ld.Resources, lr)
}

func (ld *LockData) Remove(qualifiedName string) {
	filtered := ld.Resources[:0]
	for _, r := range ld.Resources {
		if r.QualifiedName != qualifiedName {
			filtered = append(filtered, r)
		}
	}
	ld.Resources = filtered
}

package target

import (
	"context"

	"github.com/lovelyJason/openskills/internal/resource"
)

type Adapter interface {
	Name() string
	Detect() bool
	SupportedResources() []resource.Type
	Install(ctx context.Context, res *resource.Resource, mode resource.InstallMode, sourcePath string) (*InstallResult, error)
	Uninstall(ctx context.Context, res *resource.Resource) error
	IsInstalled(res *resource.Resource) (bool, error)
}

type InstallResult struct {
	Paths         []string
	ConfigEntries []string
	NativeRef     string
}

type MarketplaceHook interface {
	OnMarketplaceAdd(ctx context.Context, url, name, repoDir string) error
	OnMarketplaceRemove(ctx context.Context, name string) error
	OnMarketplaceUpdate(ctx context.Context, name, repoDir string) error
}

type VersionChecker interface {
	CheckVersion() error
}

type Registry struct {
	adapters []Adapter
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(a Adapter) {
	r.adapters = append(r.adapters, a)
}

func (r *Registry) All() []Adapter {
	return r.adapters
}

func (r *Registry) Available() []Adapter {
	var result []Adapter
	for _, a := range r.adapters {
		if a.Detect() {
			result = append(result, a)
		}
	}
	return result
}

func (r *Registry) Get(name string) Adapter {
	for _, a := range r.adapters {
		if a.Name() == name {
			return a
		}
	}
	return nil
}

func (r *Registry) AvailableNames() []string {
	var names []string
	for _, a := range r.Available() {
		names = append(names, a.Name())
	}
	return names
}

func (r *Registry) SupportsMarketplace(name string) bool {
	a := r.Get(name)
	if a == nil {
		return false
	}
	_, ok := a.(MarketplaceHook)
	return ok
}

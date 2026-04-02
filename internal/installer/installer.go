package installer

import (
	"context"
	"fmt"
	"time"

	"github.com/lovelyJason/openskills/internal/backup"
	"github.com/lovelyJason/openskills/internal/lockfile"
	"github.com/lovelyJason/openskills/internal/resource"
	"github.com/lovelyJason/openskills/internal/state"
	"github.com/lovelyJason/openskills/internal/target"
)

type Installer struct {
	stateMgr   *state.Manager
	registry   *target.Registry
	maxBackups int
}

func New(stateMgr *state.Manager, registry *target.Registry, maxBackups int) *Installer {
	return &Installer{
		stateMgr:   stateMgr,
		registry:   registry,
		maxBackups: maxBackups,
	}
}

type InstallRequest struct {
	Resource    resource.Resource
	Mode        resource.InstallMode
	TargetNames []string
	SourcePath  string
	CommitSHA   string
}

func (inst *Installer) Install(ctx context.Context, req InstallRequest) (err error) {
	tx, txErr := backup.Begin(fmt.Sprintf("install %s", req.Resource.QualifiedName()))
	if txErr != nil {
		return fmt.Errorf("backup begin: %w", txErr)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			tx.Cleanup()
		} else {
			tx.Commit()
			backup.PruneBackups(inst.maxBackups)
		}
	}()

	st, err := inst.stateMgr.Load()
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	installation := state.Installation{
		ID:           req.Resource.QualifiedName(),
		ResourceType: req.Resource.Type,
		Name:         req.Resource.Name,
		Marketplace:  req.Resource.Marketplace,
		Version:      req.Resource.Version,
		GitCommitSHA: req.CommitSHA,
		Mode:         req.Mode,
		Targets:      make(map[string]state.InstallationTarget),
	}

	for _, tName := range req.TargetNames {
		adapter := inst.registry.Get(tName)
		if adapter == nil {
			err = fmt.Errorf("unknown target: %s", tName)
			return err
		}
		if !adapter.Detect() {
			err = fmt.Errorf("target %s is not available on this system", tName)
			return err
		}

		supported := false
		for _, rt := range adapter.SupportedResources() {
			if rt == req.Resource.Type {
				supported = true
				break
			}
		}
		if !supported {
			err = fmt.Errorf("target %s does not support resource type %s", tName, req.Resource.Type)
			return err
		}

		var result *target.InstallResult
		result, err = adapter.Install(ctx, &req.Resource, req.Mode, req.SourcePath)
		if err != nil {
			err = fmt.Errorf("install to %s: %w", tName, err)
			return err
		}

		for _, p := range result.Paths {
			if bErr := tx.BackupFile(p); bErr != nil {
				err = fmt.Errorf("backup file %s: %w", p, bErr)
				return err
			}
		}

		installation.Targets[tName] = state.InstallationTarget{
			InstalledAt:   time.Now(),
			Paths:         result.Paths,
			ConfigEntries: result.ConfigEntries,
			NativeRef:     result.NativeRef,
		}
	}

	st.UpsertInstallation(installation)
	err = inst.stateMgr.Save(st)
	if err != nil {
		err = fmt.Errorf("save state: %w", err)
		return err
	}

	var lock *lockfile.LockData
	lock, err = lockfile.Load()
	if err != nil {
		err = fmt.Errorf("load lockfile: %w", err)
		return err
	}
	lock.Upsert(lockfile.LockedResource{
		QualifiedName: req.Resource.QualifiedName(),
		Type:          req.Resource.Type,
		Version:       req.Resource.Version,
		GitCommitSHA:  req.CommitSHA,
	})
	err = lock.Save()
	if err != nil {
		err = fmt.Errorf("save lockfile: %w", err)
		return err
	}

	return nil
}

func (inst *Installer) Uninstall(ctx context.Context, qualifiedName string, targetNames []string) (err error) {
	tx, txErr := backup.Begin(fmt.Sprintf("uninstall %s", qualifiedName))
	if txErr != nil {
		return fmt.Errorf("backup begin: %w", txErr)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			tx.Cleanup()
		} else {
			tx.Commit()
			backup.PruneBackups(inst.maxBackups)
		}
	}()

	var st *state.Store
	st, err = inst.stateMgr.Load()
	if err != nil {
		err = fmt.Errorf("load state: %w", err)
		return err
	}

	installation := st.FindInstallation(qualifiedName)
	if installation == nil {
		err = fmt.Errorf("installation %q not found", qualifiedName)
		return err
	}

	removeTargets := targetNames
	if len(removeTargets) == 0 {
		for t := range installation.Targets {
			removeTargets = append(removeTargets, t)
		}
	}

	for _, tName := range removeTargets {
		tgt, ok := installation.Targets[tName]
		if !ok {
			continue
		}

		for _, p := range tgt.Paths {
			if bErr := tx.BackupFile(p); bErr != nil {
				err = fmt.Errorf("backup file %s: %w", p, bErr)
				return err
			}
		}

		adapter := inst.registry.Get(tName)
		if adapter == nil {
			err = fmt.Errorf("unknown target %s during uninstall", tName)
			return err
		}

		res := &resource.Resource{
			Name:        installation.Name,
			Type:        installation.ResourceType,
			Marketplace: installation.Marketplace,
		}
		err = adapter.Uninstall(ctx, res)
		if err != nil {
			err = fmt.Errorf("uninstall from %s: %w", tName, err)
			return err
		}

		delete(installation.Targets, tName)
	}

	if len(installation.Targets) == 0 {
		st.RemoveInstallation(qualifiedName)
		var lock *lockfile.LockData
		lock, err = lockfile.Load()
		if err != nil {
			err = fmt.Errorf("load lockfile: %w", err)
			return err
		}
		lock.Remove(qualifiedName)
		err = lock.Save()
		if err != nil {
			err = fmt.Errorf("save lockfile: %w", err)
			return err
		}
	} else {
		st.UpsertInstallation(*installation)
	}

	err = inst.stateMgr.Save(st)
	if err != nil {
		err = fmt.Errorf("save state: %w", err)
	}
	return err
}

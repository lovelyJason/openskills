package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lovelyJason/openskills/internal/config"
	"github.com/lovelyJason/openskills/internal/installer"
	"github.com/lovelyJason/openskills/internal/state"
	"github.com/lovelyJason/openskills/internal/target"
)

var Version = "dev"

type App struct {
	cfg      *config.Config
	stateMgr *state.Manager
	registry *target.Registry
	inst     *installer.Installer
}

func NewApp() *App {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	sm := state.NewManager()
	reg := target.NewRegistry()
	reg.Register(target.NewCodex())
	reg.Register(target.NewClaude())
	reg.Register(target.NewCursor())

	return &App{
		cfg:      cfg,
		stateMgr: sm,
		registry: reg,
		inst:     installer.New(sm, reg, cfg.MaxBackups),
	}
}

func Execute() {
	app := NewApp()
	root := &cobra.Command{
		Use:   "openskills",
		Short: "AI editor extension manager",
		Long: `openskills — manage plugins, skills, and marketplaces across AI editors.

Supports Codex, Claude, Cursor, and more. Install resources via symlink
or native mode, with version pinning and automatic rollback.`,
		SilenceUsage: true,
		Version:      Version,
	}

	root.AddCommand(
		app.newMarketplaceCmd(),
		app.newPluginCmd(),
		app.newSkillCmd(),
		app.newCodexCmd(),
		app.newListCmd(),
		app.newStatusCmd(),
		app.newUpdateCmd(),
		app.newDoctorCmd(),
		newCompletionCmd(),
	)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func (a *App) resolveTargets(flagTargets []string) ([]string, error) {
	available := a.registry.AvailableNames()
	if len(available) == 0 {
		return nil, fmt.Errorf("no AI editors detected on this system")
	}
	if len(flagTargets) > 0 {
		for _, t := range flagTargets {
			if a.registry.Get(t) == nil {
				return nil, fmt.Errorf("unknown target: %s (available: %v)", t, available)
			}
		}
		return flagTargets, nil
	}
	if len(a.cfg.DefaultTargets) > 0 {
		return a.cfg.DefaultTargets, nil
	}
	return nil, nil
}

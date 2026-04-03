package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lovelyJason/openskills/internal/codexmgr"
	"github.com/lovelyJason/openskills/internal/target"
	"github.com/lovelyJason/openskills/internal/ui"
)

func (a *App) newCodexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "codex",
		Short: "Codex-specific management commands",
	}

	cmd.AddCommand(
		a.codexSyncCmd(),
		a.codexCleanupCmd(),
		a.codexBuiltinListCmd(),
		a.codexInstalledListCmd(),
		a.codexVersionCmd(),
	)

	return cmd
}

func (a *App) requireCodex() (*target.Codex, error) {
	adapter := a.registry.Get("codex")
	if adapter == nil {
		return nil, fmt.Errorf("codex adapter not registered")
	}
	codex, ok := adapter.(*target.Codex)
	if !ok {
		return nil, fmt.Errorf("codex adapter type mismatch")
	}
	if !codex.Detect() {
		return nil, fmt.Errorf("codex not detected on this system (missing ~/.codex)")
	}
	return codex, nil
}

func (a *App) codexSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Re-sync the Codex aggregated marketplace",
		RunE: func(cmd *cobra.Command, args []string) error {
			codex, err := a.requireCodex()
			if err != nil {
				return err
			}

			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			var repos []codexmgr.RepoEntry
			for _, m := range st.Marketplaces {
				repos = append(repos, codexmgr.RepoEntry{Name: m.Name, RepoDir: m.LocalPath})
			}

			ui.Info("Syncing Codex marketplace...")
			if err := codex.Manager().SyncAll(repos); err != nil {
				return err
			}
			ui.Success("Codex marketplace synced")
			return nil
		},
	}
}

func (a *App) codexCleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Remove all Codex-managed state (registry, plugin state, caches)",
		RunE: func(cmd *cobra.Command, args []string) error {
			codex, err := a.requireCodex()
			if err != nil {
				return err
			}

			confirmed, err := ui.Confirm("This will remove all Codex-managed plugin data. Continue?")
			if err != nil || !confirmed {
				return fmt.Errorf("aborted")
			}

			if err := codex.Manager().CleanupAll(); err != nil {
				return err
			}
			ui.Success("Codex managed state cleaned up")
			return nil
		},
	}
}

func (a *App) codexBuiltinListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "builtin-list",
		Short: "List Codex built-in plugins from official marketplace",
		RunE: func(cmd *cobra.Command, args []string) error {
			codex, err := a.requireCodex()
			if err != nil {
				return err
			}

			plugins, err := codex.Manager().BuiltinPluginList()
			if err != nil {
				ui.Warn("Cannot read official marketplace: %v", err)
				return nil
			}

			if len(plugins) == 0 {
				ui.Info("No built-in plugins found.")
				return nil
			}

			ui.Header("Codex Built-in Plugins")
			for _, p := range plugins {
				name, _ := p["name"].(string)
				category, _ := p["category"].(string)
				if category == "" {
					category = "-"
				}
				fmt.Printf("  %-40s [%s]\n", name, category)
			}
			return nil
		},
	}
}

func (a *App) codexInstalledListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "installed-list",
		Short: "List plugins installed in Codex config.toml",
		RunE: func(cmd *cobra.Command, args []string) error {
			codex, err := a.requireCodex()
			if err != nil {
				return err
			}

			installed, err := codex.Manager().InstalledPluginList()
			if err != nil {
				return err
			}

			if len(installed) == 0 {
				ui.Info("No plugins found in Codex config.toml")
				return nil
			}

			ui.Header("Codex Installed Plugins (config.toml)")
			for name, enabled := range installed {
				status := "\033[31mdisabled\033[0m"
				if enabled {
					status = "\033[32menabled\033[0m"
				}
				fmt.Printf("  %-40s %s\n", name, status)
			}
			return nil
		},
	}
}

func (a *App) codexVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show detected Codex CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ver, err := codexmgr.DetectedVersion()
			if err != nil {
				return fmt.Errorf("cannot detect codex version: %w", err)
			}
			fmt.Printf("  Codex CLI version: %s\n", ver)
			fmt.Printf("  Minimum required:  %s\n", codexmgr.MinVersion)

			if err := codexmgr.CheckVersion(); err != nil {
				ui.Warn("%v", err)
			} else {
				ui.Success("Version OK")
			}
			return nil
		},
	}
}

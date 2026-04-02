package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lovelyJason/openskills/internal/gitutil"
	"github.com/lovelyJason/openskills/internal/marketplace"
	"github.com/lovelyJason/openskills/internal/scanner"
	"github.com/lovelyJason/openskills/internal/state"
	"github.com/lovelyJason/openskills/internal/target"
	"github.com/lovelyJason/openskills/internal/ui"
)

func (a *App) newMarketplaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "marketplace",
		Aliases: []string{"mp"},
		Short:   "Manage marketplace repositories",
	}

	cmd.AddCommand(
		a.marketplaceAddCmd(),
		a.marketplaceListCmd(),
		a.marketplaceUpdateCmd(),
		a.marketplaceRemoveCmd(),
		a.marketplacePinCmd(),
		a.marketplaceUnpinCmd(),
	)

	return cmd
}

func (a *App) marketplaceAddCmd() *cobra.Command {
	var name string
	var targets []string

	cmd := &cobra.Command{
		Use:   "add <git-url>",
		Short: "Add a marketplace repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			spin := ui.NewSpinner(fmt.Sprintf("Cloning %s ...", url))
			entry, err := marketplace.Add(st, url, name)
			if err != nil {
				spin.Stop()
				return err
			}
			spin.Stop()

			hasCodex := scanner.HasCodexPlugins(entry.LocalPath)
			hasClaude := scanner.HasClaudePlugin(entry.LocalPath)

			compat := map[string]bool{
				"codex":  hasCodex,
				"claude": hasClaude,
				"cursor": false,
			}

			targetNames, err := a.resolveTargets(targets)
			if err != nil {
				return err
			}

			if len(targetNames) > 0 {
				var filtered []string
				for _, t := range targetNames {
					if !a.registry.SupportsMarketplace(t) {
						ui.Dim("  %s — marketplace not supported, skipped", t)
						continue
					}
					if compatible, known := compat[t]; known && !compatible {
						ui.Dim("  %s — this source has no %s-compatible plugins, skipped", t, t)
						continue
					}
					filtered = append(filtered, t)
				}
				targetNames = filtered
			} else {
				var options []ui.SelectOption
				for _, tName := range a.registry.AvailableNames() {
					if !a.registry.SupportsMarketplace(tName) {
						options = append(options, ui.SelectOption{
							Label: tName, Value: tName,
							Disabled: true, DisabledMsg: "marketplace not supported",
						})
						continue
					}
					if compatible, known := compat[tName]; known && !compatible {
						options = append(options, ui.SelectOption{
							Label: tName, Value: tName,
							Disabled: true, DisabledMsg: fmt.Sprintf("this source has no %s-compatible plugins", tName),
						})
						continue
					}
					options = append(options, ui.SelectOption{Label: tName, Value: tName})
				}
				targetNames, err = ui.MultiSelectEx("Select target editors:", options)
				if err != nil {
					return err
				}
			}

			if len(targetNames) == 0 {
				ui.Warn("No compatible targets selected")
				ui.Info("Marketplace cloned to: %s", entry.LocalPath)
				if err := a.stateMgr.Save(st); err != nil {
					return err
				}
				return nil
			}

			if err := a.stateMgr.Save(st); err != nil {
				return err
			}

			plugins, _ := scanner.ScanPlugins(entry.LocalPath, entry.Name)
			claudePlugins, _ := scanner.ScanClaudePlugin(entry.LocalPath, entry.Name)
			totalPlugins := len(plugins) + len(claudePlugins)

			selectedTargets := make(map[string]bool, len(targetNames))
			for _, t := range targetNames {
				selectedTargets[t] = true
			}

			spin = ui.NewSpinner("Registering marketplace ...")
			a.fireMarketplaceHooks(func(hook target.MarketplaceHook, adapterName string) {
				if !selectedTargets[adapterName] {
					return
				}
				spin.Update(fmt.Sprintf("[%s] Registering marketplace ...", adapterName))
				if err := hook.OnMarketplaceAdd(context.Background(), url, entry.Name, entry.LocalPath); err != nil {
					ui.Warn("[%s] marketplace hook: %v", adapterName, err)
				}
			})
			spin.Stop()

			ui.Success("Marketplace added: %s", entry.Name)
			ui.Info("Targets: %v", targetNames)
			ui.Info("Path: %s", entry.LocalPath)
			ui.Info("Plugins: %d", totalPlugins)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Custom marketplace name")
	cmd.Flags().StringSliceVarP(&targets, "target", "t", nil, "Target editors (codex,claude,cursor)")
	return cmd
}

func (a *App) marketplaceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List registered marketplaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			if len(st.Marketplaces) == 0 {
				ui.Info("No marketplaces registered.")
				ui.Info("Add one with: openskills marketplace add <git-url>")
				return nil
			}

			ui.Header("Registered Sources")
			for _, m := range st.Marketplaces {
				pinned := ""
				if m.PinnedVer != "" {
					pinned = fmt.Sprintf(" (pinned: %s)", m.PinnedVer)
				}

				sourceLabel := "marketplace"
				if m.Source == state.SourceSkillRepo {
					sourceLabel = "skill repo"
				}

				plugins, _ := scanner.ScanPlugins(m.LocalPath, m.Name)
				skills, _ := scanner.ScanSkills(m.LocalPath, m.Name)

				fmt.Printf("  \033[1;33m%-25s\033[0m %s [%s]%s\n", m.Name, m.URL, sourceLabel, pinned)
				fmt.Printf("    %d plugins, %d skills | updated %s\n",
					len(plugins), len(skills), m.LastUpdated.Format("2006-01-02 15:04"))
			}
			return nil
		},
	}
}

func (a *App) marketplaceUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update [name...]",
		Short: "Update marketplace repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			targets := st.Marketplaces
			if len(args) > 0 {
				targets = nil
				for _, name := range args {
					m := st.FindMarketplace(name)
					if m == nil {
						ui.Warn("Unknown marketplace: %s", name)
						continue
					}
					targets = append(targets, *m)
				}
			}

			for _, m := range targets {
				if m.PinnedVer != "" {
					ui.Warn("%s is pinned to %s, skipping", m.Name, m.PinnedVer)
					continue
				}
				spin := ui.NewSpinner(fmt.Sprintf("Updating %s ...", m.Name))
				if err := marketplace.Update(&m); err != nil {
					spin.Stop()
					ui.Error("Failed to update %s: %v", m.Name, err)
					continue
				}
				st.UpsertMarketplace(m)

				a.fireMarketplaceHooks(func(hook target.MarketplaceHook, adapterName string) {
					spin.Update(fmt.Sprintf("[%s] Syncing %s ...", adapterName, m.Name))
					if err := hook.OnMarketplaceUpdate(context.Background(), m.Name, m.LocalPath); err != nil {
						ui.Warn("[%s] update hook: %v", adapterName, err)
					}
				})
				spin.Stop()

				ui.Success("Updated %s", m.Name)
			}

			return a.stateMgr.Save(st)
		},
	}
}

func (a *App) marketplaceRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a marketplace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			entry := st.FindMarketplace(name)
			if entry == nil {
				return fmt.Errorf("marketplace %q not found", name)
			}

			installations := st.InstallationsByMarketplace(name)
			if len(installations) > 0 {
				confirmed, err := ui.Confirm(fmt.Sprintf(
					"%d resources installed from %s will be orphaned. Continue?",
					len(installations), name))
				if err != nil || !confirmed {
					return fmt.Errorf("aborted")
				}
			}

			a.fireMarketplaceHooks(func(hook target.MarketplaceHook, adapterName string) {
				if err := hook.OnMarketplaceRemove(context.Background(), name); err != nil {
					ui.Warn("[%s] remove hook: %v", adapterName, err)
				}
			})

			os.RemoveAll(entry.LocalPath)
			st.RemoveMarketplace(name)
			ui.Success("Removed marketplace: %s", name)
			return a.stateMgr.Save(st)
		},
	}
}

func (a *App) marketplacePinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pin <name> <version>",
		Short: "Pin a marketplace to a specific version/tag/commit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, version := args[0], args[1]
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			entry := st.FindMarketplace(name)
			if entry == nil {
				return fmt.Errorf("marketplace %q not found", name)
			}

			if err := marketplace.Pin(entry, version); err != nil {
				return err
			}

			sha, _ := gitutil.CurrentCommitSHA(entry.LocalPath)
			st.UpsertMarketplace(*entry)
			shortSHA := sha
			if len(shortSHA) > 8 {
				shortSHA = shortSHA[:8]
			}
			ui.Success("Pinned %s to %s (commit: %s)", name, version, shortSHA)
			return a.stateMgr.Save(st)
		},
	}
}

func (a *App) fireMarketplaceHooks(fn func(hook target.MarketplaceHook, adapterName string)) {
	for _, adapter := range a.registry.Available() {
		if hook, ok := adapter.(target.MarketplaceHook); ok {
			fn(hook, adapter.Name())
		}
	}
}

func (a *App) marketplaceUnpinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unpin <name>",
		Short: "Unpin a marketplace, allowing updates to latest",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			entry := st.FindMarketplace(name)
			if entry == nil {
				return fmt.Errorf("marketplace %q not found", name)
			}

			if err := marketplace.Unpin(entry); err != nil {
				return err
			}

			st.UpsertMarketplace(*entry)
			ui.Success("Unpinned %s, now tracking branch: %s", name, entry.Branch)
			return a.stateMgr.Save(st)
		},
	}
}

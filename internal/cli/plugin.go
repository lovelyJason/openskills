package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/lovelyJason/openskills/internal/gitutil"
	"github.com/lovelyJason/openskills/internal/installer"
	"github.com/lovelyJason/openskills/internal/marketplace"
	"github.com/lovelyJason/openskills/internal/resolver"
	"github.com/lovelyJason/openskills/internal/resource"
	"github.com/lovelyJason/openskills/internal/target"
	"github.com/lovelyJason/openskills/internal/ui"
)

func (a *App) newPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plugin",
		Aliases: []string{"plugins"},
		Short:   "Manage plugins",
	}

	cmd.AddCommand(
		a.pluginListCmd(),
		a.pluginInstallCmd(),
		a.pluginUninstallCmd(),
		a.pluginStatusCmd(),
	)

	return cmd
}

func (a *App) pluginListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [marketplace]",
		Short: "List available plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			if len(st.Marketplaces) == 0 {
				ui.Info("No marketplaces registered. Run: openskills marketplace add <git-url>")
				return nil
			}

			ui.Header("Available Plugins")
			for _, m := range st.Marketplaces {
				if len(args) > 0 && m.Name != args[0] {
					continue
				}
				plugins, err := marketplace.ListResources(&m, resource.TypePlugin)
				if err != nil {
					ui.Warn("Error scanning %s: %v", m.Name, err)
					continue
				}
				if len(plugins) == 0 {
					continue
				}
				fmt.Printf("\n  \033[1m%s\033[0m\n", m.Name)
				for _, p := range plugins {
					installed := ""
					inst := st.FindInstallation(p.QualifiedName())
					if inst != nil {
						var targets []string
						for t := range inst.Targets {
							targets = append(targets, t)
						}
						installed = fmt.Sprintf(" \033[32m[installed: %v]\033[0m", targets)
					}
					fmt.Printf("    %-30s %s [%s]%s\n", p.Name, p.Description, p.Category, installed)
				}
			}
			return nil
		},
	}
}

func (a *App) pluginInstallCmd() *cobra.Command {
	var targets []string
	var mode string

	cmd := &cobra.Command{
		Use:   "install <plugin[@version]> [plugin...]",
		Short: "Install plugins to target editors",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			allResources, err := marketplace.ListAllResources(st)
			if err != nil {
				return err
			}

			var plugins []resource.Resource
			for _, arg := range args {
				filtered := filterByType(allResources, resource.TypePlugin)
				res, err := resolver.Resolve(arg, filtered)
				if err != nil {
					if ambErr, ok := err.(*resolver.AmbiguousError); ok {
						ui.Warn("Ambiguous plugin %q found in multiple marketplaces:", arg)
						for _, m := range ambErr.Matches {
							fmt.Printf("    - %s\n", m.QualifiedName())
						}
						ui.Info("Please specify: openskills plugin install %s@<marketplace>", arg)
						return nil
					}
					return err
				}
				plugins = append(plugins, *res)
			}

			targetNames, err := a.resolveTargets(targets)
			if err != nil {
				return err
			}
			if len(targetNames) == 0 {
				available := a.registry.AvailableNames()
				targetNames, err = ui.MultiSelect("Select target editors:", available)
				if err != nil {
					return err
				}
			}
			if len(targetNames) == 0 {
				return fmt.Errorf("no targets selected")
			}

		installMode := resource.InstallMode(mode)
		if installMode == "" {
			selected, err := ui.SelectInstallMode()
			if err != nil {
				return err
			}
			installMode = resource.InstallMode(selected)
		}

		for _, plugin := range plugins {
			mp := st.FindMarketplace(plugin.Marketplace)
			if mp == nil {
				ui.Error("Marketplace %s not found for plugin %s", plugin.Marketplace, plugin.Name)
				continue
			}

				sha, _ := gitutil.CurrentCommitSHA(mp.LocalPath)

				var succeeded []string
				for _, tName := range targetNames {
					adapter := a.registry.Get(tName)
					if adapter == nil {
						ui.Error("[%s] unknown target", tName)
						continue
					}
					if vc, ok := adapter.(target.VersionChecker); ok {
						if err := vc.CheckVersion(); err != nil {
							ui.Error("[%s] %v", tName, err)
							continue
						}
					}

					ui.Info("Installing %s to %s ...", plugin.QualifiedName(), tName)
					err := a.inst.Install(context.Background(), installer.InstallRequest{
						Resource:    plugin,
						Mode:        installMode,
						TargetNames: []string{tName},
						SourcePath:  plugin.LocalPath,
						CommitSHA:   sha,
					})
					if err != nil {
						ui.Error("[%s] %v", tName, err)
						continue
					}
					succeeded = append(succeeded, tName)
					ui.Success("Installed %s (%s) to %s", plugin.QualifiedName(), installMode, tName)
				}
				if len(succeeded) == 0 {
					ui.Error("Failed to install %s to any target", plugin.QualifiedName())
				}
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&targets, "target", "t", nil, "Target editors (codex,claude,cursor)")
	cmd.Flags().StringVarP(&mode, "mode", "m", "", "Override install mode (symlink/native)")
	return cmd
}

func (a *App) pluginUninstallCmd() *cobra.Command {
	var targets []string

	cmd := &cobra.Command{
		Use:   "uninstall <plugin> [plugin...]",
		Short: "Uninstall plugins from target editors",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			for _, arg := range args {
				st, err := a.stateMgr.Load()
				if err != nil {
					return err
				}
				allResources, _ := marketplace.ListAllResources(st)
				filtered := filterByType(allResources, resource.TypePlugin)
				res, err := resolver.Resolve(arg, filtered)
				if err != nil {
					ui.Error("Cannot resolve %s: %v", arg, err)
					continue
				}

				qn := res.QualifiedName()
				err = a.inst.Uninstall(ctx, qn, targets)
				if err != nil {
					ui.Error("Failed to uninstall %s: %v", qn, err)
					continue
				}
				ui.Success("Uninstalled %s", qn)
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&targets, "target", "t", nil, "Target editors (codex,claude,cursor)")
	return cmd
}

func (a *App) pluginStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [plugin...]",
		Short: "Show plugin installation status",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			if len(args) == 0 {
				ui.Header("Installed Plugins")
				for _, inst := range st.Installations {
					if inst.ResourceType != resource.TypePlugin {
						continue
					}
					fmt.Printf("  \033[1;33m%-35s\033[0m v%s [%s]\n",
						inst.ID, inst.Version, inst.Mode)
					for tName, t := range inst.Targets {
						fmt.Printf("    → %s (installed %s)\n", tName, t.InstalledAt.Format("2006-01-02"))
					}
				}
				return nil
			}

			allResources, _ := marketplace.ListAllResources(st)
			filtered := filterByType(allResources, resource.TypePlugin)

			for _, arg := range args {
				res, err := resolver.Resolve(arg, filtered)
				if err != nil {
					ui.Error("%v", err)
					continue
				}
				inst := st.FindInstallation(res.QualifiedName())
				if inst == nil {
					fmt.Printf("  %s: not installed\n", res.QualifiedName())
					continue
				}
				fmt.Printf("  \033[1;33m%s\033[0m\n", inst.ID)
				sha := inst.GitCommitSHA
				if len(sha) > 8 {
					sha = sha[:8]
				}
				fmt.Printf("    Version: %s (commit: %s)\n", inst.Version, sha)
				fmt.Printf("    Mode: %s\n", inst.Mode)
				for tName, t := range inst.Targets {
					fmt.Printf("    → %s: installed %s, paths: %v\n", tName, t.InstalledAt.Format("2006-01-02"), t.Paths)
				}
			}
			return nil
		},
	}
}

func filterByType(resources []resource.Resource, t resource.Type) []resource.Resource {
	var result []resource.Resource
	for _, r := range resources {
		if r.Type == t {
			result = append(result, r)
		}
	}
	return result
}

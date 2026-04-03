package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lovelyJason/openskills/internal/gitutil"
	"github.com/lovelyJason/openskills/internal/installer"
	"github.com/lovelyJason/openskills/internal/marketplace"
	"github.com/lovelyJason/openskills/internal/paths"
	"github.com/lovelyJason/openskills/internal/resolver"
	"github.com/lovelyJason/openskills/internal/resource"
	"github.com/lovelyJason/openskills/internal/scanner"
	"github.com/lovelyJason/openskills/internal/state"
	"github.com/lovelyJason/openskills/internal/ui"
)

func (a *App) newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "skill",
		Aliases: []string{"skills"},
		Short:   "Manage skills",
	}

	cmd.AddCommand(
		a.skillAddCmd(),
		a.skillListCmd(),
		a.skillInstallCmd(),
		a.skillUninstallCmd(),
		a.skillRemoveSourceCmd(),
	)

	return cmd
}

func (a *App) skillAddCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "add <git-url>",
		Short: "Add a skill repository (any Git repo with a skills/ directory)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := marketplace.ExpandURL(args[0])

			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			spin := ui.NewSpinner(fmt.Sprintf("Cloning %s ...", url))
			entry, err := marketplace.AddSkillRepo(st, url, name)
			if err != nil {
				spin.Stop()
				return err
			}

			if err := a.stateMgr.Save(st); err != nil {
				spin.Stop()
				return err
			}
			spin.Stop()

			skills, _ := scanner.ScanSkills(entry.LocalPath, entry.Name)
			ui.Success("Skill repo added: %s", entry.Name)
			ui.Info("Path: %s", entry.LocalPath)
			ui.Info("Skills found: %d", len(skills))
			for _, s := range skills {
				fmt.Printf("    %-30s %s\n", s.Name, s.Description)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Custom name for the skill repo")
	return cmd
}

func (a *App) skillRemoveSourceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-source <name>",
		Short: "Remove a skill repository source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			entry := st.FindMarketplace(name)
			if entry == nil {
				return fmt.Errorf("skill repo %q not found", name)
			}
			if !entry.Sources.Has(state.SourceSkillRepo) {
				return fmt.Errorf("%q is a marketplace, use 'openskills marketplace remove' instead", name)
			}

			installed := st.InstallationsByMarketplace(name)
			if len(installed) > 0 {
				ui.Warn("%d skill(s) still installed from %s, uninstall them first", len(installed), name)
				return fmt.Errorf("cannot remove source with active installations")
			}

			os.RemoveAll(entry.LocalPath)
			st.RemoveMarketplace(name)
			if err := a.stateMgr.Save(st); err != nil {
				return err
			}

			ui.Success("Removed skill repo: %s", name)
			return nil
		},
	}
}

func (a *App) skillListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [source]",
		Short: "List available skills from all sources (marketplaces + skill repos)",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			if len(st.Marketplaces) == 0 {
				ui.Info("No sources registered.")
				ui.Info("Add a marketplace: openskills marketplace add <git-url>")
				ui.Info("Add a skill repo:  openskills skill add <git-url>")
				return nil
			}

			ui.Header("Available Skills")
			for _, m := range st.Marketplaces {
				if len(args) > 0 && m.Name != args[0] {
					continue
				}
				skills, err := scanner.ScanSkills(m.LocalPath, m.Name)
				if err != nil {
					ui.Warn("Error scanning %s: %v", m.Name, err)
					continue
				}
				if len(skills) == 0 {
					continue
				}

				sourceLabel := m.Sources.Label()
				fmt.Printf("\n  \033[1m%s\033[0m (%s)\n", m.Name, sourceLabel)
				for _, s := range skills {
					installed := ""
					inst := st.FindInstallation(s.QualifiedName())
					if inst != nil {
						var targets []string
						for t := range inst.Targets {
							targets = append(targets, t)
						}
						installed = fmt.Sprintf(" \033[32m[installed: %v]\033[0m", targets)
					}
					fmt.Printf("    %-30s %s%s\n", s.Name, s.Description, installed)
				}
			}
			return nil
		},
	}
}

func (a *App) skillInstallCmd() *cobra.Command {
	var targets []string
	var mode string

	cmd := &cobra.Command{
		Use:   "install <skill[@version]> [skill...]",
		Short: "Install skills to target editors",
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

			var skills []resource.Resource
			for _, arg := range args {
				filtered := filterByType(allResources, resource.TypeSkill)
				res, err := resolver.Resolve(arg, filtered)
				if err != nil {
					if ambErr, ok := err.(*resolver.AmbiguousError); ok {
						ui.Warn("Ambiguous skill %q found in multiple sources:", arg)
						for _, m := range ambErr.Matches {
							fmt.Printf("    - %s\n", m.QualifiedName())
						}
						ui.Info("Please specify: openskills skill install %s@<source>", arg)
						return nil
					}
					return err
				}
				skills = append(skills, *res)
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

		for _, skill := range skills {
			mp := st.FindMarketplace(skill.Marketplace)
			if mp == nil {
				ui.Error("Source %s not found for skill %s", skill.Marketplace, skill.Name)
				continue
			}

				sourcePath := skill.LocalPath
				if installMode == resource.ModeSymlink {
					intermediate, err := stageSkillToOsk(skill.Name, sourcePath)
					if err != nil {
						ui.Error("Failed to stage %s: %v", skill.Name, err)
						continue
					}
					sourcePath = intermediate
				}

				sha, _ := gitutil.CurrentCommitSHA(mp.LocalPath)

				ui.Info("Installing %s to %v ...", skill.QualifiedName(), targetNames)
				err := a.inst.Install(context.Background(), installer.InstallRequest{
					Resource:    skill,
					Mode:        installMode,
					TargetNames: targetNames,
					SourcePath:  sourcePath,
					CommitSHA:   sha,
				})
				if err != nil {
					ui.Error("Failed to install %s: %v", skill.Name, err)
					continue
				}
				ui.Success("Installed %s (%s) to %v", skill.QualifiedName(), installMode, targetNames)
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&targets, "target", "t", nil, "Target editors (codex,claude,cursor)")
	cmd.Flags().StringVarP(&mode, "mode", "m", "", "Override install mode (symlink/native)")
	return cmd
}

func (a *App) skillUninstallCmd() *cobra.Command {
	var targets []string

	cmd := &cobra.Command{
		Use:   "uninstall <skill> [skill...]",
		Short: "Uninstall skills from target editors",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			for _, arg := range args {
				st, err := a.stateMgr.Load()
				if err != nil {
					return err
				}
				allResources, _ := marketplace.ListAllResources(st)
				filtered := filterByType(allResources, resource.TypeSkill)
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

				cleanupStagedSkill(res.Name)
				ui.Success("Uninstalled %s", qn)
			}
			return nil
		},
	}

	cmd.Flags().StringSliceVarP(&targets, "target", "t", nil, "Target editors (codex,claude,cursor)")
	return cmd
}

// stageSkillToOsk creates a symlink in ~/.osk/skills/<name> pointing to the
// source repo's skill directory. Editor adapters then symlink to this
// intermediate path, creating a two-layer chain:
//
//	repo/skills/<name> ← ~/.osk/skills/<name> ← ~/.agents/skills/<name>
func stageSkillToOsk(skillName, repoSkillPath string) (string, error) {
	dir := paths.SkillsDir()
	if err := paths.EnsureDir(dir); err != nil {
		return "", err
	}
	dest := fmt.Sprintf("%s/%s", dir, skillName)
	os.Remove(dest)
	if err := os.Symlink(repoSkillPath, dest); err != nil {
		return "", fmt.Errorf("stage skill to ~/.osk/skills: %w", err)
	}
	return dest, nil
}

func cleanupStagedSkill(skillName string) {
	staged := fmt.Sprintf("%s/%s", paths.SkillsDir(), skillName)
	os.Remove(staged)
}

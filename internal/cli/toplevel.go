package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/lovelyJason/openskills/internal/codexmgr"
	"github.com/lovelyJason/openskills/internal/discovery"
	"github.com/lovelyJason/openskills/internal/marketplace"
	"github.com/lovelyJason/openskills/internal/resource"
	"github.com/lovelyJason/openskills/internal/state"
	"github.com/lovelyJason/openskills/internal/target"
	"github.com/lovelyJason/openskills/internal/ui"
)

func (a *App) newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all installed resources across all targets",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			fmt.Println()
			fmt.Println(ui.RenderAppHeader(Version, len(st.Marketplaces), len(st.Installations)))
			fmt.Println()

			if len(st.Installations) > 0 {
				var pluginItems, skillItems []ui.ListItem
				for _, inst := range st.Installations {
					item := ui.ListItem{Name: inst.ID, IsOSK: true}
					var tNames []string
					for t := range inst.Targets {
						tNames = append(tNames, t)
					}
					item.Tag = fmt.Sprintf("→ %v", tNames)
					if inst.ResourceType == resource.TypePlugin {
						pluginItems = append(pluginItems, item)
					} else {
						skillItems = append(skillItems, item)
					}
				}
				var sections []ui.SectionData
				if len(pluginItems) > 0 {
					sections = append(sections, ui.SectionData{Title: "Plugins", Icon: "🔌", Items: pluginItems})
				}
				if len(skillItems) > 0 {
					sections = append(sections, ui.SectionData{Title: "Skills", Icon: "🎯", Items: skillItems})
				}
				if len(sections) > 0 {
					fmt.Println(ui.RenderEditorSection("OpenSkills Managed", "", sections))
				}
			}

			for _, adapter := range a.registry.All() {
				if !adapter.Detect() {
					continue
				}
				switch adapter.Name() {
				case "claude":
					renderClaudePlatform()
				case "codex":
					renderCodexPlatform()
				case "cursor":
					renderCursorPlatform()
				}
			}

			return nil
		},
	}
}

func renderClaudePlatform() {
	res := discovery.DiscoverClaude()
	var sections []ui.SectionData

	if len(res.Marketplaces) > 0 {
		items := make([]ui.ListItem, len(res.Marketplaces))
		for i, m := range res.Marketplaces {
			items[i] = ui.ListItem{Name: m.Name, Tag: m.Tag}
		}
		sections = append(sections, ui.SectionData{Title: "Marketplaces", Icon: "🏪", Items: items})
	}

	if len(res.Plugins) > 0 {
		items := make([]ui.ListItem, len(res.Plugins))
		for i, p := range res.Plugins {
			items[i] = ui.ListItem{Name: p.Name, Tag: p.Source, IsOSK: p.IsOSK}
		}
		sections = append(sections, ui.SectionData{Title: "Plugins", Icon: "🔌", Items: items})
	}

	if len(res.Skills) > 0 {
		items := make([]ui.ListItem, len(res.Skills))
		for i, s := range res.Skills {
			items[i] = ui.ListItem{Name: s.Name, IsOSK: s.IsOSK}
		}
		sections = append(sections, ui.SectionData{Title: "Skills", Icon: "🎯", Items: items})
	}

	if len(sections) == 0 {
		return
	}
	fmt.Println(ui.RenderEditorSection("Claude", "", sections))
}

func renderCodexPlatform() {
	res := discovery.DiscoverCodex()
	var sections []ui.SectionData

	extra := ""
	if ver, err := codexmgr.DetectedVersion(); err == nil {
		extra = "v" + ver
	}

	if len(res.Plugins) > 0 {
		items := make([]ui.ListItem, len(res.Plugins))
		for i, p := range res.Plugins {
			items[i] = ui.ListItem{Name: p.Name, Tag: p.Tag, IsOSK: p.IsOSK}
		}
		sections = append(sections, ui.SectionData{Title: "Installed Plugins", Icon: "🔌", Items: items})
	}

	if len(res.Skills) > 0 {
		items := make([]ui.ListItem, len(res.Skills))
		for i, s := range res.Skills {
			items[i] = ui.ListItem{Name: s.Name, Tag: s.Tag, IsOSK: s.IsOSK}
		}
		sections = append(sections, ui.SectionData{Title: "Skills", Icon: "🎯", Items: items})
	}

	if len(sections) == 0 {
		return
	}
	fmt.Println(ui.RenderEditorSection("Codex", extra, sections))
}

func renderCursorPlatform() {
	res := discovery.DiscoverCursor()
	var sections []ui.SectionData

	if len(res.Skills) > 0 {
		items := make([]ui.ListItem, len(res.Skills))
		for i, s := range res.Skills {
			items[i] = ui.ListItem{Name: s.Name, IsOSK: s.IsOSK}
		}
		sections = append(sections, ui.SectionData{Title: "Skills", Icon: "🎯", Items: items})
	}

	if len(sections) == 0 {
		return
	}
	fmt.Println(ui.RenderEditorSection("Cursor", "", sections))
}

func (a *App) newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show system status and health",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			ui.Header("OpenSkills Status")

			fmt.Printf("  Version: %s\n", Version)
			fmt.Printf("  Config:  %s\n", a.cfg.DefaultInstallMode)
			fmt.Println()

			fmt.Println("  \033[1mDetected Editors:\033[0m")
			for _, adapter := range a.registry.All() {
				status := "\033[31m✗ not found\033[0m"
				extra := ""
				if adapter.Detect() {
					status = "\033[32m✓ detected\033[0m"
					if adapter.Name() == "codex" {
						if ver, err := codexmgr.DetectedVersion(); err == nil {
							extra = fmt.Sprintf(" (v%s)", ver)
						}
					}
				}
				fmt.Printf("    %-10s %s%s\n", adapter.Name(), status, extra)
			}

			fmt.Println()
			mpCount, srCount := 0, 0
			for _, m := range st.Marketplaces {
				if m.Source == state.SourceSkillRepo {
					srCount++
				} else {
					mpCount++
				}
			}
			fmt.Printf("  \033[1mMarketplaces:\033[0m %d registered\n", mpCount)
			if srCount > 0 {
				fmt.Printf("  \033[1mSkill Repos:\033[0m  %d registered\n", srCount)
			}
			fmt.Printf("  \033[1mInstallations:\033[0m %d resources\n", len(st.Installations))

			return nil
		},
	}
}

func (a *App) newUpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update all marketplaces and installed resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			if len(st.Marketplaces) == 0 {
				ui.Info("No sources to update.")
				return nil
			}

			for i := range st.Marketplaces {
				m := &st.Marketplaces[i]
				sourceLabel := "marketplace"
				if m.Source == state.SourceSkillRepo {
					sourceLabel = "skill repo"
				}
				if m.PinnedVer != "" {
					ui.Warn("%s (%s) pinned to %s, skipping", m.Name, sourceLabel, m.PinnedVer)
					continue
				}
				ui.Info("Updating %s (%s) ...", m.Name, sourceLabel)
				if err := marketplace.Update(m); err != nil {
					ui.Error("Failed to update %s: %v", m.Name, err)
					continue
				}
				ui.Success("Updated %s", m.Name)
			}

			return a.stateMgr.Save(st)
		},
	}
}

func (a *App) newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check system health and detect issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			st, err := a.stateMgr.Load()
			if err != nil {
				return err
			}

			ui.Header("Health Check")
			issues := 0

			for _, adapter := range a.registry.Available() {
				if vc, ok := adapter.(target.VersionChecker); ok {
					if err := vc.CheckVersion(); err != nil {
						ui.Warn("[%s] %v", adapter.Name(), err)
						issues++
					} else {
						ui.Success("[%s] version OK", adapter.Name())
					}
				}
			}

			for _, m := range st.Marketplaces {
				sourceLabel := "Marketplace"
				if m.Source == state.SourceSkillRepo {
					sourceLabel = "Skill repo"
				}
				_, err := marketplace.ListResources(&m, "")
				if err != nil {
					ui.Error("Cannot scan %s %s: %v", sourceLabel, m.Name, err)
					issues++
				} else {
					ui.Success("%s %s: OK", sourceLabel, m.Name)
				}
			}

			for _, inst := range st.Installations {
				for tName, t := range inst.Targets {
					for _, p := range t.Paths {
						if _, err := os.Lstat(p); err != nil {
							ui.Warn("Missing path for %s@%s: %s", inst.ID, tName, p)
							issues++
						}
					}
				}
			}

			if issues == 0 {
				ui.Success("All checks passed!")
			} else {
				ui.Warn("%d issue(s) found", issues)
			}

			return nil
		},
	}
}

func filterInstallations(installations []state.Installation, t resource.Type) []state.Installation {
	var result []state.Installation
	for _, inst := range installations {
		if inst.ResourceType == t {
			result = append(result, inst)
		}
	}
	return result
}

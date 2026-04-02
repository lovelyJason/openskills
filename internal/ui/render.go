package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF8C42")).
			PaddingLeft(1)

	sectionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#6C9EFF"))

	editorTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#3D5A80")).
				Padding(0, 1)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	greenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4EC970"))

	yellowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD93D"))

	cyanStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6ECFF6"))

	tagOfficialStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4EC970")).
				Bold(true)

	tagCommunityStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6ECFF6"))

	tagSystemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#B39DDB"))

	tagAgentsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB74D"))

	tagEnabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4EC970"))

	tagDisabledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF5350"))

	oskBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF8C42")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3D5A80")).
			Padding(0, 1).
			MarginBottom(1)

	oskBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF8C42")).
			Padding(0, 1).
			MarginBottom(1)
)

type ListItem struct {
	Name   string
	Tag    string
	Source string
	IsOSK  bool
}

func RenderAppHeader(version string, mpCount, installCount int) string {
	var sb strings.Builder

	title := titleStyle.Render("⚡ OpenSkills")
	sb.WriteString(title + "  " + dimStyle.Render("v"+version))
	sb.WriteString("\n")

	if mpCount == 0 && installCount == 0 {
		sb.WriteString(dimStyle.Render("  Marketplaces: 0  |  Installed: 0"))
		sb.WriteString("\n")
		sb.WriteString(cyanStyle.Render("  ℹ ") + dimStyle.Render("Try: ") + "openskills marketplace add <git-url>")
	} else {
		sb.WriteString(fmt.Sprintf("  Marketplaces: %s  |  Installed: %s",
			greenStyle.Render(fmt.Sprintf("%d", mpCount)),
			greenStyle.Render(fmt.Sprintf("%d", installCount)),
		))
	}

	return oskBoxStyle.Width(60).Render(sb.String())
}

func RenderEditorSection(name string, extra string, sections []SectionData) string {
	var sb strings.Builder

	header := editorTitleStyle.Render(name)
	if extra != "" {
		header += "  " + dimStyle.Render(extra)
	}
	sb.WriteString(header)
	sb.WriteString("\n")

	for i, sec := range sections {
		if len(sec.Items) == 0 {
			continue
		}

		icon := sec.Icon
		if icon == "" {
			icon = "●"
		}
		countStr := dimStyle.Render(fmt.Sprintf("(%d)", len(sec.Items)))
		sb.WriteString("\n")
		sb.WriteString(sectionHeaderStyle.Render(fmt.Sprintf(" %s %s ", icon, sec.Title)) + " " + countStr)
		sb.WriteString("\n")

		maxNameLen := 0
		for _, item := range sec.Items {
			nameLen := len(item.Name)
			if item.IsOSK {
				nameLen += 4
			}
			if nameLen > maxNameLen {
				maxNameLen = nameLen
			}
		}
		if maxNameLen > 36 {
			maxNameLen = 36
		}

		for j, item := range sec.Items {
			isLast := j == len(sec.Items)-1
			prefix := " ├─"
			if isLast {
				prefix = " └─"
			}

			name := item.Name
			if len(name) > 36 {
				name = name[:33] + "..."
			}
			line := dimStyle.Render(prefix) + " " + name
			if item.IsOSK {
				line += " " + oskBadgeStyle.Render("osk")
			}

			nameLen := len(item.Name)
			if item.IsOSK {
				nameLen += 4
			}
			if nameLen > 36 {
				nameLen = 36
			}

			tagStr := ""
			if item.Tag != "" {
				tagStr = renderTag(item.Tag)
			} else if item.Source != "" {
				tagStr = dimStyle.Render(item.Source)
			}
			if tagStr != "" {
				pad := maxNameLen - nameLen + 2
				if pad < 2 {
					pad = 2
				}
				line += strings.Repeat(" ", pad) + tagStr
			}

			sb.WriteString(line)
			if j < len(sec.Items)-1 || i < len(sections)-1 {
				sb.WriteString("\n")
			}
		}
	}

	return boxStyle.Width(60).Render(sb.String())
}

type SectionData struct {
	Title string
	Icon  string
	Items []ListItem
}

func renderTag(tag string) string {
	switch tag {
	case "official":
		return tagOfficialStyle.Render("官方")
	case "community":
		return tagCommunityStyle.Render("社区")
	case "author":
		return yellowStyle.Render("作者")
	case "System":
		return tagSystemStyle.Render("System")
	case "Agents":
		return tagAgentsStyle.Render("Agents")
	case "enabled":
		return tagEnabledStyle.Render("✓ enabled")
	case "disabled":
		return tagDisabledStyle.Render("✗ disabled")
	default:
		return tagAgentsStyle.Render(tag)
	}
}

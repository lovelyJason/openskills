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

		displayItems := sec.Items
		totalCount := len(sec.Items)
		truncated := false
		if sec.MaxShow > 0 && len(sec.Items) > sec.MaxShow {
			displayItems = sec.Items[:sec.MaxShow]
			truncated = true
			if sec.Total > 0 {
				totalCount = sec.Total
			}
		}

		maxNameLen := 0
		for _, item := range displayItems {
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

		lastIdx := len(displayItems) - 1
		if truncated {
			lastIdx = -1
		}

		for j, item := range displayItems {
			isLast := j == len(displayItems)-1 && !truncated
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
			sb.WriteString("\n")
			_ = lastIdx
		}

		if truncated {
			remaining := totalCount - len(displayItems)
			moreText := dimStyle.Render(" └─ ") + dimStyle.Render(fmt.Sprintf("... and %d more", remaining))
			sb.WriteString(moreText)
			if i < len(sections)-1 {
				sb.WriteString("\n")
			}
		} else if len(displayItems) > 0 {
			content := sb.String()
			if strings.HasSuffix(content, "\n") && i == len(sections)-1 {
				sb.Reset()
				sb.WriteString(strings.TrimRight(content, "\n"))
			}
		}
	}

	return boxStyle.Width(60).Render(sb.String())
}

type SectionData struct {
	Title    string
	Icon     string
	Items    []ListItem
	MaxShow  int // if > 0, truncate and show "... and N more"
	Total    int // total count (used with MaxShow for truncation message)
}

type SourceInfo struct {
	Name       string
	SourceType string // "marketplace" or "skill repo"
	Plugins    int
	Skills     int
	SampleNames []string
}

func RenderSourcesSection(sources []SourceInfo) string {
	var sb strings.Builder

	header := editorTitleStyle.Render("Registered Sources")
	sb.WriteString(header)
	sb.WriteString("\n")

	for i, src := range sources {
		isLast := i == len(sources)-1
		prefix := " ├─"
		connector := " │ "
		if isLast {
			prefix = " └─"
			connector = "   "
		}

		typeTag := tagCommunityStyle.Render(src.SourceType)
		counts := ""
		if src.Plugins > 0 && src.Skills > 0 {
			counts = fmt.Sprintf("%dp/%ds", src.Plugins, src.Skills)
		} else if src.Plugins > 0 {
			counts = fmt.Sprintf("%d plugins", src.Plugins)
		} else if src.Skills > 0 {
			counts = fmt.Sprintf("%d skills", src.Skills)
		}

		namePart := greenStyle.Render(src.Name)
		metaPart := typeTag + "  " + dimStyle.Render(counts)
		sb.WriteString(dimStyle.Render(prefix) + " " + namePart + "  " + metaPart)
		sb.WriteString("\n")

		maxSample := 5
		if len(src.SampleNames) < maxSample {
			maxSample = len(src.SampleNames)
		}
		total := src.Plugins + src.Skills
		for j := 0; j < maxSample; j++ {
			samplePrefix := connector + dimStyle.Render("├─")
			isLastSample := j == maxSample-1 && maxSample >= total
			if isLastSample {
				samplePrefix = connector + dimStyle.Render("└─")
			}
			sb.WriteString(samplePrefix + " " + dimStyle.Render(src.SampleNames[j]))
			sb.WriteString("\n")
		}
		if total > maxSample {
			remaining := total - maxSample
			sb.WriteString(connector + dimStyle.Render("└─") + " " + dimStyle.Render(fmt.Sprintf("... and %d more", remaining)))
			sb.WriteString("\n")
		}
	}

	content := strings.TrimRight(sb.String(), "\n")
	return boxStyle.Width(60).Render(content)
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
	case "installed":
		return tagEnabledStyle.Render("✓ Installed")
	case "enabled":
		return tagEnabledStyle.Render("✓ enabled")
	case "disabled":
		return tagDisabledStyle.Render("✗ disabled")
	case "available":
		return dimStyle.Render("Available")
	default:
		return tagAgentsStyle.Render(tag)
	}
}

package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.screen == screenSetup {
		return m.setupView()
	}
	return m.downloadView()
}

func (m Model) setupView() string {
	termWidth := m.width
	if termWidth <= 0 {
		termWidth = 120
	}
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("25")).
		Padding(0, 1)
	subtitleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	titleBlock := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("YouTube Downloader"),
		subtitleStyle.Render("Channel or single video URL, pure Go download engine"),
	)

	modeText := m.renderModeSelector()
	startLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Press Enter on this row to start")
	startButton := lipgloss.NewStyle().Padding(0, 1).Render("[ Start Download ]")
	if m.focusIndex == 3 {
		startButton = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("35")).Padding(0, 1).Render("[ Start Download ]")
	}

	left := []string{titleBlock, "", modeText, "", m.focusLine(1, m.urlInput.View()), m.focusLine(2, m.outputInput.View()), "", m.focusLine(3, startButton), startLabel}
	if strings.TrimSpace(m.errText) != "" {
		left = append(left, "")
		left = append(left, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("Error: "+m.errText))
	}
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)
	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(1, 2)

	mainCard := cardStyle.Render(strings.Join(left, "\n"))
	helpCard := helpStyle.Render(m.setupHelpTxt)
	return lipgloss.JoinVertical(lipgloss.Left, helpCard, "", mainCard)
}

func (m Model) downloadView() string {
	mode := strings.ToUpper(string(m.request.Mode))
	target := "Target: " + short(m.request.TargetURL, 120)
	failed := 0
	for _, v := range m.videos {
		if v.hasError {
			failed++
		}
	}
	progress := fmt.Sprintf("%s %.1f%%", bar(m.overall, 34), m.overall*100)
	statusLine := fmt.Sprintf("Completed %d/%d   Failed %d   Active %s   Queue %d", m.completed, m.totalVideos, failed, mode, len(m.pending))
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Padding(0, 1).Render("Download Session")
	summary := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("239")).
		Padding(0, 1).
		Render(strings.Join([]string{target, progress, statusLine}, "\n"))
	lines := []string{header, summary, "", m.table.View()}
	if m.addingURL {
		lines = append(lines, "", m.addURLComposerView())
	}
	if strings.TrimSpace(m.lastLog) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Log: "+short(m.lastLog, 140)))
	}
	if strings.TrimSpace(m.errText) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("Error: "+short(m.errText, 140)))
	}
	if m.addingURL {
		lines = append(lines, "", lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("type URL   left/right mode   enter queue/start   esc close composer"))
	} else {
		lines = append(lines, "", lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render("↑/↓/j/k move row   a add link   q quit   esc cancel + back to setup"))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderModeSelector() string {
	modeChip := func(index int, label, color string) string {
		style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("248"))
		if m.modeIndex == index {
			style = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color(color))
		}
		return style.Render(strings.ToUpper(label))
	}
	prefix := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Mode:")
	row := strings.Join([]string{
		prefix,
		modeChip(0, "audio", "31"),
		modeChip(1, "video", "127"),
		modeChip(2, "both", "64"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("(left/right to switch)"),
	}, " ")
	if m.focusIndex == 0 {
		return lipgloss.NewStyle().Bold(true).Render("● " + row)
	}
	return "○ " + row
}

func (m Model) focusLine(index int, value string) string {
	if m.focusIndex == index {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render("● " + value)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("○ ") + value
}

func (m Model) addURLComposerView() string {
	chip := func(index int, label, color string) string {
		base := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("248"))
		if m.addModeIndex == index {
			base = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color(color))
		}
		return base.Render(strings.ToUpper(label))
	}
	modeRow := strings.Join([]string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Mode:"),
		chip(0, "audio", "31"),
		chip(1, "video", "127"),
		chip(2, "both", "64"),
	}, " ")
	body := strings.Join([]string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render("Add URL to Queue"),
		modeRow,
		m.addInput.View(),
	}, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Render(body)
}

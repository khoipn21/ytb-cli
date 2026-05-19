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
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("YouTube Channel/Video Downloader")
	modeLabel := []string{"audio", "video", "both"}[m.modeIndex]
	modeText := fmt.Sprintf("Mode: [%s]  (left/right to toggle)", strings.ToUpper(modeLabel))
	if m.focusIndex == 0 {
		modeText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render(modeText)
	}
	startLabel := "Start download"
	if m.focusIndex == 3 {
		startLabel = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("[ " + startLabel + " ]")
	} else {
		startLabel = "[ " + startLabel + " ]"
	}
	left := []string{
		title,
		"",
		modeText,
		m.urlInput.View(),
		m.outputInput.View(),
		m.ytDLPInput.View(),
		"",
		startLabel,
	}
	if strings.TrimSpace(m.errText) != "" {
		left = append(left, "")
		left = append(left, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.errText))
	}
	right := lipgloss.NewStyle().PaddingLeft(2).Render(m.setupHelpTxt)
	return lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(left, "\n"), right)
}

func (m Model) downloadView() string {
	mode := strings.ToUpper(string(m.request.Mode))
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Running: " + mode + " download")
	target := "Target: " + m.request.TargetURL
	overall := fmt.Sprintf("Overall: [%s] %.1f%% (%d/%d completed)", bar(m.overall, 36), m.overall*100, m.completed, m.totalVideos)
	lines := []string{header, target, overall, "", m.table.View()}
	if strings.TrimSpace(m.lastLog) != "" {
		lines = append(lines, "Last log: "+short(m.lastLog, 120))
	}
	if strings.TrimSpace(m.errText) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+short(m.errText, 120)))
	}
	lines = append(lines, "", "↑/↓/j/k: move row  q: quit  esc: cancel to setup")
	return strings.Join(lines, "\n")
}

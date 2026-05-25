package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"youtube-channel-audio-downloader/internal/downloader"
)

var filterLabels = []string{"All", "Active", "Completed", "Failed", "Queued"}

func (m Model) downloadView() string {
	if m.width > 0 && m.height > 0 && (m.width < 80 || m.height < 24) {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Render("Terminal too small. Resize to at least 80x24.")
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4")).Padding(0, 1).Render("Download Session")
	summary := panelStyle(false).Render(m.renderDownloadSummary())
	filterBar := panelStyle(false).Render(m.renderTopFilterBar())
	tablePanel := panelStyle(true).Render(m.table.View())

	lines := []string{header, summary, filterBar}
	if m.detailOpen {
		detailPanel := panelStyle(m.activePanel == panelDetail).Width(44).Render(m.renderItemDetail())
		if m.width >= 145 {
			lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, tablePanel, detailPanel))
		} else {
			lines = append(lines, tablePanel, detailPanel)
		}
	} else {
		lines = append(lines, tablePanel)
	}

	if m.addingURL {
		lines = append(lines, panelStyle(true).Render(m.addURLComposerView()))
	}
	if strings.TrimSpace(m.lastLog) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render("Log: "+short(m.lastLog, 160)))
	}
	if strings.TrimSpace(m.errText) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render("Error: "+short(m.errText, 160)))
	}
	if m.showHelp {
		lines = append(lines, panelStyle(false).Render(m.renderDownloadHelp()))
	}
	lines = append(lines, m.renderDownloadFooter())
	return strings.Join(lines, "\n")
}

func (m Model) renderDownloadSummary() string {
	mode := strings.ToUpper(string(m.request.Mode))
	target := "Target: " + short(m.request.TargetURL, 96)
	failed := 0
	for i := range m.videos {
		if m.videos[i].hasError {
			failed++
		}
	}
	progress := fmt.Sprintf("%s %.1f%%", bar(m.overall, 28), m.overall*100)
	statusLine := fmt.Sprintf("Completed %d/%d  Failed %d  Queue %d  Mode %s", m.completed, m.totalVideos, failed, len(m.pending), mode)
	return strings.Join([]string{target, progress, statusLine}, "\n")
}

func (m Model) renderTopFilterBar() string {
	counts := m.filterCounts()
	parts := make([]string, 0, len(filterLabels)+1)
	parts = append(parts, lipgloss.NewStyle().Bold(true).Render("Status"))
	for i, label := range filterLabels {
		tag := fmt.Sprintf("%s %d", label, counts[i])
		style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("7"))
		if m.filterIndex == i {
			style = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("6"))
		}
		parts = append(parts, style.Render(tag))
	}
	parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("h/l switch filter"))
	return strings.Join(parts, "  ")
}

func (m Model) renderItemDetail() string {
	item, index, ok := m.detailVideoWithIndex()
	if !ok {
		return strings.Join([]string{"Item Detail", "", "No row selected."}, "\n")
	}
	status := m.resolveVideoStatus(index, item)
	speed := item.speed
	if speed == "" {
		speed = "-"
	}
	eta := item.eta
	if eta == "" {
		eta = "-"
	}
	mode := strings.ToUpper(string(detectRowMode(item.video.Title, m.request.Mode)))
	lines := []string{
		lipgloss.NewStyle().Bold(true).Render("Item Detail"),
		"",
		"Row: " + fmt.Sprintf("%d", index+1),
		"Status: " + status,
		"Progress: " + fmt.Sprintf("%.1f%%", item.percent*100),
		"Mode: " + mode,
		"Speed: " + speed,
		"ETA: " + eta,
		"Type: " + string(downloader.DetectTargetType(item.video.URL)),
		"",
		"Title:",
		short(item.video.Title, 220),
		"",
		"URL:",
		short(item.video.URL, 220),
	}
	if item.errorText != "" {
		lines = append(lines, "", lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("Error: "+short(item.errorText, 200)))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderDownloadFooter() string {
	if m.addingURL {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("[enter] queue/start  [left/right] mode  [esc] close composer")
	}
	if m.detailOpen {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("[esc/enter] close detail  [a] add URL  [?] help  [q] quit")
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("[h/l] filter  [↑/↓ j/k] row  [enter] open detail  [a] add URL  [esc] back")
}

func (m Model) renderDownloadHelp() string {
	return strings.Join([]string{
		"Keyboard",
		"[h]/[l] change status filter in top bar",
		"[j]/[k] or arrows move selected row",
		"[enter] open detail panel for selected row",
		"[esc] close detail or cancel run and return setup",
		"[a] add URL to queue, [q] quit",
	}, "\n")
}

func panelStyle(focused bool) lipgloss.Style {
	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1)
	if focused {
		return style.BorderForeground(lipgloss.Color("6"))
	}
	return style
}

func (m Model) filterCounts() [5]int {
	counts := [5]int{}
	counts[0] = len(m.videos)
	for i := range m.videos {
		status := m.resolveVideoStatus(i, m.videos[i])
		switch status {
		case "downloading":
			counts[1]++
		case "done":
			counts[2]++
		case "error":
			counts[3]++
		case "queued":
			counts[4]++
		}
	}
	return counts
}

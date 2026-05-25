package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"youtube-channel-audio-downloader/internal/downloader"
)

func (m *Model) initDownloadTable() {
	m.table = table.New(table.WithFocused(true))
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("4")).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder())
	styles.Selected = styles.Selected.
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("6")).
		Bold(true)
	m.table.SetStyles(styles)
	m.updateTableDimensions()
	m.refreshTableRows()
	m.activePanel = panelTable
}

func (m *Model) refreshTableRows() {
	rows := make([]table.Row, 0, len(m.videos))
	visible := make([]int, 0, len(m.videos))
	for i, item := range m.videos {
		status := m.resolveVideoStatus(i, item)
		if !m.matchesFilter(status) {
			continue
		}
		mode := detectRowMode(item.video.Title, m.request.Mode)
		visible = append(visible, i)
		if m.tableCompact {
			rows = append(rows, table.Row{
				fmt.Sprintf("%d", i+1),
				renderMode(mode),
				renderStatus(status),
				fmt.Sprintf("%.0f%%", item.percent*100),
				item.video.Title,
			})
			continue
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			renderMode(mode),
			renderStatus(status),
			fmt.Sprintf("%.1f%%", item.percent*100),
			item.speed,
			item.eta,
			item.video.Title,
		})
	}
	cursor := m.table.Cursor()
	m.visibleRows = visible
	m.table.SetRows(rows)
	if len(rows) == 0 {
		m.table.SetCursor(0)
		return
	}
	if cursor >= len(rows) {
		cursor = len(rows) - 1
	}
	if cursor < 0 {
		cursor = 0
	}
	m.table.SetCursor(cursor)
}

func (m *Model) updateTableDimensions() {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 120
	}
	if height <= 0 {
		height = 35
	}
	tableHeight := int(math.Max(8, float64(height-16)))
	if m.detailOpen && width < 145 {
		tableHeight = int(math.Max(8, float64(height-29)))
	}
	tableWidth := width - 6
	if m.detailOpen && width >= 145 {
		tableWidth = width - 50
	} else if width >= 110 {
		tableWidth = width - 30
	}
	if tableWidth < 52 {
		tableWidth = 52
	}

	m.tableCompact = tableWidth < 96
	if m.tableCompact {
		remaining := tableWidth - 28
		if remaining < 16 {
			remaining = 16
		}
		columns := []table.Column{
			{Title: "#", Width: 3},
			{Title: "Mode", Width: 5},
			{Title: "Status", Width: 12},
			{Title: "Prog", Width: 6},
			{Title: "Title", Width: remaining},
		}
		m.table.SetColumns(columns)
		m.table.SetHeight(tableHeight)
		m.table.SetWidth(tableWidth)
		return
	}
	remaining := tableWidth - 52
	if remaining < 18 {
		remaining = 18
	}
	columns := []table.Column{
		{Title: "#", Width: 3},
		{Title: "Mode", Width: 6},
		{Title: "Status", Width: 12},
		{Title: "Progress", Width: 9},
		{Title: "Speed", Width: 8},
		{Title: "ETA", Width: 6},
		{Title: "Title", Width: remaining},
	}
	m.table.SetColumns(columns)
	m.table.SetHeight(tableHeight)
	m.table.SetWidth(tableWidth)
}

func renderMode(mode downloader.DownloadMode) string {
	switch mode {
	case downloader.ModeVideo:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Render("video")
	case downloader.ModeBoth:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("111")).Render("both")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render("audio")
	}
}

func renderStatus(status string) string {
	switch status {
	case "done":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("done")
	case "error":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("error")
	case "downloading":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("downloading")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Render("queued")
	}
}

func detectRowMode(title string, fallback downloader.DownloadMode) downloader.DownloadMode {
	low := strings.ToLower(strings.TrimSpace(title))
	if strings.HasSuffix(low, "[video]") {
		return downloader.ModeVideo
	}
	if strings.HasSuffix(low, "[audio]") {
		return downloader.ModeAudio
	}
	if fallback == downloader.ModeBoth {
		return downloader.ModeAudio
	}
	return fallback
}

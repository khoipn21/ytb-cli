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
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder())
	styles.Selected = styles.Selected.
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("25")).
		Bold(true)
	m.table.SetStyles(styles)
	m.updateTableDimensions()
	m.refreshTableRows()
}

func (m *Model) refreshTableRows() {
	rows := make([]table.Row, 0, len(m.videos))
	for i, item := range m.videos {
		status := "queued"
		if item.done {
			status = "done"
		}
		if item.hasError {
			status = "error"
		}
		if i == m.currentIndex && !item.done && !item.hasError {
			status = "downloading"
		}
		mode := detectRowMode(item.video.Title, m.request.Mode)
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
	m.table.SetRows(rows)
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
	tableHeight := int(math.Max(8, float64(height-13)))
	m.tableCompact = width < 104
	if m.tableCompact {
		remaining := width - 34
		if remaining < 16 {
			remaining = 16
		}
		columns := []table.Column{
			{Title: "#", Width: 3},
			{Title: "Mode", Width: 5},
			{Title: "Status", Width: 10},
			{Title: "Prog", Width: 6},
			{Title: "Title", Width: remaining},
		}
		m.table.SetColumns(columns)
		m.table.SetHeight(tableHeight)
		m.table.SetWidth(width - 2)
		return
	}
	remaining := width - 52
	if remaining < 24 {
		remaining = 24
	}
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Mode", Width: 6},
		{Title: "Status", Width: 11},
		{Title: "Progress", Width: 9},
		{Title: "Speed", Width: 10},
		{Title: "ETA", Width: 6},
		{Title: "Title", Width: remaining},
	}
	m.table.SetColumns(columns)
	m.table.SetHeight(tableHeight)
	m.table.SetWidth(width - 2)
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
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("done")
	case "error":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("error")
	case "downloading":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Render("running")
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Render("queued")
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

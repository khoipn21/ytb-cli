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
	styles.Header = styles.Header.Bold(true).BorderBottom(true).BorderStyle(lipgloss.NormalBorder())
	styles.Selected = styles.Selected.Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62")).Bold(true)
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
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			string(mode),
			status,
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
	tableHeight := int(math.Max(8, float64(height-10)))
	remaining := width - 46
	if remaining < 24 {
		remaining = 24
	}
	columns := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Mode", Width: 6},
		{Title: "Status", Width: 12},
		{Title: "Progress", Width: 10},
		{Title: "Speed", Width: 8},
		{Title: "ETA", Width: 6},
		{Title: "Title", Width: remaining},
	}
	m.table.SetColumns(columns)
	m.table.SetHeight(tableHeight)
	m.table.SetWidth(width - 2)
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

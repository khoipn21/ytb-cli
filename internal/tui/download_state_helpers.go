package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"youtube-channel-audio-downloader/internal/downloader"
)

func startDownloadCmd(options downloader.Options, runID int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		events := downloader.NewService(options).Run(ctx)
		return startDownloadMsg{events: events, cancel: cancel, options: options, runID: runID}
	}
}

func waitForEvent(events <-chan downloader.Event, runID int) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		return eventMsg{event: event, events: events, runID: runID, ok: ok}
	}
}

func modeToIndex(mode downloader.DownloadMode) int {
	switch mode {
	case downloader.ModeVideo:
		return 1
	case downloader.ModeBoth:
		return 2
	default:
		return 0
	}
}

func indexToMode(index int) downloader.DownloadMode {
	switch index {
	case 1:
		return downloader.ModeVideo
	case 2:
		return downloader.ModeBoth
	default:
		return downloader.ModeAudio
	}
}

func (m *Model) moveFilter(delta int) {
	totalFilters := len(filterLabels)
	m.filterIndex = (m.filterIndex + delta + totalFilters) % totalFilters
	m.refreshTableRows()
}

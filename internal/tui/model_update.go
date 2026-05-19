package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"youtube-channel-audio-downloader/internal/downloader"
)

func (m Model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "tab", "down":
		m.focusIndex = (m.focusIndex + 1) % 4
		m.focusCurrentField()
		return m, nil
	case "shift+tab", "up":
		m.focusIndex = (m.focusIndex + 3) % 4
		m.focusCurrentField()
		return m, nil
	case "left", "right":
		if m.focusIndex == 0 {
			m.modeIndex = (m.modeIndex + 1) % 3
			return m, nil
		}
	case "enter":
		if m.focusIndex == 3 {
			return m.startDownloadFlow()
		}
		m.focusIndex = (m.focusIndex + 1) % 4
		m.focusCurrentField()
		return m, nil
	}
	var cmd tea.Cmd
	if m.focusIndex == 1 {
		m.urlInput, cmd = m.urlInput.Update(msg)
		return m, cmd
	}
	if m.focusIndex == 2 {
		m.outputInput, cmd = m.outputInput.Update(msg)
		return m, cmd
	}
	if m.focusIndex == 3 {
		m.ytDLPInput, cmd = m.ytDLPInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateDownload(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		if m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
	case "esc":
		if m.cancel != nil {
			m.cancel()
		}
		m = NewModel(Config{
			InitialURL:    m.request.TargetURL,
			InitialMode:   string(m.request.Mode),
			InitialOutput: m.request.OutputDir,
			InitialYTDLP:  m.request.YTDLPBin,
		})
		return m, nil
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) startDownloadFlow() (tea.Model, tea.Cmd) {
	targetURL := strings.TrimSpace(m.urlInput.Value())
	outputDir := strings.TrimSpace(m.outputInput.Value())
	ytBin := strings.TrimSpace(m.ytDLPInput.Value())
	if targetURL == "" || outputDir == "" {
		m.errText = "URL and output directory are required."
		return m, nil
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		m.errText = fmt.Sprintf("cannot create output directory: %v", err)
		return m, nil
	}
	mode := downloader.ModeAudio
	if m.modeIndex == 1 {
		mode = downloader.ModeVideo
	} else if m.modeIndex == 2 {
		mode = downloader.ModeBoth
	}
	m.request = downloader.Options{
		TargetURL: targetURL,
		OutputDir: outputDir,
		YTDLPBin:  ytBin,
		Mode:      mode,
	}
	return m, startDownloadCmd(m.request)
}

func startDownloadCmd(options downloader.Options) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		events := downloader.NewService(options).Run(ctx)
		return startDownloadMsg{events: events, cancel: cancel}
	}
}

func waitForEvent(events <-chan downloader.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		return eventMsg{event: event, ok: ok}
	}
}

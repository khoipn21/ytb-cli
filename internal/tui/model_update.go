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
	return m, nil
}

func (m Model) updateDownload(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addingURL {
		return m.updateAddURL(msg)
	}
	switch msg.String() {
	case "q":
		if m.cancel != nil {
			m.cancel()
		}
		return m, tea.Quit
	case "a":
		m.addingURL = true
		m.addModeIndex = modeToIndex(m.request.Mode)
		m.addInput.SetValue("")
		m.addInput.Focus()
		m.errText = ""
		return m, nil
	case "esc":
		if m.cancel != nil {
			m.cancel()
		}
		m = NewModel(Config{
			InitialURL:    m.request.TargetURL,
			InitialMode:   string(m.request.Mode),
			InitialOutput: m.request.OutputDir,
		})
		return m, nil
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) updateAddURL(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.addingURL = false
		m.addInput.Blur()
		return m, nil
	case "left":
		m.addModeIndex = (m.addModeIndex + 2) % 3
		return m, nil
	case "right":
		m.addModeIndex = (m.addModeIndex + 1) % 3
		return m, nil
	case "enter":
		targetURL := strings.TrimSpace(m.addInput.Value())
		if targetURL == "" {
			m.errText = "added URL cannot be empty"
			return m, nil
		}
		req := downloader.Options{
			TargetURL: targetURL,
			OutputDir: m.request.OutputDir,
			Mode:      indexToMode(m.addModeIndex),
		}
		m.addingURL = false
		m.addInput.Blur()
		m.addInput.SetValue("")
		if m.running {
			m.pending = append(m.pending, req)
			m.lastLog = fmt.Sprintf("Queued %s (%s)", short(req.TargetURL, 80), strings.ToUpper(string(req.Mode)))
			return m, nil
		}
		m.lastLog = fmt.Sprintf("Starting %s (%s)", short(req.TargetURL, 80), strings.ToUpper(string(req.Mode)))
		return m, startDownloadCmd(req)
	}
	var cmd tea.Cmd
	m.addInput, cmd = m.addInput.Update(msg)
	return m, cmd
}

func (m Model) startDownloadFlow() (tea.Model, tea.Cmd) {
	targetURL := strings.TrimSpace(m.urlInput.Value())
	outputDir := strings.TrimSpace(m.outputInput.Value())
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
		Mode:      mode,
	}
	m.addingURL = false
	m.addInput.Blur()
	m.addInput.SetValue("")
	m.videos = nil
	m.pending = nil
	m.totalVideos = 0
	m.completed = 0
	m.overall = 0
	m.currentIndex = -1
	return m, startDownloadCmd(m.request)
}

func startDownloadCmd(options downloader.Options) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(context.Background())
		events := downloader.NewService(options).Run(ctx)
		return startDownloadMsg{events: events, cancel: cancel, options: options}
	}
}

func waitForEvent(events <-chan downloader.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-events
		return eventMsg{event: event, ok: ok}
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

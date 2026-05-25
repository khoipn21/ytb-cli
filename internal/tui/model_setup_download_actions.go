package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"youtube-channel-audio-downloader/internal/downloader"
)

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
		m.runID++
		return m, startDownloadCmd(req, m.runID)
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
	m.detailOpen = false
	m.detailIndex = -1
	m.activePanel = panelTable
	m.runID++
	return m, startDownloadCmd(m.request, m.runID)
}

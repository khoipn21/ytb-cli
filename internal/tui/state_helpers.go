package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/glamour"
	"youtube-channel-audio-downloader/internal/downloader"
)

func (m *Model) applyEvent(event downloader.Event) {
	switch event.Type {
	case downloader.EventPlaylistReady:
		m.request.TargetURL = event.RequestURL
		m.request.Mode = event.Mode
		m.batchStart = len(m.videos)
		m.totalVideos += event.TotalVideos
		for _, v := range event.Videos {
			m.videos = append(m.videos, videoState{video: v})
		}
		if m.table.Width() == 0 {
			m.initDownloadTable()
		} else {
			m.updateTableDimensions()
			m.refreshTableRows()
		}
	case downloader.EventVideoStart:
		index := m.batchStart + event.VideoIndex
		m.currentIndex = index
		m.refreshTableRows()
		if visibleIndex := m.findVisibleRowByVideoIndex(index); visibleIndex >= 0 {
			m.table.SetCursor(visibleIndex)
		}
	case downloader.EventVideoProgress:
		index := m.batchStart + event.VideoIndex
		if index < 0 || index >= len(m.videos) {
			return
		}
		item := &m.videos[index]
		item.speed = event.Speed
		item.eta = event.ETA
		item.percent = clamp(event.Percent)
		m.recomputeOverall()
		m.refreshTableRows()
	case downloader.EventVideoDone:
		index := m.batchStart + event.VideoIndex
		if index >= 0 && index < len(m.videos) {
			item := &m.videos[index]
			item.done = true
			item.percent = 1.0
			item.speed = ""
			item.eta = ""
		}
		m.completed++
		m.recomputeOverall()
		m.refreshTableRows()
	case downloader.EventVideoError:
		index := m.batchStart + event.VideoIndex
		if index >= 0 && index < len(m.videos) {
			item := &m.videos[index]
			item.hasError = true
			item.errorText = event.Message
		}
		m.errText = event.Message
		m.refreshTableRows()
	case downloader.EventLog:
		m.lastLog = strings.TrimSpace(event.Message)
	case downloader.EventFinished:
		m.currentIndex = -1
		m.recomputeOverall()
		m.lastLog = fmt.Sprintf("Finished batch: %d/%d completed overall", m.completed, m.totalVideos)
		m.refreshTableRows()
	}
}

func (m *Model) focusCurrentField() {
	m.urlInput.Blur()
	m.outputInput.Blur()
	if m.focusIndex == 1 {
		m.urlInput.Focus()
	}
	if m.focusIndex == 2 {
		m.outputInput.Focus()
	}
}

func (m *Model) recomputeOverall() {
	if m.totalVideos <= 0 {
		m.overall = 0
		return
	}
	total := float64(m.completed)
	if m.currentIndex >= 0 && m.currentIndex < len(m.videos) && !m.videos[m.currentIndex].done {
		total += m.videos[m.currentIndex].percent
	}
	m.overall = clamp(total / float64(m.totalVideos))
}

func newInput(prompt, value, placeholder string) textinput.Model {
	input := textinput.New()
	input.Prompt = prompt + ": "
	input.Placeholder = placeholder
	input.SetValue(value)
	input.CharLimit = 1024
	input.Width = 72
	return input
}

func (m *Model) updateSetupDimensions() {
	width := m.width
	if width <= 0 {
		width = 120
	}
	inputWidth := width - 48
	if inputWidth < 26 {
		inputWidth = 26
	}
	if inputWidth > 96 {
		inputWidth = 96
	}
	m.urlInput.Width = inputWidth
	m.outputInput.Width = inputWidth
	m.addInput.Width = inputWidth

	helpWrap := 52
	if width < 130 {
		helpWrap = 48
	}
	if width < 110 {
		helpWrap = 40
	}
	if width < 95 {
		helpWrap = 34
	}
	if width < 78 {
		helpWrap = 30
	}
	m.setupHelpTxt = renderMarkdown(m.setupHelpMD, helpWrap)
}

func setupMarkdown() string {
	return `
# YouTube Downloader Setup

- Select **mode**: audio, video, or both
- Paste a **channel URL**, **playlist URL**, or **single video URL**
- Choose output directory, then press **Enter** on Start

## Keys

- Tab / Shift+Tab or j/k: move between controls
- h / l: cycle mode when mode row is focused
- Enter: continue / start
- ?: toggle help panel
- q: quit
`
}

func renderMarkdown(raw string, wrap int) string {
	if wrap < 24 {
		wrap = 24
	}
	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(wrap))
	if err != nil {
		return raw
	}
	out, err := renderer.Render(raw)
	if err != nil {
		return raw
	}
	return out
}

func bar(percent float64, width int) string {
	done := int(clamp(percent) * float64(width))
	return strings.Repeat("=", done) + strings.Repeat(".", width-done)
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func short(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}
	return value[:max-3] + "..."
}

func (m Model) findVisibleRowByVideoIndex(videoIndex int) int {
	for i, idx := range m.visibleRows {
		if idx == videoIndex {
			return i
		}
	}
	return -1
}

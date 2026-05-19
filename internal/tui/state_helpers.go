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
		m.totalVideos = event.TotalVideos
		m.videos = make([]videoState, len(event.Videos))
		for i, v := range event.Videos {
			m.videos[i] = videoState{video: v}
		}
		m.initDownloadTable()
	case downloader.EventVideoStart:
		m.currentIndex = event.VideoIndex
		if event.VideoIndex >= 0 {
			m.table.SetCursor(event.VideoIndex)
		}
	case downloader.EventVideoProgress:
		if event.VideoIndex < 0 || event.VideoIndex >= len(m.videos) {
			return
		}
		item := &m.videos[event.VideoIndex]
		item.speed = event.Speed
		item.eta = event.ETA
		item.percent = clamp(event.Percent)
		m.recomputeOverall()
		m.refreshTableRows()
	case downloader.EventVideoDone:
		if event.VideoIndex >= 0 && event.VideoIndex < len(m.videos) {
			item := &m.videos[event.VideoIndex]
			item.done = true
			item.percent = 1.0
			item.speed = ""
			item.eta = ""
		}
		m.completed++
		m.recomputeOverall()
		m.refreshTableRows()
	case downloader.EventVideoError:
		if event.VideoIndex >= 0 && event.VideoIndex < len(m.videos) {
			item := &m.videos[event.VideoIndex]
			item.hasError = true
			item.errorText = event.Message
		}
		m.errText = event.Message
		m.refreshTableRows()
	case downloader.EventLog:
		m.lastLog = strings.TrimSpace(event.Message)
	case downloader.EventFinished:
		m.overall = 1
		m.lastLog = fmt.Sprintf("Finished: %d/%d completed", m.completed, m.totalVideos)
		m.refreshTableRows()
	}
}

func (m *Model) focusCurrentField() {
	m.urlInput.Blur()
	m.outputInput.Blur()
	m.ytDLPInput.Blur()
	if m.focusIndex == 1 {
		m.urlInput.Focus()
	}
	if m.focusIndex == 2 {
		m.outputInput.Focus()
	}
	if m.focusIndex == 3 {
		m.ytDLPInput.Focus()
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

func setupMarkdown() string {
	return `
# YouTube Downloader Setup

- Select **mode**: audio, video, or both
- Paste a **channel URL** or **single video URL**
- Choose output directory, then press **Enter** on Start

## Keys

- Tab / Shift+Tab: move between controls
- Left / Right: cycle mode
- Enter: continue
- q: quit
`
}

func renderMarkdown(raw string) string {
	renderer, err := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(56))
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

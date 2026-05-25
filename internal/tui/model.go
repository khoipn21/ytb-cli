package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"youtube-channel-audio-downloader/internal/downloader"
)

type screen string

const (
	screenSetup    screen = "setup"
	screenDownload screen = "download"
)

type videoState struct {
	video     downloader.Video
	percent   float64
	speed     string
	eta       string
	done      bool
	hasError  bool
	errorText string
}

type eventMsg struct {
	event downloader.Event
	ok    bool
}

type startDownloadMsg struct {
	events  <-chan downloader.Event
	cancel  context.CancelFunc
	options downloader.Options
	err     error
}

type Model struct {
	screen       screen
	width        int
	height       int
	focusIndex   int
	modeIndex    int
	errText      string
	lastLog      string
	setupHelpMD  string
	setupHelpTxt string
	urlInput     textinput.Model
	outputInput  textinput.Model
	addInput     textinput.Model
	cancel       context.CancelFunc
	events       <-chan downloader.Event
	request      downloader.Options
	pending      []downloader.Options
	videos       []videoState
	table        table.Model
	tableCompact bool
	addingURL    bool
	addModeIndex int
	running      bool
	batchStart   int
	totalVideos  int
	currentIndex int
	completed    int
	overall      float64
}

func NewModel(config Config) Model {
	urlInput := newInput("YouTube URL", config.InitialURL, "https://www.youtube.com/@channel/videos or https://youtu.be/...")
	outputInput := newInput("Output directory", config.InitialOutput, "./downloads")
	addInput := newInput("Add URL", "", "Paste channel/video/playlist URL then Enter")
	modeIndex := 0
	initialMode := strings.ToLower(strings.TrimSpace(config.InitialMode))
	if initialMode == string(downloader.ModeVideo) {
		modeIndex = 1
	} else if initialMode == string(downloader.ModeBoth) {
		modeIndex = 2
	}
	m := Model{
		screen:       screenSetup,
		modeIndex:    modeIndex,
		urlInput:     urlInput,
		outputInput:  outputInput,
		addInput:     addInput,
		currentIndex: -1,
		setupHelpMD:  setupMarkdown(),
	}
	m.focusCurrentField()
	m.setupHelpTxt = renderMarkdown(m.setupHelpMD, 52)
	m.updateSetupDimensions()
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSetupDimensions()
		if m.screen == screenDownload {
			m.updateTableDimensions()
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
		return m, nil
	case startDownloadMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
			return m, nil
		}
		m.request = msg.options
		m.cancel = msg.cancel
		m.events = msg.events
		m.screen = screenDownload
		m.running = true
		m.errText = ""
		return m, waitForEvent(m.events)
	case eventMsg:
		if !msg.ok {
			m.running = false
			if len(m.pending) > 0 {
				next := m.pending[0]
				m.pending = m.pending[1:]
				m.lastLog = fmt.Sprintf("Starting queued target: %s", short(next.TargetURL, 80))
				return m, startDownloadCmd(next)
			}
			return m, nil
		}
		m.applyEvent(msg.event)
		if msg.event.Type == downloader.EventFinished {
			m.running = false
			if len(m.pending) > 0 {
				next := m.pending[0]
				m.pending = m.pending[1:]
				m.lastLog = fmt.Sprintf("Starting queued target: %s", short(next.TargetURL, 80))
				return m, startDownloadCmd(next)
			}
		}
		return m, waitForEvent(m.events)
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		}
		if m.screen == screenSetup {
			return m.updateSetup(msg)
		}
		return m.updateDownload(msg)
	}
	return m, nil
}

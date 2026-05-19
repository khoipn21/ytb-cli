package tui

import (
	"context"
	"strings"

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
	events <-chan downloader.Event
	cancel context.CancelFunc
	err    error
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
	ytDLPInput   textinput.Model
	cancel       context.CancelFunc
	events       <-chan downloader.Event
	request      downloader.Options
	videos       []videoState
	totalVideos  int
	currentIndex int
	completed    int
	overall      float64
}

func NewModel(config Config) Model {
	urlInput := newInput("YouTube URL", config.InitialURL, "https://www.youtube.com/@channel/videos or https://youtu.be/...")
	outputInput := newInput("Output directory", config.InitialOutput, "./downloads")
	ytDLPInput := newInput("yt-dlp executable", config.InitialYTDLP, "yt-dlp")
	modeIndex := 0
	if strings.EqualFold(strings.TrimSpace(config.InitialMode), string(downloader.ModeVideo)) {
		modeIndex = 1
	}
	m := Model{
		screen:       screenSetup,
		modeIndex:    modeIndex,
		urlInput:     urlInput,
		outputInput:  outputInput,
		ytDLPInput:   ytDLPInput,
		currentIndex: -1,
		setupHelpMD:  setupMarkdown(),
	}
	m.focusCurrentField()
	m.setupHelpTxt = renderMarkdown(m.setupHelpMD)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.setupHelpTxt = renderMarkdown(m.setupHelpMD)
		return m, nil
	case startDownloadMsg:
		if msg.err != nil {
			m.errText = msg.err.Error()
			return m, nil
		}
		m.cancel = msg.cancel
		m.events = msg.events
		m.screen = screenDownload
		m.errText = ""
		return m, waitForEvent(m.events)
	case eventMsg:
		if !msg.ok {
			return m, nil
		}
		m.applyEvent(msg.event)
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

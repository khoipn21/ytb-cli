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
	event  downloader.Event
	events <-chan downloader.Event
	runID  int
	ok     bool
}

type startDownloadMsg struct {
	events  <-chan downloader.Event
	cancel  context.CancelFunc
	options downloader.Options
	runID   int
	err     error
}

type Model struct {
	screen          screen
	width           int
	height          int
	focusIndex      int
	modeIndex       int
	errText         string
	lastLog         string
	setupHelpMD     string
	setupHelpTxt    string
	urlInput        textinput.Model
	outputInput     textinput.Model
	addInput        textinput.Model
	cancel          context.CancelFunc
	events          <-chan downloader.Event
	request         downloader.Options
	pending         []downloader.Options
	videos          []videoState
	table           table.Model
	tableCompact    bool
	showHelp        bool
	activePanel     downloadPanel
	filterIndex     int
	visibleRows     []int
	detailOpen      bool
	detailIndex     int
	addingURL       bool
	addModeIndex    int
	running         bool
	batchStart      int
	totalVideos     int
	currentIndex    int
	completed       int
	overall         float64
	setupIntroFrame int
	setupIdleFrame  int
	setupTickGen    int
	runID           int
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
		detailIndex:  -1,
		currentIndex: -1,
		setupHelpMD:  setupMarkdown(),
		setupTickGen: 1,
	}
	m.focusCurrentField()
	m.setupHelpTxt = renderMarkdown(m.setupHelpMD, 52)
	m.updateSetupDimensions()
	return m
}

func (m Model) Init() tea.Cmd { return setupAnimationTick(m.setupTickGen) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case setupAnimationTickMsg:
		if m.screen != screenSetup || msg.generation != m.setupTickGen {
			return m, nil
		}
		if m.setupIntroFrame < setupIntroFrameCount {
			m.setupIntroFrame++
		} else {
			m.setupIdleFrame++
		}
		return m, setupAnimationTick(m.setupTickGen)
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
		if msg.runID != m.runID {
			return m, nil
		}
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
		return m, waitForEvent(m.events, m.runID)
	case eventMsg:
		if m.screen != screenDownload || msg.runID != m.runID || msg.events != m.events {
			return m, nil
		}
		if !msg.ok {
			m.running = false
			if len(m.pending) > 0 {
				next := m.pending[0]
				m.pending = m.pending[1:]
				m.lastLog = fmt.Sprintf("Starting queued target: %s", short(next.TargetURL, 80))
				m.runID++
				return m, startDownloadCmd(next, m.runID)
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
				m.runID++
				return m, startDownloadCmd(next, m.runID)
			}
		}
		return m, waitForEvent(m.events, m.runID)
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

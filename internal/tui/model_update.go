package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.focusIndex == 1 || m.focusIndex == 2 {
		switch msg.String() {
		case "tab", "shift+tab", "up", "down", "enter":
		default:
			var cmd tea.Cmd
			if m.focusIndex == 1 {
				m.urlInput, cmd = m.urlInput.Update(msg)
			} else {
				m.outputInput, cmd = m.outputInput.Update(msg)
			}
			return m, cmd
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "tab", "down", "j":
		m.focusIndex = (m.focusIndex + 1) % 4
		m.focusCurrentField()
		return m, nil
	case "shift+tab", "up", "k":
		m.focusIndex = (m.focusIndex + 3) % 4
		m.focusCurrentField()
		return m, nil
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "left", "h":
		if m.focusIndex == 0 {
			m.modeIndex = (m.modeIndex + 2) % 3
			return m, nil
		}
	case "right", "l":
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
	return m, nil
}

func (m Model) updateDownload(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.addingURL {
		return m.updateAddURL(msg)
	}
	if m.detailOpen {
		switch msg.String() {
		case "q":
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "a":
			m.detailOpen = false
			m.activePanel = panelTable
			m.updateTableDimensions()
			m.addingURL = true
			m.addModeIndex = modeToIndex(m.request.Mode)
			m.addInput.SetValue("")
			m.addInput.Focus()
			m.errText = ""
			return m, nil
		case "esc", "enter":
			m.detailOpen = false
			m.activePanel = panelTable
			m.updateTableDimensions()
			return m, nil
		}
		return m, nil
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
	case "?":
		m.showHelp = !m.showHelp
		return m, nil
	case "left", "h":
		m.moveFilter(-1)
		return m, nil
	case "right", "l":
		m.moveFilter(1)
		return m, nil
	case "enter":
		if idx, ok := m.selectedVideoIndex(); ok {
			m.detailOpen = true
			m.detailIndex = idx
			m.activePanel = panelDetail
			m.updateTableDimensions()
		}
		return m, nil
	case "esc":
		if m.cancel != nil {
			m.cancel()
		}
		nextRunID := m.runID + 1
		nextTickGen := m.setupTickGen + 1
		m = NewModel(Config{
			InitialURL:    m.request.TargetURL,
			InitialMode:   string(m.request.Mode),
			InitialOutput: m.request.OutputDir,
		})
		m.runID = nextRunID
		m.setupTickGen = nextTickGen
		return m, setupAnimationTick(m.setupTickGen)
	}
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

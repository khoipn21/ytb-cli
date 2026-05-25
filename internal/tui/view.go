package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"youtube-channel-audio-downloader/internal/downloader"
)

func (m Model) View() string {
	if m.screen == screenSetup {
		return m.setupView()
	}
	return m.downloadView()
}

func (m Model) setupView() string {
	if m.width > 0 && m.height > 0 && (m.width < 80 || m.height < 24) {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Render("Terminal too small. Resize to at least 80x24.")
	}
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("4")).
		Padding(0, 1).
		Render("YouTube Downloader Setup")
	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Render("Prepare request, then press Enter on Start Download")
	pixelArt := setupPixelArt(m.setupIntroFrame, m.setupIdleFrame)

	formPanel := panelStyle(true).Render(m.renderSetupForm())
	lines := []string{header, subtitle, pixelArt, formPanel}
	if m.showHelp {
		lines = append(lines, panelStyle(false).Render(m.setupHelpTxt))
	}
	if strings.TrimSpace(m.errText) != "" {
		lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true).Render("Error: "+m.errText))
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("[tab] next field  [j/k] move  [h/l] mode  [enter] next/start  [?] help  [q] quit"))
	return strings.Join(lines, "\n")
}

func (m Model) renderSetupForm() string {
	startButton := lipgloss.NewStyle().Padding(0, 1).Render("[ Start Download ]")
	if m.focusIndex == 3 {
		startButton = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("2")).Padding(0, 1).Render("[ Start Download ]")
	}
	modeValue := strings.Join([]string{
		setupModeChip(0, m.modeIndex, "AUDIO", "31"),
		setupModeChip(1, m.modeIndex, "VIDEO", "127"),
		setupModeChip(2, m.modeIndex, "BOTH", "64"),
	}, " ")
	lines := []string{
		lipgloss.NewStyle().Bold(true).Render("Form"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("Field              | Value"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("-------------------+--------------------------------------------"),
		m.setupTableRow(0, "Mode", modeValue),
		m.setupTableRow(1, "Target URL", m.urlInput.View()),
		m.setupTableRow(-1, "Detected", renderTargetTypeTag(downloader.DetectTargetType(m.urlInput.Value()))),
		m.setupTableRow(2, "Output", m.outputInput.View()),
		m.setupTableRow(3, "Action", startButton),
	}
	return strings.Join(lines, "\n")
}

func (m Model) setupTableRow(index int, field, value string) string {
	label := lipgloss.NewStyle().Width(18).Render(field)
	if index == m.focusIndex {
		label = lipgloss.NewStyle().Width(18).Bold(true).Foreground(lipgloss.Color("15")).Render("► " + field)
	}
	return label + " | " + value
}

func setupModeChip(index, active int, label, color string) string {
	style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("248"))
	if index == active {
		style = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color(color))
	}
	return style.Render(label)
}

func (m Model) renderModeSelector() string {
	modeChip := func(index int, label, color string) string {
		style := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("248"))
		if m.modeIndex == index {
			style = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color(color))
		}
		return style.Render(strings.ToUpper(label))
	}
	prefix := lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Mode:")
	row := strings.Join([]string{
		prefix,
		modeChip(0, "audio", "31"),
		modeChip(1, "video", "127"),
		modeChip(2, "both", "64"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("(left/right to switch)"),
	}, " ")
	if m.focusIndex == 0 {
		return lipgloss.NewStyle().Bold(true).Render("● " + row)
	}
	return "○ " + row
}

func (m Model) focusLine(index int, value string) string {
	if m.focusIndex == index {
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render("● " + value)
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color("243")).Render("○ ") + value
}

func (m Model) addURLComposerView() string {
	chip := func(index int, label, color string) string {
		base := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("248"))
		if m.addModeIndex == index {
			base = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color(color))
		}
		return base.Render(strings.ToUpper(label))
	}
	modeRow := strings.Join([]string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Mode:"),
		chip(0, "audio", "31"),
		chip(1, "video", "127"),
		chip(2, "both", "64"),
	}, " ")
	body := strings.Join([]string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render("Add URL to Queue"),
		modeRow,
		m.addInput.View(),
		m.renderLinkTypeLine(m.addInput.Value()),
	}, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Render(body)
}

func (m Model) renderLinkTypeLine(rawURL string) string {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	return labelStyle.Render("Link Type: ") + renderTargetTypeTag(downloader.DetectTargetType(rawURL))
}

func renderTargetTypeTag(targetType downloader.TargetType) string {
	style := lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(lipgloss.Color("230"))
	switch targetType {
	case downloader.TargetTypeChannel:
		return style.Background(lipgloss.Color("33")).Render("CHANNEL")
	case downloader.TargetTypeVideo:
		return style.Background(lipgloss.Color("99")).Render("VIDEO")
	case downloader.TargetTypePlaylist:
		return style.Background(lipgloss.Color("35")).Render("PLAYLIST")
	default:
		return style.Background(lipgloss.Color("239")).Render("UNKNOWN")
	}
}

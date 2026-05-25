package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const setupIntroFrameCount = 11

type setupAnimationTickMsg struct {
	generation int
}

var ytbcliPixelRows = []string{
	"██    ██ ████████ ██████   ██████ ██      ██",
	" ██  ██     ██    ██   ██ ██      ██      ██",
	"  ████      ██    ██████  ██      ██      ██",
	"   ██       ██    ██   ██ ██      ██      ██",
	"   ██       ██    ██████   ██████ ███████ ██",
}

func setupPixelArt(introFrame, idleFrame int) string {
	return colorizePixelArt(setupPixelArtRaw(introFrame, idleFrame), idleFrame)
}

func setupPixelArtRaw(introFrame, idleFrame int) string {
	rows := make([]string, 0, len(ytbcliPixelRows)+2)
	rows = append(rows, revealRows(entranceFlashRows(introFrame), introFrame)...)
	rows = append(rows, "", "  ░ youtube terminal downloader ░")
	return strings.Join(rows, "\n")
}

func colorizePixelArt(raw string, idleFrame int) string {
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		if i == len(lines)-1 {
			lines[i] = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Render(line)
			continue
		}
		lines[i] = colorizePixelLine(line, i, idleFrame)
	}
	return strings.Join(lines, "\n")
}

func colorizePixelLine(line string, row, idleFrame int) string {
	palette := []lipgloss.Color{"39", "45", "81", "75", "111"}
	var out strings.Builder
	for col, char := range []rune(line) {
		style := lipgloss.NewStyle().Foreground(palette[(row+col/7+idleFrame/10)%len(palette)])
		switch char {
		case '▓':
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231"))
		case '▒':
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
		case '░':
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("218"))
		case ' ':
			out.WriteRune(char)
			continue
		}
		if char == '█' && shimmerPixel(row, col, idleFrame) {
			style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("231"))
		}
		out.WriteString(style.Render(string(char)))
	}
	return out.String()
}

func shimmerPixel(row, col, idleFrame int) bool {
	return (row*17+col*11+idleFrame*3)%31 == 0
}

func revealRows(rows []string, frame int) []string {
	if frame >= setupIntroFrameCount {
		return rows
	}
	maxWidth := 0
	for _, row := range rows {
		if len([]rune(row)) > maxWidth {
			maxWidth = len([]rune(row))
		}
	}
	visibleColumns := frame * maxWidth / setupIntroFrameCount
	revealed := make([]string, 0, len(rows))
	for _, row := range rows {
		runes := []rune(row)
		rowColumns := visibleColumns
		if rowColumns > len(runes) {
			rowColumns = len(runes)
		}
		revealed = append(revealed, string(runes[:rowColumns]))
	}
	return revealed
}

func entranceFlashRows(introFrame int) []string {
	rows := append([]string(nil), ytbcliPixelRows...)
	if introFrame >= setupIntroFrameCount || introFrame%3 != 0 {
		return rows
	}
	for i, row := range rows {
		runes := []rune(row)
		for j := range runes {
			if runes[j] == '█' {
				runes[j] = '▓'
			}
		}
		rows[i] = string(runes)
	}
	return rows
}

func maxPixelArtWidth(rows []string) int {
	maxWidth := 1
	for _, row := range rows {
		if width := len([]rune(row)); width > maxWidth {
			maxWidth = width
		}
	}
	return maxWidth
}

func setupAnimationTick(generation int) tea.Cmd {
	return tea.Tick(90*time.Millisecond, func(time.Time) tea.Msg {
		return setupAnimationTickMsg{generation: generation}
	})
}

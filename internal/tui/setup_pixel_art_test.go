package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestSetupPixelArtKeepsStableHeightAfterIntro(t *testing.T) {
	first := setupPixelArtRaw(setupIntroFrameCount, 1)
	second := setupPixelArtRaw(setupIntroFrameCount, 8)

	if got, want := lineCount(first), lineCount(second); got != want {
		t.Fatalf("line count changed during idle animation: got %d, want %d", got, want)
	}
}

func TestSetupPixelArtRevealExpandsTowardFullArt(t *testing.T) {
	firstFrame := setupPixelArtRaw(1, 0)
	fullFrame := setupPixelArtRaw(setupIntroFrameCount, 0)

	if len([]rune(firstFrame)) >= len([]rune(fullFrame)) {
		t.Fatalf("intro frame should be shorter than full frame")
	}
	if !strings.Contains(fullFrame, "youtube terminal downloader") {
		t.Fatalf("full frame should include the setup tagline")
	}
}

func TestSetupPixelArtRemovesSweepLineGlyphs(t *testing.T) {
	frame := setupPixelArtRaw(setupIntroFrameCount, 4)

	if strings.Contains(frame, "*") {
		t.Fatalf("idle frame should not include radiant sweep line markers")
	}
	if strings.ContainsAny(frame, "▓▒") {
		t.Fatalf("idle frame should not include glass shimmer sweep glyphs")
	}
}

func TestSetupPixelArtUsesANSIColors(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(previousProfile)
	})

	frame := setupPixelArt(setupIntroFrameCount, 4)

	if !strings.Contains(frame, "\x1b[") {
		t.Fatalf("rendered frame should include ANSI color sequences")
	}
}

func TestSetupPixelArtIdleShimmerChangesColorFrame(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(previousProfile)
	})

	first := setupPixelArt(setupIntroFrameCount, 1)
	second := setupPixelArt(setupIntroFrameCount, 2)

	if first == second {
		t.Fatalf("idle shimmer should change rendered color frames")
	}
	if setupPixelArtRaw(setupIntroFrameCount, 1) != setupPixelArtRaw(setupIntroFrameCount, 2) {
		t.Fatalf("idle shimmer should not change raw glyph layout")
	}
}

func lineCount(value string) int {
	if value == "" {
		return 0
	}
	return strings.Count(value, "\n") + 1
}

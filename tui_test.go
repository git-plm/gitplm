package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func TestOverlayCenter(t *testing.T) {
	base := strings.Repeat("abcdefghij\n", 4) + "abcdefghij"
	box := "XX\nYY"
	out := overlayCenter(base, box, 10, 5)
	lines := strings.Split(ansi.Strip(out), "\n")
	want := []string{"abcdefghij", "abcdXXghij", "abcdYYghij", "abcdefghij", "abcdefghij"}
	for i, w := range want {
		if lines[i] != w {
			t.Errorf("line %d: got %q, want %q", i, lines[i], w)
		}
	}
}

func TestOverlayCenterPadsShortBase(t *testing.T) {
	out := overlayCenter("short", "XX\nYY", 10, 5)
	lines := strings.Split(ansi.Strip(out), "\n")
	if len(lines) != 5 {
		t.Fatalf("got %d lines, want 5", len(lines))
	}
	if lines[1] != "    XX" || lines[2] != "    YY" {
		t.Errorf("box not centered on padded rows: %q %q", lines[1], lines[2])
	}
}

func TestReleasePopoverWrapsAndFits(t *testing.T) {
	m := &modelNew{width: 100, height: 30, mode: modeRelease}
	long := strings.Repeat("word ", 80) + "\n" + strings.Repeat("x", 250) + "\nshort"
	m.releaseLog = long

	w, h := m.releasePopoverSize()
	for _, line := range m.releaseLines() {
		if lipgloss.Width(line) > w-popoverFrameWidth {
			t.Errorf("wrapped line too wide (%d): %q", lipgloss.Width(line), line)
		}
	}

	m.releaseScroll = m.releaseMaxScroll()
	box := m.renderReleasePopover()
	boxLines := strings.Split(box, "\n")
	if len(boxLines) != h {
		t.Errorf("popover height %d, want %d", len(boxLines), h)
	}
	for _, line := range boxLines {
		if lipgloss.Width(line) != w {
			t.Errorf("popover line width %d, want %d: %q", lipgloss.Width(line), w, ansi.Strip(line))
		}
	}

	screen := overlayCenter(strings.Repeat("base\n", 29)+"base", box, m.width, m.height)
	for _, line := range strings.Split(screen, "\n") {
		if lipgloss.Width(line) > m.width {
			t.Errorf("screen line wider than terminal: %d", lipgloss.Width(line))
		}
	}

	m.scrollRelease(1000)
	if m.releaseScroll != m.releaseMaxScroll() {
		t.Errorf("scroll past end not clamped: %d", m.releaseScroll)
	}
	m.scrollRelease(-1000)
	if m.releaseScroll != 0 {
		t.Errorf("scroll before start not clamped: %d", m.releaseScroll)
	}
}

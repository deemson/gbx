package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// branchStyle is a grouping cue: the same name must always map to the same
// color, and distinct names should not collapse to the same one.
func TestBranchStyleDeterministic(t *testing.T) {
	a := branchStyle("main").GetForeground()
	b := branchStyle("main").GetForeground()
	if a != b {
		t.Fatalf("same name gave different colors: %v != %v", a, b)
	}
}

// The full-spectrum hash spreads names across the whole hue range, so distinct
// names should get distinct colors — the common main/trunk pair especially.
func TestBranchStyleMainTrunkDistinct(t *testing.T) {
	main := branchStyle("main").GetForeground()
	trunk := branchStyle("trunk").GetForeground()
	if main == trunk {
		t.Fatalf("main and trunk collided on color: %v", main)
	}
}

// renderHighlight must never alter the visible text — only layer styling on the
// matched runes. Stripping ANSI should round-trip the original string, whatever
// the run-coalescing does (asserted profile-independently via ansi.Strip).
func TestRenderHighlightPreservesText(t *testing.T) {
	s := "api-gateway"
	hl := map[int]bool{0: true, 1: true, 2: true, 8: true} // a, p, i, w
	got := ansi.Strip(renderHighlight(s, hl, lipgloss.NewStyle()))
	if got != s {
		t.Fatalf("visible text changed: %q != %q", got, s)
	}
	// Empty highlight set is the plain-render fast path; text must survive too.
	if got := ansi.Strip(renderHighlight(s, nil, lipgloss.NewStyle())); got != s {
		t.Fatalf("empty-highlight text changed: %q != %q", got, s)
	}
}

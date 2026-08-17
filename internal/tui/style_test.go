package tui

import (
	"fmt"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// assertSlot asserts that name resolved to branchPalette[idx]. Colors are
// compared by foreground, the only property the palette sets.
func assertSlot(t *testing.T, bc branchColors, name string, idx int) {
	t.Helper()
	got, want := bc.style(name).GetForeground(), branchPalette[idx].GetForeground()
	if got != want {
		t.Fatalf("%q got color %v, want branchPalette[%d] = %v", name, got, idx, want)
	}
}

// Rank drives the color: the most-checked-out branch takes the first palette
// slot, the next-most the second, and so on.
func TestBranchColorsRankByFrequency(t *testing.T) {
	bc := newBranchColors([]string{"main", "main", "main", "dev", "dev", "feat"})
	assertSlot(t, bc, "main", 0)
	assertSlot(t, bc, "dev", 1)
	assertSlot(t, bc, "feat", 2)
}

// Names tied on frequency are ordered by name hash, so they get distinct colors
// and the assignment doesn't depend on the order the repos were scanned in.
func TestBranchColorsTiebreakStable(t *testing.T) {
	forward := newBranchColors([]string{"main", "trunk"})
	reversed := newBranchColors([]string{"trunk", "main"})
	if forward.style("main").GetForeground() == forward.style("trunk").GetForeground() {
		t.Fatal("equally common names collided on one color")
	}
	for _, name := range []string{"main", "trunk"} {
		if forward.style(name).GetForeground() != reversed.style(name).GetForeground() {
			t.Fatalf("%q changed color when the input order changed", name)
		}
	}
}

// Past the palette's end the ranks cycle through the bright half only, leaving
// the plain hues to the most common branches. Counts here are all distinct so
// the ranking is fixed by frequency alone, with no hash tiebreak involved.
func TestBranchColorsWrapAtBrightBlue(t *testing.T) {
	var branches []string
	names := make([]string, 13)
	for rank := range names {
		names[rank] = fmt.Sprintf("branch-%d", rank)
		// Rank 0 is the most common, so counts descend with the rank.
		for range len(names) - rank {
			branches = append(branches, names[rank])
		}
	}
	bc := newBranchColors(branches)
	for rank := range 10 {
		assertSlot(t, bc, names[rank], rank)
	}
	assertSlot(t, bc, names[10], 5) // wraps to bright blue, not blue
	assertSlot(t, bc, names[11], 6)
	assertSlot(t, bc, names[12], 7)
}

// The same repo set must always produce the same colors — the grouping cue is
// worthless if it moves between renders.
func TestBranchColorsDeterministic(t *testing.T) {
	branches := []string{"main", "main", "dev", "feat", "release"}
	first, second := newBranchColors(branches), newBranchColors(branches)
	for _, name := range branches {
		if first.style(name).GetForeground() != second.style(name).GetForeground() {
			t.Fatalf("%q got a different color on rebuild", name)
		}
	}
}

// A name the resolver never counted renders unstyled rather than borrowing some
// other branch's color.
func TestBranchColorsUnknownNameUnstyled(t *testing.T) {
	bc := newBranchColors([]string{"main"})
	if got := bc.style("never-seen"); got.Render("x") != "x" {
		t.Fatalf("uncounted name was styled: %q", got.Render("x"))
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

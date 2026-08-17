package tui

import (
	"cmp"
	"crypto/md5"
	"encoding/binary"
	"maps"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
)

// Status signals are colored from the terminal's own ANSI 16-color palette
// (indices as strings), so the shades are theme-relative and adapt to light or
// dark backgrounds for free, without per-mode tuning.
var (
	colorGreen     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	colorRed       = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	colorYellow    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	colorMagenta   = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	colorCyan      = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	colorBrightRed = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	colorDim       = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// branchPalette is a curated set of the terminal's chromatic ANSI colors —
// theme-relative (they adapt to light/dark for free). Red and bright red are
// left out so a branch name never reads as an error. Order matters: the five
// plain hues come first and go to the most common branches, their bright
// variants after — see branchColors.
var branchPalette = []lipgloss.Style{
	lipgloss.NewStyle().Foreground(lipgloss.Color("4")),  // blue
	lipgloss.NewStyle().Foreground(lipgloss.Color("2")),  // green
	lipgloss.NewStyle().Foreground(lipgloss.Color("5")),  // magenta
	lipgloss.NewStyle().Foreground(lipgloss.Color("6")),  // cyan
	lipgloss.NewStyle().Foreground(lipgloss.Color("3")),  // yellow
	lipgloss.NewStyle().Foreground(lipgloss.Color("12")), // bright blue
	lipgloss.NewStyle().Foreground(lipgloss.Color("10")), // bright green
	lipgloss.NewStyle().Foreground(lipgloss.Color("13")), // bright magenta
	lipgloss.NewStyle().Foreground(lipgloss.Color("14")), // bright cyan
	lipgloss.NewStyle().Foreground(lipgloss.Color("11")), // bright yellow
}

// branchPaletteWrap is the index of bright blue — the slot the palette cycles
// back to once the ranks run past its end. Everything below it (the plain hues)
// stays reserved for the most common branches, so they never get taken by a
// one-off branch just because the list is long.
const branchPaletteWrap = 5

// branchColors assigns each branch name a palette slot by *rank*: the name
// checked out in the most repos takes branchPalette[0], the next-most [1], and
// so on. The color is a grouping cue that doubles as a frequency cue — the
// biggest cluster of identical rows always reads blue. A name it never counted
// resolves to the zero Style and renders unstyled; every rendered row's branch
// is counted, so that case means something upstream is wrong and should look
// like it.
type branchColors map[string]lipgloss.Style

// newBranchColors ranks branches by frequency, one vote per element (pass the
// checked-out branch of every repo, duplicates included). Names tied on count
// are ordered by their name hash, so the assignment is stable and never depends
// on the order the repos were scanned in.
func newBranchColors(branches []string) branchColors {
	counts := map[string]int{}
	for _, b := range branches {
		counts[b]++
	}
	names := slices.Collect(maps.Keys(counts))
	slices.SortFunc(names, func(a, b string) int {
		if c := cmp.Compare(counts[b], counts[a]); c != 0 {
			return c
		}
		if c := cmp.Compare(nameHash(a), nameHash(b)); c != 0 {
			return c
		}
		return cmp.Compare(a, b)
	})
	bc := make(branchColors, len(names))
	for rank, name := range names {
		bc[name] = branchPalette[paletteIndex(rank)]
	}
	return bc
}

func (bc branchColors) style(name string) lipgloss.Style {
	return bc[name]
}

// paletteIndex maps a frequency rank to a palette slot. Ranks within the
// palette map straight through; past its end they cycle through the bright half
// only, so with more than len(branchPalette) distinct names the colors repeat.
func paletteIndex(rank int) int {
	if rank < len(branchPalette) {
		return rank
	}
	return branchPaletteWrap + (rank-len(branchPalette))%(len(branchPalette)-branchPaletteWrap)
}

// nameHash is the tiebreak ordering key for names of equal frequency: the full
// 32-bit md5 prefix, so similar names still scatter rather than clumping into
// adjacent colors the way a lexicographic tiebreak would.
func nameHash(name string) uint32 {
	hash := md5.Sum([]byte(name))
	return binary.BigEndian.Uint32(hash[0:4])
}

// renderHighlight renders s over base, layering bold + underline on the runes
// whose starting byte offset is in hl (the filter-matched characters). The
// highlight is attribute-only, so base's foreground — default for names, the
// rank hue for branches — shows through on matched and unmatched runes alike.
// Contiguous runes of the same state are coalesced into one styled segment to
// keep the escape count down.
func renderHighlight(s string, hl map[int]bool, base lipgloss.Style) string {
	if len(hl) == 0 {
		return base.Render(s)
	}
	hi := base.Bold(true).Underline(true)
	var out strings.Builder
	var seg []rune
	segHL := false
	flush := func() {
		if len(seg) == 0 {
			return
		}
		st := base
		if segHL {
			st = hi
		}
		out.WriteString(st.Render(string(seg)))
		seg = seg[:0]
	}
	for bi, r := range s {
		if hl[bi] != segHL {
			flush()
			segHL = hl[bi]
		}
		seg = append(seg, r)
	}
	flush()
	return out.String()
}

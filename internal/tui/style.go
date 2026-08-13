package tui

import (
	"crypto/md5"
	"encoding/binary"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/lucasb-eyer/go-colorful"
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

// branchStyle hashes a branch name into a full-spectrum true-color hue, so the
// same name always reads the same color across rows — a grouping cue. Each of
// hue, saturation and lightness is drawn from a distinct 32-bit slice of the
// hash, so names scatter uniformly across the whole HSL space and even
// dissimilar names rarely land close. Unlike the status signals these RGB
// values are fixed, not theme-relative. Adapted from lazygit's author colors.
func branchStyle(name string) lipgloss.Style {
	hash := md5.Sum([]byte(name))
	c := colorful.Hsl(
		hashFrac(hash[0:4])*360.0,     // hue: full 0–360
		0.6+0.4*hashFrac(hash[4:8]),   // saturation: 0.6–1.0
		0.4+0.25*hashFrac(hash[8:12]), // lightness: 0.4–0.65
	)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex()))
}

// hashFrac maps four hash bytes to a fraction in [0, 1) using their full 32-bit
// entropy — a uniform spread, unlike a byte-sum modulo which clusters.
func hashFrac(b []byte) float64 {
	return float64(binary.BigEndian.Uint32(b)) / (float64(math.MaxUint32) + 1)
}

// renderHighlight renders s over base, layering bold + underline on the runes
// whose starting byte offset is in hl (the filter-matched characters). The
// highlight is attribute-only, so base's foreground — default for names, the
// hash hue for branches — shows through on matched and unmatched runes alike.
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

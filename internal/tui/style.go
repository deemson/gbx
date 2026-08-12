package tui

import (
	"crypto/md5"
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
// same name always reads the same color across rows — a grouping cue. The hash
// spreads names over the whole HSL space, so distinct branches get distinct
// colors with near-zero collisions. Unlike the status signals these RGB values
// are fixed, not theme-relative. Ported from lazygit's author-color generator.
func branchStyle(name string) lipgloss.Style {
	hash := md5.Sum([]byte(name))
	c := colorful.Hsl(
		randFloat(hash[0:4])*360.0,
		0.6+0.4*randFloat(hash[4:8]),
		0.4+randFloat(hash[8:12])*0.2,
	)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex()))
}

func randFloat(hash []byte) float64 {
	return float64(randInt(hash, 100)) / 100
}

func randInt(hash []byte, max int) int {
	sum := 0
	for _, b := range hash {
		sum = (sum + int(b)) % max
	}
	return sum
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

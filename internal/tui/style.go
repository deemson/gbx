package tui

import (
	"crypto/md5"
	"encoding/binary"

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
// left out so a branch name never reads as an error; the remaining five hues
// plus their bright variants give ten slots to spread names across.
var branchPalette = []lipgloss.Style{
	lipgloss.NewStyle().Foreground(lipgloss.Color("2")),  // green
	lipgloss.NewStyle().Foreground(lipgloss.Color("3")),  // yellow
	lipgloss.NewStyle().Foreground(lipgloss.Color("4")),  // blue
	lipgloss.NewStyle().Foreground(lipgloss.Color("5")),  // magenta
	lipgloss.NewStyle().Foreground(lipgloss.Color("6")),  // cyan
	lipgloss.NewStyle().Foreground(lipgloss.Color("10")), // bright green
	lipgloss.NewStyle().Foreground(lipgloss.Color("11")), // bright yellow
	lipgloss.NewStyle().Foreground(lipgloss.Color("12")), // bright blue
	lipgloss.NewStyle().Foreground(lipgloss.Color("13")), // bright magenta
	lipgloss.NewStyle().Foreground(lipgloss.Color("14")), // bright cyan
}

// branchStyle hashes a branch name to a fixed slot in branchPalette, so the
// same name always reads the same color across rows — a grouping cue. With only
// a handful of distinct branches on screen, collisions are rare.
func branchStyle(name string) lipgloss.Style {
	hash := md5.Sum([]byte(name))
	return branchPalette[binary.BigEndian.Uint32(hash[0:4])%uint32(len(branchPalette))]
}

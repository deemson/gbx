package tui

import "charm.land/lipgloss/v2"

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

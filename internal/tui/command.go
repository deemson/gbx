package tui

// cmdState is the result state of the last command run on a repo, rendered as a
// glyph in the row's result cell. Shared by all command bindings.
type cmdState int

const (
	cmdNone cmdState = iota
	cmdRunning
	cmdOK
	cmdFailed
)

func (c cmdState) glyph() string {
	switch c {
	case cmdRunning:
		return "⟳"
	case cmdOK:
		return "✓"
	case cmdFailed:
		return "✗"
	default:
		return ""
	}
}

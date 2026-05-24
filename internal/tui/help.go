package tui

import (
	"fmt"
	"strings"
)

// keyBinding is one row of the help overlay. This slice is the source of truth
// for the bindings.
type keyBinding struct {
	keys string
	desc string
}

var keyBindings = []keyBinding{
	{"type", "filter repos (see filter syntax)"},
	{"ctrl+1", "filter by name + branch (default)"},
	{"ctrl+2", "filter by name"},
	{"ctrl+3", "filter by branch"},
	{"↑ / ↓", "move the cursor (also ctrl+k / ctrl+j)"},
	{"tab", "toggle git command mode (runs on the filtered repos)"},
	{"enter", "command mode: run the command on the filtered repos"},
	{"ctrl+r", "refresh status of the filtered repos"},
	{"ctrl+g", "toggle this help"},
	{"esc", "command mode: cancel; list: quit"},
	{"ctrl+c", "quit"},
}

// filterSyntax documents the fzf-style filter DSL: space-separated terms ANDed
// together, each fuzzy by default unless anchored or negated.
var filterSyntax = []keyBinding{
	{"foo", "fuzzy match"},
	{"^foo", "starts with foo"},
	{"foo$", "ends with foo"},
	{"!foo", "exclude foo"},
}

func helpContent() string {
	var b strings.Builder
	b.WriteString("gbx — keys\n\n")
	for _, kb := range keyBindings {
		fmt.Fprintf(&b, "  %-8s  %s\n", kb.keys, kb.desc)
	}
	b.WriteString("\nfilter syntax (space = AND)\n\n")
	for _, fs := range filterSyntax {
		fmt.Fprintf(&b, "  %-8s  %s\n", fs.keys, fs.desc)
	}
	b.WriteString("\nesc: back\n")
	return b.String()
}

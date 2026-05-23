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
	{"type", "filter repos (fuzzy)"},
	{"ctrl+1", "toggle include / exclude filter"},
	{"ctrl+2", "filter by name + branch (default)"},
	{"ctrl+3", "filter by name"},
	{"ctrl+4", "filter by branch"},
	{"↑ / ↓", "move the cursor (also ctrl+k / ctrl+j)"},
	{"tab", "toggle git command mode (runs on the filtered repos)"},
	{"enter", "command mode: run the command on the filtered repos"},
	{"ctrl+r", "refresh status of the filtered repos"},
	{"ctrl+g", "toggle this help"},
	{"esc", "command mode: cancel; list: quit"},
	{"ctrl+c", "quit"},
}

func helpContent() string {
	var b strings.Builder
	b.WriteString("gbx — keys\n\n")
	for _, kb := range keyBindings {
		fmt.Fprintf(&b, "  %-8s  %s\n", kb.keys, kb.desc)
	}
	b.WriteString("\nesc: back\n")
	return b.String()
}

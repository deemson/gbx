package tui

import (
	"fmt"
	"strings"
)

// keyBinding is one row of the help overlay. This slice is the source of truth
// for the bindings; KEYMAP.md mirrors it.
type keyBinding struct {
	keys string
	desc string
}

var keyBindings = []keyBinding{
	{"type", "filter repos (fuzzy)"},
	{"↑ / ↓", "move the cursor (also ctrl+k / ctrl+j)"},
	{"enter", "drill into the repo under the cursor"},
	{"ctrl+p", "pull the filtered repos"},
	{"ctrl+o", "checkout a branch on the filtered repos"},
	{"ctrl+r", "refresh status of the filtered repos"},
	{"ctrl+g", "toggle this help"},
	{"esc", "back, or quit from the list"},
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

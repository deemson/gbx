package tui

import (
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
)

// The command line wears a free-form costume over a fixed grammar:
//
//	checkout <ref>      checkout -b <name>      fetch      pull
//
// Autocomplete is position-aware: token 0 completes the command word; both
// checkout arg slots complete from the union of every branch across the visible
// repos ("-b" is also offered in the <ref> slot). A picked branch absent from
// some repos just errors on those rows. fetch/pull take no argument.

var commandWords = []string{"checkout", "fetch", "pull"}

// splitActive divides the command line into the completed head (keeping its
// trailing space) and the active token being typed at the end. The active token
// is what autocomplete filters on; the head determines which slot we're in.
func splitActive(line string) (head, active string) {
	if i := strings.LastIndex(line, " "); i >= 0 {
		return line[:i+1], line[i+1:]
	}
	return "", line
}

// candidates returns the full suggestion set for the slot identified by the
// already-completed tokens (head), before filtering by the active token.
func (m model) candidates(head []string) []string {
	switch {
	case len(head) == 0:
		return commandWords
	case len(head) == 1 && head[0] == "checkout":
		return append([]string{"-b"}, m.visibleBranches()...)
	case len(head) == 2 && head[0] == "checkout" && head[1] == "-b":
		return m.visibleBranches()
	default:
		return nil
	}
}

// visibleBranches returns every distinct branch across the repos passing the
// filter, deduped and sorted.
func (m model) visibleBranches() []string {
	seen := map[string]bool{}
	var out []string
	for _, r := range m.matched() {
		for _, b := range r.branches {
			if !seen[b] {
				seen[b] = true
				out = append(out, b)
			}
		}
	}
	sort.Strings(out)
	return out
}

// suggestionsFor returns the suggestions for the current command line: the
// slot's candidates ranked by fuzzy match against the active token (or all of
// them, in order, when nothing is typed yet for that token).
func (m model) suggestionsFor(line string) []string {
	head, active := splitActive(line)
	cands := m.candidates(strings.Fields(head))
	if active == "" {
		return cands
	}
	matches := fuzzy.Find(active, cands)
	out := make([]string, len(matches))
	for i, mt := range matches {
		out[i] = mt.Str
	}
	return out
}

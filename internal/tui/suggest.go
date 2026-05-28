package tui

import "github.com/sahilm/fuzzy"

// fuzzyPick ranks cands by fuzzy match against pattern and returns the matched
// strings in best-first order. Used by the c prompt to narrow branch
// suggestions as the user types.
func fuzzyPick(pattern string, cands []string) []string {
	matches := fuzzy.Find(pattern, cands)
	out := make([]string, len(matches))
	for i, m := range matches {
		out[i] = m.Str
	}
	return out
}

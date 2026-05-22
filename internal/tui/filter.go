package tui

import "github.com/sahilm/fuzzy"

// rankFilter returns the indexes of names matching pattern as a fzf-style fuzzy
// subsequence, ranked best-match-first. An empty pattern matches every name, in
// the original order.
func rankFilter(pattern string, names []string) []int {
	if pattern == "" {
		idx := make([]int, len(names))
		for i := range names {
			idx[i] = i
		}
		return idx
	}
	matches := fuzzy.Find(pattern, names)
	idx := make([]int, len(matches))
	for i, m := range matches {
		idx[i] = m.Index
	}
	return idx
}

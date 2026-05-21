package tui

import "strings"

// fuzzyMatch reports whether pattern occurs in target as a case-insensitive
// subsequence (fzf-style). An empty pattern matches everything.
func fuzzyMatch(pattern, target string) bool {
	if pattern == "" {
		return true
	}
	p := []rune(strings.ToLower(pattern))
	t := []rune(strings.ToLower(target))
	pi := 0
	for ti := 0; ti < len(t) && pi < len(p); ti++ {
		if t[ti] == p[pi] {
			pi++
		}
	}
	return pi == len(p)
}

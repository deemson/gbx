package tui

import "testing"

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		pattern, target string
		want            bool
	}{
		{"", "anything", true},
		{"api", "api-gateway", true},
		{"agw", "api-gateway", true},     // subsequence, not contiguous
		{"API", "api-gateway", true},     // case-insensitive
		{"gateway", "api-gateway", true}, // contiguous tail
		{"xyz", "api-gateway", false},
		{"wag", "api-gateway", false}, // characters present but out of order
	}
	for _, c := range cases {
		if got := fuzzyMatch(c.pattern, c.target); got != c.want {
			t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", c.pattern, c.target, got, c.want)
		}
	}
}

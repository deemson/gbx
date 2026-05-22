package tui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// matchedNames runs the filter over names and returns the matched names in
// ranked order, for concise assertions.
func matchedNames(pattern string, names []string) []string {
	idx := rankFilter(pattern, names)
	out := make([]string, len(idx))
	for i, j := range idx {
		out[i] = names[j]
	}
	return out
}

func TestRankFilterMembership(t *testing.T) {
	names := []string{"api-gateway"}
	cases := []struct {
		pattern string
		want    []string
	}{
		{"", []string{"api-gateway"}},        // empty matches everything
		{"api", []string{"api-gateway"}},     // contiguous head
		{"agw", []string{"api-gateway"}},     // subsequence, not contiguous
		{"API", []string{"api-gateway"}},     // case-insensitive
		{"gateway", []string{"api-gateway"}}, // contiguous tail
		{"xyz", []string{}},                  // no match
		{"wag", []string{}},                  // chars present but out of order
	}
	for _, c := range cases {
		require.Equal(t, c.want, matchedNames(c.pattern, names), "pattern %q", c.pattern)
	}
}

func TestRankFilterRanksBestFirst(t *testing.T) {
	// "api" matches both, but the contiguous head match outranks the trailing one.
	names := []string{"legacy-api", "api-gateway"}
	require.Equal(t, []string{"api-gateway", "legacy-api"}, matchedNames("api", names))
}

func TestRankFilterEmptyKeepsOrder(t *testing.T) {
	names := []string{"charlie", "alpha", "bravo"}
	require.Equal(t, []int{0, 1, 2}, rankFilter("", names))
}

package tui

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// pick returns the names at the given indexes, for concise assertions.
func pick(names []string, idx []int) []string {
	out := make([]string, len(idx))
	for i, j := range idx {
		out[i] = names[j]
	}
	return out
}

// matchedNames runs a name-only include filter and returns matched names in
// ranked order.
func matchedNames(pattern string, names []string) []string {
	return pick(names, rankFilter(pattern, names, make([]string, len(names)), fieldName, polarityInclude))
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
	require.Equal(t, []int{0, 1, 2}, rankFilter("", names, make([]string, 3), fieldNameBranch, polarityInclude))
}

func TestRankFilterNameBranchMatchesEitherField(t *testing.T) {
	names := []string{"api-gateway", "auth-service"}
	branches := []string{"develop", "main"}
	// "main" matches no name but the second repo's branch.
	got := pick(names, rankFilter("main", names, branches, fieldNameBranch, polarityInclude))
	require.Equal(t, []string{"auth-service"}, got)
}

func TestRankFilterNameBranchSumsScores(t *testing.T) {
	// "main" matches repo A's branch only; repo B's name AND branch. B's summed
	// score floats it above A.
	names := []string{"service", "main-app"}
	branches := []string{"main", "main"}
	got := pick(names, rankFilter("main", names, branches, fieldNameBranch, polarityInclude))
	require.Equal(t, []string{"main-app", "service"}, got)
}

func TestRankFilterBranchOnly(t *testing.T) {
	names := []string{"main-app", "auth-service"}
	branches := []string{"develop", "main"}
	// Field is branch, so the name "main-app" is ignored; only the branch matches.
	got := pick(names, rankFilter("main", names, branches, fieldBranch, polarityInclude))
	require.Equal(t, []string{"auth-service"}, got)
}

func TestRankFilterUnloadedBranchNeverMatches(t *testing.T) {
	names := []string{"api-gateway"}
	branches := []string{""} // status not loaded yet
	got := rankFilter("main", names, branches, fieldBranch, polarityInclude)
	require.Empty(t, got)
}

func TestRankFilterExcludeKeepsNonMatches(t *testing.T) {
	names := []string{"api-gateway", "auth-service", "billing"}
	branches := make([]string, 3)
	// Exclude name matches of "api": the matching repo drops, the rest stay in
	// name order.
	got := pick(names, rankFilter("api", names, branches, fieldName, polarityExclude))
	require.Equal(t, []string{"auth-service", "billing"}, got)
}

func TestRankFilterExcludeEmptyShowsAll(t *testing.T) {
	names := []string{"a", "b", "c"}
	branches := make([]string, 3)
	got := rankFilter("", names, branches, fieldNameBranch, polarityExclude)
	require.Equal(t, []int{0, 1, 2}, got)
}

func TestRankFilterExcludeNameBranchDropsEitherFieldMatch(t *testing.T) {
	names := []string{"api-gateway", "auth-service"}
	branches := []string{"develop", "main"}
	// Exclude "main": the second repo matches via its branch, so only the first
	// survives.
	got := pick(names, rankFilter("main", names, branches, fieldNameBranch, polarityExclude))
	require.Equal(t, []string{"api-gateway"}, got)
}

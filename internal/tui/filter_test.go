package tui

import (
	"sort"
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

// matchedNames runs a name-only filter and returns matched names in ranked order.
func matchedNames(pattern string, names []string) []string {
	return pick(names, rankFilter(pattern, names, make([]string, len(names)), fieldName))
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
	require.Equal(t, []int{0, 1, 2}, rankFilter("", names, make([]string, 3), fieldNameBranch))
}

func TestRankFilterNameBranchMatchesEitherField(t *testing.T) {
	names := []string{"api-gateway", "auth-service"}
	branches := []string{"develop", "main"}
	// "main" matches no name but the second repo's branch.
	got := pick(names, rankFilter("main", names, branches, fieldNameBranch))
	require.Equal(t, []string{"auth-service"}, got)
}

func TestRankFilterNameBranchSumsScores(t *testing.T) {
	// "main" matches repo A's branch only; repo B's name AND branch. B's summed
	// score floats it above A.
	names := []string{"service", "main-app"}
	branches := []string{"main", "main"}
	got := pick(names, rankFilter("main", names, branches, fieldNameBranch))
	require.Equal(t, []string{"main-app", "service"}, got)
}

func TestRankFilterBranchOnly(t *testing.T) {
	names := []string{"main-app", "auth-service"}
	branches := []string{"develop", "main"}
	// Field is branch, so the name "main-app" is ignored; only the branch matches.
	got := pick(names, rankFilter("main", names, branches, fieldBranch))
	require.Equal(t, []string{"auth-service"}, got)
}

func TestRankFilterUnloadedBranchNeverMatches(t *testing.T) {
	names := []string{"api-gateway"}
	branches := []string{""} // status not loaded yet
	got := rankFilter("main", names, branches, fieldBranch)
	require.Empty(t, got)
}

func TestRankFilterPrefixAnchorIsExact(t *testing.T) {
	names := []string{"api-gateway", "legacy-api"}
	// ^api keeps only the repo whose name starts with "api".
	require.Equal(t, []string{"api-gateway"}, matchedNames("^api", names))
	// ^agw must NOT match api-gateway: the prefix anchor is exact, not fuzzy.
	require.Equal(t, []string{}, matchedNames("^agw", names))
}

func TestRankFilterSuffixAnchor(t *testing.T) {
	names := []string{"api-gateway", "legacy-api"}
	require.Equal(t, []string{"legacy-api"}, matchedNames("api$", names))
}

func TestRankFilterEqualsAnchor(t *testing.T) {
	names := []string{"api", "api-gateway"}
	// ^api$ is equality: "api-gateway" only starts with "api", so just "api".
	require.Equal(t, []string{"api"}, matchedNames("^api$", names))
}

func TestRankFilterNegateIsExactSubstring(t *testing.T) {
	names := []string{"api-gateway", "billing"}
	// !api drops the repo containing "api"; substring, not fuzzy.
	require.Equal(t, []string{"billing"}, matchedNames("!api", names))
}

func TestRankFilterAndsTerms(t *testing.T) {
	names := []string{"api-gateway", "legacy-api", "billing"}
	// "api !legacy": fuzzy-api AND not-legacy → only api-gateway survives.
	require.Equal(t, []string{"api-gateway"}, matchedNames("api !legacy", names))
}

func TestRankFilterNegateNameBranchMissesBothFields(t *testing.T) {
	names := []string{"api-gateway", "auth-service"}
	branches := []string{"develop", "main"}
	// !main drops the repo whose branch is main, even though no name has "main".
	got := pick(names, rankFilter("!main", names, branches, fieldNameBranch))
	require.Equal(t, []string{"api-gateway"}, got)
}

func TestRankFilterDegenerateTermsIgnored(t *testing.T) {
	names := []string{"api-gateway", "billing"}
	// A half-typed "!" (lone operator) is dropped: "api !" behaves like "api".
	require.Equal(t, matchedNames("api", names), matchedNames("api !", names))
	// Only lone operators => no effective term => everything matches in order.
	require.Equal(t, []int{0, 1}, rankFilter("! ^ $", names, make([]string, 2), fieldName))
}

func TestRankFilterNoFuzzyKeepsInputOrder(t *testing.T) {
	names := []string{"api-two", "api-one"} // intentionally not sorted
	// ^api matches both; with no fuzzy term every survivor scores 0, so they keep
	// input order (which the model feeds name-sorted).
	got := pick(names, rankFilter("^api", names, make([]string, 2), fieldName))
	require.Equal(t, []string{"api-two", "api-one"}, got)
}

// positions runs the pattern's terms against s and returns the highlighted byte
// offsets sorted, for concise assertions (offsets == char indices for ASCII).
func positions(pattern, s string) []int {
	hl := matchPositions(parseTerms(pattern), s)
	out := make([]int, 0, len(hl))
	for o := range hl {
		out = append(out, o)
	}
	sort.Ints(out)
	return out
}

func TestMatchPositions(t *testing.T) {
	cases := []struct {
		pattern string
		s       string
		want    []int
	}{
		{"", "api-gateway", []int{}},                                              // empty: nothing lit
		{"api", "api-gateway", []int{0, 1, 2}},                                    // fuzzy, contiguous head
		{"agw", "api-gateway", []int{0, 4, 8}},                                    // fuzzy subsequence: a, g, w
		{"API", "api-gateway", []int{0, 1, 2}},                                    // case-insensitive
		{"^api", "api-gateway", []int{0, 1, 2}},                                   // prefix range
		{"way$", "api-gateway", []int{8, 9, 10}},                                  // suffix range
		{"^api-gateway$", "api-gateway", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}, // equals: whole
		{"!api", "api-gateway", []int{}},                                          // negated: nothing lit
		{"^api way$", "api-gateway", []int{0, 1, 2, 8, 9, 10}},                    // union across terms
		{"xyz", "api-gateway", []int{}},                                           // no match: nothing lit
	}
	for _, c := range cases {
		require.Equal(t, c.want, positions(c.pattern, c.s), "pattern %q", c.pattern)
	}
}

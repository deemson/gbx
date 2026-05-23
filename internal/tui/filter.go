package tui

import (
	"sort"

	"github.com/sahilm/fuzzy"
)

// filterField selects which repo text the pattern matches against.
type filterField int

const (
	fieldNameBranch filterField = iota // ctrl+2, default: name OR branch
	fieldName                          // ctrl+3
	fieldBranch                        // ctrl+4
)

// filterPolarity selects whether the pattern keeps matches or non-matches.
type filterPolarity int

const (
	polarityInclude filterPolarity = iota // default
	polarityExclude                       // ctrl+1
)

// rankFilter returns the indexes into the parallel names/branches slices that
// survive the filter, ordered for display. branches[i] may be "" when a repo's
// status has not loaded yet; an empty "" never matches a non-empty pattern.
//
// An empty pattern returns every index in original order, regardless of field
// or polarity. Otherwise the include-set is the fzf-style fuzzy match over the
// selected field(s): in fieldNameBranch a repo is in the set if name OR branch
// matches, ranked by the summed score (so a both-fields match outranks a
// single-field one). polarityExclude returns the complement of that set, in
// original (name) order, since non-matches have no score.
func rankFilter(pattern string, names, branches []string, field filterField, polarity filterPolarity) []int {
	if pattern == "" {
		return seq(len(names))
	}
	scores := includeScores(pattern, names, branches, field)
	if polarity == polarityExclude {
		return complement(scores, len(names))
	}
	return rankByScore(scores)
}

// includeScores maps each matched index to its summed fuzzy score across the
// selected field(s). A repo missing from the map did not match.
func includeScores(pattern string, names, branches []string, field filterField) map[int]int {
	scores := make(map[int]int)
	add := func(data []string) {
		for _, m := range fuzzy.Find(pattern, data) {
			scores[m.Index] += m.Score
		}
	}
	switch field {
	case fieldName:
		add(names)
	case fieldBranch:
		add(branches)
	default: // fieldNameBranch
		add(names)
		add(branches)
	}
	return scores
}

// rankByScore returns the matched indexes ordered by score (best first), with
// ties broken by ascending index for a stable order.
func rankByScore(scores map[int]int) []int {
	idx := make([]int, 0, len(scores))
	for i := range scores {
		idx = append(idx, i)
	}
	sort.Slice(idx, func(a, b int) bool {
		if scores[idx[a]] != scores[idx[b]] {
			return scores[idx[a]] > scores[idx[b]]
		}
		return idx[a] < idx[b]
	})
	return idx
}

// complement returns the indexes in [0,n) absent from scores, in ascending order.
func complement(scores map[int]int, n int) []int {
	idx := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if _, matched := scores[i]; !matched {
			idx = append(idx, i)
		}
	}
	return idx
}

// seq returns 0..n-1.
func seq(n int) []int {
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	return idx
}

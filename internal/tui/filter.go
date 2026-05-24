package tui

import (
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
)

// filterField selects which repo text the pattern matches against.
type filterField int

const (
	fieldNameBranch filterField = iota // ctrl+1, default: name OR branch
	fieldName                          // ctrl+2
	fieldBranch                        // ctrl+3
)

// termKind is how a parsed filter term matches: fuzzy (the bare default) or one
// of the exact anchored forms.
type termKind int

const (
	termFuzzy  termKind = iota // bare: fzf-style subsequence, scored
	termPrefix                 // ^foo: starts-with (exact)
	termSuffix                 // foo$: ends-with (exact)
	termEquals                 // ^foo$: equals (exact)
)

// term is one space-separated unit of the filter line. negate flips membership
// (a "!"-prefixed term keeps repos the body does NOT match). Only positive bare
// terms contribute to ranking; everything else only gates membership.
type term struct {
	kind   termKind
	body   string
	negate bool
}

// parseTerms splits the pattern on whitespace and classifies each unit. A term
// is fuzzy by default; a leading "!" negates, a leading "^" anchors the prefix,
// a trailing "$" anchors the suffix (both anchors => equality). Operators are
// positional — "^" matters only at the start, "$" only at the end. Degenerate
// terms (empty body after stripping operators) are dropped, so a half-typed "!"
// or "^" never filters everything away mid-keystroke.
func parseTerms(pattern string) []term {
	fields := strings.Fields(pattern)
	terms := make([]term, 0, len(fields))
	for _, f := range fields {
		t := term{kind: termFuzzy}
		if strings.HasPrefix(f, "!") {
			t.negate = true
			f = f[1:]
		}
		prefix := strings.HasPrefix(f, "^")
		if prefix {
			f = f[1:]
		}
		suffix := strings.HasSuffix(f, "$")
		if suffix {
			f = f[:len(f)-1]
		}
		switch {
		case prefix && suffix:
			t.kind = termEquals
		case prefix:
			t.kind = termPrefix
		case suffix:
			t.kind = termSuffix
		}
		t.body = f
		if t.body == "" {
			continue // degenerate: lone !, ^, $, ^$
		}
		terms = append(terms, t)
	}
	return terms
}

// rankFilter returns the indexes into the parallel names/branches slices that
// survive the filter, ordered for display. branches[i] may be "" when a repo's
// status has not loaded yet; an empty "" never matches a non-empty term.
//
// The pattern is an fzf-style DSL: space-separated terms, all ANDed. A repo
// survives only if every term passes. Survivors are ranked by the summed fuzzy
// score of the positive bare terms (anchored and negated terms score 0), ties
// and no-fuzzy patterns falling back to input order. An empty pattern (or one of
// only degenerate terms) returns every index in original order.
func rankFilter(pattern string, names, branches []string, field filterField) []int {
	terms := parseTerms(pattern)
	if len(terms) == 0 {
		return seq(len(names))
	}
	type scored struct{ idx, score int }
	var matched []scored
	for i := range names {
		if score, ok := evalRepo(terms, names[i], branches[i], field); ok {
			matched = append(matched, scored{i, score})
		}
	}
	// Stable sort over a slice built in ascending index order: ties (and the
	// all-zero no-fuzzy case) keep input order, which is name order in practice.
	sort.SliceStable(matched, func(a, b int) bool {
		return matched[a].score > matched[b].score
	})
	idx := make([]int, len(matched))
	for i, m := range matched {
		idx[i] = m.idx
	}
	return idx
}

// evalRepo runs every term against one repo (AND): the repo is in the set only
// if all terms pass. The returned score sums the fuzzy scores of the positive
// bare terms; anchored and negated terms contribute nothing.
func evalRepo(terms []term, name, branch string, field filterField) (int, bool) {
	score := 0
	for _, t := range terms {
		matched, s := matchTerm(t, name, branch, field)
		pass := matched
		if t.negate {
			pass = !matched
		}
		if !pass {
			return 0, false
		}
		if !t.negate {
			score += s
		}
	}
	return score, true
}

// matchTerm reports whether term t matches the repo's active field(s) and, for a
// positive bare term, the summed match score. In name+branch the term may match
// either field; the score sums both, so a repo matching in both outranks a
// single-field match. The returned bool is the raw match — evalRepo inverts it
// for negated terms.
func matchTerm(t term, name, branch string, field filterField) (bool, int) {
	switch field {
	case fieldName:
		return matchString(t, name)
	case fieldBranch:
		return matchString(t, branch)
	default:
		nm, ns := matchString(t, name)
		bm, bs := matchString(t, branch)
		return nm || bm, ns + bs
	}
}

// matchString reports whether term t matches a single string and its fuzzy score
// (0 for the exact kinds). Comparison is case-insensitive. A negated bare term
// tests exact substring containment, not fuzzy — fuzzy negation would exclude
// almost everything (any subsequence) and is unpredictable.
func matchString(t term, s string) (bool, int) {
	ls := strings.ToLower(s)
	lb := strings.ToLower(t.body)
	switch t.kind {
	case termPrefix:
		return strings.HasPrefix(ls, lb), 0
	case termSuffix:
		return strings.HasSuffix(ls, lb), 0
	case termEquals:
		return ls == lb, 0
	default: // bare term
		if t.negate {
			return strings.Contains(ls, lb), 0
		}
		ms := fuzzy.Find(t.body, []string{s})
		if len(ms) == 0 {
			return false, 0
		}
		return true, ms[0].Score
	}
}

// seq returns 0..n-1.
func seq(n int) []int {
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	return idx
}

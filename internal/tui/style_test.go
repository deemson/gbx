package tui

import "testing"

// branchStyle is a grouping cue: the same name must always map to the same
// color, and distinct names should not collapse to the same one.
func TestBranchStyleDeterministic(t *testing.T) {
	a := branchStyle("main").GetForeground()
	b := branchStyle("main").GetForeground()
	if a != b {
		t.Fatalf("same name gave different colors: %v != %v", a, b)
	}
}

// The curated palette has only six slots, so collisions between arbitrary names
// are expected — but the common main/trunk pair must read as distinct colors.
func TestBranchStyleMainTrunkDistinct(t *testing.T) {
	main := branchStyle("main").GetForeground()
	trunk := branchStyle("trunk").GetForeground()
	if main == trunk {
		t.Fatalf("main and trunk collided on color: %v", main)
	}
}

// branchStyle must only ever return a color from the curated palette.
func TestBranchStyleInPalette(t *testing.T) {
	got := branchStyle("anything").GetForeground()
	for _, s := range branchPalette {
		if s.GetForeground() == got {
			return
		}
	}
	t.Fatalf("color %v is not in branchPalette", got)
}

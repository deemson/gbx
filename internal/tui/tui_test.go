package tui

import "testing"

func TestEmptyShowsNoRepos(t *testing.T) {
	tp := runTestProgram(t, t.TempDir())
	tp.waitForContent("no repos")
}

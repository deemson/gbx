package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/require"
)

func mkRepo(t *testing.T, dir, name string) gitest.Repo {
	t.Helper()
	p := filepath.Join(dir, name)
	require.NoError(t, os.Mkdir(p, 0755))
	return gitest.Init(t, p)
}

func TestEmptyShowsNoRepos(t *testing.T) {
	tp := runTestProgram(t, t.TempDir())
	tp.waitForContent("no repos")
}

func TestSingleRepoAppears(t *testing.T) {
	dir := t.TempDir()
	mkRepo(t, dir, "myrepo")
	tp := runTestProgram(t, dir)
	tp.waitForContent("myrepo")
}

func TestMultipleReposAllAppear(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"zebra", "apple", "monkey"} {
		mkRepo(t, dir, name)
	}
	tp := runTestProgram(t, dir)
	tp.waitForContent("apple", "monkey", "zebra")
}

func TestNonRepoDirsIgnored(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "plain-dir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "loose-file"), []byte("x"), 0644))
	mkRepo(t, dir, "realrepo")

	tp := runTestProgram(t, dir)
	tp.waitForContent("realrepo")

	time.Sleep(200 * time.Millisecond)
	out := tp.out.String()
	require.NotContains(t, out, "plain-dir")
	require.NotContains(t, out, "loose-file")
}

func TestRepoShowsCleanState(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "withcommit")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")
	branch := repo.BranchShowCurrent()

	tp := runTestProgram(t, dir)
	tp.waitForContent("withcommit", branch, "↑0 ↓0", "clean")
}

func TestRepoShowsChangedCount(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "dirty")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("a", "1")
	repo.Commit("init")
	repo.WriteFileAdd("b", "2") // staged, uncommitted

	tp := runTestProgram(t, dir)
	tp.waitForContent("dirty", "1 changed")
}

func TestFilterExcludingAllShowsNoMatches(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"apple", "banana"} {
		mkRepo(t, dir, name)
	}
	tp := runTestProgram(t, dir)
	tp.waitForContent("apple", "banana")

	tp.send("zzzz")
	tp.waitForContent("no matches")
}

package tui2

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/require"
)

func TestEmptyDirShowsDiscovering(t *testing.T) {
	tp := runTestProgram(t, t.TempDir())
	tp.waitForContent("discovering repos")
}

func TestSingleRepoAppears(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "myrepo")
	require.NoError(t, os.Mkdir(repoDir, 0755))
	gitest.Init(t, repoDir)

	tp := runTestProgram(t, dir)
	tp.waitForContent("myrepo")
}

func TestMultipleReposAllAppear(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"zebra", "apple", "monkey"} {
		repoDir := filepath.Join(dir, name)
		require.NoError(t, os.Mkdir(repoDir, 0755))
		gitest.Init(t, repoDir)
	}

	tp := runTestProgram(t, dir)
	tp.waitForContent("apple", "monkey", "zebra")
}

func TestNonRepoDirsIgnored(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "plain-dir"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "loose-file"), []byte("x"), 0644))

	repoDir := filepath.Join(dir, "realrepo")
	require.NoError(t, os.Mkdir(repoDir, 0755))
	gitest.Init(t, repoDir)

	tp := runTestProgram(t, dir)
	tp.waitForContent("realrepo")

	time.Sleep(200 * time.Millisecond)
	out := tp.out.String()
	require.NotContains(t, out, "plain-dir")
	require.NotContains(t, out, "loose-file")
}

func TestRepoWithCommitShowsBranchAndZeroDiff(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "withcommit")
	require.NoError(t, os.Mkdir(repoDir, 0755))
	repo := gitest.Init(t, repoDir)
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "data")
	repo.Commit("initial")

	branch := repo.BranchShowCurrent()

	tp := runTestProgram(t, dir)
	tp.waitForContent("withcommit", branch, "+0 -0")
}

func TestRepoWithStagedChangesShowsDiff(t *testing.T) {
	dir := t.TempDir()
	repoDir := filepath.Join(dir, "withdiff")
	require.NoError(t, os.Mkdir(repoDir, 0755))
	repo := gitest.Init(t, repoDir)
	repo.SetupCommitConfig()
	repo.WriteFileAdd("file", "line1\nline2\n")
	repo.Commit("initial")
	repo.WriteFileAdd("file", "line1\nline2\nline3\nline4\n")

	tp := runTestProgram(t, dir)
	tp.waitForContent("withdiff", "+2 -0")

	out := tp.out.String()
	require.False(t, strings.Contains(out, "discovering repos") &&
		!strings.Contains(out, "withdiff"),
		"expected table to be rendered, got:\n%s", out)
}

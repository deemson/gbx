package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/require"
)

var (
	keyTab      = tea.KeyPressMsg{Code: tea.KeyTab}
	keyShiftTab = tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	keyEnter    = tea.KeyPressMsg{Code: tea.KeyEnter}
	keyEsc      = tea.KeyPressMsg{Code: tea.KeyEscape}
	keyQuestion = tea.KeyPressMsg{Code: '?', Text: "?"}
	keyCtrlF    = tea.KeyPressMsg{Code: 'f', Mod: tea.ModCtrl}
	ctrl1       = tea.KeyPressMsg{Code: '1', Mod: tea.ModCtrl}
	ctrl2       = tea.KeyPressMsg{Code: '2', Mod: tea.ModCtrl}
	ctrl3       = tea.KeyPressMsg{Code: '3', Mod: tea.ModCtrl}
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
	tp.waitForContent("withcommit", branch) // clean tree → silent state column
}

func TestRepoShowsChangedCount(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "dirty")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("a", "1")
	repo.Commit("init")
	repo.WriteFileAdd("b", "2") // staged, uncommitted

	tp := runTestProgram(t, dir)
	tp.waitForContent("dirty", "✚1") // one staged-added file
}

func TestRunPullFailureShowsCross(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "lonely")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("f", "x")
	repo.Commit("c1")

	tp := runTestProgram(t, dir)
	tp.waitForContent("lonely")

	tp.send("p") // no upstream → fails
	tp.waitForContent("✗")
}

func TestRunCheckoutSwitchesToBranch(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "proj")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("f", "x")
	repo.Commit("c1")
	start := repo.BranchShowCurrent()
	repo.CheckoutBranch("feature")
	repo.Checkout(start) // leave "feature" existing but not current

	tp := runTestProgram(t, dir)
	tp.waitForContent("proj", start)

	// `c` opens the checkout prompt; type the ref; Enter runs.
	tp.send("c")
	tp.send("feature")
	tp.sendKey(keyEnter)
	tp.waitForContent("feature")
}

func TestRunCheckoutUnknownRefShowsCross(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "proj")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("f", "x")
	repo.Commit("c1")

	tp := runTestProgram(t, dir)
	tp.waitForContent("proj")

	tp.send("c")
	tp.send("nope-not-real")
	tp.sendKey(keyEnter)
	tp.waitForContent("✗")
}

func TestRowShowsLineChanges(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "proj")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("a.txt", "1\n2\n3\n")
	repo.Commit("c1")
	repo.WriteFile("a.txt", "1\n2\n3\n4\n") // append one tracked line → +1 -0

	tp := runTestProgram(t, dir)
	tp.waitForContent("proj", "~1", "+1 -0") // one unstaged-modified file
}

func TestHeaderShowsVersionPidAndChips(t *testing.T) {
	dir := t.TempDir()
	mkRepo(t, dir, "proj")

	tp := runTestProgram(t, dir)
	// The active chip's label is bold (an SGR the Ascii profile keeps), so assert
	// an inactive chip — "<C-2> name" stays plain/contiguous — plus the "<C-1>"
	// key prefix, "gbx dev", and "PID:".
	tp.waitForContent("gbx dev", "PID:", "<C-1>", "<C-2> name")
}

func TestHelpOverlayShowsBindings(t *testing.T) {
	dir := t.TempDir()
	mkRepo(t, dir, "proj")

	tp := runTestProgram(t, dir)
	tp.waitForContent("proj")

	tp.sendKey(keyQuestion)
	tp.waitForContent("gbx — keys", "list mode", "ctrl+f", "filter prompt")
}

func TestRefreshPicksUpExternalChange(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "proj")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("a", "1")
	repo.Commit("c1")
	branch := repo.BranchShowCurrent()

	tp := runTestProgram(t, dir)
	tp.waitForContent("proj", branch) // branch column confirms initial status loaded

	repo.WriteFileAdd("b", "2") // change made after the initial status load
	tp.send("r")
	tp.waitForContent("✚1")
}

func TestFilterExcludingAllShowsNoMatches(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"apple", "banana"} {
		mkRepo(t, dir, name)
	}
	tp := runTestProgram(t, dir)
	tp.waitForContent("apple", "banana")

	tp.sendKey(keyCtrlF)
	tp.send("zzzz")
	tp.sendKey(keyEnter)
	tp.waitForContent("no matches")
}

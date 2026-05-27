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
	ctrlR       = tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}
	ctrlG       = tea.KeyPressMsg{Code: 'g', Mod: tea.ModCtrl}
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
	tp.waitForContent("withcommit", branch, "✓") // clean tree → check glyph
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

func TestRunPullSuccessShowsCheck(t *testing.T) {
	dir := t.TempDir()

	remoteDir := t.TempDir()
	gitest.InitBare(t, remoteDir)

	producer := gitest.Init(t, t.TempDir())
	producer.RemoteAdd("origin", remoteDir)
	producer.SetupCommitConfig()
	producer.WriteFileAdd("file", "v1\n")
	producer.Commit("c1")
	producer.PushSetUpstream("origin", producer.BranchShowCurrent())

	// consumer lives inside the scanned dir, starts at c1, tracks origin.
	consumer := gitest.Clone(t, remoteDir, filepath.Join(dir, "consumer"))

	// producer advances the remote; consumer fetches → it now has something to pull.
	producer.WriteFileAdd("file", "v1\nv2\n")
	producer.Commit("c2")
	producer.Push()
	consumer.Fetch()

	tp := runTestProgram(t, dir)
	tp.waitForContent("consumer", "↓1")

	// enter → command mode → type → enter; success renders a fresh ✓ glyph.
	// (The behind→0 status change is an in-place cursor update the raw stream
	// doesn't show contiguously, so the refresh is asserted at model level.)
	tp.sendKey(keyEnter)
	tp.send("pull")
	tp.sendKey(keyEnter)
	tp.waitForContent("✓")
}

func TestRunPullFailureShowsCross(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "lonely")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("f", "x")
	repo.Commit("c1")

	tp := runTestProgram(t, dir)
	tp.waitForContent("lonely")

	tp.sendKey(keyEnter)
	tp.send("pull") // no upstream → fails
	tp.sendKey(keyEnter)
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

	tp.sendKey(keyEnter)
	tp.send("checkout feature")
	tp.sendKey(keyEnter)
	// the post-command status refresh repaints the branch column with the new
	// current branch — text not on screen before, so the raw stream shows it.
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

	tp.sendKey(keyEnter)
	tp.send("checkout nope-not-real")
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

func TestHelpOverlayShowsBindings(t *testing.T) {
	dir := t.TempDir()
	mkRepo(t, dir, "proj")

	tp := runTestProgram(t, dir)
	tp.waitForContent("proj")

	tp.sendKey(ctrlG)
	tp.waitForContent("gbx — keys", "tab", "ctrl+r", "ctrl+g")
}

func TestRefreshPicksUpExternalChange(t *testing.T) {
	dir := t.TempDir()
	repo := mkRepo(t, dir, "proj")
	repo.SetupCommitConfig()
	repo.WriteFileAdd("a", "1")
	repo.Commit("c1")

	tp := runTestProgram(t, dir)
	tp.waitForContent("proj", "✓")

	repo.WriteFileAdd("b", "2") // change made after the initial status load
	tp.sendKey(ctrlR)
	tp.waitForContent("✚1")
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

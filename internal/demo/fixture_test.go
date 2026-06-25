//go:build fixture

package demo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/deemson/gbx/internal/git/gitest"
	"github.com/stretchr/testify/require"
)

// TestGenerateDemoFixture writes a tree of git repositories in varied states
// under $GBX_FIXTURE_DIR for the VHS demo to film. Bare "remotes" live in a
// sibling $GBX_FIXTURE_DIR-remotes tree so they stay outside the filmed root
// (gbx would otherwise try to open them as repos). The justfile rm -rf's both
// before each run, so this assumes a clean slate.
func TestGenerateDemoFixture(t *testing.T) {
	root := os.Getenv("GBX_FIXTURE_DIR")
	require.NotEmpty(t, root, "GBX_FIXTURE_DIR must be set")
	remotes := root + "-remotes"

	mkRepo := func(name string) gitest.Repo {
		dir := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		r := gitest.Init(t, dir)
		r.SetupCommitConfig()
		return r
	}
	mkBare := func(name string) string {
		dir := filepath.Join(remotes, name+".git")
		require.NoError(t, os.MkdirAll(dir, 0o755))
		gitest.InitBare(t, dir)
		return dir
	}

	// payments-api: clean, on the default branch, up to date with its remote.
	api := mkRepo("payments-api")
	api.RemoteAdd("origin", mkBare("payments-api"))
	api.WriteFileAdd("main.go", "package main\n")
	api.Commit("initial commit")
	api.PushSetUpstream("origin", api.BranchShowCurrent())

	// web-frontend: on a feature branch with a dirty working tree (one modified
	// tracked file + one untracked file).
	web := mkRepo("web-frontend")
	web.WriteFileAdd("app.js", "console.log('v1')\n")
	web.Commit("initial commit")
	web.CheckoutBranch("feature/checkout")
	web.WriteFile("app.js", "console.log('wip')\n")
	web.WriteFile("TODO.md", "- finish checkout flow\n")

	// auth-service: two local commits ahead of its remote.
	auth := mkRepo("auth-service")
	auth.RemoteAdd("origin", mkBare("auth-service"))
	auth.WriteFileAdd("auth.go", "package auth\n")
	auth.Commit("initial commit")
	auth.PushSetUpstream("origin", auth.BranchShowCurrent())
	auth.WriteFileAdd("oauth.go", "package auth\n")
	auth.Commit("feat: add oauth")
	auth.WriteFileAdd("session.go", "package auth\n")
	auth.Commit("feat: add sessions")

	// notifications: one commit behind its remote (a sibling clone advanced the
	// remote, then this repo fetched). Pressing `p` fast-forwards it.
	notifRemote := mkBare("notifications")
	notif := mkRepo("notifications")
	notif.RemoteAdd("origin", notifRemote)
	notif.WriteFileAdd("notify.go", "package notify\n")
	notif.Commit("initial commit")
	notif.PushSetUpstream("origin", notif.BranchShowCurrent())
	adv := gitest.Clone(t, notifRemote, filepath.Join(remotes, "notifications-advance"))
	adv.SetupCommitConfig()
	adv.WriteFileAdd("email.go", "package notify\n")
	adv.Commit("feat: email channel")
	adv.Push()
	notif.Fetch()

	// data-pipeline: clean, on a non-default branch, up to date.
	data := mkRepo("data-pipeline")
	data.RemoteAdd("origin", mkBare("data-pipeline"))
	data.WriteFileAdd("pipeline.py", "print('etl')\n")
	data.Commit("initial commit")
	data.PushSetUpstream("origin", data.BranchShowCurrent())
	data.CheckoutBranch("develop")
	data.WriteFileAdd("transform.py", "print('transform')\n")
	data.Commit("feat: add transform step")
	data.PushSetUpstream("origin", "develop")

	// legacy-scripts: clean, no remote at all.
	legacy := mkRepo("legacy-scripts")
	legacy.WriteFileAdd("backup.sh", "#!/bin/sh\n")
	legacy.Commit("initial commit")

	t.Logf("demo fixture written to %s (remotes in %s)", root, remotes)
}

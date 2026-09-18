package git_test

import (
	"errors"
	"testing"

	"github.com/deemson/gbx/internal/git"
	gitexec "github.com/deemson/gbx/internal/git/exec"
	"github.com/stretchr/testify/require"
)

func TestRunErrorPreservesRecognizedFailureDetails(t *testing.T) {
	processErr := errors.New("exit status 1")
	res := gitexec.Result{
		Args:     []string{"-C", "/tmp/repo", "pull", "--ff-only"},
		ExitCode: 1,
		Stdout:   []byte("out\n"),
		Stderr:   []byte("full git explanation\n"),
	}
	err := git.NewRunErr(res, processErr, git.ErrNoUpstream)

	require.Equal(t, git.ErrNoUpstream.Error(), err.Error())
	require.Equal(t, git.ErrNoUpstream.Error(), err.Summary())
	require.ErrorIs(t, err, git.ErrNoUpstream)
	require.Equal(t, res, err.Res)
	require.Equal(t, processErr, err.Err)
}

func TestRunErrorKeepsUnknownDiagnosticAndUnwrapsProcessError(t *testing.T) {
	processErr := errors.New("exit status 42")
	res := gitexec.Result{Args: []string{"status"}, ExitCode: 42, Stderr: []byte("boom\n")}
	err := git.NewUnknownRunErr(res, processErr)

	require.Contains(t, err.Error(), "status: exit status 42")
	require.Contains(t, err.Error(), "stderr=`boom`")
	require.Equal(t, processErr.Error(), err.Summary())
	require.ErrorIs(t, err, processErr)
}

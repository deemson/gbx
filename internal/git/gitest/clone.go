package gitest

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/exec"
	"github.com/stretchr/testify/require"
)

func Clone(t *testing.T, source, target string) Repo {
	ctx := context.Background()
	res, err := exec.Git{}.Run(ctx, "clone", source, target)
	if err != nil {
		require.NoError(t, git.NewUnknownRunErr(res, err))
	}
	if !bytes.Contains(res.Stderr, []byte("Cloning into")) || !bytes.Contains(res.Stderr, []byte("done.")) {
		require.NoError(t, git.NewUnknownRunErr(res, errors.New("unexpected output")))
	}
	return Open(t, target)
}

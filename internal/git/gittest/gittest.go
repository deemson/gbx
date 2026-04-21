package gittest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deemson/gbx/internal/git"
	"github.com/deemson/gbx/internal/git/gitexec"
)

func Init(ctx context.Context, path string) (git.Repo, error) {

	res, err := gitexec.Run(ctx, path, "init")
	if err != nil {
		return git.Repo{}, err
	}
	if !strings.Contains(res.Stdout, "Initialized empty Git repository") {
		return git.Repo{}, fmt.Errorf("unexpected init output: %s", res.Stdout)
	}
	return git.Open(ctx, path)
}

func WriteFile(r git.Repo, name, content string) error {
	full := filepath.Join(r.Path(), name)
	return os.WriteFile(full, []byte(content), 0644)
}

func Add(ctx context.Context, r git.Repo, paths ...string) error {
	args := append([]string{"add"}, paths...)
	_, err := gitexec.Run(ctx, r.Path(), args...)
	return err
}

func Rm(ctx context.Context, r git.Repo, paths ...string) error {
	args := append([]string{"rm"}, paths...)
	_, err := gitexec.Run(ctx, r.Path(), args...)
	return err
}

func Commit(ctx context.Context, r git.Repo, msg string) error {
	_, err := gitexec.Run(ctx, r.Path(),
		"-c", "user.email=test@example.com",
		"-c", "user.name=test",
		"commit", "-m", msg,
	)
	return err
}

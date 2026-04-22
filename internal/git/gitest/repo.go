package gitest

import (
	"os"
	"path"

	"github.com/deemson/gbx/internal/git"
)

type Repo struct {
	git.Repo
}

func (r Repo) WriteFile(name string, data []byte) error {
	return os.WriteFile(path.Join(r.Path(), name), data, 0644)
}

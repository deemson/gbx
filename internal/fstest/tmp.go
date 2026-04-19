package fstest

import (
	"os"
)

func CreateTempDir(name string) (string, error) {
	return os.MkdirTemp("", name)
}

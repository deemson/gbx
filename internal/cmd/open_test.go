package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenRequiresDirectory(t *testing.T) {
	cmd := openCmd("")
	cmd.SetArgs(nil)
	require.Error(t, cmd.Execute())
}

func TestValidateDirectoriesPreservesOccurrencesAndLabels(t *testing.T) {
	dir := t.TempDir()
	args := []string{dir, dir}
	dirs, err := validateDirectories(args)
	require.NoError(t, err)
	require.Len(t, dirs, 2)
	require.Equal(t, args, []string{dirs[0].Label, dirs[1].Label})
}

func TestValidateDirectoriesNamesInvalidArgument(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	_, err := validateDirectories([]string{missing})
	require.ErrorContains(t, err, missing)

	file := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
	_, err = validateDirectories([]string{file})
	require.ErrorContains(t, err, file)
	require.ErrorContains(t, err, "not a directory")
}

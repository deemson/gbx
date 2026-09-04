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

func TestValidateDirectoriesHidesHeadingForSingleCurrentDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	for _, arg := range []string{".", "./", cwd} {
		dirs, err := validateDirectories([]string{arg})
		require.NoError(t, err)
		require.True(t, dirs[0].HideHeading, arg)
	}
}

func TestValidateDirectoriesKeepsHeadingForOtherOrMultipleDirectories(t *testing.T) {
	other := t.TempDir()
	dirs, err := validateDirectories([]string{other})
	require.NoError(t, err)
	require.False(t, dirs[0].HideHeading)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	dirs, err = validateDirectories([]string{cwd, cwd})
	require.NoError(t, err)
	require.False(t, dirs[0].HideHeading)
	require.False(t, dirs[1].HideHeading)
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

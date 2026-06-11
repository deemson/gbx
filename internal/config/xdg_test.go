package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deemson/gbx/internal/config"
	"github.com/stretchr/testify/require"
)

// useTempConfigHome points XDG_CONFIG_HOME at a fresh temp dir for the duration
// of the test and returns the resolved config.toml / schema.json paths.
func useTempConfigHome(t *testing.T) (configPath, schemaPath string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return filepath.Join(dir, "gbx", "config.toml"), filepath.Join(dir, "gbx", "schema.json")
}

func TestWriteDefaultFresh(t *testing.T) {
	configPath, schemaPath := useTempConfigHome(t)

	paths, err := config.WriteDefault(false)
	require.NoError(t, err)
	require.Equal(t, []string{configPath, schemaPath}, paths)
	require.FileExists(t, configPath)
	require.FileExists(t, schemaPath)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(data), "#:schema ./schema.json\n"),
		"config.toml should open with the schema directive, got:\n%s", data)
	// The directive is a TOML comment, so the file still parses.
	_, err = config.FromBytes(data)
	require.NoError(t, err)
}

func TestWriteDefaultRefusesWhenExists(t *testing.T) {
	configPath, _ := useTempConfigHome(t)

	_, err := config.WriteDefault(false)
	require.NoError(t, err)

	// Mark the file so we can prove the refused call touched nothing.
	require.NoError(t, os.WriteFile(configPath, []byte("sentinel"), 0644))

	_, err = config.WriteDefault(false)
	var existsErr *config.FileExistsError
	require.ErrorAs(t, err, &existsErr)
	require.Equal(t, configPath, existsErr.Path)

	data, readErr := os.ReadFile(configPath)
	require.NoError(t, readErr)
	require.Equal(t, "sentinel", string(data))
}

func TestWriteDefaultForceOverwrites(t *testing.T) {
	configPath, _ := useTempConfigHome(t)

	_, err := config.WriteDefault(false)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, []byte("sentinel"), 0644))

	paths, err := config.WriteDefault(true)
	require.NoError(t, err)
	require.Equal(t, []string{configPath}, paths[:1])

	data, readErr := os.ReadFile(configPath)
	require.NoError(t, readErr)
	require.NotEqual(t, "sentinel", string(data))
}

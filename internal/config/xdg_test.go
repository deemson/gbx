package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/deemson/gbx/internal/config"
	"github.com/stretchr/testify/require"
)

// useTempConfigHome points XDG at a fresh temp dir for the duration of the test
// and returns the resolved config.toml / schema.json paths.
func useTempConfigHome(t *testing.T) (configPath, schemaPath string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	xdg.Reload()
	return filepath.Join(dir, "gbx", "config.toml"), filepath.Join(dir, "gbx", "schema.json")
}

func TestWriteDefaultFresh(t *testing.T) {
	configPath, schemaPath := useTempConfigHome(t)

	paths, err := config.WriteDefault(false)
	require.NoError(t, err)
	require.Equal(t, []string{configPath, schemaPath}, paths)
	require.FileExists(t, configPath)
	require.FileExists(t, schemaPath)
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

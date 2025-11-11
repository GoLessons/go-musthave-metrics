package config

import (
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func writeJSON(t *testing.T, v any) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "server.json")
	bs, err := json.Marshal(v)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, bs, 0o644))
	return path
}

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("ADDRESS", "")
	t.Setenv("CONFIG", "")
	os.Args = []string{"prog"}

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "localhost:8080", cfg.Address)
	require.False(t, cfg.DumpConfig.Restore)
	require.EqualValues(t, 300, cfg.DumpConfig.StoreInterval)
	require.Equal(t, "metric-storage.json", cfg.DumpConfig.FileStoragePath)
}

func TestLoadConfig_FileOnly(t *testing.T) {
	v := map[string]any{
		"Address":     "127.0.0.1:9999",
		"DatabaseDsn": "dsn_file",
		"DumpConfig": map[string]any{
			"Restore":         true,
			"StoreInterval":   123,
			"FileStoragePath": "from-file.json",
		},
		"Key": "filekey",
	}
	path := writeJSON(t, v)

	t.Setenv("ADDRESS", "")
	t.Setenv("CONFIG", path)
	os.Args = []string{"prog"}

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1:9999", cfg.Address)
	require.Equal(t, "dsn_file", cfg.DatabaseDsn)
	require.Equal(t, "filekey", cfg.Key)
	require.True(t, cfg.DumpConfig.Restore)
	require.EqualValues(t, 123, cfg.DumpConfig.StoreInterval)
	require.Equal(t, "from-file.json", cfg.DumpConfig.FileStoragePath)
}

func TestLoadConfig_FlagsOverrideFile(t *testing.T) {
	v := map[string]any{
		"Address": "127.0.0.1:9999",
		"DumpConfig": map[string]any{
			"Restore":         true,
			"StoreInterval":   123,
			"FileStoragePath": "from-file.json",
		},
	}
	path := writeJSON(t, v)

	t.Setenv("ADDRESS", "")
	t.Setenv("CONFIG", path)
	os.Args = []string{
		"prog",
		"-address", "0.0.0.0:8081",
		"-restore", "false",
		"-store-interval", "777",
		"-file-storage-path", "flags.json",
	}

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0:8081", cfg.Address)
	require.False(t, cfg.DumpConfig.Restore)
	require.EqualValues(t, 777, cfg.DumpConfig.StoreInterval)
	require.Equal(t, "flags.json", cfg.DumpConfig.FileStoragePath)
}

func TestLoadConfig_EnvOverridesAll(t *testing.T) {
	v := map[string]any{
		"Address": "127.0.0.1:9999",
		"DumpConfig": map[string]any{
			"Restore":         false,
			"StoreInterval":   10,
			"FileStoragePath": "file.json",
		},
	}
	path := writeJSON(t, v)

	t.Setenv("CONFIG", path)
	t.Setenv("ADDRESS", "env:9090")
	t.Setenv("RESTORE", "true")
	t.Setenv("STORE_INTERVAL", "5")
	t.Setenv("FILE_STORAGE_PATH", "env.json")

	os.Args = []string{
		"prog",
		"-address", "flags:8080",
		"-restore", "false",
		"-store-interval", "777",
		"-file-storage-path", "flags.json",
	}

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "env:9090", cfg.Address)
	require.True(t, cfg.DumpConfig.Restore)
	require.EqualValues(t, 5, cfg.DumpConfig.StoreInterval)
	require.Equal(t, "env.json", cfg.DumpConfig.FileStoragePath)
}

func TestLoadConfig_ConfigPathEnvPreferredOverFlag(t *testing.T) {
	vEnv := map[string]any{
		"Address": "env-file:8082",
	}
	vFlag := map[string]any{
		"Address": "flag-file:8083",
	}
	pathEnv := writeJSON(t, vEnv)
	pathFlag := writeJSON(t, vFlag)

	t.Setenv("ADDRESS", "")
	t.Setenv("CONFIG", pathEnv)
	os.Args = []string{"prog", "-config", pathFlag}

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "env-file:8082", cfg.Address)
}

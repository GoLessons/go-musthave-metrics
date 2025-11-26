package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/require"
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

func prepareConfigEnv(t *testing.T, configPath string, args ...string) {
	t.Helper()
	envs := []string{
		"ADDRESS",
		"CONFIG",
		"DATABASE_DSN",
		"RESTORE",
		"STORE_INTERVAL",
		"FILE_STORAGE_PATH",
		"KEY",
		"CRYPTO_KEY",
		"TRUSTED_SUBNET",
		"AUDIT_FILE",
		"AUDIT_URL",
		"PPROF_ON_SHUTDOWN",
		"PPROF_DIR",
		"PPROF_FILENAME",
		"PPROF_HTTP",
		"PPROF_HTTP_ADDR",
	}
	for _, e := range envs {
		t.Setenv(e, "")
	}
	if configPath != "" {
		t.Setenv("CONFIG", configPath)
	}
	if len(args) == 0 {
		os.Args = []string{"prog"}
	} else {
		os.Args = append([]string{"prog"}, args...)
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	prepareConfigEnv(t, "")

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "localhost:8080", cfg.Address)
	require.False(t, cfg.DumpConfig.Restore)
	require.EqualValues(t, 300, cfg.DumpConfig.StoreInterval)
	require.Equal(t, "metric-storage.json", cfg.DumpConfig.FileStoragePath)
	require.Equal(t, "", cfg.TrustedSubnet)
}

func TestLoadConfig_FileOnly(t *testing.T) {
	v := map[string]any{
		"Address":       "127.0.0.1:9999",
		"DatabaseDsn":   "dsn_file",
		"TrustedSubnet": "127.0.0.0/8",
		"DumpConfig": map[string]any{
			"Restore":         true,
			"StoreInterval":   123,
			"FileStoragePath": "from-file.json",
		},
		"Key": "filekey",
	}
	path := writeJSON(t, v)

	prepareConfigEnv(t, path)

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1:9999", cfg.Address)
	require.Equal(t, "dsn_file", cfg.DatabaseDsn)
	require.Equal(t, "filekey", cfg.Key)
	require.True(t, cfg.DumpConfig.Restore)
	require.EqualValues(t, 123, cfg.DumpConfig.StoreInterval)
	require.Equal(t, "from-file.json", cfg.DumpConfig.FileStoragePath)
	require.Equal(t, "127.0.0.0/8", cfg.TrustedSubnet)
}

func TestLoadConfig_FlagsOverrideFile(t *testing.T) {
	v := map[string]any{
		"Address":       "127.0.0.1:9999",
		"TrustedSubnet": "10.0.0.0/8",
		"DumpConfig": map[string]any{
			"Restore":         true,
			"StoreInterval":   123,
			"FileStoragePath": "from-file.json",
		},
	}
	path := writeJSON(t, v)

	prepareConfigEnv(t, path,
		"-address=0.0.0.0:8081",
		"-restore=false",
		"-store-interval=777",
		"-file-storage-path=flags.json",
		"-t=192.168.0.0/16",
	)

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "0.0.0.0:8081", cfg.Address)
	require.False(t, cfg.DumpConfig.Restore)
	require.EqualValues(t, 777, cfg.DumpConfig.StoreInterval)
	require.Equal(t, "flags.json", cfg.DumpConfig.FileStoragePath)
	require.Equal(t, "192.168.0.0/16", cfg.TrustedSubnet)
}

func TestLoadConfig_EnvOverridesAll(t *testing.T) {
	v := map[string]any{
		"Address":       "127.0.0.1:9999",
		"TrustedSubnet": "10.0.0.0/8",
		"DumpConfig": map[string]any{
			"Restore":         false,
			"StoreInterval":   10,
			"FileStoragePath": "file.json",
		},
	}
	path := writeJSON(t, v)

	prepareConfigEnv(t, path,
		"-address=flags:8080",
		"-restore=false",
		"-store-interval=777",
		"-file-storage-path=flags.json",
		"-trusted_subnet=192.168.0.0/16",
	)

	t.Setenv("ADDRESS", "env:9090")
	t.Setenv("RESTORE", "true")
	t.Setenv("STORE_INTERVAL", "5")
	t.Setenv("FILE_STORAGE_PATH", "env.json")
	t.Setenv("TRUSTED_SUBNET", "fd00::/8")

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "env:9090", cfg.Address)
	require.True(t, cfg.DumpConfig.Restore)
	require.EqualValues(t, 5, cfg.DumpConfig.StoreInterval)
	require.Equal(t, "env.json", cfg.DumpConfig.FileStoragePath)
	require.Equal(t, "fd00::/8", cfg.TrustedSubnet)
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

	prepareConfigEnv(t, pathEnv, "-config", pathFlag)

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "env-file:8082", cfg.Address)
}

func TestLoadConfig_TrustedSubnet_FlagOnly_Long(t *testing.T) {
	prepareConfigEnv(t, "",
		"-trusted_subnet=10.10.0.0/16",
	)

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "10.10.0.0/16", cfg.TrustedSubnet)
}

func TestLoadConfig_TrustedSubnet_EnvOnly(t *testing.T) {
	prepareConfigEnv(t, "")
	t.Setenv("TRUSTED_SUBNET", "2001:db8::/32")

	cfg, err := LoadConfig(nil)
	require.NoError(t, err)

	require.Equal(t, "2001:db8::/32", cfg.TrustedSubnet)
}

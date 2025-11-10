package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	DumpConfig      DumpConfig
	Key             string `env:"KEY"`
	CryptoKey       string `env:"CRYPTO_KEY"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	PprofOnShutdown bool   `env:"PPROF_ON_SHUTDOWN"`
	PprofDir        string `env:"PPROF_DIR"`
	PprofFilename   string `env:"PPROF_FILENAME"`
	PprofHTTP       bool   `env:"PPROF_HTTP"`
	PprofHTTPAddr   string `env:"PPROF_HTTP_ADDR"`
}

type DumpConfig struct {
	StoreInterval   uint64 `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
}

type ConfigError struct {
	Msg string
	err error
}

func Error(format string, a ...any) error {
	return &ConfigError{
		Msg: fmt.Sprintf(format, a...),
	}
}

func wrapError(msg string, err error) error {
	return &ConfigError{msg, err}
}

func (e *ConfigError) Error() string {
	if e.err != nil && e.err.Error() != e.Error() {
		return fmt.Sprintf("[CONFIG] %s (previous: %v)", e.Msg, e.err)
	}

	return fmt.Sprintf("[CONFIG] %s", e.Msg)
}

func (e *ConfigError) Unwrap() error {
	return e.err
}

func LoadConfig(args *map[string]any) (*Config, error) {
	flags := flag.NewFlagSet("app-config", flag.ContinueOnError)

	envAddress := os.Getenv("ADDRESS")
	if envAddress == "" {
		envAddress = "localhost:8080"
	}

	address := flags.String("address", envAddress, "HTTP server address")
	restore := flags.Bool("restore", false, "Restore metrics before starting")
	storeInterval := flags.Uint64("store-interval", 300, "Store interval in seconds")
	fileStoragePath := flags.String("file-storage-path", "metric-storage.json", "File storage path")
	databaseDsn := flags.String("database-dsn", "", "Database DSN")
	key := flags.String("key", "", "Key for signature verification")
	cryptoKey := flags.String("crypto-key", "", "Path to RSA private key for request decryption")
	auditFile := flags.String("audit-file", "", "Audit log file path")
	auditURL := flags.String("audit-url", "", "Audit log URL")

	pprofOnShutdown := flags.Bool("pprof-on-shutdown", false, "Enable heap profile write on shutdown")
	pprofDir := flags.String("pprof-dir", "profiles", "Directory to store pprof files")
	pprofFilename := flags.String("pprof-filename", "base.pprof", "Heap profile filename")
	pprofHTTP := flags.Bool("pprof-http", false, "Expose net/http/pprof endpoints")
	pprofHTTPAddr := flags.String("pprof-http-addr", ":6060", "pprof HTTP listen address")

	flags.StringVar(address, "a", *address, "HTTP server address (short)")
	flags.BoolVar(restore, "r", *restore, "Restore metrics before starting (short)")
	flags.Uint64Var(storeInterval, "i", *storeInterval, "Store interval in seconds (short)")
	flags.StringVar(fileStoragePath, "f", *fileStoragePath, "File storage path (short)")
	flags.StringVar(databaseDsn, "d", *databaseDsn, "Database DSN")
	flags.StringVar(key, "k", *key, "Key for signature verification (short)")

	filteredArgs := filterArgs(flags, os.Args[1:])

	err := flags.Parse(filteredArgs)
	if err != nil {
		if strings.Contains(err.Error(), "flag provided but not defined") {
			return nil, err
		}
	}

	cfg := &Config{
		Address:     *address,
		DatabaseDsn: *databaseDsn,
		Key:         *key,
		CryptoKey:   *cryptoKey,
		AuditFile:   *auditFile,
		AuditURL:    *auditURL,
		DumpConfig: DumpConfig{
			Restore:         *restore,
			StoreInterval:   *storeInterval,
			FileStoragePath: *fileStoragePath,
		},
		PprofOnShutdown: *pprofOnShutdown,
		PprofDir:        *pprofDir,
		PprofFilename:   *pprofFilename,
		PprofHTTP:       *pprofHTTP,
		PprofHTTPAddr:   *pprofHTTPAddr,
	}

	// ENV overrides
	if databaseDsn := os.Getenv("DATABASE_DSN"); databaseDsn != "" {
		cfg.DatabaseDsn = databaseDsn
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restoreVal, err := strconv.ParseBool(envRestore)
		if err != nil {
			return nil, wrapError("ошибка парсинга RESTORE", err)
		} else {
			cfg.DumpConfig.Restore = restoreVal
		}
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		interval, err := strconv.ParseUint(envStoreInterval, 10, 64)
		if err != nil {
			return nil, wrapError("ошибка парсинга STORE_INTERVAL", err)
		} else {
			cfg.DumpConfig.StoreInterval = interval
		}
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		cfg.DumpConfig.FileStoragePath = envFileStoragePath
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}

	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		cfg.AuditURL = envAuditURL
	}

	if v := os.Getenv("PPROF_ON_SHUTDOWN"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, wrapError("ошибка парсинга PPROF_ON_SHUTDOWN", err)
		}
		cfg.PprofOnShutdown = b
	}
	if v := os.Getenv("PPROF_DIR"); v != "" {
		cfg.PprofDir = v
	}
	if v := os.Getenv("PPROF_FILENAME"); v != "" {
		cfg.PprofFilename = v
	}
	if v := os.Getenv("PPROF_HTTP"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, wrapError("ошибка парсинга PPROF_HTTP", err)
		}
		cfg.PprofHTTP = b
	}
	if v := os.Getenv("PPROF_HTTP_ADDR"); v != "" {
		cfg.PprofHTTPAddr = v
	}

	if args != nil {
		redefineLocal(args, cfg)
	}

	return cfg, nil
}

func filterArgs(flags *flag.FlagSet, args []string) []string {
	var filteredArgs []string
	validFlags := make(map[string]bool)

	flags.VisitAll(func(f *flag.Flag) {
		validFlags["-"+f.Name] = true
	})

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(arg) > 1 && arg[0] == '-' {
			splitArg := strings.SplitN(arg, "=", 2)
			flagName := splitArg[0]

			if _, valid := validFlags[flagName]; valid {
				filteredArgs = append(filteredArgs, flagName)

				if len(splitArg) > 1 {
					// Добавляем значение если оно присутствует
					filteredArgs = append(filteredArgs, splitArg[1])
				} else if i+1 < len(args) && args[i+1][0] != '-' {
					// Если значение передается отдельным аргументом, добавляем его
					filteredArgs = append(filteredArgs, args[i+1])
					i++ // Пропустить значение
				}
			}
		}
	}

	return filteredArgs
}

func redefineLocal(args *map[string]any, cfg *Config) {
	if val, ok := (*args)["Address"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.Address = strVal
		}
	}

	if val, ok := (*args)["DatabaseDsn"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.DatabaseDsn = strVal
		}
	}

	if val, ok := (*args)["DumpConfig.Restore"]; ok {
		if boolVal, ok := val.(bool); ok {
			cfg.DumpConfig.Restore = boolVal
		}
	}

	if val, ok := (*args)["DumpConfig.StoreInterval"]; ok {
		if intValue, ok := val.(uint64); ok {
			cfg.DumpConfig.StoreInterval = intValue
		}
	}

	if val, ok := (*args)["DumpConfig.FileStoragePath"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.DumpConfig.FileStoragePath = strVal
		}
	}

	if val, ok := (*args)["Key"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.Key = strVal
		}
	}

	if val, ok := (*args)["AuditFile"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.AuditFile = strVal
		}
	}

	if val, ok := (*args)["AuditURL"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.AuditURL = strVal
		}
	}

	if val, ok := (*args)["PprofOnShutdown"]; ok {
		if boolVal, ok := val.(bool); ok {
			cfg.PprofOnShutdown = boolVal
		}
	}
	if val, ok := (*args)["PprofDir"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.PprofDir = strVal
		}
	}
	if val, ok := (*args)["PprofFilename"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.PprofFilename = strVal
		}
	}
	if val, ok := (*args)["PprofHTTP"]; ok {
		if boolVal, ok := val.(bool); ok {
			cfg.PprofHTTP = boolVal
		}
	}
	if val, ok := (*args)["PprofHTTPAddr"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.PprofHTTPAddr = strVal
		}
	}
	if val, ok := (*args)["CryptoKey"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.CryptoKey = strVal
		}
	}
}

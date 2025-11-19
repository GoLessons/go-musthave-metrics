package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	fileconfig "github.com/GoLessons/go-musthave-metrics/pkg/file-config"
)

type Config struct {
	Address         string `env:"ADDRESS"`
	DatabaseDsn     string `env:"DATABASE_DSN"`
	DumpConfig      DumpConfig
	Key             string `env:"KEY"`
	CryptoKey       string `env:"CRYPTO_KEY"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET"`
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

	cfgDefaults := &Config{
		Address:     envAddress,
		DatabaseDsn: "",
		Key:         "",
		CryptoKey:   "",
		AuditFile:   "",
		AuditURL:    "",
		DumpConfig: DumpConfig{
			Restore:         false,
			StoreInterval:   300,
			FileStoragePath: "metric-storage.json",
		},
		PprofOnShutdown: false,
		PprofDir:        "profiles",
		PprofFilename:   "base.pprof",
		PprofHTTP:       false,
		PprofHTTPAddr:   ":6060",
	}

	if configPath := getFileConfigPath(); configPath != "" {
		if err := fileconfig.LoadInto(configPath, cfgDefaults); err != nil {
			return nil, wrapError("ошибка чтения файла конфигурации", err)
		}
	}

	address := flags.String("address", cfgDefaults.Address, "HTTP server address")
	restore := flags.Bool("restore", cfgDefaults.DumpConfig.Restore, "Restore metrics before starting")
	storeInterval := flags.Uint64("store-interval", cfgDefaults.DumpConfig.StoreInterval, "Store interval in seconds")
	fileStoragePath := flags.String("file-storage-path", cfgDefaults.DumpConfig.FileStoragePath, "File storage path")
	databaseDsn := flags.String("database-dsn", cfgDefaults.DatabaseDsn, "Database DSN")
	key := flags.String("key", cfgDefaults.Key, "Key for signature verification")
	cryptoKey := flags.String("crypto-key", cfgDefaults.CryptoKey, "Path to RSA private key for request decryption")
	auditFile := flags.String("audit-file", cfgDefaults.AuditFile, "Audit log file path")
	auditURL := flags.String("audit-url", cfgDefaults.AuditURL, "Audit log URL")
	trustedSubnet := flags.String("trusted_subnet", cfgDefaults.TrustedSubnet, "Trusted subnet CIDR")

	pprofOnShutdown := flags.Bool("pprof-on-shutdown", cfgDefaults.PprofOnShutdown, "Enable heap profile write on shutdown")
	pprofDir := flags.String("pprof-dir", cfgDefaults.PprofDir, "Directory to store pprof files")
	pprofFilename := flags.String("pprof-filename", cfgDefaults.PprofFilename, "Heap profile filename")
	pprofHTTP := flags.Bool("pprof-http", cfgDefaults.PprofHTTP, "Expose net/http/pprof endpoints")
	pprofHTTPAddr := flags.String("pprof-http-addr", cfgDefaults.PprofHTTPAddr, "pprof HTTP listen address")

	flags.StringVar(address, "a", *address, "HTTP server address (short)")
	flags.BoolVar(restore, "r", *restore, "Restore metrics before starting (short)")
	flags.Uint64Var(storeInterval, "i", *storeInterval, "Store interval in seconds (short)")
	flags.StringVar(fileStoragePath, "f", *fileStoragePath, "File storage path (short)")
	flags.StringVar(databaseDsn, "d", *databaseDsn, "Database DSN")
	flags.StringVar(key, "k", *key, "Key for signature verification (short)")
	flags.StringVar(trustedSubnet, "t", *trustedSubnet, "Trusted subnet CIDR (short)")

	filteredArgs := filterArgs(flags, os.Args[1:])

	err := flags.Parse(filteredArgs)
	if err != nil {
		if strings.Contains(err.Error(), "flag provided but not defined") {
			return nil, err
		}
	}

	cfg := &Config{
		Address:       *address,
		DatabaseDsn:   *databaseDsn,
		Key:           *key,
		CryptoKey:     *cryptoKey,
		TrustedSubnet: *trustedSubnet,
		AuditFile:     *auditFile,
		AuditURL:      *auditURL,
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

	if v := os.Getenv("ADDRESS"); v != "" {
		cfg.Address = v
	}
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
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		cfg.TrustedSubnet = envTrustedSubnet
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
			eq := strings.Index(arg, "=")
			flagName := arg
			if eq != -1 {
				flagName = arg[:eq]
			}

			if _, valid := validFlags[flagName]; valid {
				if eq != -1 {
					filteredArgs = append(filteredArgs, arg)
					continue
				}

				filteredArgs = append(filteredArgs, flagName)

				if i+1 < len(args) && args[i+1][0] != '-' {
					filteredArgs = append(filteredArgs, args[i+1])
					i++
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

	if val, ok := (*args)["TrustedSubnet"]; ok {
		if strVal, ok := val.(string); ok {
			cfg.TrustedSubnet = strVal
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

func getFileConfigPath() string {
	if v := os.Getenv("CONFIG"); v != "" {
		return v
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-c=") {
			return strings.TrimPrefix(a, "-c=")
		}
		if strings.HasPrefix(a, "-config=") {
			return strings.TrimPrefix(a, "-config=")
		}
		if a == "-c" || a == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}

	return ""
}

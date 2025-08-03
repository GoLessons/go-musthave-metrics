package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address     string `env:"ADDRESS"`
	DatabaseDsn string `env:"DATABASE_DSN"`
	DumpConfig  DumpConfig
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
		return fmt.Sprintf("[CONFIG] %s (previous: %w)", e.Msg, e.err)
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

	flags.StringVar(address, "a", *address, "HTTP server address (short)")
	flags.BoolVar(restore, "r", *restore, "Restore metrics before starting (short)")
	flags.Uint64Var(storeInterval, "i", *storeInterval, "Store interval in seconds (short)")
	flags.StringVar(fileStoragePath, "f", *fileStoragePath, "File storage path (short)")
	flags.StringVar(databaseDsn, "d", *databaseDsn, "Database DSN")

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
		DumpConfig: DumpConfig{
			Restore:         *restore,
			StoreInterval:   *storeInterval,
			FileStoragePath: *fileStoragePath,
		},
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
}

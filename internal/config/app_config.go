package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
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

func LoadConfig(args *map[string]any) (*Config, error) {
	flags := flag.NewFlagSet("app-config", flag.ExitOnError)

	address := flags.String("address", "localhost:8080", "HTTP server address")
	restore := flags.Bool("restore", false, "Restore metrics before starting")
	storeInterval := flags.Uint64("store-interval", 300, "Store interval in seconds")
	fileStoragePath := flags.String("file-storage-path", "metric-storage.json", "File storage path")
	databaseDsn := flags.String("database-dsn", "", "Database DSN")

	flags.StringVar(address, "a", *address, "HTTP server address (short)")
	flags.BoolVar(restore, "r", *restore, "Restore metrics before starting (short)")
	flags.Uint64Var(storeInterval, "i", *storeInterval, "Store interval in seconds (short)")
	flags.StringVar(fileStoragePath, "f", *fileStoragePath, "File storage path (short)")
	flags.StringVar(databaseDsn, "d", *databaseDsn, "Database DSN")

	err := flags.Parse(flags.Args())
	if err != nil {
		return nil, err
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

	if envAddress := os.Getenv("ADDRESS"); envAddress != "" {
		cfg.Address = envAddress
	}

	if databaseDsn := os.Getenv("DATABASE_DSN"); databaseDsn != "" {
		cfg.DatabaseDsn = databaseDsn
	}

	/*if cfg.DatabaseDsn == "" {
		return nil, errors.New("укажите DSN для подключения к базе данных")
	}*/

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		restoreVal, err := strconv.ParseBool(envRestore)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга RESTORE: %v", err)
		} else {
			cfg.DumpConfig.Restore = restoreVal
		}
	}

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		interval, err := strconv.ParseUint(envStoreInterval, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга STORE_INTERVAL: %v", err)
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

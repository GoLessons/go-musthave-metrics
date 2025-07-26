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

func LoadConfig() (*Config, error) {
	address := flag.String("address", "localhost:8080", "HTTP server address")
	restore := flag.Bool("restore", false, "Restore metrics before starting")
	storeInterval := flag.Uint64("store-interval", 300, "Store interval in seconds")
	fileStoragePath := flag.String("file-storage-path", "metric-storage.json", "File storage path")
	databaseDsn := flag.String("database-dsn", "", "Database DSN")

	flag.StringVar(address, "a", *address, "HTTP server address (short)")
	flag.BoolVar(restore, "r", *restore, "Restore metrics before starting (short)")
	flag.Uint64Var(storeInterval, "i", *storeInterval, "Store interval in seconds (short)")
	flag.StringVar(fileStoragePath, "f", *fileStoragePath, "File storage path (short)")
	flag.StringVar(databaseDsn, "d", *databaseDsn, "Database DSN")

	flag.Parse()

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

	return cfg, nil
}

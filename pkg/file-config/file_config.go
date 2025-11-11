package fileconfig

import (
	"os"
	"path/filepath"

	"github.com/goccy/go-json"
)

func Load[T any](path string) (T, error) {
	var conf T

	bs, err := os.ReadFile(path)
	if err != nil {
		return conf, err
	}

	if err := json.Unmarshal(bs, &conf); err != nil {
		return conf, err
	}

	return conf, nil
}

func Save[T any](path string, v T) error {
	bs, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	tmp, err := os.CreateTemp(dir, ".fileconfig-*")
	if err != nil {
		return err
	}

	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(bs); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), path)
}

func LoadInto(path string, v any) error {
	bs, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(bs, v)
}

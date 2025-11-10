package resettool

import (
	"fmt"
	"os"
	"path/filepath"
)

func Run(dir string) error {
	pkgs, err := ScanPackages(dir)
	if err != nil {
		return err
	}

	for _, p := range pkgs {
		if len(p.Structs) == 0 {
			continue
		}
		content := BuildPackageContent(p)
		outPath := filepath.Join(p.Dir, "reset.gen.go")
		if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}

	return nil
}

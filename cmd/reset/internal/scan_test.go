package resettool

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestScan_IgnoresTestsAndGen(t *testing.T) {
	dir := t.TempDir()
	ok := `package p

// generate:reset
type X struct{}
`
	ignoredTest := `package p

// generate:reset
type T struct{}
`
	ignoredGen := `package p

// generate:reset
type G struct{}
`

	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(ok), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "x_test.go"), []byte(ignoredTest), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "g.gen.go"), []byte(ignoredGen), 0o644); err != nil {
		t.Fatal(err)
	}

	pkgs, err := ScanPackages(dir)
	if err != nil {
		t.Fatalf("ScanPackages error: %v", err)
	}

	var names []string
	for _, p := range pkgs {
		for _, s := range p.Structs {
			names = append(names, s.Name)
		}
	}
	sort.Strings(names)

	if len(names) != 1 || names[0] != "X" {
		t.Fatalf("expected only X, got: %v", names)
	}
}

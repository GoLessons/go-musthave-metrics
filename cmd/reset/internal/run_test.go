package resettool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_WritesAndOverwritesResetGen(t *testing.T) {
	dir := t.TempDir()
	src := `package mypkg

// generate:reset
type A struct{}
`
	if err := os.WriteFile(filepath.Join(dir, "a.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	// Предсуществующий файл должен перезаписываться
	if err := os.WriteFile(filepath.Join(dir, "reset.gen.go"), []byte("package mypkg\n// old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run(dir); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(dir, "reset.gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "package mypkg") {
		t.Fatalf("reset.gen.go should contain package header, got: %s", string(out))
	}
	if !strings.Contains(string(out), "func (x *A) Reset() {}") {
		t.Fatalf("reset.gen.go should contain Reset() for A, got: %s", string(out))
	}
	if strings.Contains(string(out), "old content") {
		t.Fatalf("reset.gen.go should be overwritten, still contains old content")
	}
}

func TestRun_SkipStructsWithExistingReset(t *testing.T) {
	dir := t.TempDir()
	src := `package p

// generate:reset
type B struct{}

func (x *B) Reset() {}


// generate:reset
type C struct{}
`
	if err := os.WriteFile(filepath.Join(dir, "b_and_c.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run(dir); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(dir, "reset.gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(out)
	if strings.Contains(body, "func (x *B) Reset() {}") {
		t.Fatalf("B already has Reset(), generated file must NOT include it")
	}
	if !strings.Contains(body, "func (x *C) Reset() {}") {
		t.Fatalf("Generated file must include Reset() for C, got: %s", body)
	}
}

func TestRun_GenericsTypeParams(t *testing.T) {
	dir := t.TempDir()
	src := `package p

// generate:reset
type Foo[T any, U comparable] struct{}
`
	if err := os.WriteFile(filepath.Join(dir, "foo.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run(dir); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	out, err := os.ReadFile(filepath.Join(dir, "reset.gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "func (x *Foo[T, U]) Reset() {}") {
		t.Fatalf("Reset() should carry type params T, U, got: %s", string(out))
	}
}

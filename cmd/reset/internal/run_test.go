package resettool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_GeneratesResetBodyWithRules(t *testing.T) {
	dir := t.TempDir()
	src := `package p

// generate:reset
type Child struct{ V int }
func (c *Child) Reset() { c.V = 0 }

// generate:reset
type Holder struct{
	I  int
	S  string
	B  bool
	SL []string
	M  map[string]int
	C  Child
	PC *Child
	PI *int
}
`
	if err := os.WriteFile(filepath.Join(dir, "types.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run(dir); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(dir, "reset.gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	out := string(body)

	// Примитивы
	wantPrims := []string{"x.I = 0", `x.S = ""`, "x.B = false"}
	for _, w := range wantPrims {
		if !strings.Contains(out, w) {
			t.Fatalf("missing primitive reset %q in generated body:\n%s", w, out)
		}
	}

	// Слайсы и мапы
	if !strings.Contains(out, "x.SL = x.SL[:0]") {
		t.Fatalf("missing slice trim in generated body:\n%s", out)
	}
	if !strings.Contains(out, "clear(x.M)") {
		t.Fatalf("missing map clear in generated body:\n%s", out)
	}

	// Вложенные структуры с Reset()
	if !strings.Contains(out, "_maybeReset(&x.C)") {
		t.Fatalf("nested struct with Reset must be called via helper:\n%s", out)
	}

	// Указатели: сброс значения и зануление
	if !strings.Contains(out, "if x.PC != nil") || !strings.Contains(out, "_maybeReset(x.PC)") || !strings.Contains(out, "x.PC = nil") {
		t.Fatalf("pointer to Child must call Reset and set nil:\n%s", out)
	}
	if !strings.Contains(out, "if x.PI != nil") || !strings.Contains(out, "*x.PI = 0") || !strings.Contains(out, "x.PI = nil") {
		t.Fatalf("pointer to int must set pointee zero and set nil:\n%s", out)
	}
}

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

func TestRun_ResetableStructRules(t *testing.T) {
	dir := t.TempDir()
	src := `package p

// generate:reset
type ResetableStruct struct {
	i     int
	str   string
	strP  *string
	s     []int
	m     map[string]string
	child *ResetableStruct
}
`
	if err := os.WriteFile(filepath.Join(dir, "types.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := Run(dir); err != nil {
		t.Fatalf("Run error: %v", err)
	}

	body, err := os.ReadFile(filepath.Join(dir, "reset.gen.go"))
	if err != nil {
		t.Fatal(err)
	}
	out := string(body)

	wants := []string{
		"x.i = 0",
		`x.str = ""`,
		"if x.strP != nil",
		`*x.strP = ""`,
		"x.s = x.s[:0]",
		"clear(x.m)",
		"if x.child != nil",
		"_maybeReset(x.child)",
	}
	for _, w := range wants {
		if !strings.Contains(out, w) {
			t.Fatalf("missing expected snippet %q in generated body:\n%s", w, out)
		}
	}
}

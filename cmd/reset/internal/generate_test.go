package resettool

import (
	"strings"
	"testing"
)

func TestBuildPackageContent_Render(t *testing.T) {
	p := PackageInfo{
		Package: "p",
		Structs: []StructInfo{
			{Name: "A"},
			{Name: "Foo", TypeParams: []string{"T", "U"}},
		},
	}
	out := BuildPackageContent(p)
	if !strings.HasPrefix(out, "package p\n\n") {
		t.Fatalf("must start with package header, got: %q", out[:min(20, len(out))])
	}
	if !strings.Contains(out, "func (x *A) Reset() {}") {
		t.Fatalf("must include A.Reset(), got: %s", out)
	}
	if !strings.Contains(out, "func (x *Foo[T, U]) Reset() {}") {
		t.Fatalf("must include Foo[T, U].Reset(), got: %s", out)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

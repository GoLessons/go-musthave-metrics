package resettool

import (
	"go/ast"
	"strings"
	"testing"
)

func TestBuildPackageContent_RenderWithHelpers(t *testing.T) {
	p := PackageInfo{
		Package: "p",
		Structs: []StructInfo{
			{
				Name: "A",
				Fields: []FieldInfo{
					{Names: []string{"I"}, Type: &ast.Ident{Name: "int"}},
					{Names: []string{"S"}, Type: &ast.Ident{Name: "string"}},
					{Names: []string{"B"}, Type: &ast.Ident{Name: "bool"}},
					{Names: []string{"M"}, Type: &ast.MapType{Key: &ast.Ident{Name: "string"}, Value: &ast.Ident{Name: "int"}}},
					{Names: []string{"SL"}, Type: &ast.ArrayType{Elt: &ast.Ident{Name: "string"}}}, // slice: Len == nil
					{Names: []string{"P"}, Type: &ast.StarExpr{X: &ast.Ident{Name: "int"}}},
				},
			},
		},
	}
	out := BuildPackageContent(p)

	if !strings.Contains(out, "type _resetter interface{ Reset() }") {
		t.Fatalf("missing helper resetter interface")
	}
	if !strings.Contains(out, "func _maybeReset[T any](v *T)") {
		t.Fatalf("missing helper _maybeReset")
	}
	if !strings.Contains(out, "x.I = 0") || !strings.Contains(out, `x.S = ""`) || !strings.Contains(out, "x.B = false") {
		t.Fatalf("primitives must be zeroed, out: %s", out)
	}
	if !strings.Contains(out, "clear(x.M)") {
		t.Fatalf("map must be cleared, out: %s", out)
	}
	if !strings.Contains(out, "x.SL = x.SL[:0]") {
		t.Fatalf("slice must be trimmed to [:0], out: %s", out)
	}
	if !strings.Contains(out, "if x.P != nil") || !strings.Contains(out, "*x.P = 0") || !strings.Contains(out, "x.P = nil") {
		t.Fatalf("pointer must set pointee to zero and then nil, out: %s", out)
	}
}

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

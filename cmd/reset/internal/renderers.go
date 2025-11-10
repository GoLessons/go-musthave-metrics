package resettool

import (
	"go/ast"
	"strings"
)

type exprRenderer interface {
	TypeString() string
	RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool)
}

func newExprRenderer(e ast.Expr) exprRenderer {
	switch t := e.(type) {
	case *ast.Ident:
		return &identRenderer{t}
	case *ast.ArrayType:
		// Слайс: ArrayType с Len == nil; иначе — массив.
		if t.Len == nil {
			return &sliceRenderer{t}
		}
		return &arrayRenderer{t}
	case *ast.MapType:
		return &mapRenderer{t}
	case *ast.StarExpr:
		return &starRenderer{t}
	case *ast.InterfaceType:
		return &interfaceRenderer{t}
	case *ast.FuncType:
		return &funcRenderer{t}
	case *ast.ChanType:
		return &chanRenderer{t}
	case *ast.SelectorExpr:
		return &selectorRenderer{t}
	case *ast.IndexExpr:
		return &indexRenderer{t}
	case *ast.IndexListExpr:
		return &indexListRenderer{t}
	case *ast.StructType:
		return &structRenderer{t}
	default:
		return &unknownRenderer{t}
	}
}

type identRenderer struct{ n *ast.Ident }

func (r *identRenderer) TypeString() string { return r.n.Name }

func (r *identRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	if isPrimitive(r.n.Name) {
		b.WriteString(indent)
		b.WriteString(fieldRef)
		b.WriteString(" = ")
		b.WriteString(zeroLiteralForIdent(r.n.Name))
		b.WriteString("\n")
		return
	}

	if callMaybeReset {
		b.WriteString(indent)
		b.WriteString("_maybeReset(&")
		b.WriteString(fieldRef)
		b.WriteString(")\n")
	}

	ts := r.TypeString()
	if ts != "" {
		b.WriteString(indent)
		b.WriteString("var _z ")
		b.WriteString(ts)
		b.WriteString("\n")
		b.WriteString(indent)
		b.WriteString(fieldRef)
		b.WriteString(" = _z\n")
	}
}

type selectorRenderer struct{ s *ast.SelectorExpr }

func (r *selectorRenderer) TypeString() string {
	prefix := newExprRenderer(r.s.X).TypeString()
	if prefix == "" {
		return ""
	}
	return prefix + "." + r.s.Sel.Name
}

func (r *selectorRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	if callMaybeReset {
		b.WriteString(indent)
		b.WriteString("_maybeReset(&")
		b.WriteString(fieldRef)
		b.WriteString(")\n")
	}
	ts := r.TypeString()
	if ts != "" {
		b.WriteString(indent)
		b.WriteString("var _z ")
		b.WriteString(ts)
		b.WriteString("\n")
		b.WriteString(indent)
		b.WriteString(fieldRef)
		b.WriteString(" = _z\n")
	}
}

type sliceRenderer struct{ s *ast.ArrayType }

func (r *sliceRenderer) TypeString() string {
	elt := newExprRenderer(r.s.Elt).TypeString()
	if elt == "" {
		return ""
	}
	return "[]" + elt
}

func (r *sliceRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	b.WriteString(indent)
	target := wrapIfStar(fieldRef)
	b.WriteString(fieldRef)
	b.WriteString(" = ")
	b.WriteString(target)
	b.WriteString("[:0]\n")
}

type mapRenderer struct{ m *ast.MapType }

func (r *mapRenderer) TypeString() string {
	k := newExprRenderer(r.m.Key).TypeString()
	v := newExprRenderer(r.m.Value).TypeString()
	if k == "" || v == "" {
		return ""
	}
	return "map[" + k + "]" + v
}

func (r *mapRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	b.WriteString(indent)
	b.WriteString("clear(")
	b.WriteString(fieldRef)
	b.WriteString(")\n")
}

type starRenderer struct{ s *ast.StarExpr }

func (r *starRenderer) TypeString() string {
	inner := newExprRenderer(r.s.X).TypeString()
	if inner == "" {
		return ""
	}
	return "*" + inner
}

func (r *starRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	b.WriteString(indent)
	b.WriteString("if ")
	b.WriteString(fieldRef)
	b.WriteString(" != nil {\n")

	// вызвать Reset у указателя, если он его реализует
	b.WriteString(indent)
	b.WriteString("\t_maybeReset(")
	b.WriteString(fieldRef)
	b.WriteString(")\n")

	inner := newExprRenderer(r.s.X)
	inner.RenderReset(b, "*"+fieldRef, indent+"\t", false)

	b.WriteString(indent)
	b.WriteString("\t")
	b.WriteString(fieldRef)
	b.WriteString(" = nil\n")
	b.WriteString(indent)
	b.WriteString("}\n")
}

type interfaceRenderer struct{ _ *ast.InterfaceType }

func (r *interfaceRenderer) TypeString() string { return "" }

func (r *interfaceRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	b.WriteString(indent)
	b.WriteString(fieldRef)
	b.WriteString(" = nil\n")
}

type funcRenderer struct{ _ *ast.FuncType }

func (r *funcRenderer) TypeString() string { return "" }

func (r *funcRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	b.WriteString(indent)
	b.WriteString(fieldRef)
	b.WriteString(" = nil\n")
}

type chanRenderer struct{ _ *ast.ChanType }

func (r *chanRenderer) TypeString() string { return "" }

func (r *chanRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	b.WriteString(indent)
	b.WriteString(fieldRef)
	b.WriteString(" = nil\n")
}

type arrayRenderer struct{ a *ast.ArrayType }

func (r *arrayRenderer) TypeString() string {
	elt := newExprRenderer(r.a.Elt).TypeString()
	if elt == "" {
		return ""
	}
	var lenStr string
	if r.a.Len != nil {
		lenStr = lenExprString(r.a.Len)
	}
	if lenStr == "" {
		return "[]" + elt
	}
	return "[" + lenStr + "]" + elt
}

func (r *arrayRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	ts := r.TypeString()
	if ts == "" {
		return
	}
	b.WriteString(indent)
	b.WriteString("var _z ")
	b.WriteString(ts)
	b.WriteString("\n")
	b.WriteString(indent)
	b.WriteString(fieldRef)
	b.WriteString(" = _z\n")
}

type indexRenderer struct{ i *ast.IndexExpr }

func (r *indexRenderer) TypeString() string {
	x := newExprRenderer(r.i.X).TypeString()
	idx := newExprRenderer(r.i.Index).TypeString()
	if x == "" || idx == "" {
		return ""
	}
	return x + "[" + idx + "]"
}

func (r *indexRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	if callMaybeReset {
		ts := r.TypeString()
		b.WriteString(indent)
		b.WriteString("_maybeReset(&")
		b.WriteString(fieldRef)
		b.WriteString(")\n")
		if ts != "" {
			b.WriteString(indent)
			b.WriteString("var _z ")
			b.WriteString(ts)
			b.WriteString("\n")
			b.WriteString(indent)
			b.WriteString(fieldRef)
			b.WriteString(" = _z\n")
		}
	}
}

type indexListRenderer struct{ i *ast.IndexListExpr }

func (r *indexListRenderer) TypeString() string {
	x := newExprRenderer(r.i.X).TypeString()
	if x == "" {
		return ""
	}
	var args []string
	for _, a := range r.i.Indices {
		s := newExprRenderer(a).TypeString()
		if s == "" {
			return ""
		}
		args = append(args, s)
	}
	return x + "[" + strings.Join(args, ", ") + "]"
}

func (r *indexListRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	if callMaybeReset {
		ts := r.TypeString()
		b.WriteString(indent)
		b.WriteString("_maybeReset(&")
		b.WriteString(fieldRef)
		b.WriteString(")\n")
		if ts != "" {
			b.WriteString(indent)
			b.WriteString("var _z ")
			b.WriteString(ts)
			b.WriteString("\n")
			b.WriteString(indent)
			b.WriteString(fieldRef)
			b.WriteString(" = _z\n")
		}
	}
}

type structRenderer struct{ _ *ast.StructType }

func (r *structRenderer) TypeString() string { return "" }

func (r *structRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	if callMaybeReset {
		b.WriteString(indent)
		b.WriteString("_maybeReset(&")
		b.WriteString(fieldRef)
		b.WriteString(")\n")
	}
}

type unknownRenderer struct{ _ ast.Expr }

func (r *unknownRenderer) TypeString() string { return "" }

func (r *unknownRenderer) RenderReset(b *strings.Builder, fieldRef string, indent string, callMaybeReset bool) {
	if callMaybeReset {
		b.WriteString(indent)
		b.WriteString("_maybeReset(&")
		b.WriteString(fieldRef)
		b.WriteString(")\n")
	}
}

func isPrimitive(name string) bool {
	switch name {
	case "string", "bool",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		"float32", "float64",
		"complex64", "complex128",
		"byte", "rune":
		return true
	default:
		return false
	}
}

func lenExprString(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.BasicLit:
		return v.Value
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		p := newExprRenderer(v.X).TypeString()
		if p == "" {
			return ""
		}
		return p + "." + v.Sel.Name
	default:
		return ""
	}
}

func wrapIfStar(s string) string {
	if strings.HasPrefix(s, "*") {
		return "(" + s + ")"
	}
	return s
}

func zeroLiteralForIdent(name string) string {
	switch name {
	case "string":
		return `""`
	case "bool":
		return "false"
	default:
		return "0"
	}
}

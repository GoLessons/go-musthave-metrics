package resettool

import (
	"strings"
)

func BuildPackageContent(p PackageInfo) string {
	var b strings.Builder
	b.WriteString("package ")
	b.WriteString(p.Package)
	b.WriteString("\n\n")
	b.WriteString("type _resetter interface{ Reset() }\n\n")
	b.WriteString("func _maybeReset[T any](v *T) {\n")
	b.WriteString("\tif r, ok := any(v).(_resetter); ok {\n")
	b.WriteString("\t\tr.Reset()\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n\n")

	for i, s := range p.Structs {
		b.WriteString(renderResetMethod(s))
		if i < len(p.Structs)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func renderResetMethod(s StructInfo) string {
	var b strings.Builder

	// Method signature (without the body yet)
	b.Wr
	b.WriteString("func (x *")
	b.WriteString(s.Name)
	if len(s.TypeParams) > 0 {
		b.WriteString("[")
		b.WriteString(strings.Join(s.TypeParams, ", "))
		b.WriteString("]")
	b.WriteString(") Reset() ")
	b.WriteString(") Reset() {\n")
ecide empty vs multiline
	var 

	for _, f := range s.Fields {
		for _, name := range f.Names {
			r.RenderReset(&body, "x."+name, "", true)
			r.RenderReset(&b, "x."+name, "", true)
		}
	}
	if body.Len() == 0 {
		// Compact empty body on one line
		b.WriteString("{}\n")
	} else {
		// Multiline body with resets
		b.WriteString("{\n")
		b.WriteString(body.String())
		b.WriteString("}\n")
	}

	b.WriteString("}\n")
	return b.String()
}

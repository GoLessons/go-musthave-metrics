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

	b.WriteString("func (x *")
	b.WriteString(s.Name)
	if len(s.TypeParams) > 0 {
		b.WriteString("[")
		b.WriteString(strings.Join(s.TypeParams, ", "))
		b.WriteString("]")
	}
	b.WriteString(") Reset() ")

	// Формируем тело отдельно, чтобы решить — пустое или многострочное
	var body strings.Builder
	for _, f := range s.Fields {
		for _, name := range f.Names {
			r := newExprRenderer(f.Type)
			r.RenderReset(&body, "x."+name, "", true)
		}
	}

	if body.Len() == 0 {
		// Компактное пустое тело в одну строку
		b.WriteString("{}\n")
	} else {
		// Многострочное тело со сбросами полей
		b.WriteString("{\n")
		b.WriteString(body.String())
		b.WriteString("}\n")
	}

	return b.String()
}

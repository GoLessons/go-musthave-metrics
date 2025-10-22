package resettool

import "strings"

func BuildPackageContent(p PackageInfo) string {
	var b strings.Builder
	b.WriteString("package ")
	b.WriteString(p.Package)
	b.WriteString("\n\n")

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
	b.WriteString(") Reset() {}\n")

	return b.String()
}

package resettool

import "go/ast"

type FieldInfo struct {
	Names    []string
	Type     ast.Expr
	Embedded bool
}

type StructInfo struct {
	Name       string
	TypeParams []string
	Fields     []FieldInfo
}

type PackageInfo struct {
	Dir     string
	Package string
	Structs []StructInfo
}

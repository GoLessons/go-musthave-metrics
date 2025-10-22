package resettool

type StructInfo struct {
	Name       string
	TypeParams []string
}

type PackageInfo struct {
	Dir     string
	Package string
	Structs []StructInfo
}

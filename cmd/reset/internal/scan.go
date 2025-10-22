package resettool

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

func ScanPackages(dir string) ([]PackageInfo, error) {
	fset := token.NewFileSet()
	var result []PackageInfo

	walkErr := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if shouldSkipDir(d.Name()) {
			return filepath.SkipDir
		}

		pkgs, err := parser.ParseDir(fset, path, includeFile, parser.ParseComments)
		if err != nil {
			return nil
		}
		for _, pkg := range pkgs {
			if pinfo, ok := buildPackageInfo(path, pkg, fset); ok {
				result = append(result, pinfo)
			}
		}
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	return result, nil
}

func shouldSkipDir(name string) bool {
	if name == ".git" || name == "vendor" || name == "node_modules" {
		return true
	}
	if strings.HasPrefix(name, ".") && name != "." {
		return true
	}
	return false
}

func includeFile(info fs.FileInfo) bool {
	n := info.Name()
	if !strings.HasSuffix(n, ".go") {
		return false
	}
	if strings.HasSuffix(n, "_test.go") {
		return false
	}
	if strings.HasSuffix(n, ".gen.go") {
		return false
	}
	if n == "reset.gen.go" {
		return false
	}
	return true
}

func containsResetTag(cg *ast.CommentGroup) bool {
	if cg == nil {
		return false
	}
	for _, c := range cg.List {
		if strings.Contains(c.Text, "generate:reset") {
			return true
		}
	}
	return false
}

func hasGenerateResetTag(f *ast.File, fset *token.FileSet, ts *ast.TypeSpec) bool {
	if containsResetTag(ts.Doc) {
		return true
	}

	specLine := fset.Position(ts.Pos()).Line
	var nearest *ast.CommentGroup
	nearestEnd := 0

	for _, cg := range f.Comments {
		endLine := fset.Position(cg.End()).Line
		if endLine < specLine && endLine > nearestEnd {
			nearest = cg
			nearestEnd = endLine
		}
	}

	if nearest != nil && specLine-nearestEnd <= 1 {
		return containsResetTag(nearest)
	}

	return false
}

func collectExistingResetMethods(pkg *ast.Package) map[string]bool {
	m := make(map[string]bool)
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv == nil {
				continue
			}
			if fd.Name.Name != "Reset" {
				continue
			}
			if len(fd.Recv.List) == 0 {
				continue
			}
			base := receiverBaseName(fd.Recv.List[0].Type)
			if base != "" {
				m[base] = true
			}
		}
	}
	return m
}

func receiverBaseName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverBaseName(t.X)
	case *ast.IndexExpr:
		return receiverBaseName(t.X)
	case *ast.IndexListExpr:
		return receiverBaseName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

func buildPackageInfo(path string, pkg *ast.Package, fset *token.FileSet) (PackageInfo, bool) {
	pinfo := PackageInfo{
		Dir:     path,
		Package: pkg.Name,
	}

	existing := collectExistingResetMethods(pkg)
	seen := make(map[string]struct{})

	for _, f := range pkg.Files {
		collectStructs(f, fset, existing, seen, &pinfo)
	}

	if len(pinfo.Structs) == 0 {
		return PackageInfo{}, false
	}

	sort.Slice(pinfo.Structs, func(i, j int) bool {
		return pinfo.Structs[i].Name < pinfo.Structs[j].Name
	})

	return pinfo, true
}

func collectStructs(f *ast.File, fset *token.FileSet, existing map[string]bool, seen map[string]struct{}, pinfo *PackageInfo) {
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			continue
		}

		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, ok := ts.Type.(*ast.StructType); !ok {
				continue
			}
			if !hasGenerateResetTag(f, fset, ts) {
				continue
			}

			name := ts.Name.Name
			if _, dup := seen[name]; dup {
				continue
			}
			if existing[name] {
				continue
			}

			pinfo.Structs = append(pinfo.Structs, StructInfo{
				Name:       name,
				TypeParams: extractTypeParams(ts),
			})
			seen[name] = struct{}{}
		}
	}
}

func extractTypeParams(ts *ast.TypeSpec) []string {
	if ts.TypeParams == nil {
		return nil
	}
	var params []string
	for _, fld := range ts.TypeParams.List {
		for _, n := range fld.Names {
			params = append(params, n.Name)
		}
	}
	return params
}

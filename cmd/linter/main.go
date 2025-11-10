package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
)

var ExitCheckAnalyzer = &analysis.Analyzer{
	Name:     "exitcheck",
	Doc:      "использование panic и os.Exit/log.Fatal вне main.main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func main() {
	singlechecker.Main(ExitCheckAnalyzer)
}

func run(pass *analysis.Pass) (interface{}, error) {
	ins, _ := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if ins == nil {
		return nil, nil
	}

	ins.WithStack([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}
		call := n.(*ast.CallExpr)

		// 1) panic (встроенная функция)
		if isCall(pass, call.Fun, "panic") {
			pass.Reportf(call.Pos(), "использование panic запрещено")
		}

		// 2) os.Exit / log.Fatal (пакетные функции)
		isExit := isCall(pass, call.Fun, "Exit", "os")
		isFatal := isCall(pass, call.Fun, "Fatal", "log")
		if isExit || isFatal {
			allowed := pass.Pkg.Name() == "main" && insideFunc(stack, "main")
			if !allowed {
				fn := "os.Exit"
				if isFatal {
					fn = "log.Fatal"
				}

				pass.Reportf(call.Pos(), "вызов %s вне функции main пакета main", fn)
			}
		}

		return true
	})

	return nil, nil
}

func insideFunc(stack []ast.Node, name string) bool {
	for i := len(stack) - 1; i >= 0; i-- {
		if fd, ok := stack[i].(*ast.FuncDecl); ok {
			return fd.Name.Name == name
		}
	}

	return false
}

func isCall(pass *analysis.Pass, expr ast.Expr, name string, pkgPathOpt ...string) bool {
	// Без pkgPath: ожидаем встроенную функцию (например, panic).
	if len(pkgPathOpt) == 0 {
		id, ok := expr.(*ast.Ident)
		if !ok {
			return false
		}
		if b, ok := pass.TypesInfo.Uses[id].(*types.Builtin); ok {
			return b.Name() == name
		}
		return false
	}

	// С pkgPath: ожидаем пакетную функцию (например, os.Exit или log.Fatal).
	pkgPath := pkgPathOpt[0]
	switch f := expr.(type) {
	case *ast.Ident:
		// dot-импорт: вызывается как идентификатор
		if fn, ok := pass.TypesInfo.Uses[f].(*types.Func); ok {
			return fn.Pkg() != nil && fn.Pkg().Path() == pkgPath && fn.Name() == name
		}
	case *ast.SelectorExpr:
		// обычный селектор pkg.Func или метод; семантика как в текущем коде
		if fn, ok := pass.TypesInfo.Uses[f.Sel].(*types.Func); ok {
			return fn.Pkg() != nil && fn.Pkg().Path() == pkgPath && fn.Name() == name
		}
	}

	return false
}

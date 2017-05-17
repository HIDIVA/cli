package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type warning struct {
	message string
	token.Position
}

type visitor struct {
	fileSet *token.FileSet

	constSpecs []string
	funcDecls  []string
	typeSpecs  []string
	varSpecs   []string
	warnings   []warning
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch typedNode := node.(type) {
	case *ast.File:
		return v
	case *ast.Package:
		return v
	case *ast.GenDecl:
		if typedNode.Tok == token.CONST {
			v.checkConst(typedNode)
		} else if typedNode.Tok == token.VAR {
			v.checkVar(typedNode)
		}
		return v
	case *ast.FuncDecl:
		v.checkFunc(typedNode)
	case *ast.TypeSpec:
		v.checkType(typedNode)
	}

	return nil
}

func (v *visitor) checkConst(node *ast.GenDecl) {
	constName := node.Specs[0].(*ast.ValueSpec).Names[0].Name
	v.constSpecs = append(v.constSpecs, constName)
	if len(v.funcDecls) != 0 {
		v.addWarning(fmt.Sprintf("constant '%s' comes after a function declaration", constName), node.Pos())
	}
	if len(v.typeSpecs) != 0 {
		v.addWarning(fmt.Sprintf("constant '%s' comes after a type declaration", constName), node.Pos())
	}
	if len(v.varSpecs) != 0 {
		v.addWarning(fmt.Sprintf("constant '%s' comes after a variable declaration", constName), node.Pos())
	}
}

func (v *visitor) checkVar(node *ast.GenDecl) {
	varName := node.Specs[0].(*ast.ValueSpec).Names[0].Name
	v.varSpecs = append(v.varSpecs, varName)
	if len(v.funcDecls) != 0 {
		v.addWarning(fmt.Sprintf("variable '%s' comes after a function declaration", varName), node.Pos())
	}
	if len(v.typeSpecs) != 0 {
		v.addWarning(fmt.Sprintf("variable '%s' comes after a type declaration", varName), node.Pos())
	}
}

func (v *visitor) checkFunc(node *ast.FuncDecl) {
	funcName := node.Name.Name

	if node.Recv != nil {
		var receiver string
		switch typedType := node.Recv.List[0].Type.(type) {
		case *ast.Ident:
			receiver = typedType.Name
		case *ast.StarExpr:
			receiver = typedType.X.(*ast.Ident).Name
		}
		if len(v.typeSpecs) > 0 {
			lastTypeSpec := v.typeSpecs[len(v.typeSpecs)-1]
			if receiver != lastTypeSpec {
				v.addWarning(fmt.Sprintf("method '%s' of '%s' must be defined immediately after type '%s'", funcName, receiver, receiver), node.Pos())
			}
		}
	} else {
		v.funcDecls = append(v.funcDecls, funcName)
	}
}

func (v *visitor) checkType(node *ast.TypeSpec) {
	typeName := node.Name.Name
	v.typeSpecs = append(v.typeSpecs, typeName)
	if len(v.funcDecls) != 0 {
		v.addWarning(fmt.Sprintf("type declaration for '%s' comes after a function declaration", typeName), node.Pos())
	}
}

func (v *visitor) addWarning(message string, pos token.Pos) {
	v.warnings = append(v.warnings, warning{
		message:  message,
		Position: v.fileSet.Position(pos),
	})
}

func shouldParseFile(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

func main() {
	fileSet := token.NewFileSet()

	v := visitor{
		fileSet: fileSet,
	}

	err := filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		base := filepath.Base(path)
		if base == "vendor" || base == ".git" || strings.HasSuffix(base, "fakes") {
			return filepath.SkipDir
		}

		packages, err := parser.ParseDir(fileSet, path, shouldParseFile, 0)
		if err != nil {
			return err
		}

		for _, packag := range packages {
			// fmt.Println(ast.Print(fileSet, packag))
			ast.Walk(&v, packag)
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, warning := range v.warnings {
		fmt.Printf("%s:%d %s\n", warning.Position.Filename, warning.Position.Line, warning.message)
	}

	if len(v.warnings) > 0 {
		os.Exit(1)
	}
}

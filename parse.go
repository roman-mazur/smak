package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func parse(p string) (*ast.File, error) {
	i, err := os.Stat(p)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	if !i.IsDir() {
		return parseFile(fset, p)
	}
	pkgs, err := parser.ParseDir(fset, p, func(info fs.FileInfo) bool {
		return filepath.Ext(info.Name()) == ".go" && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	pkg := pkgs[filepath.Base(p)]
	return ast.MergePackageFiles(pkg, 0), nil
}

func parseFile(fset *token.FileSet, p string) (*ast.File, error) {
	content, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return parser.ParseFile(fset, filepath.Base(p), content, 0)
}

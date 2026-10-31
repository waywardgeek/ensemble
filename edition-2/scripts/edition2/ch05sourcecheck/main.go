// Source properties discovered across every nested module, not a fixed spoke list.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type finding struct {
	File string `json:"file"`
	Why  string `json:"why"`
}

func inspect(root string) ([]finding, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, err
	}
	module := ""
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) > 1 && f[0] == "module" {
			module = f[1]
		}
	}
	if module == "" {
		return nil, fmt.Errorf("module declaration absent")
	}
	out := []finding{}
	set := token.NewFileSet()
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") || d.Name() == "evidence" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		add := func(why string) { out = append(out, finding{filepath.ToSlash(rel), why}) }
		for _, spec := range file.Imports {
			p, _ := strconv.Unquote(spec.Path.Value)
			if file.Name.Name == "main" && strings.HasPrefix(p, module+"/internal/") {
				add("executable reaches private first-party package: " + p)
			}
			if !strings.HasPrefix(filepath.ToSlash(rel), "gui/") && (strings.Contains(p, "websocket") || p == module+"/gui") {
				add("headless library requires GUI transport: " + p)
			}
		}
		for _, decl := range file.Decls {
			g, ok := decl.(*ast.GenDecl)
			if !ok || g.Tok != token.VAR {
				continue
			}
			for _, spec := range g.Specs {
				v := spec.(*ast.ValueSpec)
				for i, name := range v.Names {
					if name.Name == "_" {
						continue
					} // compile-time interface assertion
					if i >= len(v.Values) {
						add("package-owned mutable zero value: " + name.Name)
						continue
					}
					// Literal, read-only lookup tables are permitted. Session allocation,
					// constructors, channels and injected services are never lookup tables.
					mutable := false
					ast.Inspect(v.Values[i], func(n ast.Node) bool {
						switch n.(type) {
						case *ast.CallExpr, *ast.FuncLit:
							mutable = true
						}
						return true
					})
					ast.Inspect(file, func(n ast.Node) bool {
						var targets []ast.Expr
						switch node := n.(type) {
						case *ast.AssignStmt:
							targets = node.Lhs
						case *ast.IncDecStmt:
							targets = []ast.Expr{node.X}
						}
						for _, target := range targets {
							ast.Inspect(target, func(n ast.Node) bool {
								id, ok := n.(*ast.Ident)
								if ok && id.Obj == name.Obj {
									mutable = true
								}
								return true
							})
						}
						return true
					})
					if mutable {
						add("package-owned runtime mutable value: " + name.Name)
					}
				}
			}
		}
		return nil
	})
	return out, err
}
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: ch05sourcecheck SOURCE")
		os.Exit(2)
	}
	findings, err := inspect(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	json.NewEncoder(os.Stdout).Encode(findings)
	if len(findings) > 0 {
		os.Exit(1)
	}
}

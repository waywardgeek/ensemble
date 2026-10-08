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
		file, err := parser.ParseFile(set, path, nil, parser.ParseComments)
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
					if immutableDeclaration(file, g, v, i) && !writtenAnywhere(root, path, name.Name) {
						continue
					}
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
		fmt.Fprintln(os.Stderr, "usage: ch07sourcecheck SOURCE")
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

// These are immutable library values, not session objects. Recognition follows
// the actual imported package and declaration, never a conventional variable name.
func importedAs(file *ast.File, name, path string) bool {
	for _, spec := range file.Imports {
		value, _ := strconv.Unquote(spec.Path.Value)
		alias := filepath.Base(value)
		if spec.Name != nil {
			alias = spec.Name.Name
		}
		if alias == name && value == path {
			return true
		}
	}
	return false
}
func immutableDeclaration(file *ast.File, g *ast.GenDecl, v *ast.ValueSpec, index int) bool {
	if index < len(v.Values) {
		call, ok := v.Values[index].(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return false
		}
		fun, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		pkg, ok := fun.X.(*ast.Ident)
		if !ok || !((fun.Sel.Name == "New" && importedAs(file, pkg.Name, "errors")) ||
			(fun.Sel.Name == "Errorf" && importedAs(file, pkg.Name, "fmt"))) {
			return false
		}
		literal, ok := call.Args[0].(*ast.BasicLit)
		return ok && literal.Kind == token.STRING
	}
	if len(v.Names) != 1 || len(v.Values) != 0 {
		return false
	}
	typ, ok := v.Type.(*ast.SelectorExpr)
	if !ok || typ.Sel.Name != "FS" {
		return false
	}
	pkg, ok := typ.X.(*ast.Ident)
	if !ok || !importedAs(file, pkg.Name, "embed") {
		return false
	}
	for _, group := range []*ast.CommentGroup{g.Doc, v.Doc} {
		if group != nil {
			for _, c := range group.List {
				if strings.HasPrefix(c.Text, "//go:embed ") {
					return true
				}
			}
		}
	}
	return false
}
func packageImport(root, dir string) string {
	for current := dir; ; current = filepath.Dir(current) {
		if data, err := os.ReadFile(filepath.Join(current, "go.mod")); err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				f := strings.Fields(line)
				if len(f) > 1 && f[0] == "module" {
					rel, _ := filepath.Rel(current, dir)
					if rel == "." {
						return f[1]
					}
					return f[1] + "/" + filepath.ToSlash(rel)
				}
			}
		}
		if current == root || current == filepath.Dir(current) {
			return ""
		}
	}
}

// Reject direct writes and address escape in every production file, including
// writes from another file or an importing package. Local shadows are distinct.
func writtenAnywhere(root, declaration, name string) bool {
	ownerDir := filepath.Dir(declaration)
	ownerImport := packageImport(root, ownerDir)
	written := false
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			written = true
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") || d.Name() == "evidence" || d.Name() == "vendor" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			written = true
			return err
		}
		globals := map[*ast.Object]bool{}
		for _, decl := range file.Decls {
			g, ok := decl.(*ast.GenDecl)
			if !ok || g.Tok != token.VAR {
				continue
			}
			for _, spec := range g.Specs {
				v := spec.(*ast.ValueSpec)
				for _, n := range v.Names {
					if n.Name == name {
						globals[n.Obj] = true
					}
				}
			}
		}
		same := filepath.Dir(path) == ownerDir
		targetContains := func(target ast.Expr) bool {
			found := false
			ast.Inspect(target, func(node ast.Node) bool {
				switch x := node.(type) {
				case *ast.SelectorExpr:
					if id, ok := x.X.(*ast.Ident); ok && x.Sel.Name == name && importedAs(file, id.Name, ownerImport) {
						found = true
					}
				case *ast.Ident:
					if x.Name == name && ((same && (x.Obj == nil || globals[x.Obj])) || (!same && x.Obj == nil && importedAs(file, ".", ownerImport))) {
						found = true
					}
				}
				return !found
			})
			return found
		}
		ast.Inspect(file, func(node ast.Node) bool {
			var targets []ast.Expr
			switch x := node.(type) {
			case *ast.AssignStmt:
				targets = x.Lhs
			case *ast.IncDecStmt:
				targets = []ast.Expr{x.X}
			case *ast.UnaryExpr:
				if x.Op == token.AND {
					targets = []ast.Expr{x.X}
				}
			}
			for _, target := range targets {
				if targetContains(target) {
					written = true
				}
			}
			return !written
		})
		return nil
	})
	return written
}

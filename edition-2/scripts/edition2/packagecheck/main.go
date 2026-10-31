// packagecheck checks source-level package boundaries, not ownership or behavior.
// Constructor paths and interface dispatch still require independent review.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type finding struct {
	ID   string `json:"id"`
	File string `json:"file"`
	Why  string `json:"why"`
}

func inspect(root string) ([]finding, error) {
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, err
	}
	var module string
	for _, line := range strings.Split(string(mod), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == "module" {
			module = strings.Trim(fields[1], "\"")
		}
	}
	if module == "" {
		return nil, fmt.Errorf("missing module declaration")
	}
	findings := []finding{}
	fset := token.NewFileSet()
	var commonAST []*ast.File
	commonFunctions := map[string]*ast.FuncDecl{}
	commonFiles := map[string]string{}
	methodFiles := map[*ast.FuncDecl]string{}
	var interfaceMethods []*ast.FuncDecl
	invalidCommonImports := false
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root {
				if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "vendor" || entry.Name() == "testdata" {
					return filepath.SkipDir
				}
				if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
					return filepath.SkipDir // a different module needs its own check
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		dir := filepath.ToSlash(filepath.Dir(rel))
		file, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		add := func(id, why string) { findings = append(findings, finding{id, rel, why}) }
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if importPath != module && !strings.HasPrefix(importPath, module+"/") {
				continue
			}
			if file.Name.Name == "main" && strings.HasPrefix(importPath, module+"/internal/") {
				add("cli-public", "executable imports private implementation: "+importPath)
			}
			if dir == "internal/common" {
				add("common-import", "vocabulary package imports first-party implementation: "+importPath)
				invalidCommonImports = true
			} else if strings.HasPrefix(dir, "internal/") && importPath != module+"/internal/common" {
				// Discover every internal package; a new spoke must not escape the rule.
				add("star-import", "implementation imports outside common: "+importPath)
			}
		}
		if dir != "internal/common" {
			return nil
		}
		commonAST = append(commonAST, file)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv == nil {
				commonFunctions[fn.Name.Name], commonFiles[fn.Name.Name] = fn, rel
				continue
			}
			interfaceMethods = append(interfaceMethods, fn)
			methodFiles[fn] = rel
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Resolve symbols only after imports are sound. No student initializer runs.
	if invalidCommonImports {
		return findings, nil
	}
	info := &types.Info{Defs: map[*ast.Ident]types.Object{}, Uses: map[*ast.Ident]types.Object{}}
	config := &types.Config{Importer: importer.Default()}
	if len(commonAST) > 0 {
		if _, err := config.Check(module+"/internal/common", fset, commonAST, info); err != nil {
			return nil, fmt.Errorf("common type analysis incomplete: %w", err)
		}
	}
	helpers := map[types.Object]*ast.FuncDecl{}
	for _, fn := range commonFunctions {
		helpers[info.Defs[fn.Name]] = fn
	}
	// A helper used by permitted interface dispatch may stay beside that method.
	// Reachability alone does not prove that its semantics belong here: review does.
	reached := map[string]bool{}
	var visit func(*ast.FuncDecl)
	visit = func(fn *ast.FuncDecl) {
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			id, ok := node.(*ast.Ident)
			if ok {
				if helper := helpers[info.Uses[id]]; helper != nil && !reached[helper.Name.Name] {
					reached[helper.Name.Name] = true
					visit(helper)
				}
			}
			return true
		})
	}
	for _, fn := range interfaceMethods {
		method := info.Defs[fn.Name].(*types.Func)
		if standardMethod(method.Name(), method.Type().(*types.Signature)) {
			visit(fn)
		} else {
			findings = append(findings, finding{"common-behavior", methodFiles[fn], "method is not permitted standard-library dispatch: " + fn.Name.Name})
		}
	}
	for name := range commonFunctions {
		if !reached[name] {
			findings = append(findings, finding{"common-behavior", commonFiles[name], "free function outside interface-dispatch helpers: " + name})
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		return a.ID+a.File+a.Why < b.ID+b.File+b.Why
	})
	return findings, nil
}

func standardMethod(name string, sig *types.Signature) bool {
	if sig.Variadic() {
		return false
	}
	var params, results []types.Type
	bytesType := types.NewSlice(types.Typ[types.Byte])
	errorType := types.Universe.Lookup("error").Type()
	switch name {
	case "String", "Error":
		results = []types.Type{types.Typ[types.String]}
	case "MarshalJSON", "MarshalText":
		results = []types.Type{bytesType, errorType}
	case "UnmarshalJSON", "UnmarshalText":
		params, results = []types.Type{bytesType}, []types.Type{errorType}
	default:
		return false
	}
	match := func(tuple *types.Tuple, want []types.Type) bool {
		if tuple.Len() != len(want) {
			return false
		}
		for i, typ := range want {
			if !types.Identical(tuple.At(i).Type(), typ) {
				return false
			}
		}
		return true
	}
	return match(sig.Params(), params) && match(sig.Results(), results)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: packagecheck MODULE_ROOT")
		os.Exit(2)
	}
	root, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	findings, err := inspect(root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := json.NewEncoder(os.Stdout).Encode(struct {
		Scope    string    `json:"scope"`
		Passed   bool      `json:"passed"`
		Findings []finding `json:"findings"`
	}{"module imports and common behavior placement; ownership and semantics require review", len(findings) == 0, findings}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if len(findings) != 0 {
		os.Exit(1)
	}
}

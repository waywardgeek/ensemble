package grade

// Framework-blind structural analysis of a student source tree.
//
// This scanner backs two checks: Chapter 5's logger-accessible check and
// Chapter 22's back-pointer-chain check. They share it on purpose, so that
// the two chapters cannot drift apart on what "back-pointer" means.
//
// Nothing in this file knows the identifiers common.Host, common.Agent or
// common.Call. A tree that names its hub interface Parent and its dispatch
// struct Invocation is scored on exactly the same structural property as the
// reference implementation. That is the entire point: the version of the
// Chapter 5 check this replaces grepped the tree for the strings "Logf" and
// "Host" and reported "Host interface with Logf found, embedded in Call" --
// a message describing a structure it never looked at. It passed any tree
// containing a logging helper and an unrelated identifier, and it would have
// failed the reference implementation the moment the interface was renamed.
// A check that a rename can kill is a check on vocabulary.

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// astField is one struct field or one function parameter.
type astField struct {
	Name     string // "" for an embedded field
	Type     string // type name with the qualifier stripped
	Qual     string // package qualifier, "" when declared locally
	Embedded bool
	IsFunc   bool // the type is a func literal type: a closure staple
	Ptr      bool
}

// Key returns the "pkg.Name" lookup key for the field's type.
func (f astField) Key(localPkg string) string {
	if f.Qual != "" {
		return f.Qual + "." + f.Type
	}
	return localPkg + "." + f.Type
}

type astMethod struct {
	Name     string
	Variadic bool
	Results  int
	FmtFirst bool // first parameter is a plain string: a format string
	// ResultTypes are the method's result types. Chapter 22 needs these to
	// find the SECOND link of a back-pointer chain: a method whose result is
	// itself an interface is how a parent hands you its own parent without
	// the dispatch struct growing a field.
	ResultTypes []astField
}

type astType struct {
	Pkg     string
	Name    string
	File    string
	Line    int
	IsIface bool
	Methods []astMethod // interfaces only
	Fields  []astField  // structs only
}

func (t *astType) Key() string { return t.Pkg + "." + t.Name }

type astFunc struct {
	Pkg      string
	Name     string
	Recv     string // receiver type name, "" for plain functions
	File     string
	Line     int
	Params   []astField
	body     *ast.BlockStmt
	IsMethod bool
}

type astScan struct {
	Root  string
	Types map[string]*astType
	Funcs []*astFunc
	fset  *token.FileSet
}

// scanTree parses every non-test .go file under dir.
func scanTree(dir string) (*astScan, error) {
	s := &astScan{
		Root:  dir,
		Types: map[string]*astType{},
		fset:  token.NewFileSet(),
	}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if base == "node_modules" || base == "vendor" || base == ".git" ||
				base == "testdata" || strings.HasPrefix(base, ".") && base != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, perr := parser.ParseFile(s.fset, path, nil, parser.SkipObjectResolution)
		if perr != nil {
			return nil // an unparseable file is the build check's problem
		}
		s.addFile(path, file)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(s.Types) == 0 {
		return nil, fmt.Errorf("no Go types found under %s", dir)
	}
	return s, nil
}

func (s *astScan) addFile(path string, file *ast.File) {
	pkg := file.Name.Name
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				t := &astType{
					Pkg:  pkg,
					Name: ts.Name.Name,
					File: path,
					Line: s.fset.Position(ts.Pos()).Line,
				}
				switch u := ts.Type.(type) {
				case *ast.InterfaceType:
					t.IsIface = true
					for _, m := range u.Methods.List {
						ft, ok := m.Type.(*ast.FuncType)
						if !ok || len(m.Names) == 0 {
							continue
						}
						t.Methods = append(t.Methods, astMethod{
							Name:        m.Names[0].Name,
							Variadic:    isVariadic(ft),
							Results:     resultCount(ft),
							FmtFirst:    firstParamIsString(ft),
							ResultTypes: resultTypes(ft),
						})
					}
				case *ast.StructType:
					for _, f := range u.Fields.List {
						t.Fields = append(t.Fields, fieldsOf(f)...)
					}
				default:
					continue
				}
				s.Types[t.Key()] = t
			}
		case *ast.FuncDecl:
			fn := &astFunc{
				Pkg:  pkg,
				Name: d.Name.Name,
				File: path,
				Line: s.fset.Position(d.Pos()).Line,
				body: d.Body,
			}
			if d.Recv != nil && len(d.Recv.List) > 0 {
				fn.IsMethod = true
				rf := fieldsOf(d.Recv.List[0])
				if len(rf) > 0 {
					fn.Recv = rf[0].Type
				}
			}
			if d.Type.Params != nil {
				for _, p := range d.Type.Params.List {
					fn.Params = append(fn.Params, fieldsOf(p)...)
				}
			}
			s.Funcs = append(s.Funcs, fn)
		}
	}
}

// fieldsOf flattens one ast.Field (which may declare several names) into
// astField values.
func fieldsOf(f *ast.Field) []astField {
	base := astField{}
	typ := f.Type
	if star, ok := typ.(*ast.StarExpr); ok {
		base.Ptr = true
		typ = star.X
	}
	switch t := typ.(type) {
	case *ast.Ident:
		base.Type = t.Name
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			base.Qual = x.Name
		}
		base.Type = t.Sel.Name
	case *ast.FuncType:
		base.IsFunc = true
		base.Type = "func"
	default:
		base.Type = ""
	}
	if len(f.Names) == 0 {
		base.Embedded = true
		return []astField{base}
	}
	var out []astField
	for _, n := range f.Names {
		g := base
		g.Name = n.Name
		out = append(out, g)
	}
	return out
}

func isVariadic(ft *ast.FuncType) bool {
	if ft.Params == nil || len(ft.Params.List) == 0 {
		return false
	}
	last := ft.Params.List[len(ft.Params.List)-1]
	_, ok := last.Type.(*ast.Ellipsis)
	return ok
}

func resultCount(ft *ast.FuncType) int {
	if ft.Results == nil {
		return 0
	}
	n := 0
	for _, r := range ft.Results.List {
		if len(r.Names) == 0 {
			n++
		} else {
			n += len(r.Names)
		}
	}
	return n
}

// resultTypes describes a method's result types, reusing the same field
// parser used for struct fields so that a result and a field are judged by
// identical rules.
func resultTypes(ft *ast.FuncType) []astField {
	if ft.Results == nil {
		return nil
	}
	var out []astField
	for _, r := range ft.Results.List {
		out = append(out, fieldsOf(r)...)
	}
	return out
}

// isLogShaped identifies a formatted-logging method by its SIGNATURE rather
// than by its name: it takes a format string, is variadic, and returns
// nothing. Printf, Logf, Debugf, Emitf and a student's own Tracef all match;
// an accessor such as Logger() does not, because it returns a value and
// merely hands out a dependency instead of being the dependency.
//
// Matching on shape is deliberate. The check this file backs replaced one
// that grepped for the literal string "Logf", and anchoring on the name
// here, even partially, would reintroduce the same defect in a quieter form.
func isLogShaped(m astMethod) bool {
	return m.Variadic && m.Results == 0 && m.FmtFirst
}

func firstParamIsString(ft *ast.FuncType) bool {
	if ft.Params == nil || len(ft.Params.List) == 0 {
		return false
	}
	id, ok := ft.Params.List[0].Type.(*ast.Ident)
	return ok && id.Name == "string"
}

// loggingInterfaces returns every interface in the tree exposing a
// formatted-logging method.
func (s *astScan) loggingInterfaces() []*astType {
	var out []*astType
	for _, t := range s.Types {
		if !t.IsIface {
			continue
		}
		for _, m := range t.Methods {
			if isLogShaped(m) {
				out = append(out, t)
				break
			}
		}
	}
	return out
}

// loggingMethods returns the names of the logging methods on an interface.
func loggingMethods(t *astType) []string {
	var out []string
	for _, m := range t.Methods {
		if isLogShaped(m) {
			out = append(out, m.Name)
		}
	}
	return out
}

// storedBackPointerUse looks for the parent chain actually being walked: a
// struct in some OTHER package that stores the interface as a field (which
// is what a back-pointer constructor parameter turns into) and a method on
// that struct that calls the logging method through the stored field.
//
// Matching is on the selector shape x.<field>.<LogMethod>(...) rather than on
// the receiver's name, because receiver names are a style choice.
func (s *astScan) storedBackPointerUse(iface *astType) (string, bool) {
	methods := loggingMethods(iface)
	for _, t := range s.Types {
		if t.IsIface || t.Pkg == iface.Pkg {
			continue // the holder must be a different package than the hub
		}
		for _, f := range t.Fields {
			if f.IsFunc || f.Embedded || f.Key(t.Pkg) != iface.Key() {
				continue
			}
			for _, fn := range s.Funcs {
				if !fn.IsMethod || fn.Pkg != t.Pkg || fn.Recv != t.Name || fn.body == nil {
					continue
				}
				if m, ok := callsFieldMethod(fn.body, f.Name, methods); ok {
					return fmt.Sprintf("%s stores %s as field %q and calls .%s.%s in %s at %s:%d",
						t.Key(), iface.Key(), f.Name, f.Name, m, fn.Name,
						filepath.Base(fn.File), fn.Line), true
				}
			}
		}
	}
	return "", false
}

// callsFieldMethod reports whether the body contains a call of the shape
// x.field.Method(...) for one of the named methods.
func callsFieldMethod(body *ast.BlockStmt, field string, methods []string) (string, bool) {
	found := ""
	ast.Inspect(body, func(n ast.Node) bool {
		if found != "" {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		inner, ok := sel.X.(*ast.SelectorExpr)
		if !ok || inner.Sel.Name != field {
			return true
		}
		for _, m := range methods {
			if sel.Sel.Name == m {
				found = m
				return false
			}
		}
		return true
	})
	return found, found != ""
}

// dispatchStructFor finds the tool dispatch struct: a struct declared in the
// hub package alongside the interface, carrying that interface as a field,
// and handed to functions in at least one other package. That last clause is
// what makes it a DISPATCH struct rather than just another struct -- it is
// the value the hub passes outward to code that knows nothing else.
func (s *astScan) dispatchStructFor(iface *astType) (string, bool) {
	for _, t := range s.Types {
		if t.IsIface || t.Pkg != iface.Pkg {
			continue
		}
		carried := ""
		for _, f := range t.Fields {
			if f.IsFunc || f.Key(t.Pkg) != iface.Key() {
				continue
			}
			if f.Embedded {
				carried = "embedded"
			} else {
				carried = "field " + f.Name
			}
			break
		}
		if carried == "" {
			continue
		}
		for _, fn := range s.Funcs {
			if fn.Pkg == t.Pkg {
				continue
			}
			for _, p := range fn.Params {
				if p.Key(fn.Pkg) == t.Key() {
					return fmt.Sprintf("%s carries %s (%s) and is passed to %s.%s at %s:%d",
						t.Key(), iface.Key(), carried, fn.Pkg, fn.Name,
						filepath.Base(fn.File), fn.Line), true
				}
			}
		}
	}
	return "", false
}

// (A helper that matched logging calls made directly on a dispatch-struct
// parameter was removed with the first draft of this check: no tool in the
// reference implementation logs, so the clause it supported was never true.
// That absence is itself a finding -- see the Chapter 22 report.)

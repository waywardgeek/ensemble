package grade

import (
	"fmt"
	"path/filepath"
)

// backPointerChain reports whether the student's tree exhibits a back-pointer
// CHAIN, and returns the evidence it found.
//
// This check is chapter 22's thesis, so it is deliberately framework-blind:
// no clause below names a type, a package, or a method. A student who calls
// the interface common.Parent, or Context, or Spine, passes if the structural
// property holds; a student whose tree merely CONTAINS the word "Agent" does
// not. That distinction is the whole reason this function exists. Chapter 5
// originally shipped a check that grepped for the literal strings "Logf" and
// "Host" and then reported a structure it had never inspected, and it sat
// green for sixteen chapters while the invariant it claimed to protect rotted
// underneath it.
//
// Three clauses:
//
//  1. FIRST LINK. A dispatch struct -- a struct that carries an interface
//     declared in its own package and is passed to another package -- carries
//     at least one interface-typed field.
//
//  2. SECOND LINK. At least one interface carried by that struct has a method
//     whose RESULT is another interface declared in the same package. This is
//     what makes it a chain rather than a lone back-pointer: the parent can
//     hand you its own parent, or a further capability, without the dispatch
//     struct having to grow a field for it.
//
//  3. NO CLOSURES. The dispatch struct carries no func-typed fields. This is
//     the distinction the chapter turns on. A closure answers exactly one
//     question forever; a back-pointer answers the next one too. A tree that
//     reaches its parent through a stapled-on func field has reproduced the
//     very thing the chapter is about.
func (s *astScan) backPointerChain() (string, bool) {
	disp, carried := s.bestDispatchStruct()
	if disp == nil {
		return "no dispatch struct found: no struct carries an interface from its own package and is passed to another package", false
	}

	// Clause 3. Checked before clause 2 so the diagnostic names the closure
	// rather than complaining about a missing chain the closure is hiding.
	for _, f := range disp.Fields {
		if f.IsFunc {
			return fmt.Sprintf("%s carries closure field %q; a dispatch struct must reach capabilities through an interface, not a stapled-on function",
				disp.Key(), f.Name), false
		}
	}

	// Clause 2.
	for _, iface := range carried {
		for _, m := range iface.Methods {
			for _, r := range m.ResultTypes {
				rt, ok := s.Types[r.Key(iface.Pkg)]
				if !ok || !rt.IsIface || rt.Pkg != iface.Pkg {
					continue
				}
				return fmt.Sprintf("%s carries %s, and %s.%s() returns %s: a chain of depth two, reached without a closure (%s)",
					disp.Key(), iface.Key(), iface.Key(), m.Name, rt.Key(),
					filepath.Base(disp.File)), true
			}
		}
	}

	names := make([]string, 0, len(carried))
	for _, c := range carried {
		names = append(names, c.Key())
	}
	return fmt.Sprintf("%s carries %v, but none of those interfaces has a method returning another interface from the same package: a back-pointer exists, but nothing can be reached THROUGH it",
		disp.Key(), names), false
}

// bestDispatchStruct returns the struct that most looks like the tree's
// dispatch type, together with the same-package interfaces it carries.
//
// "Most looks like" means: carries the greatest number of same-package
// interfaces, among structs that are actually passed to another package. The
// count is the tiebreaker rather than a name, because a dispatch struct is
// precisely the thing that gathers several capabilities and hands them across
// a package boundary.
func (s *astScan) bestDispatchStruct() (*astType, []*astType) {
	var best *astType
	var bestCarried []*astType

	for _, t := range s.Types {
		if t.IsIface {
			continue
		}
		var carried []*astType
		for _, f := range t.Fields {
			if f.IsFunc {
				continue
			}
			ft, ok := s.Types[f.Key(t.Pkg)]
			if ok && ft.IsIface && ft.Pkg == t.Pkg {
				carried = append(carried, ft)
			}
		}
		if len(carried) == 0 || !s.passedOutward(t) {
			continue
		}
		if best == nil || len(carried) > len(bestCarried) ||
			(len(carried) == len(bestCarried) && t.Key() < best.Key()) {
			best, bestCarried = t, carried
		}
	}
	return best, bestCarried
}

// passedOutward reports whether a type is passed as a parameter to a function
// in a different package. That is what distinguishes a dispatch struct from an
// ordinary local aggregate: it exists in order to cross a package boundary.
func (s *astScan) passedOutward(t *astType) bool {
	for _, fn := range s.Funcs {
		if fn.Pkg == t.Pkg {
			continue
		}
		for _, p := range fn.Params {
			if p.Key(fn.Pkg) == t.Key() {
				return true
			}
		}
	}
	return false
}

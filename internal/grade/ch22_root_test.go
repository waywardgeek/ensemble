package grade

import (
	"sort"
	"strings"
	"testing"
)

// Fixtures for single-composition-root.
//
// Written before the reference tree was ever run through the check, and in
// a vocabulary chosen to share no word with it: a kitchen, assembled from
// a pantry, a stove, a sink and a larder. Nothing here is called New,
// Agent, Engine, Host or Build. Chapter 5's fake check passed because it
// recognised the reference implementation's words; the only way to know a
// replacement is structural is to make the words unrecognisable first.

// kitchenTree builds a tree with one library composition root and one
// binary. The options bend it into each shape the check must separate.
type kitchenOpts struct {
	stapleInMain  bool // main builds a part the root also builds, then sets its fields
	configPrep    bool // main fills in a request value from the root's own package
	outsideReach  bool // main configures something the root never touches
	tinyBinary    bool // a helper binary that assembles nothing
	callRootTwice bool // one binary, two assembly sites
	secondBinary  bool // a second main package, legitimately
	delegate      bool // the root delegates its wiring to a helper
	noRoot        bool // main wires the parts itself; no library root
}

func kitchenTree(o kitchenOpts) map[string]string {
	files := map[string]string{
		"pantry/pantry.go": "package pantry\n\n" +
			"type Shelf struct{ Jars int }\n\n" +
			"func Open() *Shelf { return &Shelf{} }\n",
		"stove/stove.go": "package stove\n\n" +
			"type Burner struct{ Lit bool }\n\n" +
			"func Light() *Burner { return &Burner{Lit: true} }\n",
		"sink/sink.go": "package sink\n\n" +
			"type Basin struct{ Full bool }\n\n" +
			"func Fill() *Basin { return &Basin{Full: true} }\n",
		"larder/larder.go": "package larder\n\n" +
			"type Crate struct{ Items int }\n\n" +
			"func Stock() *Crate { return &Crate{} }\n",
		// A package the composition root never reaches. Standing in for
		// the reference tree's GUI server, which the binary owns and the
		// agent's constructor knows nothing about.
		"tray/tray.go": "package tray\n\n" +
			"type Tray struct{ Label string }\n\n" +
			"func Carry() *Tray { return &Tray{} }\n",
	}

	// The library's one assembly site. It reaches four other packages in
	// the tree, either directly or -- when delegate is set -- through an
	// unexported helper in its own package.
	body := "\tr := &Room{}\n" +
		"\tr.Shelf = pantry.Open()\n" +
		"\tr.Burner = stove.Light()\n" +
		"\tr.Basin = sink.Fill()\n" +
		"\tr.Crate = larder.Stock()\n" +
		"\treturn r\n"
	kitchen := "package kitchen\n\n" +
		"import (\n\t\"x/larder\"\n\t\"x/pantry\"\n\t\"x/sink\"\n\t\"x/stove\"\n)\n\n" +
		"type Room struct {\n" +
		"\tShelf  *pantry.Shelf\n\tBurner *stove.Burner\n" +
		"\tBasin  *sink.Basin\n\tCrate  *larder.Crate\n\tHeat   int\n}\n\n" +
		"type Order struct{ Heat int }\n\n" +
		"func Request() *Order { return &Order{} }\n\n"
	if o.delegate {
		kitchen += "func Assemble() *Room { return furnish() }\n\n" +
			"func furnish() *Room {\n" + body + "}\n"
	} else {
		kitchen += "func Assemble() *Room {\n" + body + "}\n"
	}
	files["kitchen/kitchen.go"] = kitchen

	// The binary. In the healthy shape it calls the root once, and also
	// calls a leaf constructor directly -- reading a value out of one part
	// is not assembling the program, and the check must not say it is.
	main := "package main\n\n" +
		"import (\n\t\"x/kitchen\"\n\t\"x/sink\"\n\t\"x/stove\"\n\t\"x/tray\"\n)\n\n" +
		"var _ = stove.Light\nvar _ = tray.Carry\n\n" +
		"func main() {\n"
	if o.noRoot {
		files["cmd/main.go"] = "package main\n\n" +
			"import (\n\t\"x/larder\"\n\t\"x/pantry\"\n\t\"x/sink\"\n\t\"x/stove\"\n)\n\n" +
			"func main() {\n" +
			"\t_ = pantry.Open()\n\t_ = stove.Light()\n" +
			"\t_ = sink.Fill()\n\t_ = larder.Stock()\n}\n"
	} else {
		switch {
		case o.callRootTwice:
			main += "\tr := kitchen.Assemble()\n\tr2 := kitchen.Assemble()\n\t_, _ = r, r2\n"
		case o.stapleInMain:
			// stove is inside the root's reach: the binary is building a part
			// the root already builds, and configuring it.
			main += "\tb := stove.Light()\n\tb.Lit = false\n\tr := kitchen.Assemble()\n\t_, _ = b, r\n"
		case o.configPrep:
			// kitchen is the root's own package: an Order is a request to the
			// root, not one of its parts.
			main += "\tq := kitchen.Request()\n\tq.Heat = 3\n\tr := kitchen.Assemble()\n\t_, _ = q, r\n"
		case o.outsideReach:
			// tray is outside the root's reach entirely.
			main += "\tt := tray.Carry()\n\tt.Label = \"out\"\n\tr := kitchen.Assemble()\n\t_, _ = t, r\n"
		default:
			main += "\tr := kitchen.Assemble()\n\t_ = sink.Fill()\n\t_ = r\n"
		}
		main += "}\n"
		files["cmd/main.go"] = main
	}

	// A helper binary that assembles nothing: it touches a single package
	// and is not required to call a root.
	if o.tinyBinary {
		files["probe/main.go"] = "package main\n\n" +
			"import \"x/tray\"\n\n" +
			"func main() { _ = tray.Carry() }\n"
	}

	if o.secondBinary {
		files["tool/main.go"] = "package main\n\n" +
			"import \"x/kitchen\"\n\n" +
			"func main() { _ = kitchen.Assemble() }\n"
	}
	return files
}

func rootResult(t *testing.T, files map[string]string) (bool, string) {
	t.Helper()
	s, err := scanTree(writeTree(t, files))
	if err != nil {
		t.Fatalf("scanTree: %v", err)
	}
	return detectSingleCompositionRoot(s)
}

func TestSingleCompositionRootAcceptsHealthyTree(t *testing.T) {
	cases := []struct {
		label string
		opts  kitchenOpts
	}{
		{"one binary, one root", kitchenOpts{}},
		{"root delegates to a helper", kitchenOpts{delegate: true}},
		{"a repository with two binaries", kitchenOpts{secondBinary: true}},
		{"binary fills in a request to the root", kitchenOpts{configPrep: true}},
		{"binary configures what the root never touches", kitchenOpts{outsideReach: true}},
		{"a helper binary that assembles nothing", kitchenOpts{tinyBinary: true}},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			ok, detail := rootResult(t, kitchenTree(tc.opts))
			if !ok {
				t.Fatalf("healthy tree rejected: %s", detail)
			}
		})
	}
}

// Each of these is a distinct decay. The check must reject every one, and
// the message must name the file so a student can act on it.
func TestSingleCompositionRootRejectsDecay(t *testing.T) {
	cases := []struct {
		label string
		opts  kitchenOpts
		want  string
	}{
		{"capability stapled on by the caller", kitchenOpts{stapleInMain: true}, "stapled"},
		{"one binary assembles twice", kitchenOpts{callRootTwice: true}, "2 places"},
		{"binary wires the parts itself", kitchenOpts{noRoot: true}, "calls no composition root"},
		// The regression that fixtures alone did not catch. An earlier
		// draft of this check counted root calls across the whole tree, so
		// a healthy second binary satisfied the count on behalf of a
		// decayed first one. Running it against the real pre-repair tree
		// was what exposed it: the broken binary was not the one being
		// counted.
		{"a healthy second binary masks a decayed first",
			kitchenOpts{noRoot: true, secondBinary: true}, "calls no composition root"},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			ok, detail := rootResult(t, kitchenTree(tc.opts))
			if ok {
				t.Fatalf("decayed tree accepted: %s", detail)
			}
			if !strings.Contains(detail, tc.want) {
				t.Errorf("message does not explain the failure:\n%s", detail)
			}
			if !tc.opts.noRoot && !strings.Contains(detail, "main.go") {
				t.Errorf("message names no file:\n%s", detail)
			}
		})
	}
}

// The threshold in compositionRootFanout is the check's one tuned number.
// This test measures the two populations in the reference tree and fails if
// they ever stop being clearly separated -- which is the only thing that
// would make the constant arbitrary.
func TestCompositionRootFanout(t *testing.T) {
	s, err := scanTree("../../agent")
	if err != nil {
		t.Fatalf("scanTree: %v", err)
	}

	reach := map[string]int{}
	for _, fn := range s.Funcs {
		if fn.Pkg != "main" {
			continue
		}
		for _, c := range s.callsIn(fn) {
			if c.Pkg == "" || !s.isTreePkg(fn.File, c.Pkg) {
				continue
			}
			if target := s.findFunc(c.Pkg, c.Name); target != nil {
				reach[c.Pkg+"."+c.Name] = len(s.treeFanout(target))
			}
		}
	}
	if len(reach) == 0 {
		t.Fatal("no in-tree calls found from any main package")
	}

	var roots, leaves []string
	for name, n := range reach {
		if n >= compositionRootFanout {
			roots = append(roots, name)
		} else {
			leaves = append(leaves, name)
		}
	}
	sort.Strings(roots)
	sort.Strings(leaves)
	t.Logf("composition roots (reach >= %d): %v", compositionRootFanout, roots)
	t.Logf("leaf constructors called from main: %d", len(leaves))

	if len(roots) == 0 {
		t.Fatalf("no composition root found; binaries call only %v", leaves)
	}

	// The gap is what justifies the constant. The reference tree has more
	// than one root -- the full constructor and the bare one documented in
	// chapter 13 -- but they belong to different binaries, which is why
	// the check counts per main package rather than per tree. What must
	// hold for the threshold to be meaningful is separation: the LEAST
	// reaching root must still out-reach the MOST reaching leaf, so the
	// two populations do not touch.
	worstRoot, bestLeaf := 1<<30, 0
	for _, name := range roots {
		if reach[name] < worstRoot {
			worstRoot = reach[name]
		}
	}
	for _, name := range leaves {
		if reach[name] > bestLeaf {
			bestLeaf = reach[name]
		}
	}
	if worstRoot <= bestLeaf {
		t.Fatalf("no separation: least-reaching root reaches %d, "+
			"most-reaching non-root reaches %d", worstRoot, bestLeaf)
	}
	t.Logf("separation: least root reaches %d, greatest non-root reaches %d",
		worstRoot, bestLeaf)
}

// The reference tree, last.
func TestSingleCompositionRootReferenceTree(t *testing.T) {
	s, err := scanTree("../../agent")
	if err != nil {
		t.Fatalf("scanTree: %v", err)
	}
	ok, detail := detectSingleCompositionRoot(s)
	if !ok {
		t.Fatalf("reference tree fails its own check:\n%s", detail)
	}
	t.Logf("%s", detail)
}

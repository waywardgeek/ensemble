package skills

import (
	"crypto/sha256"
	"example.com/ensemble/internal/persistence"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"example.com/ensemble/internal/common"
)

type testApplication struct {
	common.Ensemble
	logs strings.Builder
}

func (a *testApplication) Logf(format string, values ...any) { fmt.Fprintf(&a.logs, format, values...) }

type testAgent struct {
	parent    *testApplication
	config    *common.SkillConfig
	ceiling   []string
	workspace string
}

func (a *testAgent) Ensemble() common.Ensemble               { return a.parent }
func (a *testAgent) Config() common.Config                   { return common.Config{} }
func (a *testAgent) Workspace() string                       { return a.workspace }
func (a *testAgent) ModelReady(common.ModelOperation)        {}
func (a *testAgent) SkillConfiguration() *common.SkillConfig { return a.config }
func (a *testAgent) SkillCeiling() []string                  { return append([]string{}, a.ceiling...) }

func definition(name, kind, extra, body string) []byte {
	return []byte("---\nname: " + name + "\ndescription: Description " + name + "\ntype: " + kind + "\n" + extra + "---\n" + body)
}
func newOwner(t *testing.T, config *common.SkillConfig) *testAgent {
	t.Helper()
	a := &testAgent{parent: &testApplication{}, workspace: t.TempDir(), ceiling: []string{"load_skill", "unload_skill", "read_file", "write_file", "list_directory", "search_files"}}
	var err error
	a.config, err = CopyConfiguration(a, config)
	if err != nil {
		t.Fatal(err)
	}
	return a
}
func service(t *testing.T, config *common.SkillConfig) *Service {
	t.Helper()
	s, err := New(newOwner(t, config))
	if err != nil {
		t.Fatal(err)
	}
	return s
}
func fixture() *common.SkillConfig {
	return &common.SkillConfig{Primary: "base", Variables: map[string]string{"PROJECT": "demo-$TOOLS"}, Catalog: map[string][]byte{
		"base":   definition("base", "primary", "tools: read_file\nloadable-skills: edit review bad\n", "Project=${PROJECT}; cash=$$5\n$TOOLS\n$SKILLS\n"),
		"read":   definition("read", "dependency", "tools: list_directory\n", "Read manual"),
		"edit":   definition("edit", "loadable", "tools: write_file\ndepends: read\nloadable-skills: search\n", "$TOOLS\n$SKILLS"),
		"review": definition("review", "loadable", "depends: read\n", "Review manual"),
		"search": definition("search", "loadable", "tools: search_files\n", "Search manual"),
		"bad":    definition("bad", "loadable", "depends: missing\n", "Bad manual"),
		"hidden": definition("hidden", "loadable", "", "Hidden manual"),
	}}
}
func prepare(t *testing.T, s *Service, action, name string) common.SkillCandidate {
	t.Helper()
	c, err := s.Prepare(common.SkillOperation{Action: action, Name: name})
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func commit(t *testing.T, s *Service, action, name string) common.SkillCandidate {
	t.Helper()
	c := prepare(t, s, action, name)
	s.Apply(c, uint64(len(s.Inspect().Material)+1))
	return c
}
func code(t *testing.T, err error, want string) {
	t.Helper()
	e, ok := err.(*common.SkillError)
	if !ok || e.Code != want {
		t.Fatalf("wanted %s, got %v", want, err)
	}
}

func TestParserPreservesBodyAndEquivalentLists(t *testing.T) {
	s := &Service{parent: newOwner(t, fixture())}
	base := "---\nname: edit\ndescription: 'Editor''s manual: safe'\ntype: loadable\n"
	a, err := s.parse("edit", []byte(base+"tools: write_file read_file\ndepends: []\n---\n\nBody\n---\n"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.parse("edit", []byte(base+"tools:\n  - 'read_file'\n  - \"write_file\"\ndepends: []\n---\n\nBody\n---\n"))
	if err != nil || !reflect.DeepEqual(a, b) {
		t.Fatalf("equivalent lists differ: %#v %#v %v", a, b, err)
	}
	if a.Body != "\nBody\n---\n" || a.Description != "Editor's manual: safe" {
		t.Fatalf("body/header changed: %#v", a)
	}
	crlf := strings.ReplaceAll(base+"---\n\nBody\n---\n", "\n", "\r\n")
	c, err := s.parse("edit", []byte(crlf))
	if err != nil || c.Body != "\r\nBody\r\n---\r\n" {
		t.Fatalf("CRLF body changed: %q %v", c.Body, err)
	}
	for _, desc := range []string{"\"123\"", "'false'", "A: colon is text", "Édition"} {
		source := strings.Replace(base, "'Editor''s manual: safe'", desc, 1) + "---"
		if _, err = s.parse("edit", []byte(source)); err != nil {
			t.Fatalf("valid description %q: %v", desc, err)
		}
	}
}

func TestParserRejectsMalformedSource(t *testing.T) {
	s := &Service{parent: newOwner(t, fixture())}
	valid := string(definition("edit", "loadable", "tools: read_file\n", "Body"))
	cases := map[string]string{
		"unknown":             strings.Replace(valid, "tools:", "mcp_servers:", 1),
		"duplicate":           strings.Replace(valid, "type:", "name: edit\ntype:", 1),
		"path":                strings.Replace(valid, "name: edit", "name: ../edit", 1),
		"mismatch":            strings.Replace(valid, "name: edit", "name: other", 1),
		"numeric description": strings.Replace(valid, "Description edit", "123", 1),
		"reserved":            strings.Replace(valid, "Description edit", "TrUe", 1),
		"comment":             strings.Replace(valid, "Description edit", "Text # comment", 1),
		"quoted extra":        strings.Replace(valid, "Description edit", "\"text\" other", 1),
		"decoded newline":     strings.Replace(valid, "Description edit", "\"text\\nother\"", 1),
		"flow":                strings.Replace(valid, "read_file", "[read_file]", 1),
		"comma":               strings.Replace(valid, "read_file", "read_file,write_file", 1),
		"duplicate item":      strings.Replace(valid, "read_file", "read_file read_file", 1),
		"empty item":          strings.Replace(valid, "read_file", "read_file  write_file", 1),
		"mixed list":          strings.Replace(valid, "tools: read_file", "tools: read_file\n  - write_file", 1),
		"indent":              strings.Replace(valid, "tools: read_file", "tools:\n - read_file", 1),
		"tab":                 strings.Replace(valid, "tools: read_file", "tools:\n\t- read_file", 1),
		"empty block":         strings.Replace(valid, "tools: read_file", "tools:", 1),
		"empty quoted item":   strings.Replace(valid, "tools: read_file", "tools:\n  - ''", 1),
		"anchor":              strings.Replace(valid, "Description edit", "&x text", 1),
		"alias":               strings.Replace(valid, "Description edit", "*x", 1),
		"tag":                 strings.Replace(valid, "Description edit", "!!str text", 1),
		"block scalar":        strings.Replace(valid, "Description edit", "|\n  text", 1),
		"bare CR":             valid + "\r",
		"BOM":                 "\ufeff" + valid,
		"NUL":                 valid + "\x00",
		"invalid UTF8":        valid + "\xff",
		"unterminated":        "---\nname: edit\n",
		"missing header":      "Body",
		"empty primary":       string(definition("edit", "primary", "", "")),
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := s.parse("edit", []byte(source)); err == nil {
				t.Fatal("malformed source accepted")
			}
		})
	}
	if strings.Contains(s.parent.(*testAgent).parent.logs.String(), "!!str") {
		t.Fatal("source leaked into diagnostics")
	}
}

func TestSourceAndDescriptionLimits(t *testing.T) {
	s := &Service{parent: newOwner(t, fixture())}
	prefix := definition("base", "primary", "", "")
	exact := append(prefix, []byte(strings.Repeat("x", sourceLimit-len(prefix)))...)
	if _, err := s.parse("base", exact); err != nil {
		t.Fatal(err)
	}
	_, err := s.parse("base", append(exact, 'x'))
	code(t, err, "skill_too_large")
	for _, n := range []int{256, 257} {
		data := strings.Replace(string(prefix), "Description base", strings.Repeat("a", n), 1) + "Body"
		_, err := s.parse("base", []byte(data))
		if (err == nil) != (n == 256) {
			t.Fatalf("description bytes %d: %v", n, err)
		}
	}
	for _, n := range []int{128, 129} {
		catalog := map[string][]byte{}
		for i := 0; i < n; i++ {
			name := fmt.Sprintf("s%d", i)
			prefix := definition(name, "loadable", "", "")
			catalog[name] = append(prefix, []byte(strings.Repeat("x", sourceLimit-len(prefix)))...)
		}
		a := &testAgent{parent: &testApplication{}, workspace: t.TempDir()}
		_, err := CopyConfiguration(a, &common.SkillConfig{Primary: "s0", Catalog: catalog})
		if (err == nil) != (n == 128) {
			t.Fatalf("catalog total boundary %d: %v", n, err)
		}
	}
}

func TestConfigurationOwnershipAndScalarLimits(t *testing.T) {
	input := fixture()
	a := newOwner(t, input)
	input.Catalog["base"][0] = 'x'
	input.Variables["PROJECT"] = "changed"
	delete(input.Catalog, "edit")
	if a.config.Catalog["base"][0] != '-' || a.config.Variables["PROJECT"] != "demo-$TOOLS" || a.config.Catalog["edit"] == nil {
		t.Fatal("configuration borrowed input")
	}
	for _, mutate := range []func(*common.SkillConfig){
		func(c *common.SkillConfig) { c.Directory = "other" },
		func(c *common.SkillConfig) { c.Primary = "../bad" },
		func(c *common.SkillConfig) { c.Variables["TOOLS"] = "bad" },
		func(c *common.SkillConfig) { c.Variables["lower"] = "bad" },
		func(c *common.SkillConfig) { c.Variables["PROJECT"] = "\xff" },
		func(c *common.SkillConfig) { c.Variables["PROJECT"] = "\x00" },
		func(c *common.SkillConfig) { c.Variables["PROJECT"] = strings.Repeat("x", 4097) },
		func(c *common.SkillConfig) {
			for i := 0; i < 33; i++ {
				c.Variables[fmt.Sprintf("V%d", i)] = ""
			}
		},
		func(c *common.SkillConfig) {
			for i := 0; i < 17; i++ {
				c.Variables[fmt.Sprintf("V%d", i)] = strings.Repeat("x", 4096)
			}
		},
	} {
		c := fixture()
		mutate(c)
		if _, err := CopyConfiguration(a, c); err == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
	valid := fixture()
	valid.Variables = map[string]string{}
	for i := 0; i < 16; i++ {
		valid.Variables[fmt.Sprintf("V%d", i)] = strings.Repeat("x", 4096)
	}
	if _, err := CopyConfiguration(a, valid); err != nil {
		t.Fatal("exact scalar aggregate refused", err)
	}
	valid = fixture()
	valid.Variables = map[string]string{}
	for i := 0; i < 32; i++ {
		valid.Variables[fmt.Sprintf("V%d", i)] = ""
	}
	if _, err := CopyConfiguration(a, valid); err != nil {
		t.Fatal("32 empty scalars refused", err)
	}
}

func TestFrozenDirectoryCatalogAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	for name, data := range fixture().Catalog {
		if err := os.Mkdir(filepath.Join(dir, name), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name, "SKILL.md"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored.txt"), []byte("not a skill"), 0600); err != nil {
		t.Fatal(err)
	}
	s := service(t, &common.SkillConfig{Directory: dir, Primary: "base", Variables: fixture().Variables})
	commit(t, s, "initialize", "base")
	path := filepath.Join(dir, "edit", "SKILL.md")
	changed := definition("edit", "loadable", "", "NEW BODY")
	if err := os.WriteFile(path, changed, 0600); err != nil {
		t.Fatal(err)
	}
	c := prepare(t, s, "load", "edit")
	if !strings.Contains(c.Transition().Activated[1].Body, "write_file") {
		t.Fatal("dynamic load reread changed file")
	}
	fresh := service(t, &common.SkillConfig{Directory: dir, Primary: "base", Variables: fixture().Variables})
	commit(t, fresh, "initialize", "base")
	if got := prepare(t, fresh, "load", "edit").Transition().Activated[0].Body; got != "NEW BODY" {
		t.Fatal(got)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Prepare(common.SkillOperation{Action: "load", Name: "edit"}); err != nil {
		t.Fatal("deleted file changed frozen catalog", err)
	}
	if _, err := New(s.parent); err == nil {
		t.Fatal("missing SKILL.md accepted by fresh construction")
	}
	if err := os.Symlink(filepath.Join(dir, "base", "SKILL.md"), path); err != nil {
		t.Fatal(err)
	}
	if _, err := New(s.parent); err == nil {
		t.Fatal("symlink skill file accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, changed, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(dir, "base"), filepath.Join(dir, "linked")); err != nil {
		t.Fatal(err)
	}
	if _, err := New(s.parent); err == nil {
		t.Fatal("symlink child accepted")
	}
}

func TestGraphCommitDiscoveryAndRetirement(t *testing.T) {
	s := service(t, fixture())
	initial := prepare(t, s, "initialize", "base")
	if s.Inspect().State != nil {
		t.Fatal("prepare published initial state")
	}
	s.Apply(initial, 1)
	primary := s.Inspect().Material[0].Record
	if !strings.HasPrefix(primary.Body, "Project=demo-$TOOLS; cash=$5\n- load_skill\n- read_file\n- unload_skill\n") {
		t.Fatal(primary.Body)
	}
	if primary.SHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(primary.Body))) {
		t.Fatal("wrong material digest")
	}
	loaded := commit(t, s, "load", "edit").Transition()
	if got := []string{loaded.Activated[0].Name, loaded.Activated[1].Name}; !reflect.DeepEqual(got, []string{"read", "edit"}) {
		t.Fatal(got)
	}
	if !reflect.DeepEqual(loaded.Activated[1].Dependencies, []uint64{2}) {
		t.Fatal("dependency does not name its activation")
	}
	commit(t, s, "load", "search")
	commit(t, s, "load", "review")
	before := s.Inspect()
	noop := commit(t, s, "load", "edit")
	if noop.Result().Changed || !reflect.DeepEqual(before, s.Inspect()) {
		t.Fatal("no-op changed state")
	}
	commit(t, s, "unload", "edit")
	after := s.Inspect()
	if strings.Contains(strings.Join(after.State.Tools, ","), "write_file") {
		t.Fatal("unload retained sole grant")
	}
	if !reflect.DeepEqual(after.State.Roots, []string{"review", "search"}) {
		t.Fatal("advertiser removal lost explicit roots")
	}
	if !reflect.DeepEqual(after.State.Retired, []common.SkillRetired{{Name: "edit", Activation: 3}}) {
		t.Fatal(after.State.Retired)
	}
	if !reflect.DeepEqual(primary, after.Material[0].Record) {
		t.Fatal("primary rerendered")
	}
	for _, c := range after.Contributors {
		if c.Tool == "list_directory" && !reflect.DeepEqual(c.Activations, []uint64{2}) {
			t.Fatal(c)
		}
	}
	commit(t, s, "unload", "review")
	if len(s.Inspect().State.Retired) != 3 {
		t.Fatal("unused dependency not retired")
	}
	reloaded := commit(t, s, "load", "edit").Transition()
	if reloaded.Activated[0].Activation != 6 || reloaded.Activated[1].Activation != 7 {
		t.Fatal("reload reused identities", reloaded.Activated)
	}
	if len(s.Inspect().State.Retired) != 3 {
		t.Fatal("reload erased retirement")
	}
}

func TestFailedCandidatesAreAtomicAndUnreachableGraphsAreLazy(t *testing.T) {
	for _, bad := range []struct{ name, extra, body, want string }{
		{"missing", "depends: missing\n", "Body", "skill_dependency"},
		{"cycle", "depends: loop\n", "Body", "skill_dependency"},
		{"tool", "tools: missing_tool\n", "Body", "skill_tool_unavailable"},
		{"offer", "loadable-skills: read\n", "Body", "skill_dependency"},
		{"variable", "depends: read\n", "$UNKNOWN", "skill_variable"},
	} {
		t.Run(bad.name, func(t *testing.T) {
			config := fixture()
			config.Catalog["bad"] = definition("bad", "loadable", bad.extra, bad.body)
			config.Catalog["loop"] = definition("loop", "dependency", "depends: loop\n", "")
			s := service(t, config)
			commit(t, s, "initialize", "base")
			old := s.Inspect()
			_, err := s.Prepare(common.SkillOperation{Action: "load", Name: "bad"})
			code(t, err, bad.want)
			if !reflect.DeepEqual(old, s.Inspect()) {
				t.Fatal("failed candidate changed authority")
			}
			good := prepare(t, s, "load", "edit")
			if good.Transition().Activated[0].Activation != 2 {
				t.Fatal("failure consumed activation IDs")
			}
		})
	}
	s := service(t, fixture())
	commit(t, s, "initialize", "base")
	for _, name := range []string{"hidden", "read", "base", "unknown"} {
		_, err := s.Prepare(common.SkillOperation{Action: "load", Name: name})
		code(t, err, "skill_unavailable")
	}
	if c := prepare(t, s, "unload", "hidden"); c.Result().Changed {
		t.Fatal("known inactive unload is not no-op")
	}
	_, err := s.Prepare(common.SkillOperation{Action: "load", Name: "../bad"})
	code(t, err, "invalid_skill_arguments")
}

func TestSinglePassVariablesAndExpandedByteLimit(t *testing.T) {
	s := &Service{parent: newOwner(t, fixture())}
	for _, body := range []string{"$", "$5", "${NAME", "${}", "${NAME-X}", "$UNKNOWN", "$_BAD", "$TOOLSsuffix"} {
		_, err := s.expand(body, nil, nil, nil, "edit")
		code(t, err, "skill_variable")
	}
	for _, n := range []int{sourceLimit, sourceLimit + 1} {
		body, err := s.expand("${PROJECT}", nil, nil, map[string]string{"PROJECT": strings.Repeat("x", n)}, "edit")
		if n == sourceLimit {
			if err != nil || len(body) != n {
				t.Fatal("exact rendered boundary failed", err)
			}
		} else {
			code(t, err, "skill_too_large")
		}
	}
	body, err := s.expand("$TOOLS\n${SKILLS}\n${PROJECT} $$5\n", nil, nil, map[string]string{"PROJECT": "$NO_SUCH_VARIABLE"}, "edit")
	if err != nil || body != "(none)\n(none)\n$NO_SUCH_VARIABLE $5\n" {
		t.Fatalf("literal substitution changed: %q %v", body, err)
	}
}

func TestCounterExhaustionAndOwnedCopies(t *testing.T) {
	s := service(t, fixture())
	commit(t, s, "initialize", "base")
	commit(t, s, "load", "edit")
	old := s.Inspect()
	changed := s.Inspect()
	changed.State.Roots[0] = "bad"
	changed.Material[0].Record.Tools[0] = "write_file"
	changed.Contributors[0].Activations = append(changed.Contributors[0].Activations, 99)
	if !reflect.DeepEqual(old, s.Inspect()) {
		t.Fatal("inspection aliases committed state")
	}
	prepared := prepare(t, s, "load", "review")
	transition := prepared.Transition()
	transition.State.Roots[0] = "bad"
	transition.Activated[0].Body = "changed"
	if prepared.Transition().Activated[0].Body == "changed" {
		t.Fatal("candidate aliases transition copy")
	}
	// Explicit owner-state test seam, not a fabricated log claiming consecutive
	// allocation from 1 to uint64 max. The valid control above uses normal IDs.
	s.committed.state.Revision = math.MaxUint64
	if c := prepare(t, s, "load", "edit"); c.Result().Changed {
		t.Fatal("maximum revision no-op changed")
	}
	old = s.Inspect()
	_, err := s.Prepare(common.SkillOperation{Action: "load", Name: "review"})
	code(t, err, "skill_too_large")
	if !reflect.DeepEqual(old, s.Inspect()) {
		t.Fatal("revision exhaustion mutated state")
	}
	s = service(t, fixture())
	commit(t, s, "initialize", "base")
	s.committed.lastID = math.MaxUint64 - 1
	old = s.Inspect()
	_, err = s.Prepare(common.SkillOperation{Action: "load", Name: "edit"})
	code(t, err, "skill_too_large")
	if !reflect.DeepEqual(old, s.Inspect()) || s.committed.lastID != math.MaxUint64-1 {
		t.Fatal("whole-group exhaustion partially allocated")
	}
	s.committed.lastID = math.MaxUint64 - 2
	one := prepare(t, s, "load", "review")
	if got := one.Transition().Activated; got[0].Activation != math.MaxUint64-1 || got[1].Activation != math.MaxUint64 {
		t.Fatal("whole group fitting exactly was refused", got)
	}
}

func TestDiamondSecondBranchFailureLeavesWholeLedger(t *testing.T) {
	for _, broken := range []bool{false, true} {
		config := fixture()
		config.Catalog["edit"] = definition("edit", "loadable", "depends: left right\n", "Editor")
		config.Catalog["left"] = definition("left", "dependency", "depends: read\n", "Left")
		right := "depends: read\n"
		if broken {
			right = "depends: missing\n"
		}
		config.Catalog["right"] = definition("right", "dependency", right, "Right")
		s := service(t, config)
		commit(t, s, "initialize", "base")
		before := s.Inspect()
		c, err := s.Prepare(common.SkillOperation{Action: "load", Name: "edit"})
		if broken {
			code(t, err, "skill_dependency")
			if !reflect.DeepEqual(before, s.Inspect()) {
				t.Fatal("half diamond published")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		s.Apply(c, 2)
		count := 0
		for _, m := range s.Inspect().Material {
			if m.Record.Name == "read" {
				count++
			}
		}
		if count != 1 {
			t.Fatal("diamond duplicated shared manual")
		}
	}
}
func TestNewMaterialAggregateExactAndOneOver(t *testing.T) {
	for _, over := range []bool{false, true} {
		config := &common.SkillConfig{Primary: "base", Variables: map[string]string{"CHUNK": strings.Repeat("x", 4096)}, Catalog: map[string][]byte{}}
		deps := []string{}
		body := strings.Repeat("${CHUNK}", 16)
		for i := 0; i < 127; i++ {
			name := fmt.Sprintf("dep-%03d", i)
			deps = append(deps, name)
			config.Catalog[name] = definition(name, "dependency", "", body)
		}
		if over {
			deps = append(deps, "extra")
			config.Catalog["extra"] = definition("extra", "dependency", "", "x")
		}
		config.Catalog["base"] = definition("base", "primary", "depends: "+strings.Join(deps, " ")+"\n", body)
		s := service(t, config)
		c, err := s.Prepare(common.SkillOperation{Action: "initialize", Name: "base"})
		if over {
			code(t, err, "skill_too_large")
			if s.Inspect().State != nil {
				t.Fatal("oversize initialized")
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			total := 0
			for _, m := range c.Transition().Activated {
				total += len(m.Body)
			}
			if total != 8388608 {
				t.Fatal(total)
			}
		}
	}
}

// Test fixture is a composition root for the owner interface.
func (a *testAgent) Codec() common.SessionCodec { return persistence.NewCodec(a) }

package ch10_test

// Independent Chapter 10 controls through the published session/Skills API.
// Reuse the external consumer's owners and local transport, never private code.
import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"example.com/ensemble"
)

func skillsCatalog() map[string][]byte {
	return map[string][]byte{
		"base":   []byte("---\nname: base\ndescription: Base\ntype: primary\ntools: read_file\nloadable-skills: edit review\n---\nBASE-PRIMARY ${PROJECT}\n$TOOLS\n$SKILLS\n"),
		"dep":    []byte("---\nname: dep\ndescription: Dependency\ntype: dependency\ntools: list_directory\n---\nDEPENDENCY-MANUAL\n"),
		"edit":   []byte("---\nname: edit\ndescription: Edit\ntype: loadable\ntools: write_file\ndepends: dep\n---\nRETIRED-EDIT-MANUAL ${PROJECT}\n"),
		"review": []byte("---\nname: review\ndescription: Review\ntype: loadable\ntools: search_files\ndepends: dep\n---\nCURRENT-REVIEW-MANUAL\n"),
		"hidden": []byte("---\nname: hidden\ndescription: Hidden\ntype: loadable\n---\nINACTIVE-DEFINITION\n"),
		"other":  []byte("---\nname: other\ndescription: Other\ntype: primary\ntools: read_file\n---\nOTHER-PRIMARY\n"),
	}
}

func skillsOptions(o ensemble.SessionOptions) ensemble.SessionOptions {
	o.System = nil
	o.Config.System = ""
	o.Config.Builtins = []string{"read_file", "write_file", "list_directory", "search_files", "load_skill", "unload_skill"}
	o.Config.Skills = &ensemble.SkillConfig{Catalog: skillsCatalog(), Primary: "base", Variables: map[string]string{"PROJECT": "project-$TOOLS-😀"}}
	return o
}

func skillView(t *testing.T, a *ensemble.Agent) ensemble.SkillInspection {
	t.Helper()
	v, err := a.InspectSkills()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func skillChange(t *testing.T, a *ensemble.Agent, load bool, name string) {
	t.Helper()
	var r ensemble.SkillResult
	var err error
	if load {
		r, err = a.LoadSkill(name)
	} else {
		r, err = a.UnloadSkill(name)
	}
	if err != nil || !r.Changed {
		t.Fatalf("expected changed Skills operation %v %s: %+v %v", load, name, r, err)
	}
}

func skillsDirectory(t *testing.T, root string, equivalent bool) {
	t.Helper()
	for name, data := range skillsCatalog() {
		if equivalent && name == "base" {
			data = bytes.ReplaceAll(data, []byte("loadable-skills: edit review"), []byte("loadable-skills:\n  - review\n  - edit"))
			data = bytes.ReplaceAll(data, []byte("description: Base"), []byte("description: \"Base\""))
		}
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func skillsStoreBytes(t *testing.T, dir string) map[string]string {
	t.Helper()
	m := map[string]string{}
	for _, name := range []string{"events.log", "checkpoint.json", "origin.json"} {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		m[name] = string(b)
	}
	return m
}

func TestCh10PublicSkillsAuthority(t *testing.T) {
	e := localEndpoint(t)
	o := skillsOptions(wireOptions(t, "openai", e))
	first := filepath.Join(o.Config.Workspace, "catalog-first")
	second := filepath.Join(o.Config.Workspace, "catalog-moved")
	skillsDirectory(t, first, false)
	skillsDirectory(t, second, true)
	o.Config.Skills.Catalog = nil
	o.Config.Skills.Directory = first
	app := application(t)
	a := opened(t, app, o)
	skillChange(t, a, true, "edit")
	skillChange(t, a, false, "edit")
	skillChange(t, a, true, "review")
	want := skillView(t, a)
	x := export(t, a)
	closed(t, a)
	// Removing the original source makes relocation and offline independence real.
	if err := os.RemoveAll(first); err != nil {
		t.Fatal(err)
	}
	i, err := app.InspectCheckpoint(x.Bytes)
	if err != nil || !reflect.DeepEqual(want, i.InspectSkills()) {
		t.Fatalf("offline Skills inspection required original files: %v", err)
	}
	o.Config.Skills.Directory = second
	b := opened(t, app, o)
	if !reflect.DeepEqual(want, skillView(t, b)) {
		t.Fatal("equivalent moved catalog changed frozen activations or grants")
	}
	closed(t, b)
	before := skillsStoreBytes(t, o.Config.DataDir)
	cases := []struct {
		name string
		edit func(*ensemble.SessionOptions)
	}{
		{"missing-skills", func(c *ensemble.SessionOptions) { c.Config.Skills = nil }},
		{"different-primary", func(c *ensemble.SessionOptions) { c.Config.Skills.Primary = "other" }},
		{"changed-binding", func(c *ensemble.SessionOptions) { c.Config.Skills.Variables["PROJECT"] = "different" }},
		{"missing-binding", func(c *ensemble.SessionOptions) { delete(c.Config.Skills.Variables, "PROJECT") }},
		{"additional-binding", func(c *ensemble.SessionOptions) { c.Config.Skills.Variables["UNUSED"] = "new" }},
		{"changed-primary-body", func(c *ensemble.SessionOptions) {
			c.Config.Skills.Catalog["base"] = append(c.Config.Skills.Catalog["base"], 'x')
		}},
		{"changed-retired-body", func(c *ensemble.SessionOptions) {
			c.Config.Skills.Catalog["edit"] = append(c.Config.Skills.Catalog["edit"], 'x')
		}},
		{"changed-inactive-body", func(c *ensemble.SessionOptions) {
			c.Config.Skills.Catalog["hidden"] = append(c.Config.Skills.Catalog["hidden"], 'x')
		}},
		{"missing-inactive-definition", func(c *ensemble.SessionOptions) { delete(c.Config.Skills.Catalog, "hidden") }},
		{"changed-ceiling", func(c *ensemble.SessionOptions) { c.Config.Builtins = append(c.Config.Builtins, "edit_file") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := skillsOptions(o)
			tc.edit(&candidate)
			bad, err := app.OpenSession(candidate)
			if err == nil {
				_ = bad.Close()
				t.Fatal("incompatible identity mounted")
			}
			code(t, err, "session_incompatible")
			if !reflect.DeepEqual(before, skillsStoreBytes(t, o.Config.DataDir)) {
				t.Fatal("refusal rewrote original session files")
			}
			noStartupHTTP(t, e)
		})
	}
	// Refusals must release reservations and preserve a usable valid parent.
	c := opened(t, app, o)
	if !reflect.DeepEqual(want, skillView(t, c)) {
		t.Fatal("identity refusals changed retained Skills state")
	}
	noStartupHTTP(t, e)
}

func skillRawTurn(t *testing.T, a *ensemble.Agent, e *endpoint, vendor string) {
	t.Helper()
	h, err := a.Submit("raw replay positive")
	if err != nil {
		t.Fatal(err)
	}
	x := nextRequest(t, e)
	// Formatting and numeric lexemes are intentionally retained in accepted raw
	// fields. The read is real local work; no private append fabricates a response.
	args := `{ "path": "\u0070robe.txt", "start_line": 1 }`
	var response string
	switch vendor {
	case "anthropic":
		response = `{"model":"claude-sonnet-4-6","content":[{"type":"thinking","thinking":"opaque fixture","signature":"signature<&>"},{"type":"tool_use","id":"raw-read","name":"read_file","input":` + args + `}],"stop_reason":"tool_use","usage":{ "input_tokens":3, "output_tokens":2, "lexemes":[9007199254740993,1.00,-0.0,1e+3] }}`
	case "openai":
		encoded, _ := json.Marshal(args)
		response = `{"model":"gpt-4.1-mini-2025-04-14","choices":[{"index":0,"message":{"role":"assistant","content":"","tool_calls":[{"id":"raw-read","type":"function","function":{"name":"read_file","arguments":` + string(encoded) + `}}]},"finish_reason":"tool_calls"}],"usage":{ "prompt_tokens":3, "completion_tokens":2, "lexemes":[9007199254740993,1.00,-0.0,1e+3] }}`
	case "gemini":
		response = `{"modelVersion":"gemini-3.8-flash","candidates":[{"content":{"role":"model","parts":[{"text":"signed fixture","thoughtSignature":"signature<&>"},{"functionCall":{"id":"raw-read","name":"read_file","args":` + args + `},"thoughtSignature":"call-signature"}]},"finishReason":"STOP"}],"usageMetadata":{ "promptTokenCount":3, "candidatesTokenCount":2, "lexemes":[9007199254740993,1.00,-0.0,1e+3] }}`
	}
	if !json.Valid([]byte(response)) {
		t.Fatal("invalid raw-response fixture")
	}
	x.reply <- []byte(response)
	continuation := nextRequest(t, e)
	if !bytes.Contains(continuation.body, []byte("READ-EFFECT-MARKER")) {
		t.Fatal("positive local read result absent")
	}
	answer(t, vendor, continuation)
	finish(t, h)
}

func skillsProjection(t *testing.T, e *ensemble.Ensemble, a *ensemble.Agent) []any {
	t.Helper()
	x := export(t, a)
	i, err := e.InspectCheckpoint(x.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	s := i.Snapshot()
	// Exclude anchor-driven LastSeq and transient runtime state explicitly, while
	// retaining the typed parts, raw responses, calls and settled guidance.
	// Compare owned typed values directly: json.Marshal would compact RawMessage
	// and could hide the very whitespace disagreement this fixture must detect.
	return []any{s.Entries, s.Responses, s.Calls, s.Ephemera, s.Hints}
}

func skillsToolNames(t *testing.T, vendor string, body []byte) []string {
	t.Helper()
	var v map[string]json.RawMessage
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatal(err)
	}
	var tools []struct {
		Name     string `json:"name"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
		Declarations []struct {
			Name string `json:"name"`
		} `json:"functionDeclarations"`
	}
	if err := json.Unmarshal(v["tools"], &tools); err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, tool := range tools {
		switch vendor {
		case "anthropic":
			names = append(names, tool.Name)
		case "openai":
			names = append(names, tool.Function.Name)
		case "gemini":
			for _, d := range tool.Declarations {
				names = append(names, d.Name)
			}
		}
	}
	sort.Strings(names)
	return names
}

func TestCh10PublicSkillsReplay(t *testing.T) {
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		t.Run(vendor, func(t *testing.T) {
			e := localEndpoint(t)
			o := skillsOptions(wireOptions(t, vendor, e))
			if err := os.WriteFile(filepath.Join(o.Config.Workspace, "probe.txt"), []byte("READ-EFFECT-MARKER\n"), 0600); err != nil {
				t.Fatal(err)
			}
			app := application(t)
			a := opened(t, app, o)
			skillChange(t, a, true, "edit")
			skillRawTurn(t, a, e, vendor)
			old := export(t, a)
			skillChange(t, a, false, "edit")
			skillChange(t, a, true, "review")
			if err := a.Ephemeral("SETTLED-SKILLS-GUIDANCE"); err != nil {
				t.Fatal(err)
			}
			latest := export(t, a)
			want := skillView(t, a)
			wantRaw := skillsProjection(t, app, a)
			usage := a.UsageByModel()
			if latest.AsOf <= old.AsOf || len(want.State.Retired) != 2 || len(want.Material) != 5 || !reflect.DeepEqual(want.State.Roots, []string{"review"}) {
				t.Fatal("fixture lacks genuine retired material/current grants/newer tail")
			}
			lexemes, err := json.Marshal(wantRaw)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(lexemes, []byte("9007199254740993")) || !bytes.Contains(lexemes, []byte("1.00")) {
				t.Fatal("raw numeric positive missing")
			}
			closed(t, a)
			var requests [][]byte
			for _, mode := range []string{"older-tail", "null-rebuild", "snapshot-only"} {
				c := o
				c.Config.DataDir = filepath.Join(o.Config.Workspace, mode)
				root := application(t)
				var b *ensemble.Agent
				var err error
				if mode == "snapshot-only" {
					b, err = root.ImportSession(latest.Bytes, c)
				} else {
					copies(t, o.Config.DataDir, c.Config.DataDir)
					checkpoint := old.Bytes
					if mode == "null-rebuild" {
						var fields map[string]json.RawMessage
						if err = json.Unmarshal(checkpoint, &fields); err != nil {
							t.Fatal(err)
						}
						fields["state"], fields["state_sha256"] = json.RawMessage("null"), json.RawMessage("null")
						checkpoint, err = json.Marshal(fields)
						if err != nil {
							t.Fatal(err)
						}
					}
					if err = os.WriteFile(filepath.Join(c.Config.DataDir, "checkpoint.json"), checkpoint, 0600); err != nil {
						t.Fatal(err)
					}
					b, err = root.OpenSession(c)
				}
				if err != nil {
					t.Fatalf("%s positive: %v", mode, err)
				}
				noStartupHTTP(t, e)
				if !reflect.DeepEqual(want, skillView(t, b)) || !reflect.DeepEqual(usage, b.UsageByModel()) || !reflect.DeepEqual(wantRaw, skillsProjection(t, root, b)) {
					t.Fatalf("%s changed Skills/raw projection/usage", mode)
				}
				body := turn(t, b, e, vendor, "next skills request")
				for _, marker := range []string{"BASE-PRIMARY", "RETIRED-EDIT-MANUAL", "CURRENT-REVIEW-MANUAL", "SETTLED-SKILLS-GUIDANCE"} {
					if bytes.Count(body, []byte(marker)) != 1 {
						t.Fatalf("%s missing/duplicated %s", mode, marker)
					}
				}
				if got := skillsToolNames(t, vendor, body); !reflect.DeepEqual(got, []string{"list_directory", "load_skill", "read_file", "search_files", "unload_skill"}) {
					t.Fatalf("%s wrong live grants: %v", mode, got)
				}
				var seq uint64
				for _, event := range b.Events() {
					if event.Type == "request_sent" {
						seq = event.Seq
					}
				}
				replay, err := b.ReconstructRequest(seq)
				if err != nil || !bytes.Equal(body, replay) {
					t.Fatalf("%s next request mismatch: %v", mode, err)
				}
				requests = append(requests, body)
				closed(t, b)
			}
			if !bytes.Equal(requests[0], requests[1]) || !bytes.Equal(requests[0], requests[2]) {
				t.Fatal("three restoration paths render different actual requests")
			}
		})
	}
}

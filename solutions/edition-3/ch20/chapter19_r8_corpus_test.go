package ensemble_test

import (
	"ensemble"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCh19R8TrustedCorpusAndConfinedFiles(t *testing.T) {
	for _, relative := range []bool{false, true} {
		name := "absolute"
		if relative {
			name = "base-relative"
		}
		t.Run(name, func(t *testing.T) {
			top := t.TempDir()
			workspace := filepath.Join(top, "workspace")
			corpus := filepath.Join(top, "corpus")
			if err := os.MkdirAll(workspace, 0700); err != nil {
				t.Fatal(err)
			}
			paths := map[string]string{}
			add := func(folder, name, kind, metadata, body string) {
				t.Helper()
				path := filepath.Join(corpus, folder, "SKILL.md")
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("---\nname: "+name+"\ntype: "+kind+"\ndescription: "+name+" guide\n"+metadata+"---\n"+body), 0600); err != nil {
					t.Fatal(err)
				}
				paths[name] = path
			}
			add("01-entry", "primary", "primary", "tools: read_file\nloadable-skills: manual tool-guide needs-web web-search shared\n", "R8_PRIMARY_GUIDE")
			add("02-reference", "manual", "loadable", "depends: shared\n", "R8_MANUAL_GUIDE")
			add("03-tool", "tool-guide", "loadable", "", "R8_TOOL_GUIDE")
			add("04-common", "shared", "dependency", "tools: list_directory\n", "R8_SHARED_GUIDE")
			add("05-private", "not-offered", "loadable", "", "R8_UNOFFERED_GUIDE")
			add("06-gated", "needs-web", "loadable", "depends: web-search\n", "R8_GATED_GUIDE")
			add("07-web", "web-search", "loadable", "tools: unused_web_tool\nmcp_servers:\n  - name: unused-web\n    transport: url\n    url: http://127.0.0.1:1/mcp\n    auth-env: R8_CORPUS_AUTH\n", "R8_WEB_GUIDE")
			unselected := filepath.Join(top, "unselected.md")
			if err := os.WriteFile(unselected, []byte("---\nname: unselected\ntype: loadable\n---\nR8_UNSELECTED_GUIDE"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(workspace, "inside.txt"), []byte("R8_INSIDE_FILE"), 0600); err != nil {
				t.Fatal(err)
			}
			selected := corpus
			if relative {
				selected = filepath.Join("..", "corpus")
			}
			app := ensemble.New(nil)
			defer app.Shutdown()
			a, err := app.NewAgent(ensemble.Config{Model: "offline", AuthMode: ensemble.AuthAPI, BaseURL: "http://127.0.0.1:1", BaseDir: workspace, SandboxRoot: workspace, DataDir: filepath.Join(workspace, "data"), SafeMode: true, EnableWebSearch: false, Skills: &ensemble.SkillSpec{Directory: selected, Primary: "primary"}})
			if err != nil {
				t.Fatal(err)
			}
			if a.Config().SandboxRoot != workspace || a.Config().Skills.Directory != selected {
				t.Fatal("selected root or corpus was rewritten")
			}
			initial := a.Skills().Snapshot()
			if !strings.Contains(initial.Instructions, "R8_PRIMARY_GUIDE") || strings.Contains(initial.Instructions, "R8_MANUAL_GUIDE") {
				t.Fatalf("initial admission: %s", initial.Instructions)
			}
			for _, skill := range initial.Skills {
				if skill.Name == "web-search" {
					t.Fatal("disabled web declaration exposed")
				}
				if skill.Source != paths[skill.Name] {
					t.Fatalf("source relocated or identity inferred from folder: %+v", skill)
				}
			}
			if got := a.Skills().MCPAuthEnvNames(); !reflect.DeepEqual(got, []string{"R8_CORPUS_AUTH"}) {
				t.Fatalf("hidden credential inventory: %v", got)
			}
			offered := []string{}
			for _, s := range a.Skills().LoadableSkills() {
				offered = append(offered, s.Name)
			}
			if !reflect.DeepEqual(offered, []string{"manual", "needs-web", "tool-guide"}) {
				t.Fatalf("offers changed: %v", offered)
			}
			for _, name := range []string{"primary", "shared", "not-offered", "web-search", "needs-web", "unselected"} {
				if err := a.Skills().LoadSkill(name); err == nil {
					t.Errorf("unauthorized load %q succeeded", name)
				}
			}
			if got := a.Skills().Snapshot(); got.Instructions != initial.Instructions || len(got.Definitions) != len(initial.Definitions) {
				t.Fatal("refused load leaked a partial grant")
			}
			if err := a.Skills().LoadSkill("manual"); err != nil {
				t.Fatal(err)
			}
			loaded := a.Skills().Snapshot()
			if strings.Index(loaded.Instructions, "R8_SHARED_GUIDE") < 0 || strings.Index(loaded.Instructions, "R8_MANUAL_GUIDE") < strings.Index(loaded.Instructions, "R8_SHARED_GUIDE") {
				t.Fatalf("dependency order/admission: %s", loaded.Instructions)
			}
			manual, dependency := false, false
			for _, s := range loaded.Skills {
				if s.Name == "manual" {
					manual = s.Enabled && s.Origin == "manual" && s.EventSeq == 0
				}
				if s.Name == "shared" {
					dependency = s.Enabled && s.Origin == "manual" && s.EventSeq == 0
				}
			}
			if !manual || !dependency {
				t.Fatal("manual/dependency provenance changed")
			}
			for _, path := range []string{paths["manual"], filepath.Dir(paths["manual"]), "02-reference", unselected} {
				text, failed := c3Call(t, a, "load_skill", map[string]any{"name": path})
				if !failed || !strings.Contains(text, "unknown skill") {
					t.Fatalf("load_skill accepted a path or failed before name admission: %q %t %s", path, failed, text)
				}
			}
			if text, failed := c3Call(t, a, "load_skill", map[string]any{"name": "tool-guide"}); failed {
				t.Fatalf("declared tool load: %s", text)
			}
			guideAdmitted := false
			for _, s := range a.Skills().ListSkills() {
				if s.Name == "tool-guide" {
					for _, event := range c9LoadEvents(a) {
						if s.Enabled && s.Origin == "tool" && s.EventSeq != 0 && event.Seq == s.EventSeq {
							guideAdmitted = true
						}
					}
				}
			}
			if !guideAdmitted {
				t.Fatal("named load did not retain real tool-admission provenance")
			}
			if text, failed := c3Call(t, a, "read_file", map[string]any{"path": "inside.txt"}); failed || !strings.Contains(text, "R8_INSIDE_FILE") {
				t.Fatalf("inside positive control: %t %s", failed, text)
			}
			for _, path := range []string{paths["primary"], filepath.Join("..", "corpus", "01-entry", "SKILL.md")} {
				text, failed := c3Call(t, a, "read_file", map[string]any{"path": path})
				if !failed || strings.Contains(text, "R8_PRIMARY_GUIDE") {
					t.Fatalf("ordinary outside read disclosed trusted corpus: %t %s", failed, text)
				}
			}
			if strings.Contains(a.Skills().Snapshot().Instructions, "R8_UNSELECTED_GUIDE") || strings.Contains(a.Skills().Snapshot().Instructions, "R8_WEB_GUIDE") {
				t.Fatal("unselected/disabled instructions became active")
			}
		})
	}
}

func TestCh19R8ExternalCorpusValidation(t *testing.T) {
	for _, mode := range []string{"duplicate", "malformed", "primary-is-declared-name"} {
		t.Run(mode, func(t *testing.T) {
			top := t.TempDir()
			workspace := filepath.Join(top, "workspace")
			corpus := filepath.Join(top, "corpus")
			if err := os.MkdirAll(workspace, 0700); err != nil {
				t.Fatal(err)
			}
			write := func(folder, doc string) {
				t.Helper()
				dir := filepath.Join(corpus, folder)
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(doc), 0600); err != nil {
					t.Fatal(err)
				}
			}
			write("folder", "---\nname: primary\ntype: primary\n---\nPrimary")
			primary, want := "primary", ""
			switch mode {
			case "duplicate":
				write("other", "---\nname: primary\ntype: primary\n---\nDuplicate")
				want = "duplicate skill"
			case "malformed":
				write("other", "---\nname: bad\ntype: wrong\n---\nBad")
				want = "invalid skill type"
			case "primary-is-declared-name":
				primary = "folder"
				want = "unknown skill"
			}
			app := ensemble.New(nil)
			defer app.Shutdown()
			_, err := app.NewAgent(ensemble.Config{Model: "offline", AuthMode: ensemble.AuthAPI, BaseURL: "http://127.0.0.1:1", BaseDir: workspace, SandboxRoot: workspace, DataDir: filepath.Join(workspace, "data"), Skills: &ensemble.SkillSpec{Directory: corpus, Primary: primary}})
			if err == nil || !strings.Contains(err.Error(), want) || len(app.Agents()) != 0 {
				t.Fatalf("%s: %v", mode, err)
			}
		})
	}
}

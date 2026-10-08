package ensemble

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

// This is an ordinary valid catalog, not an invented exhausted-history log.
// Many repeated escaped offer descriptions exceed the whole-fact wire budget
// even though every definition, body and decoded catalog is within its limit.
func repeatedOffersConfig(t *testing.T) Config {
	c := persistenceSkillConfig(t)
	c.Builtins = append(c.Builtins, "tool_limits")
	names, dependencies := []string{}, []string{}
	for i := 0; i < 128; i++ {
		names = append(names, fmt.Sprintf("s%03d", i))
	}
	for i := 0; i < 126; i++ {
		dependencies = append(dependencies, fmt.Sprintf("d%03d", i))
	}
	list := strings.Join(names, " ")
	body := strings.Repeat("${PAD}", 16)
	c.Skills.Variables = map[string]string{"PAD": strings.Repeat("<", 4096)}
	c.Skills.Catalog = map[string][]byte{
		"base": []byte("---\nname: base\ndescription: Base\ntype: primary\ntools: tool_limits\nloadable-skills: bomb s000\n---\nIdentity."),
		"bomb": []byte("---\nname: bomb\ndescription: Bulk\ntype: loadable\ndepends: " + strings.Join(dependencies, " ") + "\nloadable-skills: " + list + "\n---\n" + body),
	}
	for _, name := range names {
		c.Skills.Catalog[name] = []byte("---\nname: " + name + "\ndescription: \"" + strings.Repeat("<", 256) + "\"\ntype: loadable\n---\n")
	}
	for _, name := range dependencies {
		c.Skills.Catalog[name] = []byte("---\nname: " + name + "\ndescription: Dependency\ntype: dependency\nloadable-skills: " + list + "\n---\n" + body)
	}
	return c
}

func TestSkillCompleteRecordRefusalIsAtomicAndRecoverable(t *testing.T) {
	app := New(nil)
	defer app.Close()
	c := repeatedOffersConfig(t)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			fmt.Fprint(w, `{"content":[{"type":"tool_use","id":"limit","name":"tool_limits","input":{"ai_callback_delay":0,"max_output_bytes":1}},{"type":"tool_use","id":"oversize","name":"load_skill","input":{"name":"bomb"}}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		} else {
			fmt.Fprint(w, `{"content":[{"type":"text","text":"refused safely"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}
	}))
	defer server.Close()
	c.BaseURL = server.URL
	a, err := app.NewAgent(c)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := a.InspectSkills()
	rawBefore, _ := os.ReadFile(c.LogPath)
	_, err = a.LoadSkill("bomb")
	var failure *SkillError
	if !errors.As(err, &failure) || failure.Code != "skill_too_large" || failure.Revision != 0 || !strings.Contains(failure.Detail, "record") {
		t.Fatalf("wrong refusal: %v", err)
	}
	after, _ := a.InspectSkills()
	rawAfter, _ := os.ReadFile(c.LogPath)
	if !reflect.DeepEqual(before, after) || string(rawBefore) != string(rawAfter) || a.faulted {
		t.Fatal("oversize changed state, log or storage health")
	}
	if _, err = a.Prompt(context.Background(), "try the large load"); err != nil {
		t.Fatal(err)
	}
	after, _ = a.InspectSkills()
	if !reflect.DeepEqual(before, after) {
		t.Fatal("model refusal changed authority")
	}
	attempts, results := 0, 0
	for _, e := range a.Events() {
		if e.Type == "skills_changed" {
			t.Fatal("oversize committed")
		}
		if e.Tool != nil && e.Tool.CallID == "oversize" {
			if e.Type == "tool_called" {
				attempts++
			}
			if e.Type == "tool_returned" {
				results++
				if len(e.Tool.Parts) != 1 || e.Tool.Parts[0].Text == nil || !strings.Contains(*e.Tool.Parts[0].Text, "skill_too_large") {
					t.Fatalf("missing controlled result: %+v", e.Tool)
				}
				if !strings.Contains(*e.Tool.Parts[0].Text, "tool_limits consumed by load_skill:") {
					t.Fatal("oversize did not consume pending limits")
				}
			}
		}
	}
	if attempts != 1 || results != 1 || requests != 2 {
		t.Fatalf("unpaired/terminal refusal: %d %d %d", attempts, results, requests)
	}
	if _, err = a.LoadSkill("s000"); err != nil {
		t.Fatal(err)
	}
	state, _ := a.SkillState()
	if state.Revision != 1 || state.Active[1].Activation != 2 {
		t.Fatal("refusal consumed counters")
	}
}

func TestSkillOversizeInitializationRefuses(t *testing.T) {
	app := New(nil)
	defer app.Close()
	c := repeatedOffersConfig(t)
	c.Skills.Primary = "bomb"
	c.Skills.Catalog["bomb"] = []byte(strings.Replace(string(c.Skills.Catalog["bomb"]), "type: loadable", "type: primary", 1))
	_, err := app.NewAgent(c)
	var failure *SkillError
	if !errors.As(err, &failure) || failure.Code != "skill_too_large" {
		t.Fatalf("wrong initial refusal: %v", err)
	}
	data, err := os.ReadFile(c.LogPath)
	if err != nil || string(data) != "{\"log_version\":1}\n" {
		t.Fatal("initial transition appended on refusal", err)
	}
}

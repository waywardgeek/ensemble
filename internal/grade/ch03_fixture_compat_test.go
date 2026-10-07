package grade

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

func TestGeminiDeclarationSchemaForms(t *testing.T) {
	for _, field := range []string{"parameters", "parametersJsonSchema", "parameters_json_schema"} {
		t.Run(field, func(t *testing.T) {
			declaration := map[string]any{
				"name": "read_file", "description": "Read text.",
				field: map[string]any{"type": "object", "properties": map[string]any{"path": map[string]any{"type": "string"}}},
			}
			for _, defect := range []string{"none", "missing", "wrong-type", "conflicting"} {
				t.Run(defect, func(t *testing.T) {
					copy := map[string]any{}
					for k, v := range declaration {
						copy[k] = v
					}
					switch defect {
					case "missing":
						delete(copy, field)
					case "wrong-type":
						copy[field] = "not a schema object"
					case "conflicting":
						other := "parameters"
						if field == other {
							other = "parametersJsonSchema"
						}
						copy[other] = declaration[field]
					}
					body, err := json.Marshal(map[string]any{"tools": []any{map[string]any{"functionDeclarations": []any{copy}}}})
					if err != nil {
						t.Fatal(err)
					}
					decls, ok, _ := vendorToolDecls("gemini", body)
					accepted := ok && len(decls) == 1 && decls[0].Schema != nil
					if accepted != (defect == "none") {
						t.Fatalf("accepted=%v for %s", accepted, defect)
					}
				})
			}
		})
	}
}

func TestFixtureResolvedIdentityIsExplicitAndLiveOnly(t *testing.T) {
	t.Setenv("LLM_RESOLVED_MODEL", "unrelated-inherited-setting")
	for _, vendor := range Ch2Vendors {
		for _, base := range []string{"", "http://127.0.0.1:1"} {
			for _, env := range [][]string{vendorEnv(vendor, base, "/tmp"), ch3Env(vendor, base, "/tmp", "/tmp/unused.log")} {
				resolved := "missing"
				for _, item := range env {
					if strings.HasPrefix(item, "LLM_RESOLVED_MODEL=") {
						resolved = strings.TrimPrefix(item, "LLM_RESOLVED_MODEL=")
					}
				}
				want := ""
				if base != "" {
					want = fakevendor.Models[vendor]
				}
				if resolved != want {
					t.Fatalf("%s base=%q resolved=%q want=%q", vendor, base, resolved, want)
				}
			}
		}
	}
}

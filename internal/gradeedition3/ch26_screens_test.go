package gradeedition3

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestScreenshotSelectionSchema(t *testing.T) {
	valid := `{"type":"object","properties":{"display":{"type":"integer","default":0},"ai_callback_delay":{"type":"number"}},"required":[]}`
	if err := screenshotSelectionSchema(json.RawMessage(valid)); err != nil {
		t.Fatal(err)
	}
	negatives := []string{
		`null`, `[]`, `{}`, `{"type":"object","properties":{}}`,
		`{"type":"object","properties":{"display":{"type":"string"}}}`,
		`{"type":"object","properties":{"display":{"type":"integer"}},"required":["display"]}`,
	}
	for _, name := range []string{"x", "y", "width", "height", "rectangle", "window"} {
		negatives = append(negatives, `{"type":"object","properties":{"display":{"type":"integer"},"`+name+`":{"type":"integer"}}}`)
	}
	for _, raw := range negatives {
		if err := screenshotSelectionSchema(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted invalid screenshot schema: %s", raw)
		}
	}
}

func TestHostScreenSchemaCandidate(t *testing.T) {
	bin := os.Getenv("CH26_HOST_CANDIDATE")
	if bin == "" {
		t.Skip("explicit CLI candidate required; declaration only")
	}
	if err := runHostScreenSchemaCLI(bin); err != nil {
		t.Fatal(err)
	}
}

func TestHostScreenSchemaExternalNegatives(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 is required for synthetic CLI candidates")
	}
	for _, mutation := range []string{"required-display", "rectangle"} {
		t.Run(mutation, func(t *testing.T) {
			root := t.TempDir()
			candidate := filepath.Join(root, "candidate")
			source := "#!" + python + "\n" + `import os,json,sys,urllib.request
sys.stdin.readline()
schema={'type':'object','properties':{'display':{'type':'integer'}},'required':[]}
if '` + mutation + `'=='required-display': schema['required']=['display']
else: schema['properties']['rectangle']={'type':'object'}
for tools in [[],[{'name':'screenshot','input_schema':schema},{'name':'list_displays','input_schema':{'type':'object','properties':{}}}]]:
 body=json.dumps({'model':'synthetic','messages':[{'role':'user','content':'schema only'}],'tools':tools}).encode()
 request=urllib.request.Request(os.environ['LLM_BASE_URL']+'/v1/messages',data=body,headers={'Content-Type':'application/json'})
 with urllib.request.urlopen(request) as response: response.read()
print(json.dumps({'observation':'turn_ended'}),flush=True)
`
			if err := os.WriteFile(candidate, []byte(source), 0700); err != nil {
				t.Fatal(err)
			}
			err := runHostScreenSchemaCLI(candidate)
			want := "display must be optional"
			if mutation == "rectangle" {
				want = "unsupported selector"
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("expected %s refusal, got %v", mutation, err)
			}
		})
	}
}

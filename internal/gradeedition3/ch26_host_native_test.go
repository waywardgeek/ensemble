package gradeedition3

import (
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestHostNativeAuthorizationGuard(t *testing.T) {
	config := HostNativeConfig{Candidate: "/must-not-launch", Pattern: HostNativePattern{Width: 64, Height: 32}}
	if err := runHostNativeCLI(config); err == nil || !strings.Contains(err.Error(), "authorization") {
		t.Fatalf("missing authorization: %v", err)
	}
	config.NativeAuthorized = true
	if err := runHostNativeCLI(config); err == nil || !strings.Contains(err.Error(), "distinct colors") {
		t.Fatalf("degenerate pattern: %v", err)
	}
	config.Display = 1
	if err := runHostNativeCLI(config); err == nil || !strings.Contains(err.Error(), "primary display") {
		t.Fatalf("unapproved primary fallback: %v", err)
	}
}

func TestHostNativeRunnerSyntheticCLI(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("synthetic candidate requires python3")
	}
	script, pixels := hostScriptFixture(t)
	for _, mutation := range []string{"positive", "metadata", "no-job", "replay-lost"} {
		t.Run(mutation, func(t *testing.T) {
			root := t.TempDir()
			candidate := filepath.Join(root, "candidate")
			source := "#!" + python + "\n" + `# Synthetic CLI only: no OS display enumeration or capture APIs.
import os,sys,json,urllib.request
pixels='` + base64.StdEncoding.EncodeToString(pixels) + `'
mutation='` + mutation + `'
schema={'type':'object','properties':{'display':{'type':'integer'}},'required':[]}
tools=[]
messages=[]
captures=0
for line in sys.stdin:
 human=json.loads(line)
 if mutation=='replay-lost' and captures==2: messages=[]
 messages.append({'role':'user','content':human['text']})
 while True:
  body=json.dumps({'model':'synthetic','messages':messages,'tools':tools}).encode()
  request=urllib.request.Request(os.environ['LLM_BASE_URL']+'/v1/messages',data=body,headers={'Content-Type':'application/json'})
  with urllib.request.urlopen(request) as response: reply=json.load(response)
  calls=[part for part in reply['content'] if part['type']=='tool_use']
  if not calls:
   print(json.dumps({'observation':'turn_ended'}),flush=True)
   break
  messages.append({'role':'assistant','content':reply['content']})
  results=[]
  for call in calls:
   name=call['name']
   image=[{'type':'image','source':{'type':'base64','media_type':'image/png','data':pixels}}]
   if name=='load_skill':
    tools=[{'name':'screenshot','input_schema':schema},{'name':'list_displays','input_schema':{'type':'object','properties':{}}}]
    content='loaded'
   elif name=='list_displays':
    content=json.dumps({'displays':[{'display':0,'width':64,'height':32,'pixel_width':64,'pixel_height':32}]})
   elif name=='screenshot':
    captures+=1
    content='job '+str(captures)+', status: running'
    if mutation=='metadata':content='saved screenshot.png'
    if mutation=='no-job':content=image
   elif name=='wait_for_job':content=image
   else:raise RuntimeError('Unexpected native tool in synthetic candidate')
   results.append({'type':'tool_result','tool_use_id':call['id'],'content':content})
  messages.append({'role':'user','content':results})
`
			if err := os.WriteFile(candidate, []byte(source), 0700); err != nil {
				t.Fatal(err)
			}
			config := HostNativeConfig{Candidate: candidate, NativeAuthorized: true, Pattern: HostNativePattern{Width: 64, Height: 32}}
			for index, rgb := range script.challenge.colors {
				config.Pattern.Colors[index] = [3]uint8{rgb.R, rgb.G, rgb.B}
			}
			err := runHostNativeCLI(config)
			if mutation == "positive" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			want := map[string]string{"metadata": "neither pixels nor a running Job", "no-job": "did not demonstrate delayed", "replay-lost": "host replay"}[mutation]
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("expected %s refusal, got %v", mutation, err)
			}
			if strings.Contains(err.Error(), "failed joined exit") {
				t.Fatal(err)
			}
		})
	}
}

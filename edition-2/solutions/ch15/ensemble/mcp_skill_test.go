package ensemble_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"ensemble/ensemble"
)

// A tiny standalone server exercises the real subprocess and newline wire.
// The primary names forbidden, but does not grant this server that name. This
// distinguishes per-server skill admission from the Agent's union of grants.
func TestMCPSkillStdioAdmissionAndUnload(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 unavailable")
	}
	workspace := t.TempDir()
	script := `import json,os,sys
open('peer.pid','w').write(str(os.getpid()))
for line in sys.stdin:
 m=json.loads(line); method=m.get('method')
 if 'id' not in m: continue
 if method=='initialize': result={'protocolVersion':'2024-11-05','capabilities':{'tools':{}}}
 elif method=='tools/list': result={'tools':[{'name':n,'description':n,'inputSchema':{'type':'object'}} for n in ['echo','forbidden']]}
 elif method=='tools/call': result={'content':[{'type':'text','text':'stdio-result'}]}
 else: continue
 print(json.dumps({'jsonrpc':'2.0','id':m['id'],'result':result}),flush=True)
`
	args, _ := json.Marshal([]string{"-u", "-c", script})
	dir := skillFiles(t, map[string]string{
		"base":   "---\nname: base\ntype: primary\ntools: forbidden\nloadable-skills: remote\n---\nBase",
		"remote": "---\nname: remote\ntype: loadable\ntools: echo\nmcp_servers:\n  - name: code-search\n    transport: stdio\n    command: " + python + "\n    args: " + string(args) + "\n---\nRemote tools: $TOOLS",
	})
	a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "fake", Workspace: workspace, DataDir: t.TempDir()}, ensemble.AgentSpec{SkillsDir: dir, PrimarySkill: "base"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { a.Shutdown() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	body, err := a.Invoke(ctx, "load_skill", json.RawMessage(`{"name":"remote"}`))
	if err != nil || !strings.Contains(body, "Loaded skill remote") {
		t.Fatalf("load %q %v", body, err)
	}
	if !slices.Contains(toolNames(a), "echo") || slices.Contains(toolNames(a), "forbidden") {
		t.Errorf("skill discovery grants: %v", toolNames(a))
	}
	if _, err := a.Invoke(ctx, "forbidden", nil); err == nil {
		t.Fatal("guessed ungranted server tool executed")
	}
	if text, err := a.Invoke(ctx, "echo", nil); err != nil || text != "stdio-result" {
		t.Fatalf("stdio call %q %v", text, err)
	}
	pidText, err := os.ReadFile(filepath.Join(workspace, "peer.pid"))
	if err != nil {
		t.Fatal(err)
	}
	pid, _ := strconv.Atoi(string(pidText))
	if err := syscall.Kill(pid, 0); err != nil {
		t.Fatalf("server not alive before unload: %v", err)
	}
	if _, err := a.Invoke(ctx, "unload_skill", json.RawMessage(`{"name":"remote"}`)); err != nil {
		t.Fatal(err)
	}
	if slices.Contains(toolNames(a), "echo") {
		t.Fatal("unload retained executable remote tool")
	}
	if _, err := a.Invoke(ctx, "echo", nil); err == nil {
		t.Fatal("unloaded remote tool executed")
	}
	if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
		t.Fatalf("unload did not reap subprocess: %v", err)
	}
	if !slices.Contains(a.Skills().PendingUnload(), "remote") {
		t.Fatal("compiled instruction removal lost pending state")
	}
}

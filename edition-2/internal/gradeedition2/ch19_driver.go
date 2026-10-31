package gradeedition2

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Observe a real completed terminal answer before EOF. The read deadline and
// process deadline bound broken fixtures; neither is substituted for completion.
func sandboxLaunch(bin string, f sandboxFixture, safe bool, probes []sandboxProbe, current bool) ([]fakevendor.Recorded, map[string]sandboxObservation, error) {
	return sandboxLaunchFor(bin, f, safe, probes, current, false)
}

func sandboxLaunchFor(bin string, f sandboxFixture, safe bool, probes []sandboxProbe, current, executable bool) ([]fakevendor.Recorded, map[string]sandboxObservation, error) {
	replies := make([]fakevendor.Reply, 0, len(probes)+1)
	for _, p := range probes {
		r := p.reply()
		if executable {
			r.Text = "I will exercise the next sandbox boundary."
		}
		replies = append(replies, r)
	}
	// Seed a known fake credential into the provider trace. We inspect the log
	// through the real file tool; filtering arbitrary assistant text is not a
	// substitute for sanitizing that readable file at write time.
	replies[0].Text = "CREDENTIAL_ECHO_" + f.nonce + " " + f.credential
	final := "SANDBOX_COMPLETE_" + f.nonce
	replies = append(replies, fakevendor.Reply{Text: final})
	var script *sandboxScript
	var provider *fakevendor.Server
	if current && !safe {
		script = &sandboxScript{probes: probes, firstText: replies[0].Text, final: final, explained: executable, observed: map[string]sandboxObservation{}}
		provider = fakevendor.NewWithOptions(nil, fakevendor.Options{Route: script.route})
	} else {
		provider = fakevendor.New(replies)
	}
	defer provider.Close()
	// A whole-session budget must exceed a legitimate command observation.
	// Historical and short safe-mode callers retain their original deadlines.
	deadline := time.Now().Add(90 * time.Second)
	args := []string{"chat", "--sandbox", f.root}
	if safe {
		args = append(args, "--safe-mode")
	}
	var p *mcpCommand
	var err error
	if current {
		args = []string{"--sandbox", f.root}
		if safe {
			args = append(args, "--safe-mode")
		}
		args = append(args, "chat")
		if executable {
			args = []string{"chat", "--no-stream", "--sandbox", f.root,
				"--safe-mode", fmt.Sprint(safe), "--enable-web-search", "false",
				"--skills-dir", f.skills, "--primary-skill", "base", "--save", filepath.Join(f.root, "save.json")}
		}
		if script != nil {
			p, err = startMCPCommandEnvTimeout(bin, f.root, args, sandboxEnv(os.Environ(), f, provider.URL()), time.Until(deadline))
		} else {
			p, err = startMCPCommandEnv(bin, f.root, args, sandboxEnv(os.Environ(), f, provider.URL()))
		}
	} else {
		p, err = startMCPCommand(bin, f.root, args, []string{"LLM_MODEL=claude-sonnet-5", "LLM_API_KEY=" + f.credential, "OPENAI_API_KEY=" + f.credential, "ANTHROPIC_API_KEY=" + f.credential, "SERVICE_TOKEN=" + f.credential, "LLM_BASE_URL=" + provider.URL(), "EN_SKILLS_DIR=" + f.skills, "EN_PRIMARY_SKILL=base"})
	}
	if err != nil {
		return nil, nil, err
	}
	defer p.close()
	if _, err = fmt.Fprintln(p.stdin, "Exercise the local sandbox."); err == nil {
		for {
			var line string
			if script != nil {
				line, err = p.readWithin(time.Until(deadline))
			} else {
				line, err = p.read()
			}
			if err != nil || strings.Contains(line, final) {
				break
			}
		}
	}
	p.stdin.Close()
	select {
	case exit := <-p.done:
		p.done <- exit
		if err == nil && exit != nil {
			err = fmt.Errorf("CLI exit: %w", exit)
		}
	case <-time.After(5 * time.Second):
		p.cancel()
		exit := <-p.done
		p.done <- exit
		if err == nil {
			err = fmt.Errorf("CLI failed to exit after completed turn and EOF")
		}
	}
	// Quiesce routed handlers before reading their session-owned state, including
	// when the CLI timed out or abandoned an HTTP request.
	provider.Close()
	reqs := provider.Requests()
	observed := map[string]sandboxObservation{}
	expected := len(replies)
	if script != nil {
		observed = script.observed
		expected += script.totalWaits
		if err == nil {
			err = script.err
		}
		if err == nil && (!script.finished || script.next != len(probes)) {
			err = fmt.Errorf("script did not finish all probes")
		}
	} else if len(reqs) == expected {
		for i, probe := range probes {
			observed[probe.id] = sandboxObservation{{reqs[i+1], probe.id}}
		}
	}
	if err == nil && len(reqs) != expected {
		err = fmt.Errorf("got %d requests for %d scripted replies including waits", len(reqs), expected)
	}
	if err != nil {
		err = fmt.Errorf("%w; stderr: %s", err, p.stderr.String())
	}
	if executable {
		os.WriteFile(filepath.Join(f.base, "host.stderr"), []byte(p.stderr.String()), 0600)
		data, _ := json.MarshalIndent(reqs, "", "  ")
		os.WriteFile(filepath.Join(f.base, "provider-requests.json"), data, 0600)
		summary, _ := json.MarshalIndent(map[string]any{"requests": len(reqs), "error": fmt.Sprint(err)}, "", "  ")
		os.WriteFile(filepath.Join(f.base, "session.json"), summary, 0600)
	}
	return reqs, observed, err
}

// Keep the seeded credential in trusted configuration so the existing child
// environment and readable-trace assertions still exercise real redaction.
func sandboxEnv(base []string, f sandboxFixture, endpoint string) []string {
	clean := make([]string, 0, len(base))
	for _, item := range base {
		if key, _, _ := strings.Cut(item, "="); key != "SERVICE_TOKEN" && key != "HOME" && key != "TMPDIR" {
			clean = append(clean, item)
		}
	}
	env := webStatusEnv(clean, f.root, f.skills, endpoint, "claude-sonnet-4-6", "base")
	for i, item := range env {
		key, _, _ := strings.Cut(item, "=")
		switch key {
		case "LLM_API_KEY", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY":
			env[i] = key + "=" + f.credential
		}
	}
	return append(env, "SERVICE_TOKEN="+f.credential, "HOME="+f.root, "TMPDIR="+f.root)
}

// Child specifications have no spawned-agent path yet. Compile a small external
// consumer of the public contract, with explicit import selection rather than
// planting code in a frozen submission or guessing its package identifier.
func sandboxChild(core, library string, c *grade.Check, current bool) {
	module := grade.DiscoverBase(core)
	if module == "" {
		fail(c, "submitted module not found")
		return
	}
	absolute, err := filepath.Abs(core)
	if err != nil {
		fail(c, "%v", err)
		return
	}
	work, err := os.MkdirTemp("", "sandbox-child-consumer-")
	if err != nil {
		fail(c, "%v", err)
		return
	}
	defer os.RemoveAll(work)
	pkg := module
	if library != "." && library != "" {
		pkg += "/" + strings.Trim(library, "/")
	}
	mod := fmt.Sprintf("module sandboxprobe\n\ngo 1.25\nrequire %s v0.0.0\nreplace %s => %s\n", module, module, strconv.Quote(absolute))
	if err = os.WriteFile(filepath.Join(work, "go.mod"), []byte(mod), 0600); err != nil {
		fail(c, "%v", err)
		return
	}
	source := strings.Replace(sandboxChildSource, "PUBLIC_PACKAGE", strconv.Quote(pkg), 1)
	if current {
		// The restart retains its existing public Config owner.
		source = strings.ReplaceAll(source, "api.AgentSpec", "api.Config")
	}
	if err = os.WriteFile(filepath.Join(work, "main.go"), []byte(source), 0600); err != nil {
		fail(c, "%v", err)
		return
	}
	command := exec.Command("go", "run", "-mod=mod", ".")
	command.Dir = work
	command.Env = append(os.Environ(), "GOWORK=off")
	out, err := command.CombinedOutput()
	if err != nil {
		fail(c, "public Child/Clamp consumer: %v\n%s", err, out)
	}
	for _, marker := range []string{"INHERITS", "NARROWS", "REPORTS", "VALID_SUBROOT", "EXISTING_SUBROOT", "SYMLINK_REFUSED"} {
		if !strings.Contains(string(out), "CHILD_"+marker+"\n") {
			fail(c, "public consumer missing %s: %s", marker, out)
		}
	}
}

const sandboxChildSource = `package main
import("fmt";"os";"path/filepath";"strings"; api PUBLIC_PACKAGE)
// Resolve the existing prefix, including an existing leaf symlink. This consumer
// compares effective boundaries, not the implementation's returned spelling.
func physical(p string) string {
 p,err:=filepath.Abs(p);if err!=nil{return ""}
 suffix:=[]string{}
 for {if real,e:=filepath.EvalSymlinks(p);e==nil{return filepath.Join(append([]string{real},suffix...)...)};parent:=filepath.Dir(p);if parent==p{return ""};suffix=append([]string{filepath.Base(p)},suffix...);p=parent}
}
func within(root,p string) bool {if p==""{return false};r,c:=physical(root),physical(p);return r!=""&&c!=""&&(c==r||strings.HasPrefix(c,r+string(filepath.Separator)))}
func main(){
 root,err:=os.MkdirTemp("","child-boundary-");if err!=nil{panic(err)};defer os.RemoveAll(root)
 root,err=filepath.EvalSymlinks(root);if err!=nil{panic(err)}
 parent:=api.AgentSpec{DataDir:root,LogPath:"parent.log",SavePath:"parent.save",SandboxRoot:root,SafeMode:true,EnableWebSearch:false}
 derived:=parent.Child(filepath.Join(root,"child"))
 if derived.DataDir==filepath.Join(root,"child")&&derived.LogPath==""&&derived.SavePath==""&&derived.SandboxRoot==root&&derived.SafeMode&&!derived.EnableWebSearch{fmt.Println("CHILD_INHERITS")}
 greedy:=derived;greedy.SafeMode=false;greedy.EnableWebSearch=true;greedy.SandboxRoot=""
 clamped,alarm:=greedy.Clamp(parent)
 // Do not stop after one failure: enforcement and its diagnostic are independent.
 if clamped.SafeMode&&!clamped.EnableWebSearch&&within(root,clamped.SandboxRoot){fmt.Println("CHILD_NARROWS")}
 if alarm!=nil{fmt.Println("CHILD_REPORTS")}
 narrow:=derived;narrow.SandboxRoot=filepath.Join(root,"future","sub")
 got,err:=narrow.Clamp(parent);if err==nil&&physical(got.SandboxRoot)==physical(narrow.SandboxRoot)&&got.SafeMode&&!got.EnableWebSearch{fmt.Println("CHILD_VALID_SUBROOT")}
 narrow.SandboxRoot=filepath.Join(root,"existing");if err=os.Mkdir(narrow.SandboxRoot,0700);err!=nil{panic(err)}
 got,err=narrow.Clamp(parent);if err==nil&&physical(got.SandboxRoot)==physical(narrow.SandboxRoot){fmt.Println("CHILD_EXISTING_SUBROOT")}
 outside,err:=os.MkdirTemp("","child-outside-");if err!=nil{panic(err)};defer os.RemoveAll(outside)
 alias:=filepath.Join(root,"escape");if err=os.Symlink(outside,alias);err!=nil{panic(err)}
 narrow.SandboxRoot=alias;got,err=narrow.Clamp(parent)
 if err!=nil&&within(root,got.SandboxRoot){fmt.Println("CHILD_SYMLINK_REFUSED")}
}
`

// The current CLI uses a PTY; the unattended fixture must not open a pager.
func sandboxGitCommand(current bool) string {
	log := "git log -1 --format=%s"
	if current {
		log = "git --no-pager log -1 --format=%s"
	}
	return "git init -q repo && cd repo && printf git-control > note && git add note && git commit -qm git-control && " + log
}

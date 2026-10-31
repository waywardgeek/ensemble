package gradeedition2

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// The fixture is an outside application, not an in-tree test or replacement
// GUI. Only construction differs for the explicitly selected historical API.
//
//go:embed ch20_consumer.go.txt
var reuseConsumer string

func reuseGo(dir string, args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "GOFLAGS=-mod=mod")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("go %v: %w\n%s", args, err, out)
	}
	return string(out), nil
}

func reuseChecksWith(core, gui, guiBin string, legacy, current bool, checks []grade.Check) {
	core, _ = filepath.Abs(core)
	gui, _ = filepath.Abs(gui)
	module := grade.DiscoverBase(core)
	guiModule := grade.DiscoverBase(gui)
	library := module + "/ensemble"
	if current {
		library = module
	}
	guiPackage := guiModule
	if legacy {
		library = module
		guiPackage = module + "/gui"
		gui = core
		guiModule = module
	}
	if module == "" || guiModule == "" {
		reuseFail(checks, fmt.Errorf("both actual module identities are required"), 0, 1, 2, 3, 4, 5, 6)
		return
	}
	work, err := os.MkdirTemp("", "reuse-grade-")
	if err != nil {
		reuseFail(checks, err, 0, 1, 2, 3, 4, 5, 6)
		return
	}
	defer os.RemoveAll(work)
	mod := fmt.Sprintf("module outside.example/edition2reuse\n\ngo 1.25.0\nrequire %s v0.0.0\nreplace %s => %s\n", module, module, strconv.Quote(core))
	if guiModule != module {
		mod += fmt.Sprintf("require %s v0.0.0\nreplace %s => %s\n", guiModule, guiModule, strconv.Quote(gui))
	}
	setup := `root:=api.New(io.Discard);a,err:=root.NewAgent(api.Config{Vendor:api.Anthropic,Model:"claude-sonnet-5",BaseURL:os.Getenv("LLM_BASE_URL"),APIKey:"fixture",DataDir:"data"});if err!=nil{panic(err)};defer a.Shutdown()`
	plant := `if err:=a.History().Append(api.Event{Type:"message_received",Message:&api.MessageData{Actor:api.Human,Parts:[]api.Part{{Type:"text",Text:"PUBLIC_EVENT_22"}}}});err!=nil{panic(err)}`
	bind := `hub,err:=gui.New(root,a,io.Discard);if err!=nil{panic(err)};closeGUI:=func(){hub.Close()};assets:=gui.Assets();serveWS:=hub.Handler("").ServeHTTP;fullHandler:=hub.Handler`
	head := `root:=api.New(io.Discard);a,err:=root.NewAgent(api.Config{Vendor:api.Anthropic,Model:"claude-sonnet-5",DataDir:"headless-data"});if err!=nil{panic(err)};defer a.Shutdown();err=a.MCP().Connect(context.Background(),"outside",api.NewMCPTransport(c,c,c));if err!=nil{panic(err)}`
	if current {
		setup = strings.ReplaceAll(strings.ReplaceAll(setup, "api.New(io.Discard)", "api.New(nil)"), "claude-sonnet-5", "claude-sonnet-4-6")
		plant = `if err:=a.RecordFact(api.Event{Type:api.MessageReceived,Message:&api.MessageData{Actor:"human",Parts:api.PartList{{Type:"text",Text:"PUBLIC_EVENT_22"}}}});err!=nil{panic(err)}`
		head = strings.ReplaceAll(strings.ReplaceAll(head, "api.New(io.Discard)", "api.New(nil)"), "claude-sonnet-5", "claude-sonnet-4-6")
	}
	if legacy {
		setup = `a,err:=api.NewAgent(api.Config{Vendor:api.VendorAnthropic,Model:"claude-sonnet-5",BaseURL:os.Getenv("LLM_BASE_URL"),APIKey:"fixture"},api.AgentSpec{DataDir:"data"});if err!=nil{panic(err)};defer a.Shutdown()`
		plant = `if err:=a.Engine().Record(api.Event{Type:api.MessageReceived,Message:&api.MessageData{Actor:api.ActorHuman,Parts:api.PartList{api.TextPart{Text:"PUBLIC_EVENT_22"}}}});err!=nil{panic(err)}`
		// Existing historical constructor receives the real actor's Send method and
		// owned history. This does not replace prompt delivery with a scripted hook.
		bind = `actor:=a.Actor();hub:=gui.New(gui.AgentHooks{Send:actor.Send,EventLog:a.EventLog()},"");actor.Attach(hub);ctx,cancel:=context.WithCancel(context.Background());defer cancel();go actor.Run(ctx);closeGUI:=hub.Close;assets,err:=fs.Sub(gui.Assets,"web");if err!=nil{panic(err)};serveWS:=hub.ServeWS;fullHandler:=gui.StaticHandler`
		head = `a,err:=api.NewAgent(api.Config{Vendor:api.VendorAnthropic,Model:"claude-sonnet-5"},api.AgentSpec{DataDir:"headless-data"});if err!=nil{panic(err)};defer a.Shutdown();err=a.ConnectMCP(api.NewMCPRawTransport(c,c));if err!=nil{panic(err)};_ = context.Background()`
	}
	source := strings.NewReplacer("CORE_PACKAGE", strconv.Quote(library), "GUI_PACKAGE", strconv.Quote(guiPackage), "SETUP_AGENT", setup, "PLANT_EVENT", plant, "SETUP_GUI", bind).Replace(reuseConsumer)
	headSource := `package main
import("context";"fmt";"io";"net";"os";api ` + strconv.Quote(library) + `)
func main(){if err:=os.MkdirAll("headless-data",0700);err!=nil{panic(err)};c,err:=net.Dial("tcp",os.Getenv("MCP_ADDR"));if err!=nil{panic(err)};defer c.Close();` + head + `;fmt.Println("HEADLESS_MCP_22");_ = io.Discard}`
	if err = recallWrite(work, map[string]string{"go.mod": mod, "guiapp/main.go": source, "headless/main.go": headSource}); err != nil {
		reuseFail(checks, err, 0, 1, 4, 5, 6)
		return
	}
	if _, err = reuseGo(work, "build", "-o", "bin/", "./..."); err != nil {
		reuseFail(checks, err, 0, 1, 4, 5, 6)
	} else {
		reuseDeps(work, core, gui, library, guiPackage, legacy, checks)
		reuseHeadless(filepath.Join(work, "bin", "headless"), work, &checks[5], current)
		reuseApp(filepath.Join(work, "bin", "guiapp"), work, checks, current)
	}
	cleanup := func() {}
	var guiBuildErr error
	if !current {
		guiBin, cleanup, guiBuildErr = grade.Build(gui)
	}
	if guiBuildErr != nil {
		reuseFail(checks, guiBuildErr, 2, 3)
		return
	}
	defer cleanup()
	reuseWire(guiBin, core, &checks[2], current)
	reuseDeploy(guiBin, &checks[3], current)
}

func reuseDeps(work, core, gui, library, guiPackage string, legacy bool, checks []grade.Check) {
	targets := []struct {
		dir, pkg string
		wantGUI  bool
	}{{work, "./headless", false}, {work, library, false}, {work, "./guiapp", true}, {core, ".", false}, {gui, "./cmd", true}}
	if legacy {
		targets[3].pkg = "./cmd"
		targets[3].wantGUI = true
	}
	for i, t := range targets {
		out, err := reuseGo(t.dir, "list", "-deps", t.pkg)
		if err != nil {
			reuseFail(checks, err, 0, 5, 6)
			continue
		}
		set := map[string]bool{}
		for _, p := range strings.Fields(out) {
			set[p] = true
		}
		if !set[library] {
			reuseFail(checks, fmt.Errorf("dependency positive control %s lacks actual public root", t.pkg), 0, 6)
			if i == 0 {
				reuseFail(checks, fmt.Errorf("headless MCP consumer lacks actual public root"), 5)
			}
		}
		ws := set["github.com/gorilla/websocket"]
		g := set[guiPackage]
		if g != t.wantGUI || ws != t.wantGUI {
			reuseFail(checks, fmt.Errorf("resolved target %s: GUI=%v WebSocket=%v wanted=%v", t.pkg, g, ws, t.wantGUI), 0)
			// The MCP row promises this same negative boundary, not merely a
			// successful handshake; do not leave its points after a coupled import.
			if i == 0 {
				reuseFail(checks, fmt.Errorf("headless MCP consumer imports GUI/WebSocket implementation"), 5)
			}
			if i == 1 {
				reuseFail(checks, fmt.Errorf("public root imports optional client"), 6)
			}
		}
	}
	checks[6].Details = append(checks[6].Details, "Dependency/public construction only; actual owner relationships and binary factory wiring require separate source review.")
}

func reuseHeadless(bin, work string, c *grade.Check, current bool) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fail(c, "%v", err)
		return
	}
	defer listener.Close()
	observed := make(chan []string, 1)
	go func() {
		conn, e := listener.Accept()
		if e != nil {
			observed <- nil
			return
		}
		defer conn.Close()
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		var methods []string
		d := json.NewDecoder(conn)
		for {
			var rpc map[string]any
			if d.Decode(&rpc) != nil {
				break
			}
			method, _ := rpc["method"].(string)
			methods = append(methods, method)
			if rpc["id"] == nil {
				continue
			}
			result := mcpInitResult()
			if method == "tools/list" {
				result = map[string]any{"tools": []any{}}
			}
			json.NewEncoder(conn).Encode(mcpReply(rpc["id"], result))
		}
		observed <- methods
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin)
	cmd.Dir = work
	env := os.Environ()
	if current {
		env = chapter20Env(work, "", "http://127.0.0.1:1/unused")
	}
	cmd.Env = append(env, "MCP_ADDR="+listener.Addr().String())
	out, err := cmd.CombinedOutput()
	listener.Close()
	methods := <-observed
	if err != nil || !strings.Contains(string(out), "HEADLESS_MCP_22") || !slices.Contains(methods, "initialize") || !slices.Contains(methods, "tools/list") {
		fail(c, "actual external core MCP handshake: %v %s methods=%v", err, out, methods)
	}
}

func reuseHTTP(c *grade.Check, url, marker string, status int) {
	client := http.Client{Timeout: 3 * time.Second}
	r, err := client.Get(url)
	if err != nil {
		fail(c, "GET %s: %v", url, err)
		return
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil || r.StatusCode != status || (status == 200 && len(b) == 0) || !strings.Contains(string(b), marker) {
		fail(c, "GET %s status=%d marker %q present=%v error=%v", url, r.StatusCode, marker, strings.Contains(string(b), marker), err)
	}
}

// Replay wraps already projected frames; the marker must be visible dialogue,
// never request arguments or an unrelated field of the envelope.
func reuseReplayText(m map[string]any, marker string, current bool) bool {
	if !current {
		return strings.Contains(guiPart(m).text, marker)
	}
	frames := []any{m}
	if m["type"] == "replay" {
		frames = list(object(m["data"])["events"])
	}
	for _, raw := range frames {
		frame := object(raw)
		part := guiPart(frame)
		if frame["type"] == "part_final" && part.kind == "text" && strings.Contains(part.text, marker) {
			return true
		}
	}
	return false
}

func reuseApp(bin, work string, checks []grade.Check, current bool) {
	provider := fakevendor.New([]fakevendor.Reply{{Text: "PUBLIC_REPLY_22"}})
	defer provider.Close()
	disk := filepath.Join(work, "disk")
	_ = recallWrite(disk, map[string]string{"index.html": "EXPLICIT_DISK_22"})
	var p *mcpCommand
	var err error
	if current {
		p, err = startMCPCommandEnv(bin, work, nil, append(chapter20Env(work, "", provider.URL()), "DISK_ASSETS="+disk))
	} else {
		p, err = startMCPCommand(bin, work, nil, []string{"LLM_BASE_URL=" + provider.URL(), "DISK_ASSETS=" + disk})
	}
	if err != nil {
		reuseFail(checks, err, 1, 4)
		return
	}
	defer p.close()
	line, err := p.read()
	var address map[string]string
	if err != nil || json.Unmarshal([]byte(line), &address) != nil || address["http"] == "" || address["mcp"] == "" {
		reuseFail(checks, fmt.Errorf("consumer readiness: %s %v", line, err), 1, 4)
		return
	}
	url := "http://" + address["http"]
	reuseHTTP(&checks[4], url+"/artifact-scroll.js", "ArtifactScroll", 200)
	reuseHTTP(&checks[4], url+"/renderers.js", "", 200)
	reuseHTTP(&checks[4], url+"/index.html", "", 404)
	reuseHTTP(&checks[3], url+"/full/", "<script", 200)
	reuseHTTP(&checks[3], url+"/disk/", "EXPLICIT_DISK_22", 200)
	tcp, err := net.DialTimeout("tcp", address["mcp"], time.Second)
	if err != nil {
		fail(&checks[1], "relay: %v", err)
		return
	}
	defer tcp.Close()
	tcp.SetDeadline(time.Now().Add(5 * time.Second))
	decoder := json.NewDecoder(tcp)
	fmt.Fprintln(tcp, `{"jsonrpc":"2.0","id":"outside-22","method":"tools/list"}`)
	var rpc map[string]any
	err = decoder.Decode(&rpc)
	if err != nil || rpc["id"] != "outside-22" || rpc["error"] == nil {
		fail(&checks[1], "no-browser correlated refusal: %v %v", rpc, err)
		return
	}
	ws, _, err := websocket.DefaultDialer.Dial("ws://"+address["http"]+"/ws", nil)
	if err != nil {
		fail(&checks[1], "%v", err)
		return
	}
	defer ws.Close()
	ws.SetReadDeadline(time.Now().Add(5 * time.Second))
	ws.WriteJSON(map[string]string{"type": "subscribe"})
	// Public events must appear in actual rendered content, not request arguments.
	replay := false
	for !replay {
		var m map[string]any
		if err = ws.ReadJSON(&m); err != nil {
			fail(&checks[1], "replay: %v", err)
			return
		}
		replay = reuseReplayText(m, "PUBLIC_EVENT_22", current)
		reuseBrowser(ws, m, current)
	}
	ws.WriteJSON(map[string]string{"type": "prompt", "text": "PUBLIC_PROMPT_22"})
	reply := false
	for {
		var m map[string]any
		if err = ws.ReadJSON(&m); err != nil {
			fail(&checks[1], "real prompt: %v", err)
			return
		}
		reuseBrowser(ws, m, current)
		reply = reply || (m["type"] == "part_final" && strings.Contains(guiPart(m).text, "PUBLIC_REPLY_22"))
		if m["type"] == "turn_ended" {
			if problem, _ := m["error"].(string); problem != "" {
				fail(&checks[1], "real prompt completion: %s", problem)
			}
			break
		}
	}
	if !reply {
		fail(&checks[1], "actual Agent reply did not reach the outside GUI client")
	}
	reqs := provider.Requests()
	if len(reqs) != 1 || !strings.Contains(string(reqs[0].Body), "PUBLIC_PROMPT_22") || !strings.Contains(string(reqs[0].Body), "PUBLIC_EVENT_22") {
		fail(&checks[1], "public inputs did not reach actual provider request")
	}
	// Browser selection is connection-owned. Preserve the no-browser refusal,
	// then use a new session accepted after the browser is actually connected.
	if current {
		_ = tcp.Close()
		tcp, err = net.DialTimeout("tcp", address["mcp"], time.Second)
		if err != nil {
			fail(&checks[1], "browser-bound relay: %v", err)
			return
		}
		defer tcp.Close()
		tcp.SetDeadline(time.Now().Add(5 * time.Second))
		decoder = json.NewDecoder(tcp)
	}
	fmt.Fprintln(tcp, `{"jsonrpc":"2.0","id":"outside-22","method":"tools/list","params":{"marker":"RELAY_RESULT_22"}}`)
	for {
		var m map[string]any
		if err = ws.ReadJSON(&m); err != nil {
			fail(&checks[1], "relay browser: %v", err)
			return
		}
		payload := object(m["payload"])
		if object(payload["params"])["marker"] == "RELAY_RESULT_22" {
			ws.WriteJSON(reuseBrowserReply(m, mcpReply(payload["id"], map[string]any{"marker": "RELAY_RESULT_22"}), current))
			break
		}
		reuseBrowser(ws, m, current)
	}
	rpc = nil
	err = decoder.Decode(&rpc)
	if err != nil || rpc["id"] != "outside-22" || object(rpc["result"])["marker"] != "RELAY_RESULT_22" {
		fail(&checks[1], "successful relay lost caller/result correlation: %v %v", rpc, err)
	}
	tcp.Close()
	p.stdin.Close()
	select {
	case exit := <-p.done:
		p.done <- exit
		if exit != nil {
			fail(&checks[1], "external consumer exit: %v", exit)
		}
	case <-time.After(5 * time.Second):
		fail(&checks[1], "consumer did not close after EOF")
	}
}

// A browser peer answers only MCP handshake traffic. Core execution is still
// the actual Agent; no callback or fabricated GUI event replaces it.
func reuseBrowserReply(frame map[string]any, payload any, current bool) map[string]any {
	key := "source"
	if current {
		key = "agent_id"
	}
	return map[string]any{"type": "jsonrpc", key: frame[key], "payload": payload}
}

func reuseBrowser(ws *websocket.Conn, m map[string]any, current bool) {
	rpc := object(m["payload"])
	if rpc["id"] == nil {
		return
	}
	method := rpc["method"]
	if method != "initialize" && method != "tools/list" {
		return
	}
	result := mcpInitResult()
	if method == "tools/list" {
		result = map[string]any{"tools": []any{}}
	}
	_ = ws.WriteJSON(reuseBrowserReply(m, mcpReply(rpc["id"], result), current))
}

func reuseDeploy(bin string, c *grade.Check, current bool) {
	work, err := os.MkdirTemp("", "deployed-gui-")
	if err != nil {
		fail(c, "%v", err)
		return
	}
	defer os.RemoveAll(work)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fail(c, "%v", err)
		return
	}
	port := fmt.Sprint(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	var p *mcpCommand
	if current {
		p, err = startMCPCommandEnv(bin, work, []string{"--port", port}, chapter20Env(work, "", "http://127.0.0.1:1/unused"))
	} else {
		p, err = startMCPCommand(bin, work, []string{"--port", port}, []string{"LLM_MODEL=claude-sonnet-5"})
	}
	if err != nil {
		fail(c, "%v", err)
		return
	}
	defer p.close()
	url := "http://127.0.0.1:" + port
	deadline := time.Now().Add(3 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		r, e := http.Get(url)
		if e == nil {
			r.Body.Close()
			ready = true
		}
		select {
		case exit := <-p.done:
			p.done <- exit
			fail(c, "deployed child exited before readiness: %v; stderr: %s", exit, p.stderr.String())
			return
		default:
		}
		if ready {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ready {
		fail(c, "deployed child did not serve its chosen port")
		return
	}
	reuseHTTP(c, url+"/", "<script", 200)
	reuseHTTP(c, url+"/artifact-scroll.js", "ArtifactScroll", 200)
	reuseHTTP(c, url+"/renderers.js", "", 200)
	p.stdin.Close()
	select {
	case exit := <-p.done:
		p.done <- exit
		if exit != nil {
			fail(c, "deployed GUI exit: %v", exit)
		}
	case <-time.After(5 * time.Second):
		fail(c, "deployed GUI did not exit after EOF")
	}
}

func reuseChecks(core, gui string, legacy bool, checks []grade.Check) {
	reuseChecksWith(core, gui, "", legacy, false, checks)
}

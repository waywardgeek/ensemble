//go:build ignore

// Run this standalone helper explicitly from the optional GUI module. Keeping it
// out of package discovery preserves the core module's headless build.
// A local public consumer re-admits retained facts for revised browser checks.
// It does not resume provider work or describe replay as a new model session.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"example.com/ensemble"
	gui "example.com/ensemble-gui"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if len(os.Args) != 4 {
		return fmt.Errorf("supply the three retained provider logs")
	}
	app := ensemble.New(os.Stderr)
	defer app.Close()
	workspace, err := os.MkdirTemp("", "ch07-replay-browser-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workspace)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	origin := "http://" + listener.Addr().String()
	mux := http.NewServeMux()
	var assets *gui.Server
	for i, name := range []string{"anthropic", "openai", "gemini"} {
		agent, err := app.NewAgent(ensemble.Config{Vendor: name, Model: "retained " + name + " answer", APIKey: "disabled", BaseURL: "http://127.0.0.1:1", DisableStreaming: true, Workspace: workspace, LogPath: filepath.Join(workspace, name+".log")})
		if err != nil {
			return err
		}
		if err := admit(agent, os.Args[i+1]); err != nil {
			return err
		}
		server, err := gui.NewServer(app, agent.ID(), origin, nil)
		if err != nil {
			return err
		}
		defer server.Close()
		if assets == nil {
			assets = server
		}
		mux.HandleFunc("/"+name+"/ws", func(w http.ResponseWriter, r *http.Request) {
			copy := r.Clone(r.Context())
			copy.URL.Path = "/ws"
			server.ServeHTTP(w, copy)
		})
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if "http://"+r.Host != origin {
			http.Error(w, "invalid Host", 403)
			return
		}
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, page)
			return
		}
		assets.ServeHTTP(w, r)
	})
	server := &http.Server{Handler: mux}
	defer server.Close()
	go server.Serve(listener)
	fmt.Println(origin)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	<-stop
	return nil
}
func admit(agent *ensemble.Agent, path string) error {
	input, err := os.Open(path)
	if err != nil {
		return err
	}
	defer input.Close()
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 16<<20)
	if !scanner.Scan() {
		return fmt.Errorf("missing retained header")
	}
	for scanner.Scan() {
		var event ensemble.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return err
		}
		if err := agent.Append(event); err != nil {
			return err
		}
	}
	return scanner.Err()
}

const page = `<!doctype html><html lang="en"><meta charset="utf-8"><title>Revised speech ownership — retained answers</title><link rel="stylesheet" href="/style.css"><body><h1>Retained provider answers in the revised browser</h1><div id="panels"></div><template id="panel"><main><h2 data-name></h2><p data-status></p><p data-pause></p><section data-artifacts class="artifacts" tabindex="0" aria-label="Conversation"></section><button data-latest>Return to latest</button><button data-auto-speech>Auto speech: off</button><button data-cancel-speech>Cancel speech</button><label>Prompt or correction<textarea data-input></textarea></label><button data-prompt>Send prompt</button><button data-hint>Send hint</button><button data-interrupt>Interrupt</button><p data-notice role="status"></p></main></template><script type="module">
import {BrowserApplication} from '/application.js';
window.application=new BrowserApplication();window.pages=[];
for(const name of ['anthropic','openai','gemini']){const root=document.querySelector('#panel').content.firstElementChild.cloneNode(true);root.querySelector('[data-name]').textContent=name;document.querySelector('#panels').append(root);pages.push(application.createPage(root,location.origin.replace(/^http/,'ws')+'/'+name+'/ws'))}
window.addEventListener('pagehide',()=>application.close(),{once:true});
</script></body></html>`

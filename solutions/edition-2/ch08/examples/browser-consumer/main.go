// An independent consumer replaces the page layout and reuses public GUI modules.
package main

import (
	"encoding/json"
	"example.com/ensemble"
	gui "example.com/ensemble-gui"
	"example.com/ensemble/cli"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	app := ensemble.New(os.Stderr)
	defer app.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	defer listener.Close()
	origin := "http://" + listener.Addr().String()
	mux := http.NewServeMux()
	var first *gui.Server
	for _, name := range []string{"first", "second"} {
		config := cli.Configuration(app)
		config.LogPath = filepath.Join(".", name+".jsonl")
		agent, err := app.NewAgent(config)
		if err != nil {
			return err
		}
		registration, err := agent.RegisterPause()
		if err != nil {
			return err
		}
		pause, err := registration.Update(true, false)
		if err != nil {
			return err
		}
		snapshot, watch, err := agent.Watch()
		if err != nil {
			return err
		}
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"agent": agent.ID(), "pause": pause, "snapshot": snapshot})
		registration.Close()
		watch.Close()
		server, err := gui.NewServer(app, agent.ID(), origin, nil, gui.ServerOptions{PreferencesPath: filepath.Join(".", name+"-preferences.json")})
		if err != nil {
			return err
		}
		defer server.Close()
		if first == nil {
			first = server
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
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, page)
		case "/consumer.js":
			w.Header().Set("Content-Type", "text/javascript")
			fmt.Fprint(w, script)
		default:
			first.ServeHTTP(w, r)
		}
	})
	server := &http.Server{Handler: mux}
	defer server.Close()
	go server.Serve(listener)
	fmt.Fprintln(os.Stdout, origin)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)
	<-stop
	return nil
}

const page = `<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>Two Agent workbench</title><link rel="stylesheet" href="/style.css"><body><h1>Two independent Agents, one application</h1><div id="workbench"></div><template id="panel"><main><h2 data-name></h2><button data-close-view>Close view</button><button data-open-view disabled>Reconnect view</button><p data-status></p><p data-pause></p><section data-artifacts class="artifacts" tabindex="0" aria-label="Conversation"></section><button data-latest>Return to latest</button><button data-auto-speech>Auto speech: off</button><button data-cancel-speech>Cancel speech</button><label>Prompt or correction<textarea data-input></textarea></label><button data-prompt>Send prompt</button><button data-hint>Send hint</button><button data-interrupt>Interrupt</button><p data-notice role="status"></p></main></template><script type="module" src="/consumer.js"></script></body></html>`
const script = `import {BrowserApplication} from '/application.js';
const application=new BrowserApplication();
for(const name of ['first','second']){
 const panel=document.querySelector('#panel').content.firstElementChild.cloneNode(true);
 panel.querySelector('[data-name]').textContent=name+' Agent';document.querySelector('#workbench').append(panel);
 const close=panel.querySelector('[data-close-view]'),open=panel.querySelector('[data-open-view]');let page;
 const mount=()=>{page=application.createPage(panel,location.origin.replace(/^http/,'ws')+'/'+name+'/ws');close.disabled=false;open.disabled=true};
 close.addEventListener('click',()=>{page.close();close.disabled=true;open.disabled=false;panel.querySelector('[data-status]').textContent='View closed';panel.querySelector('[data-pause]').textContent='No pause held by this view'});
 open.addEventListener('click',mount);mount();
}
window.addEventListener('pagehide',()=>application.close(),{once:true});`

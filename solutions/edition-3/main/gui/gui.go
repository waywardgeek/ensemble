// Package gui is a thin HTTP/browser adapter around the same public Agent.
// It has no provider wire types, reducer, or independent conversation store.
package gui

import (
	"encoding/json"
	"ensemble"
	"net/http"
	"sync"
)

func NewChat(a ensemble.Agent) http.Handler {
	var mu sync.Mutex // One browser session serializes this synchronous Agent.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(page))
		case http.MethodPost:
			var input struct {
				User string `json:"user"`
			}
			if json.NewDecoder(r.Body).Decode(&input) != nil {
				http.Error(w, "invalid input", 400)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			reply, err := a.Ask(r.Context(), input.User)
			if err != nil {
				http.Error(w, err.Error(), 502)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(struct {
				Assistant string         `json:"assistant"`
				Usage     ensemble.Usage `json:"usage"`
			}{reply, a.Usage()})
		default:
			w.Header().Set("Allow", "GET, POST")
			http.Error(w, "method not allowed", 405)
		}
	})
}

const page = `<!doctype html><html><head><meta charset="utf-8"><title>Ensemble Chapter 4</title></head><body><h1>Ensemble</h1><pre id="conversation"></pre><form><label>Message <input id="message" required></label><button>Send</button></form><p id="status"></p><script>
const form=document.querySelector('form'),input=document.querySelector('#message'),view=document.querySelector('#conversation'),status=document.querySelector('#status');
form.onsubmit=async e=>{e.preventDefault();const user=input.value;form.querySelector('button').disabled=true;status.textContent='Waiting…';try{const r=await fetch('/',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({user})});if(!r.ok)throw new Error(await r.text());const data=await r.json();view.textContent+='You: '+user+'\nAgent: '+data.assistant+'\n';input.value='';status.textContent=JSON.stringify(data.usage)}catch(err){status.textContent=err.message}finally{form.querySelector('button').disabled=false}};
</script></body></html>`

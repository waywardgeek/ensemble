package gui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"sync"
	"time"

	"example.com/ensemble"
	"example.com/ensemble-gui/internal/common"
	"example.com/ensemble-gui/internal/preferences"
	"github.com/gorilla/websocket"
)

//go:embed web/gui/*
var assets embed.FS

// ServerOwner is a connection's actual parent and its route to application services.
type ServerOwner = common.Server
type Preferences = common.Preferences
type PreferencesSnapshot = common.PreferencesSnapshot
type PreferencesService = common.PreferencesService
type ServerOptions struct{ PreferencesPath string }
type Server struct {
	preferences     *preferences.Service
	parent          ensemble.ClientOwner
	agentID, origin string
	mu              sync.Mutex
	connections     map[*Connector]bool
	closed          bool
	next            uint64
	traceMu         sync.Mutex
	trace           io.Writer
}

func NewServer(parent ensemble.ClientOwner, agentID, origin string, trace io.Writer, options ...ServerOptions) (*Server, error) {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "http" || parsed.Host == "" || parsed.Path != "" || parsed.User != nil {
		return nil, fmt.Errorf("expected local HTTP origin")
	}
	if parsed.Hostname() != "127.0.0.1" || parsed.Port() == "" {
		return nil, fmt.Errorf("server requires 127.0.0.1 and explicit port")
	}
	if len(options) > 1 {
		return nil, fmt.Errorf("only one ServerOptions allowed")
	}
	path := ".ensemble/gui-preferences.json"
	if len(options) == 1 && options[0].PreferencesPath != "" {
		path = options[0].PreferencesPath
	}
	s := &Server{parent: parent, agentID: agentID, origin: origin, trace: trace, connections: map[*Connector]bool{}}
	s.preferences, err = preferences.New(s, path)
	if err != nil {
		return nil, err
	}
	return s, nil
}
func (s *Server) Ensemble() ensemble.ClientOwner         { return s.parent }
func (s *Server) Preferences() common.PreferencesService { return s.preferences }
func (s *Server) AgentID() string                        { return s.agentID }
func (s *Server) Origin() string                         { return s.origin }
func (s *Server) Trace(id, direction, stage string, message any) {
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	if s.trace == nil {
		return
	}
	err := json.NewEncoder(s.trace).Encode(map[string]any{"time": time.Now().UTC().Format(time.RFC3339Nano), "connection_id": id, "direction": direction, "stage": stage, "message": message})
	if err != nil {
		s.trace = nil
		s.parent.Logf("GUI trace write failed; tracing disabled")
	}
}

// Assets exposes the reusable browser modules for alternative page layouts.
func (s *Server) Assets() http.Handler {
	sub, _ := fs.Sub(assets, "web/gui")
	return http.FileServer(http.FS(sub))
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if "http://"+r.Host != s.origin {
		http.Error(w, "invalid Host", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'none'; object-src 'none'; frame-src 'none'")
	if r.URL.Path != "/ws" {
		s.Assets().ServeHTTP(w, r)
		return
	}
	if r.Header.Get("Origin") != s.origin {
		http.Error(w, "invalid Origin", http.StatusForbidden)
		return
	}
	upgrade := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return r.Header.Get("Origin") == s.origin }}
	socket, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		socket.Close()
		return
	}
	s.next++
	c := NewConnector(s, fmt.Sprintf("c%d", s.next), socket)
	s.connections[c] = true
	s.mu.Unlock()
	c.Run()
	s.mu.Lock()
	delete(s.connections, c)
	s.mu.Unlock()
}
func (s *Server) Close() error {
	s.mu.Lock()
	s.closed = true
	all := make([]*Connector, 0, len(s.connections))
	for c := range s.connections {
		all = append(all, c)
	}
	s.mu.Unlock()
	s.preferences.Close()
	for _, c := range all {
		c.Close()
	}
	for _, c := range all {
		<-c.Done()
	}
	return nil
}

// Shutdown keeps the public server composable with an embedding HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})
	go func() { s.Close(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

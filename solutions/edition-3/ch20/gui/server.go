package gui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"ensemble"
	"github.com/gorilla/websocket"
)

type Server struct {
	Root         *ensemble.Ensemble
	Hub          *Hub
	Assets       string
	DefaultAgent string
	Trace        io.Writer
	traceMu      sync.Mutex
	closeOnce    sync.Once
	relayMu      sync.Mutex
	relayClosed  bool
	relayNext    uint64
	listeners    map[net.Listener]bool
	relays       map[string]*relaySession
	relayWG      sync.WaitGroup
}

func NewServer(root *ensemble.Ensemble, assets string) (*Server, error) {
	return NewServerWithSettings(root, assets, filepath.Join("cr", "settings.json"))
}

// NewServerFromConfig applies saved GUI policy before ordinary Agent recovery
// and capability initialization, not merely to its eventual view.
func NewServerFromConfig(root *ensemble.Ensemble, config ensemble.Config, assets, path string) (*Server, ensemble.Agent, error) {
	absoluteAssets, err := serverAssets(assets)
	if err != nil {
		return nil, nil, err
	}
	if path == "" {
		path = filepath.Join("cr", "settings.json")
	}
	file, err := root.ReadStartupFile(config, path)
	if err != nil {
		return nil, nil, fmt.Errorf("load GUI settings: %w", err)
	}
	settings := settingsForLaunch(config)
	if file.Exists {
		settings, err = patchSettings(settings, file.Data)
		if err != nil {
			return nil, nil, fmt.Errorf("load GUI settings: %w", err)
		}
	}
	// Carry only construction path choices, not the bootstrap's normalized
	// authentication/model/runtime configuration. Keep the same parsed snapshot.
	config.DataDir, config.SavePath = file.DataDir, file.SavePath
	config.SafeMode, config.EnableWebSearch = settings.SafeMode, settings.EnableWebSearch
	owner, err := root.CreateAgent(context.Background(), config)
	if err != nil {
		return nil, nil, err
	}
	store := settingsStore{path: file.Path, agent: owner}
	return &Server{Root: root, Hub: newHub(root, settings, store), Assets: absoluteAssets, DefaultAgent: owner.ID()}, owner, nil
}

func settingsStoreFor(owner ensemble.Agent, path string) (settingsStore, error) {
	if path == "" {
		path = filepath.Join("cr", "settings.json")
	}
	if owner != nil && !filepath.IsAbs(path) {
		path = filepath.Join(owner.Config().BaseDir, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return settingsStore{}, err
	}
	if owner != nil {
		if _, err = owner.Sandbox().ValidatePath(absolute); err != nil {
			return settingsStore{}, err
		}
	}
	return settingsStore{path: absolute, agent: owner}, nil
}
func serverAssets(assets string) (string, error) {
	if assets == "" {
		return "", nil // the default is the embedded filesystem, not the cwd
	}
	absolute, err := filepath.Abs(assets)
	if err != nil {
		return "", err
	}
	for _, name := range []string{"index.html", "app.js", "style.css"} {
		if info, err := os.Stat(filepath.Join(absolute, name)); err != nil || info.IsDir() {
			return "", fmt.Errorf("GUI asset missing: %s", filepath.Join(absolute, name))
		}
	}
	return absolute, nil
}

// NewServerWithSettings attaches to already constructed Agents. The launcher
// uses NewServerFromConfig so saved policy also governs their construction.
func NewServerWithSettings(root *ensemble.Ensemble, assets, path string) (*Server, error) {
	absoluteAssets, err := serverAssets(assets)
	if err != nil {
		return nil, err
	}
	var owner ensemble.Agent
	if agents := root.Agents(); len(agents) > 0 {
		owner = agents[0]
	}
	store, err := settingsStoreFor(owner, path)
	if err != nil {
		return nil, err
	}
	settings, err := store.load()
	if err != nil {
		return nil, fmt.Errorf("load GUI settings: %w", err)
	}
	return &Server{Root: root, Hub: newHub(root, settings, store), Assets: absoluteAssets}, nil
}

func (s *Server) wireTrace(direction string, raw []byte) {
	if s.Trace == nil {
		return
	}
	var payload bytes.Buffer
	if err := json.Compact(&payload, raw); err != nil {
		// Invalid inbound frames still leave one readable diagnostic record.
		// Quoting keeps malformed input readable without extra record lines.
		payload.Reset()
		quoted, _ := json.Marshal(string(raw))
		payload.Write(quoted)
	}
	configs := []ensemble.Config{}
	for _, a := range s.Root.Agents() {
		configs = append(configs, a.Config())
	}
	sanitized := ensemble.RedactCredentials(payload.String(), configs...)
	s.traceMu.Lock()
	defer s.traceMu.Unlock()
	if _, err := fmt.Fprintf(s.Trace, "%s %s %s\n", time.Now().Format(time.RFC3339Nano), direction, sanitized); err != nil {
		s.Hub.en.Logger().Printf("GUI diagnostic write: %v", err)
	}
}

func (s *Server) Close() {
	s.closeOnce.Do(func() { s.closeRelays(); s.Hub.Close() })
}

// Handler mounts the existing chat/WebSocket shell. Empty directory uses the
// embedded read-only assets; a nonempty directory is an explicit disk override.
func (s *Server) Handler(directory string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.socket)
	var files http.FileSystem = http.FS(Assets())
	if directory != "" {
		files = http.Dir(directory)
	}
	mux.Handle("/", http.FileServer(files))
	return mux
}

func (s *Server) socket(w http.ResponseWriter, r *http.Request) {
	up := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		return err == nil && u.Host == r.Host
	}}
	conn, err := up.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := s.Hub.connect()
	defer s.Hub.disconnect(c)
	defer conn.Close()
	writerDone := make(chan struct{})
	// Ping payloads are at most 125 bytes. The reader admits a bounded reply
	// without writing the socket or waiting behind a slow socket writer.
	pongs := make(chan string, 16)
	// Exactly one goroutine writes every data/control frame for this socket.
	go func() {
		defer close(writerDone)
		defer conn.Close()
		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-c.done:
				return
			case raw := <-c.queue:
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				start := time.Now()
				s.wireTrace("<", raw)
				if err := conn.WriteMessage(websocket.TextMessage, raw); err != nil {
					s.Hub.trace("socket_write_error", "", err.Error())
					return
				}
				s.Hub.trace("socket_write", "", fmt.Sprintf("client=%s bytes=%d elapsed=%s", c.id, len(raw), time.Since(start)))
			case payload := <-pongs:
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PongMessage, []byte(payload)); err != nil {
					s.Hub.trace("socket_write_error", "", err.Error())
					return
				}
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()
	conn.SetReadLimit(1 << 20)
	// Avoid the library's default reader-side pong/close writers.
	conn.SetPingHandler(func(payload string) error {
		select {
		case <-c.done:
			return io.EOF
		default:
		}
		select {
		case pongs <- payload:
			return nil
		default:
			// Do not silently discard a required Pong or block command reads.
			c.stop()
			s.Hub.trace("socket_control_overflow", "", c.id)
			return errors.New("websocket pong queue full")
		}
	})
	conn.SetCloseHandler(func(int, string) error { return nil })
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		s.wireTrace(">", raw)
		var cmd command
		if err := json.Unmarshal(raw, &cmd); err != nil {
			s.Hub.sendError(c, "", err)
			continue
		}
		s.Hub.trace("command", cmd.AgentID, cmd.Type)
		if err := s.command(c, cmd); err != nil {
			s.Hub.sendError(c, cmd.AgentID, err)
		}
	}
	c.stop()
	<-writerDone
}

type command struct {
	RequestID string            `json:"request_id,omitempty"`
	Payload   json.RawMessage   `json:"payload,omitempty"`
	Name      string            `json:"name,omitempty"`
	Type      string            `json:"type"`
	Settings  json.RawMessage   `json:"settings"`
	AgentID   string            `json:"agent_id"`
	Text      string            `json:"text"`
	Parts     ensemble.PartList `json:"parts"`
	Enabled   bool              `json:"enabled"`
	Automatic bool              `json:"automatic,omitempty"`
	PartID    uint64            `json:"part_id"`
	SpeechID  string            `json:"speech_id"`
	Status    string            `json:"status"`
}

func (s *Server) command(c *client, m command) error {
	if m.Type == "jsonrpc" {
		if handled, err := s.relayResponse(c, m); handled {
			return err
		}
	}
	target := m.AgentID
	if target == "" {
		target = s.DefaultAgent
	}
	if !s.Hub.allows(target) {
		return errors.New("Agent is outside this GUI view")
	}
	if m.Type == "answer_human" {
		return s.Root.RespondToHumanQuestion(target, m.RequestID, m.Text)
	}
	if m.Type == "update_settings" {
		if err := s.Hub.updateSettings(m.Settings); err != nil {
			s.Hub.sendSettingsError(c, err)
		}
		return nil
	}
	id := m.AgentID
	if id == "" {
		id = s.DefaultAgent
	}
	a, ok := s.Root.Agent(id)
	if !ok {
		return errors.New("unknown agent")
	}
	switch m.Type {
	case "set_model":
		if err := a.SetModel(m.Text); err != nil {
			return err
		}
		s.Hub.refreshStatus(a)
		return nil
	case "agent_status":
		s.Hub.refreshStatus(a)
		return nil
	case "mcp_open", "jsonrpc":
		m.AgentID = id
		return s.mcp(c, m)
	case "save":
		return a.Save()
	case "subscribe":
		return nil
	case "input", "prompt":
		if strings.TrimSpace(m.Text) == "" {
			return errors.New("empty input")
		}
		return a.Post(ensemble.UserMessage{Text: m.Text})
	case "hint":
		return a.Post(ensemble.Hint{Text: m.Text})
	case "pause":
		return a.Pause()
	case "unpause":
		if m.Automatic {
			return a.Unpause()
		}
		// Explicit Continue is user authority in every tab. Publish it before
		// the asynchronous state acknowledgement so fresh typing after this
		// notification cannot be cleared again by an older continuation epoch.
		s.Hub.mu.Lock()
		defer s.Hub.mu.Unlock()
		s.Hub.sync(s.Hub.view(a), a.View())
		if err := a.Unpause(); err != nil {
			return err
		}
		s.Hub.speech.cancelAgent(id, false)
		s.Hub.broadcast(Wire{Type: "continued", AgentID: id, Data: map[string]any{"client": c.id, "epoch": s.Hub.agents[id].state.Epoch}})
		return nil
	case "interrupt":
		return a.Post(ensemble.Interrupt{})
	case "load_skill":
		// Discovery may need replies on this same socket; never block its reader.
		go func() {
			if err := s.prepareBrowserSkill(c, a, m.Name); err != nil {
				s.Hub.sendError(c, id, err)
				return
			}
			if err := a.Skills().LoadSkill(m.Name); err != nil {
				s.Hub.sendError(c, id, err)
			}
		}()
		return nil
	case "unload_skill":
		return a.Skills().UnloadSkill(m.Name)
	case "skill_state":
		s.Hub.mu.Lock()
		s.Hub.broadcast(skillFrame(a))
		s.Hub.mu.Unlock()
		return nil
	case "speaking":
		return a.Post(ensemble.SetSpeaking{Enabled: m.Enabled})
	case "speak":
		// The server resolves a durable card; a browser cannot inject private replay
		// payloads or make a shortened preview stand in for the complete result.
		for _, e := range a.Log().EventsRange(0, a.View().Through) {
			for _, w := range projectEvent(id, e) {
				if w.PartID == m.PartID {
					if card := w.Part; card != nil && card.Text != "" && !strings.HasPrefix(card.Text, "[opaque") {
						s.Hub.mu.Lock()
						s.Hub.sync(s.Hub.view(a), a.View())
						source := speechSource(*card)
						text := cleanSpeech(card.Text)
						if card.Kind == "tool_call" || card.Kind == "tool_result" {
							text = strings.TrimSpace(card.Text)
						}
						if text != "" {
							s.Hub.speech.enqueue(id, spokenPart{Text: text, Source: source, PartID: m.PartID}, "on demand", card.TurnID)
						}
						s.Hub.mu.Unlock()
						return nil
					}
				}
			}
		}
		return errors.New("no readable final part")
	case "claim_speech":
		return s.Hub.speech.claim(c, id, m.SpeechID)
	case "speech_done":
		return s.Hub.speech.complete(c, id, m.SpeechID, m.Status)
	default:
		return errors.New("unknown command")
	}
}

// New embeds a view of an existing application-owned Agent. It opens no
// listener, does not change the Agent's policy, and does not own its shutdown.
func New(root *ensemble.Ensemble, agent ensemble.Agent, log io.Writer) (*Server, error) {
	if root == nil || nilValue(agent) || nilValue(log) {
		return nil, errors.New("GUI requires a root, Agent, and log writer")
	}
	owned, ok := root.Agent(agent.ID())
	if !ok || owned != agent {
		return nil, errors.New("Agent does not belong to supplied Ensemble")
	}
	settings := settingsForLaunch(agent.Config())
	store := settingsStore{path: filepath.Join(agent.Config().DataDir, "settings.json"), agent: agent}
	return &Server{Root: root, Hub: newScopedHub(root, settings, store, agent.ID()), DefaultAgent: agent.ID(), Trace: log}, nil
}
func nilValue(v any) bool {
	if v == nil {
		return true
	}
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	}
	return false
}

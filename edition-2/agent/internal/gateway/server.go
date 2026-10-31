// Package gateway implements the standalone edge router. It owns no Agents and
// performs no inference. The public gateway import supplies only the reusable
// wire/cryptography boundary; core implementation spokes are not involved.
package gateway

import (
	"bytes"
	"encoding/json"
	"ensemble/internal/common"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	wire "ensemble/gateway"
	"github.com/gorilla/websocket"
)

// Group binds operator-supplied trust to a set of allowed gateway operations.
type Group struct {
	// Description explains the intended authority to an administrator.
	Description string `json:"description"`
	// Capabilities contains send-message, declare-contactable or group-admin.
	Capabilities []string `json:"capabilities"`
	// Trust is owner, trusted, or external; only accepted sender groups stamp messages.
	Trust string `json:"trust"`
}

// AgentConfig is a configured or dynamically registered public identity.
type AgentConfig struct {
	// PublicKey pins the identity's X25519 public key.
	PublicKey string `json:"public_key"`
	// DisplayName is operator-selected presentation text, not authority.
	DisplayName string `json:"display_name,omitempty"`
	// Groups are base memberships; runtime administration cannot rewrite config groups.
	Groups []string `json:"groups"`
}

// Config is the gateway's operator-owned bootstrap policy.
type Config struct {
	// Gateway selects the standalone listener and persistence directory.
	Gateway struct {
		// Listen is the HTTP listener address; WebSocket authentication is independent of TLS.
		Listen string `json:"listen"`
		// DataDir contains the gateway identity and dynamic state.
		DataDir string `json:"data_dir"`
	} `json:"gateway"`
	// Groups are immutable at runtime; dynamic groups use the runtime: namespace.
	Groups map[string]Group `json:"groups"`
	// Agents are immutable bootstrap identities, separate from registered children.
	Agents map[string]AgentConfig `json:"agents"`
}

type dynamic struct {
	// Groups holds runtime definitions only; operator configuration has a separate owner.
	Groups map[string]Group `json:"groups"`
	// Children retains public identities registered through an authenticated parent.
	Children map[string]AgentConfig `json:"children"`
	// Memberships adds runtime grants without rewriting immutable base memberships.
	Memberships map[string][]string `json:"memberships"`
	// Contactable records recipient policy even while that recipient is offline.
	Contactable map[string][]string `json:"contactable"`
}
type queue struct {
	// Next is the highest allocated sequence, retained even when the queue empties.
	Next uint64 `json:"next"`
	// Ack is a cumulative receipt, not proof of model execution or external effects.
	Ack uint64 `json:"ack"`
	// Messages keeps every unacknowledged envelope in recipient sequence order.
	Messages []wire.Envelope `json:"messages"`
}

// Server is the application root for routing policy and connection lifetimes.
// A single mutex owns policy/queues; no socket write occurs while holding it.
// Authorization and queue admission are one operation under that mutex. A
// concurrent group edit can take effect before or after a send, never halfway
// through its permission check. Connections keep no cached copy of grants.
type Server struct {
	mu     sync.Mutex
	config Config
	state  dynamic
	queues map[string]*queue
	active map[string]*connection
	// sockets includes pending handshakes; active includes authenticated sessions only.
	// Shutdown must close both, otherwise a silent peer can outlive the listener.
	sockets map[*connection]bool
	stopped bool
	joined  sync.WaitGroup
	key     wire.Keypair
	logger  *log.Logger
}

// New loads policy and durable state before admitting connections. Broken state
// is an error, never an instruction to silently overwrite authorization history.
// New takes ownership of config maps; callers must not mutate them afterward.
func New(config Config, key wire.Keypair, diagnostics io.Writer) (*Server, error) {
	s := &Server{config: config, key: key, logger: log.New(diagnostics, "gateway: ", log.LstdFlags), queues: make(map[string]*queue), active: make(map[string]*connection), sockets: make(map[*connection]bool), state: dynamic{Groups: map[string]Group{}, Children: map[string]AgentConfig{}, Memberships: map[string][]string{}, Contactable: map[string][]string{}}}
	// Validate the bootstrap before opening or rewriting state files. A misspelled
	// trust label must not silently become a different authorization policy.
	for name, g := range config.Groups {
		if !validName(name) || isRuntime(name) {
			return nil, fmt.Errorf("invalid config group %q", name)
		}
		if err := validGroup(g); err != nil {
			return nil, err
		}
	}
	for id, a := range config.Agents {
		if !validName(id) {
			return nil, errors.New("invalid agent ID")
		}
		if _, err := wire.ParsePublicKey(a.PublicKey); err != nil {
			return nil, err
		}
		for _, g := range a.Groups {
			if _, ok := config.Groups[g]; !ok {
				return nil, fmt.Errorf("unknown config group %q", g)
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(config.Gateway.DataDir, "queues"), 0700); err != nil {
		return nil, err
	}
	if err := readJSON(filepath.Join(config.Gateway.DataDir, "state.json"), &s.state); err != nil {
		return nil, err
	}
	if s.state.Groups == nil || s.state.Children == nil || s.state.Memberships == nil || s.state.Contactable == nil {
		return nil, errors.New("incomplete dynamic state")
	}
	for name, g := range s.state.Groups {
		if !isRuntime(name) || !validName(name) {
			return nil, errors.New("invalid runtime group")
		}
		if err := validGroup(g); err != nil {
			return nil, err
		}
	}
	for id, a := range s.state.Children {
		if !validName(id) {
			return nil, errors.New("invalid child ID")
		}
		if _, ok := config.Agents[id]; ok {
			return nil, errors.New("child collides with config identity")
		}
		if _, err := wire.ParsePublicKey(a.PublicKey); err != nil {
			return nil, err
		}
	}
	// Queue files are named only from validated identity components. Inspecting
	// known recipients also avoids adopting unrelated files as new identities.
	for id := range s.identities() {
		q := &queue{}
		if err := readJSON(s.queuePath(id), q); err != nil {
			return nil, err
		}
		if q.Ack > q.Next {
			return nil, errors.New("invalid queue receipt")
		}
		// A saved queue is already stamped history. Validate its ordering and
		// recipient without restamping it using today's potentially changed policy.
		previous := q.Ack
		for _, e := range q.Messages {
			if e.Seq <= previous || e.Seq > q.Next || e.ToAgent != id {
				return nil, errors.New("invalid queued sequence")
			}
			previous = e.Seq
		}
		s.queues[id] = q
	}
	return s, nil
}

// Logf gives connection children access to the logger owned by this server.
// Callers report errors and validated identities, never frames or private keys.
func (s *Server) Logf(format string, args ...any) { s.logger.Printf(format, args...) }

func validName(name string) bool {
	ok, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_.:-]*$`, name)
	return ok
}
func isRuntime(name string) bool { return len(name) > 8 && name[:8] == "runtime:" }
func validGroup(g Group) error {
	if g.Trust != "owner" && g.Trust != "trusted" && g.Trust != "external" {
		return errors.New("group trust must be owner, trusted or external")
	}
	return nil
}
func (s *Server) identities() map[string]AgentConfig {
	all := make(map[string]AgentConfig)
	for id, a := range s.config.Agents {
		all[id] = a
	}
	for id, a := range s.state.Children {
		all[id] = a
	}
	return all
}
func (s *Server) identity(id string) (AgentConfig, bool) {
	a, ok := s.config.Agents[id]
	if !ok {
		a, ok = s.state.Children[id]
	}
	return a, ok
}
func (s *Server) group(name string) (Group, bool) {
	g, ok := s.config.Groups[name]
	if !ok {
		g, ok = s.state.Groups[name]
	}
	return g, ok
}

// Memberships are projected when needed rather than cached in connections.
// Runtime revocation therefore affects the next operation on an existing socket.
func (s *Server) memberships(id string) []string {
	a, _ := s.identity(id)
	return append(slices.Clone(a.Groups), s.state.Memberships[id]...)
}
func (s *Server) capable(id, capability string) bool {
	for _, name := range s.memberships(id) {
		g, _ := s.group(name)
		if slices.Contains(g.Capabilities, capability) {
			return true
		}
	}
	return false
}

// Missing files are first-run state. Existing invalid files remain errors so
// startup cannot erase pending messages under the guise of a fresh installation.
func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("null gateway state")
	}
	return json.Unmarshal(data, value)
}
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	// Rename prevents a torn individual JSON file. The periodic multi-file snapshot
	// promises orderly restart, not a transaction surviving arbitrary power loss.
	temp := path + ".tmp"
	if err = os.WriteFile(temp, append(data, '\n'), 0600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
func (s *Server) queuePath(id string) string {
	return filepath.Join(s.config.Gateway.DataDir, "queues", id+".json")
}

// Save writes dynamic policy and every recipient queue under the state lock.
// It snapshots receipts as well as envelopes. Keeping only outstanding content
// would lose the identity of already acknowledged messages after a restart.
// Queue files contain plaintext; the operator protects this private directory.
func (s *Server) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, q := range s.queues {
		if err := writeJSON(s.queuePath(id), q); err != nil {
			return err
		}
	}
	return writeJSON(filepath.Join(s.config.Gateway.DataDir, "state.json"), s.state)
}

// Close ends admission, interrupts even incomplete handshakes, joins connections,
// and saves once more. The executable owns its HTTP listener and periodic ticker.
func (s *Server) Close() error {
	s.mu.Lock()
	s.stopped = true
	for c := range s.sockets {
		c.close()
	}
	s.mu.Unlock()
	// The admission lock prevents Add from racing this join. Close does not hold
	// it while waiting, because each exiting connection must deregister itself.
	s.joined.Wait()
	return s.Save()
}

// ServeHTTP upgrades only the gateway endpoint. agent_id selects a candidate
// key; Noise proves it, and encrypted hello must repeat the same identity.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/v1/ws" {
		http.NotFound(w, r)
		return
	}
	id := r.URL.Query().Get("agent_id")
	s.mu.Lock()
	a, known := s.identity(id)
	stopped := s.stopped
	s.mu.Unlock()
	if !known || stopped {
		http.Error(w, "unknown or unavailable identity", http.StatusForbidden)
		return
	}
	conn, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{owner: s, id: id, socket: conn, done: make(chan struct{}), wake: make(chan struct{}, 1)}
	// Upgrade and shutdown can race: recheck admission after obtaining the socket,
	// then publish both the child and its parent relationship before it runs.
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		conn.Close()
		return
	}
	s.sockets[c] = true
	s.joined.Add(1)
	s.mu.Unlock()
	defer func() {
		c.close()
		s.mu.Lock()
		delete(s.sockets, c)
		if s.active[id] == c {
			delete(s.active, id)
		}
		s.mu.Unlock()
		s.joined.Done()
	}()
	// Bounded deadlines also keep a peer that never sends hello from holding shutdown.
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	public, err := wire.ParsePublicKey(a.PublicKey)
	if err != nil {
		return
	}
	c.noise, err = wire.NewNoiseKK("responder", s.key, public, []byte("ensemble-gateway-v1:"+id))
	if err != nil {
		s.Logf("peer %s: handshake setup: %v", id, err)
		return
	}
	kind, first, err := conn.ReadMessage()
	if err != nil || kind != websocket.BinaryMessage {
		s.Logf("peer %s: expected binary handshake: %v", id, err)
		return
	}
	if _, err = c.noise.ReadHandshakeMessage(first); err != nil {
		s.Logf("peer %s: handshake authentication failed", id)
		return
	}
	second, err := c.noise.WriteHandshakeMessage(nil)
	if err != nil {
		s.Logf("peer %s: handshake response: %v", id, err)
		return
	}
	if err = conn.WriteMessage(websocket.BinaryMessage, second); err != nil {
		s.Logf("peer %s: handshake write: %v", id, err)
		return
	}
	hello, err := c.read()
	if err != nil || hello.Type != "hello" || hello.AgentID != id || hello.Version != 1 {
		s.Logf("peer %s: invalid encrypted hello", id)
		return
	}
	s.mu.Lock()
	q := s.queues[id]
	if q == nil {
		q = &queue{}
		s.queues[id] = q
	}
	// A peer may report a receipt whose ack was lost on the old socket. It cannot
	// legitimately acknowledge a future message that this queue never allocated.
	// A receipt is still not a message execution record. The router deliberately
	// cannot tell whether an accepted prompt was answered, persisted by its
	// application, or used to perform an external side effect.
	if hello.Seq > q.Next {
		s.Logf("peer %s: receipt exceeds allocated sequence", id)
		s.mu.Unlock()
		return
	}
	s.ack(id, hello.Seq)
	// Only the newest authenticated session receives new deliveries. The old
	// handler remains responsible for its own cleanup and cannot remove this one.
	if old := s.active[id]; old != nil {
		old.close()
	}
	s.active[id] = c
	// The queue retains its next sequence even when empty. Reusing sequence one
	// after a drained queue would cause a reconnecting client to skip new work.
	c.sent.Store(q.Ack)
	welcome := wire.WelcomePayload{AgentID: id, QueueDepth: len(q.Messages), ServerVersion: "1"}
	s.mu.Unlock()
	if err = c.write(wire.Frame{Type: "welcome", Welcome: &welcome}); err != nil {
		return
	}
	conn.SetReadDeadline(time.Time{})
	writerDone := make(chan struct{})
	go func() { defer close(writerDone); c.deliver() }()
	defer func() { c.close(); <-writerDone }()
	c.notify()
	for {
		frame, err := c.read()
		if err != nil {
			return
		}
		s.mu.Lock()
		current := s.active[id] == c
		s.mu.Unlock()
		if !current {
			return
		}
		switch frame.Type {
		case "ack":
			s.mu.Lock()
			q = s.queues[id]
			if frame.Seq > c.sent.Load() {
				s.mu.Unlock()
				return
			}
			s.ack(id, frame.Seq)
			s.mu.Unlock()
		case "ping":
			if c.write(wire.Frame{Type: "pong"}) != nil {
				return
			}
		case "mcp_request":
			if frame.RPC == nil {
				return
			}
			response := s.request(id, *frame.RPC)
			if response != nil && c.write(wire.Frame{Type: "mcp_response", RPC: response}) != nil {
				return
			}
		case "response":
			// Reply direction does not confer additional trust. Even a reply
			// must satisfy the target's current contactability and the sender
			// capability check, just like an initial prompt.
			if frame.Message == nil {
				return
			}
			s.mu.Lock()
			err = s.route(id, *frame.Message)
			s.mu.Unlock()
			if err != nil {
				if c.write(wire.Frame{Type: "error", Error: err.Error()}) != nil {
					return
				}
			}
		default:
			if c.write(wire.Frame{Type: "error", Error: "unknown frame type"}) != nil {
				return
			}
		}
	}
}

// Cumulative receipts rely on the client's sequential handler contract. Earlier
// messages have all been accepted when a later sequence is acknowledged. Repeated
// or stale receipts are harmless; they must never move the cursor backward.
func (s *Server) ack(id string, seq uint64) {
	q := s.queues[id]
	if seq <= q.Ack {
		return
	}
	q.Ack = seq
	i := 0
	for i < len(q.Messages) && q.Messages[i].Seq <= seq {
		i++
	}
	q.Messages = append([]wire.Envelope(nil), q.Messages[i:]...)
}

// A connection is a server child, never a second owner of policy or queues.
// Its writer mutex serializes Noise nonces as well as WebSocket frames.
type connection struct {
	owner     common.GatewayOwner
	id        string
	socket    *websocket.Conn
	noise     *wire.NoiseKK
	writeMu   sync.Mutex
	closeOnce sync.Once
	done      chan struct{}
	wake      chan struct{}
	sent      atomic.Uint64 // written by delivery, read by ack validation
}

// Parent exposes the routing root that owns this connection and its queue.
func (c *connection) Parent() common.GatewayOwner { return c.owner }
func (c *connection) close()                      { c.closeOnce.Do(func() { close(c.done); c.socket.Close() }) }

// The wake signal is disposable, while queue entries are durable state. A full
// wake channel means the writer is already scheduled to inspect that state.
func (c *connection) notify() {
	select {
	case c.wake <- struct{}{}:
	default:
	}
}
func (c *connection) write(frame wire.Frame) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	c.socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
	data, err := json.Marshal(frame)
	if err != nil {
		c.Parent().Logf("peer %s: encode frame: %v", c.id, err)
		return err
	}
	return c.writeBytesLocked(data)
}
func (c *connection) writeBytesLocked(data []byte) error {
	data, err := c.noise.Encrypt(data, nil)
	if err == nil {
		err = c.socket.WriteMessage(websocket.BinaryMessage, data)
	}
	if err != nil {
		c.Parent().Logf("peer %s: write: %v", c.id, err)
		c.close()
	}
	return err
}

// Authentication precedes JSON decoding. Invalid ciphertext must never become
// an application command, even if its bytes happen to resemble an object.
func (c *connection) read() (wire.Frame, error) {
	kind, data, err := c.socket.ReadMessage()
	if err != nil {
		select {
		case <-c.done:
		default:
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				c.Parent().Logf("peer %s: read: %v", c.id, err)
			}
		}
		return wire.Frame{}, err
	}
	if kind != websocket.BinaryMessage {
		c.Parent().Logf("peer %s: expected encrypted binary frame", c.id)
		return wire.Frame{}, errors.New("expected binary frame")
	}
	data, err = c.noise.Decrypt(data, nil)
	if err != nil {
		c.Parent().Logf("peer %s: decrypt: %v", c.id, err)
		return wire.Frame{}, err
	}
	var frame wire.Frame
	err = json.Unmarshal(data, &frame)
	if err != nil {
		c.Parent().Logf("peer %s: decode frame: %v", c.id, err)
	}
	return frame, err
}

// One writer walks this session's cursor. New queue arrivals coalesce into wake
// signals; they do not spawn a goroutine per message or borrow the server lock
// across socket I/O. Reconnect starts a new cursor at the durable receipt.
func (c *connection) deliver() {
	for {
		select {
		case <-c.done:
			return
		case <-c.wake:
		}
		for {
			seq, data, ok := c.Parent().NextGatewayDelivery(c.id, c.sent.Load())
			if !ok {
				break
			}
			c.sent.Store(seq)
			c.writeMu.Lock()
			c.socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
			err := c.writeBytesLocked(data)
			c.writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// NextGatewayDelivery snapshots one frame without lending queue storage to a child.
func (s *Server) NextGatewayDelivery(id string, after uint64) (uint64, []byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, e := range s.queues[id].Messages {
		if e.Seq > after {
			data, _ := json.Marshal(wire.Frame{Type: "deliver", Envelope: &e})
			return e.Seq, data, true
		}
	}
	return 0, nil, false
}

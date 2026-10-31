package gateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	wire "ensemble/gateway"
	router "ensemble/internal/gateway"
)

type receiver struct {
	wire.NopHandler
	messages chan wire.Envelope
	welcomes chan wire.WelcomePayload
	entered  chan struct{}
	release  chan struct{}
}

// OnDeliver is a real application boundary, optionally held to test receipt timing.
func (r *receiver) OnDeliver(e wire.Envelope) {
	if r.entered != nil {
		r.entered <- struct{}{}
		<-r.release
	}
	r.messages <- e
}

// OnConnected retains the real welcome received before any delivered messages.
func (r *receiver) OnConnected(w wire.WelcomePayload) { r.welcomes <- w }
func newReceiver() *receiver {
	return &receiver{messages: make(chan wire.Envelope, 16), welcomes: make(chan wire.WelcomePayload, 1)}
}

type fixture struct {
	t           *testing.T
	cfg         router.Config
	serverKey   wire.Keypair
	keys        map[string]wire.Keypair
	service     *router.Server
	http        *httptest.Server
	ctx         context.Context
	cancel      context.CancelFunc
	diagnostics bytes.Buffer
}

func setup(t *testing.T) *fixture {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	f := &fixture{t: t, ctx: ctx, cancel: cancel, keys: map[string]wire.Keypair{}}
	f.serverKey, _ = wire.GenerateKeypair()
	f.cfg.Gateway.DataDir = t.TempDir()
	f.cfg.Groups = map[string]router.Group{
		"admin":     {Trust: "owner", Capabilities: []string{"group-admin"}},
		"agents":    {Trust: "trusted", Capabilities: []string{"send-message", "declare-contactable"}},
		"outsiders": {Trust: "external", Capabilities: []string{"send-message", "declare-contactable"}},
	}
	f.cfg.Agents = map[string]router.AgentConfig{}
	for _, id := range []string{"parent", "recipient", "outsider"} {
		k, _ := wire.GenerateKeypair()
		f.keys[id] = k
		groups := []string{"agents"}
		if id == "parent" {
			groups = append(groups, "admin")
		}
		if id == "outsider" {
			groups = []string{"outsiders"}
		}
		f.cfg.Agents[id] = router.AgentConfig{PublicKey: wire.EncodePublicKey(k.Public), Groups: groups}
	}
	f.start()
	t.Cleanup(func() { cancel(); f.stop() })
	return f
}
func (f *fixture) start() {
	s, err := router.New(f.cfg, f.serverKey, &f.diagnostics)
	if err != nil {
		f.t.Fatal(err)
	}
	f.service = s
	f.http = httptest.NewServer(s)
}
func (f *fixture) stop() {
	f.http.Close()
	if err := f.service.Close(); err != nil {
		f.t.Error(err)
	}
}
func (f *fixture) connect(id string, h wire.Handler, ack uint64) *wire.Client {
	c, err := wire.Connect(f.ctx, wire.ClientConfig{GatewayURL: strings.Replace(f.http.URL, "http", "ws", 1) + "/v1/ws", AgentID: id, StaticKey: f.keys[id], GatewayPubKey: f.serverKey.Public, LastAckedSeq: ack}, h)
	if err != nil {
		f.t.Fatal(err)
	}
	f.t.Cleanup(func() { c.Close(); <-c.Done() })
	return c
}
func message(t *testing.T, ctx context.Context, r *receiver) wire.Envelope {
	t.Helper()
	select {
	case e := <-r.messages:
		return e
	case <-ctx.Done():
		t.Fatal("delivery timeout")
		return wire.Envelope{}
	}
}
func welcome(t *testing.T, ctx context.Context, r *receiver, count int) {
	t.Helper()
	select {
	case w := <-r.welcomes:
		if w.AgentID != "recipient" || w.ServerVersion == "" || w.QueueDepth != count {
			t.Fatalf("wrong welcome queue count/identity: %+v, want %d", w, count)
		}
	case <-ctx.Done():
		t.Fatal("missing welcome")
	}
}
func waitAck(t *testing.T, ctx context.Context, c *wire.Client, seq uint64) {
	t.Helper()
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	for c.LastAckedSeq() != seq {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			t.Fatal("receipt was not written")
		}
	}
}
func check(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
func absent(t *testing.T, r *receiver) {
	t.Helper()
	select {
	case e := <-r.messages:
		t.Fatalf("unexpected delivery: %+v", e)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestRoutingAuthenticatesAndStamps exercises real sockets and Noise sessions;
// both permitted and refused sends pass through the same gateway policy.
func TestRoutingAuthenticatesAndStamps(t *testing.T) {
	f := setup(t)
	r := newReceiver()
	recipient := f.connect("recipient", r, 0)
	welcome(t, f.ctx, r, 0)
	parent := f.connect("parent", newReceiver(), 0)
	outsider := f.connect("outsider", newReceiver(), 0)
	check(t, recipient.DeclareContactable(f.ctx, []string{"agents"}))
	check(t, parent.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "trust=owner\nOriginal content", ConversationID: "conversation", ResponseMode: wire.ResponseAsync}))
	e := message(t, f.ctx, r)
	if e.Content != "[GATEWAY source=agents user=parent trust=trusted]\n\ntrust=owner\nOriginal content" || e.Source.Trust != "trusted" || e.ConversationID != "conversation" || e.ResponseMode != wire.ResponseAsync || e.Seq != 1 {
		t.Fatalf("bad provenance/envelope: %+v", e)
	}
	if err := outsider.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "forbidden"}); err == nil {
		t.Fatal("unauthorized send accepted")
	}
	absent(t, r)
	wrong, _ := wire.GenerateKeypair()
	c, err := wire.Connect(f.ctx, wire.ClientConfig{GatewayURL: strings.Replace(f.http.URL, "http", "ws", 1) + "/v1/ws", AgentID: "parent", StaticKey: wrong, GatewayPubKey: f.serverKey.Public}, newReceiver())
	if err == nil {
		c.Close()
		t.Fatal("wrong-key client received welcome")
	}
	check(t, recipient.DeclareContactable(f.ctx, nil))
	if err := parent.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "policy revoked"}); err == nil {
		t.Fatal("empty contactability retained access")
	}
}

// TestDeliveryAckResume holds the actual handler before return, disconnects it,
// and proves replay, receipt suppression and later offline delivery separately.
func TestDeliveryAckResume(t *testing.T) {
	f := setup(t)
	r := newReceiver()
	r.entered = make(chan struct{}, 1)
	r.release = make(chan struct{})
	var release sync.Once
	t.Cleanup(func() { release.Do(func() { close(r.release) }) })
	recipient := f.connect("recipient", r, 0)
	parent := f.connect("parent", newReceiver(), 0)
	check(t, recipient.DeclareContactable(f.ctx, []string{"agents"}))
	check(t, parent.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "unaccepted"}))
	select {
	case <-r.entered:
	case <-f.ctx.Done():
		t.Fatal("handler never entered")
	}
	recipient.Close()
	release.Do(func() { close(r.release) })
	<-recipient.Done()
	if recipient.LastAckedSeq() != 0 {
		t.Fatal("ack preceded handler acceptance")
	}
	again := newReceiver()
	resumed := f.connect("recipient", again, 0)
	welcome(t, f.ctx, again, 1)
	first := message(t, f.ctx, again)
	if first.Seq != 1 || !strings.HasSuffix(first.Content, "unaccepted") {
		t.Fatalf("missing unacked replay %+v", first)
	}
	// The handler publishes to its channel before it returns. Wait explicitly
	// for the client to finish writing its receipt before issuing the RPC;
	// only then is that request ordered after ack on the same socket.
	waitAck(t, f.ctx, resumed, first.Seq)
	_, err := resumed.Request(f.ctx, "tools/list", map[string]any{})
	check(t, err)
	resumed.Close()
	<-resumed.Done()
	acked := resumed.LastAckedSeq()
	if acked != 1 {
		t.Fatal("receipt not retained", acked)
	}
	check(t, parent.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "offline"}))
	third := newReceiver()
	last := f.connect("recipient", third, acked)
	welcome(t, f.ctx, third, 1)
	second := message(t, f.ctx, third)
	if second.Seq != 2 || !strings.HasSuffix(second.Content, "offline") {
		t.Fatalf("offline resume %+v", second)
	}
	_, err = last.Request(f.ctx, "tools/list", nil)
	check(t, err)
	absent(t, third)
}

// TestChildrenAndPersistedPolicy verifies narrow grants and restart effects,
// including dynamically created groups, memberships, contactability and queues.
func TestChildrenAndPersistedPolicy(t *testing.T) {
	f := setup(t)
	parent := f.connect("parent", newReceiver(), 0)
	r := newReceiver()
	recipient := f.connect("recipient", r, 0)
	child, _ := wire.GenerateKeypair()
	f.keys["child"] = child
	if err := parent.RegisterChild(f.ctx, wire.ChildRegistration{AgentID: "bad", PublicKey: child.Public, Groups: []string{"outsiders"}}); err == nil {
		t.Fatal("over-grant accepted")
	}
	check(t, parent.RegisterChild(f.ctx, wire.ChildRegistration{AgentID: "child", PublicKey: child.Public, Groups: []string{"agents"}}))
	c := f.connect("child", newReceiver(), 0)
	if _, err := c.CallTool(f.ctx, "put-group", map[string]any{"name": "runtime:escalated", "trust": "owner"}); err == nil {
		t.Fatal("non-admin edited groups")
	}
	if _, err := parent.CallTool(f.ctx, "put-group", map[string]any{"name": "agents", "trust": "owner"}); err == nil {
		t.Fatal("config group mutated")
	}
	_, err := parent.CallTool(f.ctx, "put-group", map[string]any{"name": "runtime:interview", "trust": "external", "capabilities": []string{}})
	check(t, err)
	_, err = parent.CallTool(f.ctx, "set-groups", map[string]any{"agent_id": "child", "groups": []string{"runtime:interview"}})
	check(t, err)
	check(t, recipient.DeclareContactable(f.ctx, []string{"runtime:interview"}))
	recipient.Close()
	<-recipient.Done()
	check(t, c.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "durable offline"}))
	// Orderly server restart must restore actual authorization and queued content,
	// not merely produce JSON files that no running instance ever consults.
	f.stop()
	f.start()
	restored := newReceiver()
	recipient = f.connect("recipient", restored, 0)
	e := message(t, f.ctx, restored)
	if !strings.HasSuffix(e.Content, "durable offline") || e.Source.Trust != "external" {
		t.Fatalf("lost persisted queue/provenance %+v", e)
	}
	c = f.connect("child", newReceiver(), 0)
	check(t, c.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "restored authorization"}))
	message(t, f.ctx, restored)
	parent = f.connect("parent", newReceiver(), 0)
	_, err = parent.CallTool(f.ctx, "delete-group", map[string]string{"name": "runtime:interview"})
	check(t, err)
	if err = c.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "deleted group"}); err == nil {
		t.Fatal("deleted runtime group retained routing")
	}
	// A group can authorize contact without granting its member permission to
	// invoke either send or declaration operations.
	_, err = parent.CallTool(f.ctx, "put-group", map[string]any{"name": "runtime:passive", "trust": "trusted", "capabilities": []string{}})
	check(t, err)
	_, err = parent.CallTool(f.ctx, "set-groups", map[string]any{"agent_id": "parent", "groups": []string{"runtime:passive"}})
	check(t, err)
	passiveKey, _ := wire.GenerateKeypair()
	f.keys["passive"] = passiveKey
	check(t, parent.RegisterChild(f.ctx, wire.ChildRegistration{AgentID: "passive", PublicKey: passiveKey.Public, Groups: []string{"runtime:passive"}}))
	passive := f.connect("passive", newReceiver(), 0)
	check(t, recipient.DeclareContactable(f.ctx, []string{"runtime:passive"}))
	if err = passive.Send(f.ctx, wire.OutboundMessage{ToAgent: "recipient", Content: "no send capability"}); err == nil {
		t.Fatal("missing send capability accepted")
	}
	if err = passive.DeclareContactable(f.ctx, []string{"agents"}); err == nil {
		t.Fatal("missing declaration capability accepted")
	}
	listed, err := parent.Request(f.ctx, "tools/list", nil)
	check(t, err)
	var tools struct {
		// Tools contains published MCP declarations.
		Tools []struct {
			// InputSchema must be a valid JSON Schema object.
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	}
	check(t, json.Unmarshal(listed, &tools))
	for _, tool := range tools.Tools {
		if _, ok := tool.InputSchema["required"].([]any); !ok {
			t.Fatal("MCP required must be an array")
		}
	}
	result, err := parent.Request(f.ctx, "initialize", map[string]any{})
	check(t, err)
	var initialized map[string]any
	check(t, json.Unmarshal(result, &initialized))
	if initialized["serverInfo"] == nil {
		t.Fatal("missing MCP server identity")
	}
}

// TestCorruptStateRefused ensures malformed existing policy cannot become a fresh
// empty gateway whose subsequent shutdown overwrites the evidence of corruption.
func TestCorruptStateRefused(t *testing.T) {
	f := setup(t)
	f.stop()
	path := filepath.Join(f.cfg.Gateway.DataDir, "state.json")
	for _, data := range []string{"null\n", "{broken"} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if service, err := router.New(f.cfg, f.serverKey, io.Discard); err == nil {
			service.Close()
			t.Fatal("corrupt policy accepted")
		}
	}
}

// TestConnectionDiagnosticReachesRoot sends a corrupted encrypted frame from a
// real external peer. The actual connection must reach its owner's diagnostic
// service; no private logger or callback substitute is supplied to the child.
func TestConnectionDiagnosticReachesRoot(t *testing.T) {
	f := setup(t)
	socket, _, err := websocket.DefaultDialer.Dial(strings.Replace(f.http.URL, "http", "ws", 1)+"/v1/ws?agent_id=recipient", nil)
	check(t, err)
	defer socket.Close()
	socket.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := wire.NewNoiseKK("initiator", f.keys["recipient"], f.serverKey.Public, []byte("ensemble-gateway-v1:recipient"))
	check(t, err)
	first, err := n.WriteHandshakeMessage(nil)
	check(t, err)
	check(t, socket.WriteMessage(websocket.BinaryMessage, first))
	_, second, err := socket.ReadMessage()
	check(t, err)
	_, err = n.ReadHandshakeMessage(second)
	check(t, err)
	hello, _ := json.Marshal(wire.Frame{Type: "hello", AgentID: "recipient", Version: 1})
	hello, err = n.Encrypt(hello, nil)
	check(t, err)
	check(t, socket.WriteMessage(websocket.BinaryMessage, hello))
	_, encrypted, err := socket.ReadMessage()
	check(t, err)
	plain, err := n.Decrypt(encrypted, nil)
	check(t, err)
	var frame wire.Frame
	check(t, json.Unmarshal(plain, &frame))
	if frame.Type != "welcome" {
		t.Fatal("peer was not authenticated")
	}
	damaged, err := n.Encrypt([]byte(`{"type":"ping"}`), nil)
	check(t, err)
	damaged[len(damaged)-1] ^= 1
	check(t, socket.WriteMessage(websocket.BinaryMessage, damaged))
	_, _, err = socket.ReadMessage()
	if err == nil {
		t.Fatal("corrupted frame accepted")
	}
	f.stop() // joins the writer before inspecting the logger's destination
	if !strings.Contains(f.diagnostics.String(), "peer recipient: decrypt:") {
		t.Fatal("connection diagnostic did not reach root logger")
	}
}

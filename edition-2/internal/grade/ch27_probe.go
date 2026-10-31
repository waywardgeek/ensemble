package grade

// ch27ProbeSource is a standalone program compiled in its own module
// against the student's public client library.
//
// It reports what it observed as OBS lines rather than deciding anything.
// Keeping judgment in the grader and observation in the probe means a
// student cannot make a check pass by printing the right word, and a
// failing step names the behavior that was missing instead of a Go error
// from deep inside a library.
const ch27ProbeSource = `package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	gw "github.com/waywardgeek/ensemble/agent/gateway"
)

func obs(key, value string) { fmt.Printf("OBS %s=%s\n", key, value) }

func probeFail(key, format string, args ...interface{}) {
	fmt.Printf("PROBEFAIL %s=%s\n", key, fmt.Sprintf(format, args...))
}

func okFail(key string, cond bool, format string, args ...interface{}) bool {
	if cond {
		obs(key, "ok")
		return true
	}
	probeFail(key, format, args...)
	return false
}

// rec is a Handler implementation living entirely outside the library's
// module, which is the point of the interface.
type rec struct {
	gw.NopHandler
	mu        sync.Mutex
	welcomes  chan gw.WelcomePayload
	inbox     chan gw.Envelope
	delivered int
}

func newRec() *rec {
	return &rec{
		welcomes: make(chan gw.WelcomePayload, 16),
		inbox:    make(chan gw.Envelope, 64),
	}
}

func (r *rec) OnDeliver(env gw.Envelope) {
	r.mu.Lock()
	r.delivered++
	r.mu.Unlock()
	select {
	case r.inbox <- env:
	default:
	}
}

func (r *rec) OnConnected(w gw.WelcomePayload) {
	select {
	case r.welcomes <- w:
	default:
	}
}

func (r *rec) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.delivered
}

func (r *rec) waitWelcome(d time.Duration) (gw.WelcomePayload, bool) {
	select {
	case w := <-r.welcomes:
		return w, true
	case <-time.After(d):
		return gw.WelcomePayload{}, false
	}
}

func (r *rec) waitEnvelope(d time.Duration) (gw.Envelope, bool) {
	select {
	case e := <-r.inbox:
		return e, true
	case <-time.After(d):
		return gw.Envelope{}, false
	}
}

var (
	dir        = flag.String("dir", ".", "key directory")
	mode       = flag.String("mode", "scenario", "genkeys or scenario")
	url        = flag.String("url", "", "gateway websocket URL")
	gatewayKey = flag.String("gateway-key", "", "gateway public key")
)

var agentIDs = []string{"parent", "peer", "outsider", "child1", "child2", "child3"}

func keyPath(id string) string { return filepath.Join(*dir, id+".key") }

func main() {
	flag.Parse()
	if *mode == "genkeys" {
		for _, id := range agentIDs {
			k, err := gw.LoadOrCreateKeypair(keyPath(id))
			if err != nil {
				fmt.Fprintf(os.Stderr, "keygen %s: %v\n", id, err)
				os.Exit(1)
			}
			fmt.Printf("KEY %s %s\n", id, gw.EncodePublicKey(k.Public))
		}
		return
	}
	scenario()
}

func ctx() context.Context {
	c, _ := context.WithTimeout(context.Background(), 20*time.Second)
	return c
}

func connect(id string, h gw.Handler, lastAcked uint64) (*gw.Client, error) {
	k, err := gw.LoadOrCreateKeypair(keyPath(id))
	if err != nil {
		return nil, err
	}
	pub, err := gw.ParsePublicKey(*gatewayKey)
	if err != nil {
		return nil, err
	}
	return gw.Connect(ctx(), gw.ClientConfig{
		GatewayURL:       *url,
		AgentID:          id,
		StaticKey:        k,
		GatewayPubKey:    pub,
		LastAckedSeq:     lastAcked,
		HandshakeTimeout: 10 * time.Second,
	}, h)
}

func scenario() {
	checkHandshake()

	// --- C3: the server accepts multiple authenticated agents ---
	parentRec := newRec()
	parent, err := connect("parent", parentRec, 0)
	if err != nil {
		probeFail("c3_parent_welcome", "parent could not connect: %v", err)
		return
	}
	if w, got := parentRec.waitWelcome(10 * time.Second); !got {
		probeFail("c3_parent_welcome", "no welcome frame arrived")
		return
	} else {
		// A wrong id is a real defect, but it does not stop the run.
		// Continuing means one broken observation cannot mask every
		// later check and turn a narrow bug into a blank report.
		okFail("c3_parent_welcome", w.AgentID == "parent",
			"welcome named %q, want parent", w.AgentID)
	}

	peerRec := newRec()
	peer, err := connect("peer", peerRec, 0)
	if err != nil {
		probeFail("c3_peer_welcome", "a second agent could not connect: %v", err)
		return
	}
	if _, got := peerRec.waitWelcome(10 * time.Second); !got {
		probeFail("c3_peer_welcome", "no welcome for the second agent")
		return
	}
	obs("c3_peer_welcome", "ok")

	checkImpostor()

	// --- C2 and C4: routing, provenance, trust ---
	if err := parent.DeclareContactable(ctx(), []string{"agents"}); err != nil {
		probeFail("c4_delivered", "parent could not declare contactability: %v", err)
		return
	}

	body := "what slows you down"
	if err := peer.Send(ctx(), gw.OutboundMessage{
		ToAgent:        "parent",
		Content:        body,
		ConversationID: "conv-1",
		ResponseMode:   gw.ResponseAsync,
	}); err != nil {
		probeFail("c4_delivered", "an authorized send was refused: %v", err)
		return
	}

	env, got := parentRec.waitEnvelope(10 * time.Second)
	if !okFail("c4_delivered", got, "an authorized message never arrived") {
		return
	}
	obs("c6_handler", "ok")

	okFail("c2_provenance", strings.HasPrefix(env.Content, "[GATEWAY "),
		"delivered content does not start with the provenance header: %q", env.Content)
	okFail("c2_body", strings.Contains(env.Content, body),
		"delivered content lost the original body: %q", env.Content)
	obs("c2_trust", string(env.Source.Trust))
	okFail("c2_convid", env.ConversationID == "conv-1",
		"conversation id = %q, want conv-1", env.ConversationID)
	obs("c2_mode", string(env.ResponseMode))

	// --- C4 refusal half ---
	outRec := newRec()
	outsider, err := connect("outsider", outRec, 0)
	if err != nil {
		probeFail("c4_refused", "outsider could not connect: %v", err)
		return
	}
	outRec.waitWelcome(10 * time.Second)
	before := parentRec.count()
	sendErr := outsider.Send(ctx(), gw.OutboundMessage{ToAgent: "parent", Content: "let me in"})
	okFail("c4_refused", sendErr != nil,
		"a sender in no group the recipient accepts was allowed to send")
	time.Sleep(700 * time.Millisecond)
	okFail("c4_not_delivered", parentRec.count() == before,
		"the refused message was delivered anyway")
	outsider.Close()

	// --- C5: queue while offline, ack, resume ---
	checkDelivery(parent, peer)

	// --- C7: parent registers a child ---
	checkRegistration(peer)
}

// checkHandshake exercises the handshake directly over byte slices, with no
// server involved, including the case that must fail.
func checkHandshake() {
	a, err := gw.GenerateKeypair()
	if err != nil {
		probeFail("c1_roundtrip", "keygen: %v", err)
		return
	}
	b, err := gw.GenerateKeypair()
	if err != nil {
		probeFail("c1_roundtrip", "keygen: %v", err)
		return
	}
	prologue := []byte("probe")

	ini, err := gw.NewNoiseKK("initiator", a, b.Public, prologue)
	if err != nil {
		probeFail("c1_roundtrip", "new initiator: %v", err)
		return
	}
	res, err := gw.NewNoiseKK("responder", b, a.Public, prologue)
	if err != nil {
		probeFail("c1_roundtrip", "new responder: %v", err)
		return
	}
	msg1, err := ini.WriteHandshakeMessage(nil)
	if err != nil {
		probeFail("c1_roundtrip", "write msg1: %v", err)
		return
	}
	if _, err := res.ReadHandshakeMessage(msg1); err != nil {
		probeFail("c1_roundtrip", "responder read msg1: %v", err)
		return
	}
	msg2, err := res.WriteHandshakeMessage(nil)
	if err != nil {
		probeFail("c1_roundtrip", "write msg2: %v", err)
		return
	}
	if _, err := ini.ReadHandshakeMessage(msg2); err != nil {
		probeFail("c1_roundtrip", "initiator read msg2: %v", err)
		return
	}
	if !ini.IsHandshakeComplete() || !res.IsHandshakeComplete() {
		probeFail("c1_roundtrip", "handshake not complete after two messages")
		return
	}
	ct, err := ini.Encrypt([]byte("secret"), nil)
	if err != nil {
		probeFail("c1_roundtrip", "encrypt: %v", err)
		return
	}
	pt, err := res.Decrypt(ct, nil)
	if err != nil {
		probeFail("c1_roundtrip", "decrypt: %v", err)
		return
	}
	okFail("c1_roundtrip", string(pt) == "secret",
		"decrypted %q, want secret", string(pt))

	// The negative half: a peer expecting a different static key must not
	// complete the handshake.
	wrong, _ := gw.GenerateKeypair()
	ini2, err := gw.NewNoiseKK("initiator", a, wrong.Public, prologue)
	if err != nil {
		obs("c1_wrongkey", "rejected")
		return
	}
	res2, err := gw.NewNoiseKK("responder", b, a.Public, prologue)
	if err != nil {
		obs("c1_wrongkey", "rejected")
		return
	}
	bad1, err := ini2.WriteHandshakeMessage(nil)
	if err != nil {
		obs("c1_wrongkey", "rejected")
		return
	}
	if _, err := res2.ReadHandshakeMessage(bad1); err != nil {
		obs("c1_wrongkey", "rejected")
		return
	}
	bad2, err := res2.WriteHandshakeMessage(nil)
	if err != nil {
		obs("c1_wrongkey", "rejected")
		return
	}
	if _, err := ini2.ReadHandshakeMessage(bad2); err != nil {
		obs("c1_wrongkey", "rejected")
		return
	}
	obs("c1_wrongkey", "accepted")
}

// checkImpostor connects with a valid agent id but the wrong private key.
func checkImpostor() {
	wrong, err := gw.GenerateKeypair()
	if err != nil {
		probeFail("c3_impostor", "keygen: %v", err)
		return
	}
	pub, err := gw.ParsePublicKey(*gatewayKey)
	if err != nil {
		probeFail("c3_impostor", "parse gateway key: %v", err)
		return
	}
	c, err := gw.Connect(ctx(), gw.ClientConfig{
		GatewayURL:       *url,
		AgentID:          "parent",
		StaticKey:        wrong,
		GatewayPubKey:    pub,
		HandshakeTimeout: 5 * time.Second,
	}, newRec())
	if err != nil {
		obs("c3_impostor", "rejected")
		return
	}
	// Connect may return before the gateway hangs up, so confirm the
	// connection is genuinely unusable rather than merely established.
	time.Sleep(500 * time.Millisecond)
	if err := c.DeclareContactable(ctx(), []string{"agents"}); err != nil {
		obs("c3_impostor", "rejected")
		c.Close()
		return
	}
	c.Close()
	obs("c3_impostor", "accepted")
}

func checkDelivery(parent *gw.Client, peer *gw.Client) {
	parent.Close()
	time.Sleep(500 * time.Millisecond)

	for i := 0; i < 3; i++ {
		if err := peer.Send(ctx(), gw.OutboundMessage{
			ToAgent: "parent",
			Content: fmt.Sprintf("offline-%d", i),
		}); err != nil {
			probeFail("c5_queued", "sending to an offline agent failed: %v", err)
			return
		}
	}

	r2 := newRec()
	p2, err := connect("parent", r2, 0)
	if err != nil {
		probeFail("c5_queued", "reconnect failed: %v", err)
		return
	}
	w, got := r2.waitWelcome(10 * time.Second)
	if !got {
		probeFail("c5_queued", "no welcome on reconnect")
		return
	}
	if w.QueuedCount != 3 {
		probeFail("c5_queued", "welcome reported %d queued, want 3", w.QueuedCount)
		return
	}
	var maxSeq uint64
	for i := 0; i < 3; i++ {
		env, ok := r2.waitEnvelope(10 * time.Second)
		if !ok {
			probeFail("c5_queued", "only %d of 3 queued messages were replayed", i)
			return
		}
		if env.Seq > maxSeq {
			maxSeq = env.Seq
		}
	}
	obs("c5_queued", "ok")

	// Give the client time to acknowledge, then confirm the queue drained.
	time.Sleep(1 * time.Second)
	p2.Close()
	time.Sleep(300 * time.Millisecond)

	r3 := newRec()
	p3, err := connect("parent", r3, 0)
	if err != nil {
		probeFail("c5_acked", "reconnect failed: %v", err)
		return
	}
	w3, got := r3.waitWelcome(10 * time.Second)
	if !got {
		probeFail("c5_acked", "no welcome on reconnect")
		return
	}
	if !okFail("c5_acked", w3.QueuedCount == 0,
		"after acknowledgement %d messages were still queued", w3.QueuedCount) {
		p3.Close()
		return
	}
	p3.Close()
	time.Sleep(300 * time.Millisecond)

	// Two more while offline, then resume claiming the first was handled.
	for i := 0; i < 2; i++ {
		if err := peer.Send(ctx(), gw.OutboundMessage{
			ToAgent: "parent",
			Content: fmt.Sprintf("resume-%d", i),
		}); err != nil {
			probeFail("c5_resume", "send failed: %v", err)
			return
		}
	}
	r4 := newRec()
	p4, err := connect("parent", r4, maxSeq+1)
	if err != nil {
		probeFail("c5_resume", "reconnect failed: %v", err)
		return
	}
	defer p4.Close()
	w4, got := r4.waitWelcome(10 * time.Second)
	if !got {
		probeFail("c5_resume", "no welcome on resume")
		return
	}
	if w4.QueuedCount != 1 {
		probeFail("c5_resume", "resuming after seq %d replayed %d messages, want 1", maxSeq+1, w4.QueuedCount)
		return
	}
	env, ok := r4.waitEnvelope(10 * time.Second)
	if !ok {
		probeFail("c5_resume", "the remaining message was not delivered")
		return
	}
	okFail("c5_resume", env.Seq == maxSeq+2,
		"resumed message had seq %d, want %d", env.Seq, maxSeq+2)
}

func checkRegistration(peer *gw.Client) {
	pRec := newRec()
	parent, err := connect("parent", pRec, 0)
	if err != nil {
		probeFail("c7_registered", "parent reconnect failed: %v", err)
		return
	}
	defer parent.Close()
	if _, got := pRec.waitWelcome(10 * time.Second); !got {
		probeFail("c7_registered", "no welcome")
		return
	}
	if err := parent.DeclareContactable(ctx(), []string{"agents"}); err != nil {
		probeFail("c7_registered", "declare: %v", err)
		return
	}

	childKey, err := gw.LoadOrCreateKeypair(keyPath("child1"))
	if err != nil {
		probeFail("c7_registered", "child key: %v", err)
		return
	}
	if err := parent.RegisterChild(ctx(), gw.ChildRegistration{
		AgentID:   "child1",
		PublicKey: childKey.Public,
		Groups:    []string{"agents"},
	}); err != nil {
		probeFail("c7_registered", "a parent could not register a child: %v", err)
		return
	}
	obs("c7_registered", "ok")

	cRec := newRec()
	child, err := connect("child1", cRec, 0)
	if err != nil {
		probeFail("c7_child_connected", "the registered child could not connect: %v", err)
		return
	}
	defer child.Close()
	if _, got := cRec.waitWelcome(10 * time.Second); !got {
		probeFail("c7_child_connected", "no welcome for the child")
		return
	}
	obs("c7_child_connected", "ok")

	before := pRec.count()
	if err := child.Send(ctx(), gw.OutboundMessage{ToAgent: "parent", Content: "hello parent"}); err != nil {
		probeFail("c7_child_routed", "the child could not use its granted group: %v", err)
		return
	}
	deadline := time.Now().Add(10 * time.Second)
	delivered := false
	for time.Now().Before(deadline) {
		if pRec.count() > before {
			delivered = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	okFail("c7_child_routed", delivered, "the child's message never arrived")

	// Narrow-only: peer holds agents but not admin.
	k2, err := gw.LoadOrCreateKeypair(keyPath("child2"))
	if err != nil {
		probeFail("c7_narrow_ok", "child2 key: %v", err)
		return
	}
	err = peer.RegisterChild(ctx(), gw.ChildRegistration{
		AgentID: "child2", PublicKey: k2.Public, Groups: []string{"agents"},
	})
	okFail("c7_narrow_ok", err == nil, "a parent could not grant a group it holds: %v", err)

	k3, err := gw.LoadOrCreateKeypair(keyPath("child3"))
	if err != nil {
		probeFail("c7_no_widen", "child3 key: %v", err)
		return
	}
	err = peer.RegisterChild(ctx(), gw.ChildRegistration{
		AgentID: "child3", PublicKey: k3.Public, Groups: []string{"admin"},
	})
	okFail("c7_no_widen", err != nil, "a parent granted the admin group it does not hold")
}
`

package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ensemble/ensemble"
	"ensemble/gateway"
	router "ensemble/internal/gateway"
)

type answerSink struct {
	gateway.NopHandler
	answers chan gateway.Envelope
}

// OnDeliver records the independently routed reply, not actor progress.
func (h *answerSink) OnDeliver(e gateway.Envelope) { h.answers <- e }

// TestGatewayUsesActualActor keeps the real client, routing, actor, renderer and
// provider parser. Only the external inference service is a small HTTP fake.
func TestGatewayUsesActualActor(t *testing.T) {
	requests := make(chan string, 2)
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		requests <- string(data)
		json.NewEncoder(w).Encode(map[string]any{"content": []map[string]string{{"type": "text", "text": "actor answer"}}, "usage": map[string]int{"input_tokens": 10, "output_tokens": 2}})
	}))
	defer provider.Close()
	root := ensemble.New(io.Discard)
	agent, err := root.NewAgent(ensemble.Config{Model: "fake", BaseURL: provider.URL, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer agent.Shutdown()
	cfg := router.Config{}
	cfg.Gateway.DataDir = t.TempDir()
	cfg.Groups = map[string]router.Group{"agents": {Trust: "trusted", Capabilities: []string{"send-message", "declare-contactable"}}}
	cfg.Agents = map[string]router.AgentConfig{}
	keys := map[string]gateway.Keypair{}
	for _, id := range []string{"subject", "interviewer"} {
		key, _ := gateway.GenerateKeypair()
		keys[id] = key
		cfg.Agents[id] = router.AgentConfig{PublicKey: gateway.EncodePublicKey(key.Public), Groups: []string{"agents"}}
	}
	serverKey, _ := gateway.GenerateKeypair()
	service, err := router.New(cfg, serverKey, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer service.Close()
	server := httptest.NewServer(service)
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	app := &application{root: root, agent: agent, ready: make(chan struct{}), ctx: ctx}
	connect := func(id string, h gateway.Handler) *gateway.Client {
		t.Helper()
		c, err := gateway.Connect(ctx, gateway.ClientConfig{GatewayURL: strings.Replace(server.URL, "http", "ws", 1) + "/v1/ws", AgentID: id, StaticKey: keys[id], GatewayPubKey: serverKey.Public}, h)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	subject := connect("subject", app)
	app.client = subject
	close(app.ready)
	sink := &answerSink{answers: make(chan gateway.Envelope, 2)}
	interviewer := connect("interviewer", sink)
	defer func() {
		cancel()
		subject.Close()
		interviewer.Close()
		<-subject.Done()
		<-interviewer.Done()
		app.replies.Wait()
	}()
	for _, c := range []*gateway.Client{subject, interviewer} {
		if err = c.DeclareContactable(ctx, []string{"agents"}); err != nil {
			t.Fatal(err)
		}
	}
	if err = interviewer.Send(ctx, gateway.OutboundMessage{ToAgent: "subject", Content: "real actor question", ConversationID: "specific-request", ResponseMode: gateway.ResponseAsync}); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-sink.answers:
		if e.Content != "[GATEWAY source=agents user=subject trust=trusted]\n\nactor answer" || e.ConversationID != "specific-request" {
			t.Fatalf("wrong reply %+v", e)
		}
	case <-ctx.Done():
		t.Fatal("actor reply missing")
	}
	select {
	case request := <-requests:
		if !strings.Contains(request, "[GATEWAY source=agents user=interviewer trust=trusted]") || !strings.Contains(request, "real actor question") {
			t.Fatal("provenance did not reach model")
		}
	case <-ctx.Done():
		t.Fatal("no model request")
	}
}

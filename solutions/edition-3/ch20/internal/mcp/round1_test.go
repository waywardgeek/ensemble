package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"ensemble/internal/common"
)

type emptySkills struct{ common.SkillRegistry }

func (emptySkills) Snapshot() common.CapabilitySnapshot { return common.CapabilitySnapshot{} }

type serviceAgent struct{ common.Agent }

func (serviceAgent) Skills() common.SkillRegistry { return emptySkills{} }

func TestIncomingAliasesRequireHandshake(t *testing.T) {
	manager := New(serviceAgent{})
	defer manager.Close()
	left, right := net.Pipe()
	served := make(chan error, 1)
	go func() { served <- manager.Serve(context.Background(), NewStream(left)) }()
	client := newPeer(NewStream(right), nil, false)
	defer client.close(io.EOF)
	for _, method := range []string{"tools/builtin/list", "tools/builtin/call"} {
		var result any
		err := client.call(context.Background(), method, nil, &result)
		var rpc *rpcError
		if !errors.As(err, &rpc) || rpc.Code != -32002 {
			t.Fatalf("%s before initialize: %v, want not initialized", method, err)
		}
	}
	var result any
	if err := client.call(context.Background(), "initialize", map[string]any{}, &result); err != nil {
		t.Fatal(err)
	}
	if err := client.call(context.Background(), "tools/builtin/list", nil, &result); err == nil {
		t.Fatal("initialize without initialized admitted alias")
	}
	if err := client.notify(context.Background(), "notifications/initialized", nil); err != nil {
		t.Fatal(err)
	}
	if err := client.call(context.Background(), "tools/builtin/list", nil, &result); err != nil {
		t.Fatal(err)
	}
	client.close(io.EOF)
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}

type blockedStream struct {
	entered, closed chan struct{}
	once            sync.Once
}

func (b *blockedStream) Read([]byte) (int, error) { <-b.closed; return 0, io.EOF }
func (b *blockedStream) Write([]byte) (int, error) {
	b.once.Do(func() { close(b.entered) })
	<-b.closed
	return 0, io.ErrClosedPipe
}
func (b *blockedStream) Close() error {
	select {
	case <-b.closed:
	default:
		close(b.closed)
	}
	return nil
}
func TestQueuedWriterCancellation(t *testing.T) {
	wire := &blockedStream{entered: make(chan struct{}), closed: make(chan struct{})}
	transport := NewStream(wire)
	peer := newPeer(transport, nil, false)
	defer peer.close(io.EOF)
	defer transport.Close()
	first := make(chan error, 1)
	go func() { first <- peer.send(peer.ctx, message{ID: json.RawMessage(`1`), Result: json.RawMessage(`{}`)}) }()
	<-wire.entered
	ctx, cancel := context.WithCancel(context.Background())
	queued := make(chan error, 1)
	go func() { queued <- peer.call(ctx, "ping", nil, nil) }()
	// Leave the first whole-message write blocked while cancelling its successor.
	time.Sleep(10 * time.Millisecond)
	cancel()
	select {
	case err := <-queued:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("queued send: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancelled send stuck behind live writer")
	}
	transport.Close()
	if err := <-first; err == nil {
		t.Fatal("closed writer succeeded")
	}
}
func TestOutgoingIntegerIDs(t *testing.T) {
	left, right := net.Pipe()
	p := newPeer(NewStream(left), nil, false)
	defer p.close(io.EOF)
	remote := NewStream(right)
	defer remote.Close()
	for want := 1; want <= 2; want++ {
		done := make(chan error, 1)
		go func() { var result any; done <- p.call(context.Background(), "ping", nil, &result) }()
		raw, err := remote.Recv()
		if err != nil {
			t.Fatal(err)
		}
		var request message
		if err = json.Unmarshal(raw, &request); err != nil {
			t.Fatal(err)
		}
		var got int
		if err = json.Unmarshal(request.ID, &got); err != nil || got != want {
			t.Fatalf("outgoing ID %s; want integer %d", request.ID, want)
		}
		reply, _ := json.Marshal(message{JSONRPC: "2.0", ID: request.ID, Result: json.RawMessage(`{}`)})
		if err = remote.Send(reply); err != nil {
			t.Fatal(err)
		}
		if err = <-done; err != nil {
			t.Fatal(err)
		}
	}
}

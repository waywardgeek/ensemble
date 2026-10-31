// gateway-agent is an optional Ensemble application. Gateway owns the network;
// Ensemble owns its real actor and history. Neither library owns the other.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"ensemble/ensemble"
	"ensemble/gateway"
)

type application struct {
	gateway.NopHandler
	root    *ensemble.Ensemble
	agent   ensemble.Agent
	client  *gateway.Client
	ready   chan struct{}
	ctx     context.Context
	replies sync.WaitGroup
}

// OnDeliver admits a distinct queued actor turn before acknowledging delivery.
// The reply goroutine belongs to this application and is joined during shutdown;
// it never mistakes lossy Observer progress for this request's completion.
func (a *application) OnDeliver(e gateway.Envelope) {
	// Submit is admission to the existing actor FIFO, not a second turn loop.
	// It preserves one prompt's reliable result even when another message
	// arrives while the model is working. The stamped content crosses intact.
	result := a.agent.Submit(a.ctx, e.Content)
	a.replies.Add(1)
	go func() {
		defer a.replies.Done()
		select {
		case r := <-result:
			if r.Err != nil {
				a.root.Logf("gateway turn failed: %v", r.Err)
				return
			}
			if e.ResponseMode != gateway.ResponseAsync {
				return
			}
			select {
			case <-a.ready:
			case <-a.ctx.Done():
				return
			}
			if err := a.client.Respond(a.ctx, e, r.Text); err != nil {
				a.root.Logf("gateway response failed: %v", err)
			}
		case <-a.ctx.Done():
		}
	}()
}

// OnError sends transport diagnostics through the actual Ensemble root logger.
func (a *application) OnError(err error) { a.root.Logf("gateway: %v", err) }
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	address := flag.String("gateway", "ws://127.0.0.1:8093/v1/ws", "gateway WebSocket URL")
	id := flag.String("agent", "ensemble", "configured subject identity")
	keyPath := flag.String("key", "ensemble.key", "private identity file")
	public := flag.String("gateway-key", "", "pinned gateway public key")
	groups := flag.String("groups", "agents", "comma-separated accepted sender groups")
	directory := flag.String("data", "gateway-agent-data", "this Agent's persistence directory")
	printKey := flag.Bool("print-key", false, "print this identity's public key and exit")
	flag.Parse()
	key, err := gateway.LoadOrCreateKeypair(*keyPath)
	if err != nil {
		return err
	}
	if *printKey {
		fmt.Println(gateway.EncodePublicKey(key.Public))
		return nil
	}
	peer, err := gateway.ParsePublicKey(*public)
	if err != nil {
		return err
	}
	cfg, err := ensemble.ConfigFromEnv()
	if err != nil {
		return err
	}
	cfg.DataDir = *directory
	root := ensemble.New(os.Stderr)
	agent, err := root.NewAgent(cfg, ensemble.AgentSpec{ID: *id, DataDir: *directory})
	if err != nil {
		return err
	}
	defer agent.Shutdown()
	if err = agent.Load(); err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	app := &application{root: root, agent: agent, ready: make(chan struct{}), ctx: ctx}
	c, err := gateway.Connect(ctx, gateway.ClientConfig{GatewayURL: *address, AgentID: *id, StaticKey: key, GatewayPubKey: peer}, app)
	if err != nil {
		return err
	}
	// Delivery can race Connect returning. Reply workers wait for this one-time
	// publication; the receive callback itself never waits on a client RPC.
	app.client = c
	close(app.ready)
	if err = c.DeclareContactable(ctx, strings.Split(*groups, ",")); err != nil {
		c.Close()
		cancel()
		<-c.Done()
		app.replies.Wait()
		return err
	}
	fmt.Printf("Gateway Agent %s ready\n", *id)
	select {
	case <-ctx.Done():
	case <-c.Done():
	}
	// Cancel model work and RPC waits before joining callback-created workers.
	// Done establishes that no further callback can increment the wait group.
	cancel()
	c.Close()
	<-c.Done()
	app.replies.Wait()
	return errors.Join(agent.Shutdown(), agent.Save())
}

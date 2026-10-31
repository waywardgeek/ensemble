// The gateway command owns the router, its listener, and periodic persistence.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	client "ensemble/gateway"
	router "ensemble/internal/gateway"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	configPath := flag.String("config", "gateway.json", "operator gateway configuration")
	printKey := flag.Bool("print-key", false, "print the gateway public key and exit")
	flag.Parse()
	data, err := os.ReadFile(*configPath)
	if err != nil {
		return err
	}
	var cfg router.Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	if cfg.Gateway.DataDir == "" || cfg.Gateway.Listen == "" {
		return errors.New("gateway.listen and gateway.data_dir are required")
	}
	key, err := client.LoadOrCreateKeypair(filepath.Join(cfg.Gateway.DataDir, "identity.key"))
	if err != nil {
		return err
	}
	// Printing identity helps configuration but does not announce listener
	// readiness. A launcher must also verify this new process remained alive
	// and bound its chosen port before connecting clients.
	fmt.Println(client.EncodePublicKey(key.Public))
	if *printKey {
		return nil
	}
	service, err := router.New(cfg, key, os.Stderr)
	if err != nil {
		return err
	}
	server := &http.Server{Addr: cfg.Gateway.Listen, Handler: service, ReadHeaderTimeout: 10 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	failed := make(chan error, 1)
	go func() { failed <- server.ListenAndServe() }()
	// Periodic snapshots bound the usual SIGKILL loss window without pretending
	// that shutdown hooks run after a kill. A failed save is reported, and the
	// next tick may retry; the final save error remains visible to the caller.
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			// Close the listener before joining upgraded connections; HTTP Shutdown does
			// not own WebSockets. The router then performs its final state snapshot.
			closeErr := server.Close()
			return errors.Join(closeErr, service.Close())
		case err := <-failed:
			return errors.Join(err, service.Close())
		case <-ticker.C:
			if err := service.Save(); err != nil {
				fmt.Fprintln(os.Stderr, "gateway persistence:", err)
			}
		}
	}
}

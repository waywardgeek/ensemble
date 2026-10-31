package ensemble_test

import (
	"bytes"
	"context"
	"ensemble/ensemble"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// The logger is an external boundary. Holding its first lens line leaves the
// real request worker inside observation while the Agent accepts an interrupt.
type cacheGateWriter struct {
	once    sync.Once
	entered chan struct{}
	release chan struct{}
}

// Write holds the first lens diagnostic until release and discards other logs.
func (w *cacheGateWriter) Write(p []byte) (int, error) {
	if strings.Contains(string(p), "cache lens anthropic/") {
		w.once.Do(func() { close(w.entered); <-w.release })
	}
	return len(p), nil
}

// TestCacheLensInterruptedObservation checks interrupt liveness and ordered
// artifacts while the old and replacement request workers overlap.
func TestCacheLensInterruptedObservation(t *testing.T) {
	// Repeat the reviewer's bounded scheduling window under -race. The channels
	// establish interruption before release; the sleep only lets worker two run.
	for n := 0; n < 12; n++ {
		t.Run(fmt.Sprint(n), func(t *testing.T) {
			w := &cacheGateWriter{entered: make(chan struct{}), release: make(chan struct{})}
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":10,"output_tokens":1}}`)
			}))
			defer s.Close()
			a, err := ensemble.New(w).NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: s.URL, DataDir: t.TempDir(), DisableStreaming: true})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Shutdown()
			var release sync.Once
			unblock := func() { release.Do(func() { close(w.release) }) }
			defer unblock()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			first := make(chan error, 1)
			go func() { _, err := a.Ask(ctx, "first"); first <- err }()
			select {
			case <-w.entered:
			case <-ctx.Done():
				t.Fatal("first request never reached lens")
			}
			a.Post(ensemble.Interrupt{})
			select {
			case err := <-first:
				if err == nil {
					t.Fatal("interrupt returned success")
				}
			case <-ctx.Done():
				t.Fatal("interrupt waited for the old worker")
			}
			second := make(chan error, 1)
			go func() { _, err := a.Ask(ctx, "second"); second <- err }()
			time.Sleep(20 * time.Millisecond)
			unblock()
			select {
			case err := <-second:
				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal("new request blocked")
			}
			a.Shutdown()
			current, err := os.ReadFile(filepath.Join(a.Config().DataDir, "cache-current.txt"))
			if err != nil || !bytes.Contains(current, []byte("second")) {
				t.Fatalf("current artifact lost second request: %s, %v", current, err)
			}
			prior, err := os.ReadFile(filepath.Join(a.Config().DataDir, "cache-prior.txt"))
			if err != nil || !bytes.Contains(prior, []byte("first")) || bytes.Contains(prior, []byte("second")) {
				t.Fatalf("prior artifact lost first request: %s, %v", prior, err)
			}
		})
	}
}

// TestCacheLensInterruptedAttribution preserves a delayed cold response's own
// prediction and counts its usage after the replacement turn has completed.
func TestCacheLensInterruptedAttribution(t *testing.T) {
	var logs bytes.Buffer
	var count atomic.Int32
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"content":[{"type":"text","text":"ok"}],"usage":{"input_tokens":%d,"output_tokens":1}}`, 10*count.Add(1))
	}))
	defer s.Close()
	a, err := ensemble.New(&logs).NewAgent(ensemble.Config{Model: "claude-sonnet-5", BaseURL: s.URL, DataDir: t.TempDir(), DisableStreaming: true})
	if err != nil {
		t.Fatal(err)
	}
	defer a.Shutdown()
	entered, release := make(chan struct{}), make(chan struct{})
	var gate, released sync.Once
	unblock := func() { released.Do(func() { close(release) }) }
	defer unblock()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// The old HTTP response already exists. Delay its parsing at the public
	// frame observer until the replacement turn has fully reported its usage.
	first := make(chan error, 1)
	go func() {
		_, err := a.AskWatching(ctx, "first", ensemble.StreamCallbacks{OnFrame: func(_ string, _ []byte) {
			gate.Do(func() { close(entered); <-release })
		}})
		first <- err
	}()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("first frame absent")
	}
	a.Post(ensemble.Interrupt{})
	select {
	case err := <-first:
		if err == nil {
			t.Fatal("expected interrupted turn")
		}
	case <-ctx.Done():
		t.Fatal("interrupt blocked")
	}
	if _, err := a.Ask(ctx, "second"); err != nil {
		t.Fatal(err)
	}
	unblock()
	a.Shutdown()
	found := false
	for _, line := range strings.Split(logs.String(), "\n") {
		if strings.Contains(line, "cache grant: input_tokens=10 ") {
			found = true
			if !strings.Contains(line, "predicted=0.0000") {
				t.Fatalf("old cold response attributed later prediction: %s", line)
			}
		}
	}
	if !found {
		t.Fatal("old response grant was never reported")
	}
	if usage := a.Engine().Usage(); usage.Input != 30 || usage.Output != 2 {
		t.Fatalf("overlap lost or doubled real usage: %+v", usage)
	}
}

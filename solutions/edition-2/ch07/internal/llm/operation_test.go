package llm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/ensemble/internal/common"
)

// This Agent deliberately leaves the readiness notice outstanding. Observing
// the actual store distinguishes a capacity wait from a slow HTTP reader.
type parkedAgent struct{ streamAgent }

func (*parkedAgent) ModelReady(common.ModelOperation) {}

func TestOperationDeadlineIncludesPendingWaits(t *testing.T) {
	for _, tc := range []struct {
		name    string
		chunks  int
		earlier bool
	}{
		{"capacity", 17, false}, {"final-drain", 1, false}, {"earlier-caller-deadline", 1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sent := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				fragment := strings.Repeat("x", 64<<10)
				for i := 0; i < tc.chunks; i++ {
					fmt.Fprintf(w, "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":%q}}]}\n\n", fragment)
				}
				fmt.Fprint(w, "data: {\"choices\":[{\"index\":0,\"delta\":{},\"finish_reason\":\"stop\"}]}\n\ndata: {\"choices\":[],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":1}}\n\ndata: [DONE]\n\n")
				close(sent)
			}))
			defer server.Close()
			e := New(&parkedAgent{})
			defer e.Close()
			e.client.Timeout = time.Second
			op := e.NewOperation("m1", "r1", common.Config{Vendor: "openai", Model: "fixture", BaseURL: server.URL}).(*operation)
			defer op.Discard()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.earlier {
				var earlyCancel context.CancelFunc
				ctx, earlyCancel = context.WithTimeout(ctx, 500*time.Millisecond)
				defer earlyCancel()
			}
			result := make(chan error, 1)
			go func() {
				v, err := op.Exchange(ctx, []byte(`{}`))
				if len(v.Response.Parts) > 0 {
					result <- fmt.Errorf("deadline returned accepted material")
					return
				}
				result <- err
			}()
			want := 64 << 10
			if tc.chunks > 1 {
				want = pendingLimit
			}
			barrier := time.NewTimer(400 * time.Millisecond)
			defer barrier.Stop()
			ticker := time.NewTicker(time.Millisecond)
			defer ticker.Stop()
			for {
				op.mu.Lock()
				n := op.bytes
				op.mu.Unlock()
				if n == want {
					break
				}
				select {
				case err := <-result:
					t.Fatalf("settled before pending barrier: %v", err)
				case <-barrier.C:
					t.Fatalf("pending barrier absent: %d", n)
				case <-ticker.C:
				}
			}
			select {
			case <-sent:
			case <-time.After(time.Second):
				t.Fatal("server did not finish fixture")
			}
			select {
			case err := <-result:
				if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), "timed out") {
					t.Fatalf("wrong timeout: %v", err)
				}
			case <-time.After(2 * time.Second):
				t.Fatal("operation remained parked after deadline")
			}
			if !tc.earlier && ctx.Err() != nil {
				t.Fatal("outer caller was canceled")
			}
		})
	}
}

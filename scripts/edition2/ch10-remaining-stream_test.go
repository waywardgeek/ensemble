package ensemble

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/eventlog"
	"example.com/ensemble/internal/llm"
)

// A real streaming Reader supplies every byte. This does not establish the
// separate regular-file path or its Stat/read checks. No sparse file is used.
type ch10RemainingStream struct {
	prefix                   []byte
	remaining, read, padding int64
}

func (r *ch10RemainingStream) Read(p []byte) (int, error) {
	if r.remaining == 0 {
		return 0, io.EOF
	}
	n := len(p)
	if int64(n) > r.remaining {
		n = int(r.remaining)
	}
	for i := 0; i < n; i++ {
		if len(r.prefix) != 0 {
			p[i], r.prefix = r.prefix[0], r.prefix[1:]
			continue
		}
		p[i] = ' '
		r.padding++
		if r.padding%(1<<20) == 0 {
			p[i] = '\n'
		}
	}
	r.remaining -= int64(n)
	r.read += int64(n)
	return n, nil
}

type ch10RemainingReadOwner struct {
	*Agent
	state    common.Context
	accepted int
}

func (r *ch10RemainingReadOwner) AcceptReadEvent(e common.Event, line int) error {
	if err := llm.Validate(r.engine, r.state, &e); err != nil {
		return err
	}
	llm.Apply(r.engine, &r.state, e)
	r.accepted++
	return nil
}

func TestCh10RemainingStreamingLogBytes(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	path := filepath.Join(a.Config().DataDir, "events.log")
	prefix, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasSuffix(prefix, []byte("\n")) {
		t.Fatal("genuine initializer lacks LF")
	}
	for _, n := range []int64{int64(len(prefix)), 1 << 30, (1 << 30) + 1, (1 << 30) + (64 << 20)} {
		r := &ch10RemainingStream{prefix: prefix, remaining: n}
		owner := &ch10RemainingReadOwner{Agent: a}
		err := eventlog.Stream(owner, r, true)
		if n <= 1<<30 {
			if err != nil || r.read != n || owner.accepted != 1 {
				t.Fatalf("exact streaming parent %d: read=%d accepted=%d error=%v", n, r.read, owner.accepted, err)
			}
		} else {
			if err == nil {
				t.Fatal("streaming log overflow accepted")
			}
			if n == (1<<30)+1 && !strings.Contains(err.Error(), "session log exceeds 1 GiB") {
				t.Fatalf("streaming log one-over did not identify bound: %v", err)
			}
			if r.read > (1<<30)+4096 {
				t.Fatalf("reader consumed beyond bound plus fixed read-ahead: %d", r.read)
			}
		}
	}
}

package ensemble

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"example.com/ensemble/internal/eventlog"
)

// Generate an ordinary valid initializer padded with whitespace. The tail may
// be logically huge, but this reader never materializes it. A read fence stops
// a deliberately unbounded mutant before it can exhaust the review machine.
type ch10AdmissionReader struct {
	body        []byte
	total, sent int64
	beyond      bool
}

func (r *ch10AdmissionReader) Read(p []byte) (int, error) {
	if r.sent == r.total {
		return 0, io.EOF
	}
	if r.sent >= (64<<20)+4096 {
		r.beyond = true
		return 0, errors.New("review read fence reached")
	}
	n := min(int64(len(p)), r.total-r.sent)
	for i := int64(0); i < n; i++ {
		at := r.sent + i
		switch {
		case at < int64(len(r.body)):
			p[i] = r.body[at]
		case at == r.total-1:
			p[i] = '\n'
		default:
			p[i] = ' '
		}
	}
	r.sent += n
	return int(n), nil
}

func TestCh10AdmissionBoundedAllocation(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	raw, err := os.ReadFile(filepath.Join(a.Config().DataDir, "events.log"))
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSuffix(raw, []byte{'\n'}), []byte{'\n'})
	if len(lines) != 2 {
		t.Fatal("genuine initializer parent differs")
	}
	var firstOverflow uint64
	for _, n := range []int64{64 << 20, (64 << 20) + 1, 8 << 30} {
		owner := &ch10RemainingReadOwner{Agent: a}
		stream := &ch10AdmissionReader{body: lines[1], total: n}
		input := io.MultiReader(bytes.NewReader(append(append([]byte{}, lines[0]...), '\n')), stream)
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		err = eventlog.Stream(owner, input, true)
		runtime.ReadMemStats(&after)
		allocated := after.TotalAlloc - before.TotalAlloc
		t.Logf("logical_record_bytes=%d generated_bytes=%d allocated_bytes=%d accepted=%d error=%v", n, stream.sent, allocated, owner.accepted, err)
		if stream.beyond {
			t.Fatal("reader requested beyond admission fence")
		}
		if n == 64<<20 {
			if err != nil || owner.accepted != 1 || stream.sent != n {
				t.Fatalf("exact record parent failed: %v", err)
			}
		} else {
			if err == nil || owner.accepted != 0 || !strings.Contains(err.Error(), "cannot read complete log record") {
				t.Fatalf("oversize was not refused before reduction: %v", err)
			}
			if stream.sent > (64<<20)+4096 {
				t.Fatal("overflow read exceeds fixed read-ahead")
			}
			if firstOverflow == 0 {
				firstOverflow = allocated
			} else if allocated > firstOverflow+(1<<20) {
				t.Fatalf("allocation grows with unread tail: short=%d long=%d", firstOverflow, allocated)
			}
		}
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
}

// Caller-owned input is allocated before measurement. Preparation must stop
// at its unchanged64MiB output cap, not serialize all of an oversized body.
func TestCh10AdmissionBoundedWriteAllocation(t *testing.T) {
	a, _ := ch10RemainingOwner(t)
	small := Event{Seq: 2, Time: "2026-10-08T00:00:00Z", Type: "message_received", Message: &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text("valid parent")}}}
	if _, err := a.log.Prepare(small); err != nil {
		t.Fatal("small prepared record parent refused", err)
	}
	var baseline uint64
	for _, size := range []int{64 << 20, 256 << 20} {
		candidate := small
		candidate.Message = &Entry{Actor: "system", Purpose: "instruction", Parts: []Part{Text(strings.Repeat("x", size))}}
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		_, err := a.log.Prepare(candidate)
		runtime.ReadMemStats(&after)
		allocated := after.TotalAlloc - before.TotalAlloc
		t.Logf("caller_body_bytes=%d preparation_allocated_bytes=%d error=%v", size, allocated, err)
		if err == nil {
			t.Fatal("oversized write preparation accepted")
		}
		ch10RemainingCode(t, err, "session_limit")
		if baseline == 0 {
			baseline = allocated
		} else if allocated > baseline+(1<<20) {
			t.Fatalf("preparation allocation grows with rejected body: short=%d long=%d", baseline, allocated)
		}
	}
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}

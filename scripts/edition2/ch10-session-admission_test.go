package ensemble

import (
	"io"
	"testing"

	"example.com/ensemble/internal/eventlog"
)

func TestCh10ReviewAppendFaultStopsAdmission(t *testing.T) {
	_, a, _ := ch10ReviewOpen(t)
	if err := ch10ReviewAppend(a, "healthy positive parent"); err != nil {
		t.Fatal(err)
	}
	log, ok := a.log.(*eventlog.Log)
	if !ok {
		t.Fatal("review adapter no longer matches real log")
	}
	var partial *ch10ReviewPartialWriter
	log.Ch10ReviewWrapWriter(func(real io.WriteCloser) io.WriteCloser {
		partial = &ch10ReviewPartialWriter{WriteCloser: real}
		return partial
	})
	if err := ch10ReviewAppend(a, "must fail to persist"); err == nil || partial.calls != 1 {
		t.Fatalf("real partial-write parent not established: %v", err)
	}
	if _, err := a.Submit("must refuse before new request admission"); err == nil {
		t.Fatal("terminal append failure left request admission open")
	}
	if err := a.Close(); err == nil {
		t.Fatal("terminal append failure disappeared at close")
	}
}

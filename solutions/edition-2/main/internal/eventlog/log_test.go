package eventlog

import (
	"errors"
	"example.com/ensemble/internal/common"
	"testing"
)

type owner struct{ writes int }

func (o *owner) Ensemble() common.Ensemble    { return o }
func (o *owner) Config() common.Config        { return common.Config{} }
func (o *owner) Workspace() string            { return "" }
func (o *owner) Logf(string, ...any)          {}
func (o *owner) Publish(string, common.Event) {}
func (*owner) AllocateHandle() uint64         { return 0 }
func (o *owner) Write(b []byte) (int, error) {
	o.writes++
	return len(b) / 2, errors.New("simulated partial write")
}
func (o *owner) Close() error { return nil }
func TestPartialWritePermanentlyFaultsLog(t *testing.T) {
	parent := &owner{}
	log := &Log{parent: parent, writer: parent}
	if err := log.Append(common.Event{}); err == nil {
		t.Fatal("partial write accepted")
	}
	if err := log.Append(common.Event{}); err == nil || parent.writes != 1 {
		t.Fatal("appended after partial record")
	}
}

func (*owner) Observe(common.Observation)                       {}
func (*owner) Collect([]common.RequestHandle) common.Collection { return nil }

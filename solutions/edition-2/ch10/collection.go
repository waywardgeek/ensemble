package ensemble

import (
	"context"
	"example.com/ensemble/internal/common"
	"reflect"
	"sync"
)

type collection struct {
	parent   common.Ensemble
	mu       sync.Mutex
	handles  []RequestHandle
	returned []bool
}

func (e *Ensemble) Collect(handles []RequestHandle) Collection {
	owned := append([]RequestHandle(nil), handles...)
	return &collection{parent: e, handles: owned, returned: make([]bool, len(owned))}
}
func (c *collection) Ensemble() common.Ensemble { return c.parent }
func (c *collection) Wait(ctx context.Context) ([]Completion, bool, error) {
	// A collection serializes its own drains. Handle completions are immutable and
	// closing a Done channel wakes every independent collection and waiter.
	for {
		c.mu.Lock()
		out := []Completion{}
		remaining := 0
		cases := []reflect.SelectCase{{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ctx.Done())}}
		for i, h := range c.handles {
			if c.returned[i] {
				continue
			}
			select {
			case <-h.Done():
				v, err := h.Wait(context.Background())
				if err != nil {
					c.mu.Unlock()
					return nil, false, err
				}
				out = append(out, v)
				c.returned[i] = true
			default:
				remaining++
				cases = append(cases, reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(h.Done())})
			}
		}
		c.mu.Unlock()
		if len(out) > 0 || remaining == 0 {
			return out, remaining == 0, nil
		}
		index, _, _ := reflect.Select(cases)
		if index == 0 {
			return nil, false, ctx.Err()
		}
	}
}

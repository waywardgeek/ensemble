package llm

import (
	"example.com/ensemble/internal/common"
	"math"
	"reflect"
)

func checkedUsage(a, b common.Usage) (common.Usage, bool) {
	x := []int64{a.Input, a.CacheWrite, a.CacheRead, a.Output}
	y := []int64{b.Input, b.CacheWrite, b.CacheRead, b.Output}
	for i := range x {
		if x[i] < 0 || y[i] < 0 || y[i] > math.MaxInt64-x[i] {
			return common.Usage{}, false
		}
		x[i] += y[i]
	}
	return common.Usage{Input: x[0], CacheWrite: x[1], CacheRead: x[2], Output: x[3]}, true
}
func (e *Engine) ValidateAccount(response common.Response) error {
	if response.Usage == nil {
		return &common.SessionError{Code: "session_corrupt", Detail: "missing usage"}
	}
	if _, ok := checkedUsage(e.Usage(), *response.Usage); !ok {
		return &common.SessionError{Code: "session_limit", Detail: "usage accounting overflow"}
	}
	return nil
}
func (e *Engine) RestoreUsage(accounts []common.UsageAccount, responses []common.ResponseFact) error {
	bad := func() error { return &common.SessionError{Code: "session_corrupt", Detail: "invalid semantic usage"} }
	want := map[common.Provenance]common.Usage{}
	var total common.Usage
	for _, r := range responses {
		u, ok := checkedUsage(want[r.From], r.Usage)
		if !ok {
			return bad()
		}
		want[r.From] = u
		total, ok = checkedUsage(total, r.Usage)
		if !ok {
			return bad()
		}
	}
	got := map[common.Provenance]common.Usage{}
	previous := ""
	for _, a := range accounts {
		key := a.From.Vendor + "/" + a.From.Model + "/" + a.From.Surface
		if key <= previous || !validProvenance(e, &a.From) {
			return bad()
		}
		previous = key
		got[a.From] = a.Usage
	}
	if !reflect.DeepEqual(want, got) {
		return bad()
	}
	e.usageMu.Lock()
	e.usage = got
	e.usageMu.Unlock()
	return nil
}

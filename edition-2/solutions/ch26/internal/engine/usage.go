package engine

import "ensemble/internal/common"

// Keep model identity and prompt tier with the counts that incurred them.
// A later display must not reprice large-prompt work at the ordinary rate.
type usageKey struct {
	vendor common.Vendor
	model  string
	large  bool
}

func addUsage(total *common.Usage, u common.Usage) {
	total.Input += u.Input
	total.CacheWrite += u.CacheWrite
	total.CacheRead += u.CacheRead
	total.Output += u.Output
}

func (e *engine) account(cfg common.Config, u common.Usage) {
	e.mu.Lock()
	defer e.mu.Unlock()
	features, _ := LookupModel(cfg.Model)
	key := usageKey{cfg.Vendor, cfg.Model, features.PriceThreshold > 0 && u.Input+u.CacheWrite+u.CacheRead > features.PriceThreshold}
	e.addBucket(key, u)
}

// addBucket is called with the Engine lock held. It does not touch last-response
// occupancy; both ordinary accounting and drained auxiliary import share totals.
func (e *engine) addBucket(key usageKey, u common.Usage) {
	addUsage(&e.usage, u)
	if e.byModel == nil {
		e.byModel = make(map[usageKey]common.Usage)
	}
	subtotal := e.byModel[key]
	addUsage(&subtotal, u)
	e.byModel[key] = subtotal
}

// UsageBuckets returns value copies so callers cannot mutate accounting through
// a receipt. Call after worker shutdown/drain to include late measured failures.
func (e *engine) UsageBuckets() []common.UsageBucket {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]common.UsageBucket, 0, len(e.byModel))
	for key, usage := range e.byModel {
		out = append(out, common.UsageBucket{Vendor: key.vendor, Model: key.model, Large: key.large, Usage: usage})
	}
	return out
}

// AddAuxiliaryUsage adds original model/tier buckets without repricing them at
// the parent's model or treating their aggregate as one large prompt. The actor's
// active attempt supplies once-only ownership; no unbounded dedup set lives here.
func (e *engine) AddAuxiliaryUsage(buckets []common.UsageBucket) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, bucket := range buckets {
		e.addBucket(usageKey{bucket.Vendor, bucket.Model, bucket.Large}, bucket.Usage)
	}
}

// Meter derives prices without persisting money or restoring yesterday's spend.
// Unknown contributing prices invalidate the full cost rather than showing a
// misleading partial sum. Token counts stay useful even when dollars are unknown.
func (e *engine) Meter() common.UsageMeter {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := common.UsageMeter{Usage: e.last, Total: e.usage, Priced: true}
	for key, u := range e.byModel {
		features, _ := LookupModel(key.model)
		price := features.Price
		if key.large {
			price = features.LargePrice
		}
		if price == nil {
			out.Priced = false
			continue
		}
		out.CostUSD += (float64(u.Input)*price.Input + float64(u.CacheWrite)*price.CacheWrite + float64(u.CacheRead)*price.CacheRead + float64(u.Output)*price.Output) / 1e6
	}
	// Before any request, an unknown selected model must still display a dash.
	if len(e.byModel) == 0 {
		out.Priced = e.Features().Price != nil
	}
	if !out.Priced {
		out.CostUSD = 0
	}
	input := out.Total.Input + out.Total.CacheWrite + out.Total.CacheRead
	if input > 0 {
		out.HitRate = float64(out.Total.CacheRead) / float64(input)
	}
	return out
}

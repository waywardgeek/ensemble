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
	addUsage(&e.usage, u)
	features, _ := LookupModel(cfg.Model)
	key := usageKey{cfg.Vendor, cfg.Model, features.PriceThreshold > 0 && u.Input+u.CacheWrite+u.CacheRead > features.PriceThreshold}
	if e.byModel == nil {
		e.byModel = make(map[usageKey]common.Usage)
	}
	subtotal := e.byModel[key]
	addUsage(&subtotal, u)
	e.byModel[key] = subtotal
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

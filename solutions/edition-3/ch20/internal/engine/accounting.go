package engine

import "ensemble/internal/common"

func responseCost(u common.Usage, prices common.TokenPrices, supported bool) common.Cost {
	if !supported {
		return common.Cost{Status: "unknown"}
	}
	counts := []int64{u.Input, u.Output, u.CacheRead, u.CacheWrite}
	rates := []float64{prices.Input, prices.Output, prices.CacheRead, prices.CacheWrite}
	known := 0
	cost := common.Cost{Status: "known"}
	for i, n := range counts {
		if n < 0 {
			continue
		}
		known++
		cost.USD += float64(n) * rates[i] / 1e6
	}
	if known == 0 {
		cost.Status = "unknown"
	} else if known < len(counts) {
		cost.Status = "partial"
	}
	return cost
}
func addCost(a, b common.Cost) common.Cost {
	if a.Status == "" {
		return b
	}
	status := "partial"
	if a.Status == "known" && b.Status == "known" {
		status = "known"
	}
	if a.Status == "unknown" && b.Status == "unknown" {
		status = "unknown"
	}
	return common.Cost{USD: a.USD + b.USD, Status: status}
}
func (e *engine) account(cfg common.Config, r *common.ResponseData) {
	// Returned identity wins. An absent producer can use only explicit resolved
	// request evidence, never the active model at a later UI refresh.
	model := r.From.Model
	if model == "" {
		model = cfg.ResolvedModel
	}
	prices, ok := common.LookupPrices(cfg.Vendor, cfg.Surface, model)
	cost := responseCost(r.Usage, prices, ok)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.lastModel = model
	e.last = r.Usage
	e.requests++
	e.liveUsage.Add(r.Usage)
	index := -1
	for i, b := range e.buckets {
		if b.Vendor == cfg.Vendor && b.Surface == cfg.Surface && b.Model == model {
			index = i
			break
		}
	}
	if index < 0 {
		index = len(e.buckets)
		e.buckets = append(e.buckets, common.UsageBucket{Vendor: cfg.Vendor, Surface: cfg.Surface, Model: model})
	}
	b := &e.buckets[index]
	b.Requests++
	b.Usage.Add(r.Usage)
	b.Cost = addCost(b.Cost, cost)
}
func (e *engine) Status() common.EngineStatus {
	e.mu.Lock()
	s := common.EngineStatus{LastModel: e.lastModel, Last: e.last, Session: e.liveUsage, Requests: e.requests, Buckets: append([]common.UsageBucket{}, e.buckets...)}
	for _, b := range e.buckets {
		s.Cost = addCost(s.Cost, b.Cost)
	}
	if s.Cost.Status == "" {
		s.Cost.Status = "known"
	}
	e.mu.Unlock()
	s.SelectedModel = e.parent.Config().Model
	s.LastHitRate = common.CacheHitRate(s.Last)
	s.SessionHitRate = common.CacheHitRate(s.Session)
	if e.cache != nil {
		s.Cache = e.cache.Status()
	}
	return s
}

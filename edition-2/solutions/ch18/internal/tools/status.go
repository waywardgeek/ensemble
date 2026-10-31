package tools

import (
	"encoding/json"
	"fmt"

	"ensemble/internal/common"
)

// Status is a view, not another owner of usage. The Engine already derives
// session cost from counts at each model's rate; formatting must not reprice it.
func agentStatus(call common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct{}
	if err := decode(raw, &args); err != nil {
		return "", err
	}
	engine := call.Engine()
	meter := engine.Meter()
	cost := "unknown (one or more model prices unavailable)"
	if meter.Priced {
		cost = fmt.Sprintf("$%.6f", meter.CostUSD)
	}
	return fmt.Sprintf("model: %s\nlast response: %d in, %d out, %d cache read, %d cache write\nsession total: %d in, %d out, %d cache read, %d cache write\ncache hit rate: %.2f%%\ncost: %s",
		engine.Agent().Config().Model,
		meter.Input, meter.Output, meter.CacheRead, meter.CacheWrite,
		meter.Total.Input, meter.Total.Output, meter.Total.CacheRead, meter.Total.CacheWrite,
		meter.HitRate*100, cost), nil
}

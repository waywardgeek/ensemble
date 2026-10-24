package common

// Cache is an Engine child. Preview rendering never advances it.
type Cache interface {
	Parent() Engine
	Observe(Config, []byte) CachePrediction
	Report(CachePrediction, Usage)
	Status() CacheStatus
}

type RequestSnapshot struct {
	RequestID   uint64           `json:"request_id"`
	Route       string           `json:"route"`
	PromptBytes int              `json:"prompt_bytes"`
	Digest      string           `json:"digest"`
	Bytes       int              `json:"bytes"`
	Segments    []RequestSegment `json:"segments"`
}
type RequestSegment struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Digest string `json:"digest"`
	Bytes  int    `json:"bytes"`
	Marked bool   `json:"marked"`
}
type CachePrediction struct {
	RequestID   uint64   `json:"request_id"`
	SharedBytes int      `json:"shared_bytes"`
	PromptBytes int      `json:"prompt_bytes"`
	Fraction    *float64 `json:"fraction"`
	Reason      string   `json:"reason"`
}
type CacheComparison struct {
	Prediction    CachePrediction `json:"prediction"`
	ActualHitRate *float64        `json:"actual_hit_rate"`
}
type CacheSection struct {
	Name   string `json:"name"`
	Change string `json:"change"`
	Bytes  int    `json:"bytes"`
}
type CacheStatus struct {
	Prediction      CachePrediction  `json:"prediction"`
	Comparison      *CacheComparison `json:"comparison,omitempty"`
	Sections        []CacheSection   `json:"sections"`
	OptionsChange   string           `json:"options_change"`
	DiagnosticError string           `json:"diagnostic_error,omitempty"`

	Current         *RequestSnapshot `json:"current,omitempty"`
	Previous        *RequestSnapshot `json:"previous,omitempty"`
	SharedSegments  int              `json:"shared_segments"`
	SharedBytes     int              `json:"shared_bytes"`
	FirstDifference string           `json:"first_difference,omitempty"`
	MarkerCount     int              `json:"marker_count"`
	MarkerReason    string           `json:"marker_reason"`
}

// Cost separates attributable charges from unknown work. A partial total is a
// known subtotal, never a claim that the remaining work cost nothing.
type Cost struct {
	USD    float64 `json:"usd"`
	Status string  `json:"status"` // known, partial, unknown
}
type UsageBucket struct {
	Vendor   Vendor  `json:"vendor"`
	Surface  Surface `json:"surface"`
	Model    string  `json:"model"`
	Requests int     `json:"requests"`
	Usage    Usage   `json:"usage"`
	Cost     Cost    `json:"cost"`
}
type EngineStatus struct {
	SelectedModel  string        `json:"selected_model"`
	LastModel      string        `json:"last_model,omitempty"`
	Last           Usage         `json:"last"`
	Session        Usage         `json:"session"`
	Requests       int           `json:"requests"`
	LastHitRate    *float64      `json:"last_hit_rate"`
	SessionHitRate *float64      `json:"session_hit_rate"`
	Cost           Cost          `json:"cost"`
	Buckets        []UsageBucket `json:"buckets"`
	Cache          CacheStatus   `json:"cache"`
}

// Source: https://docs.anthropic.com/en/docs/build-with-claude/prompt-caching
// Rates are USD per million tokens, for the default five-minute Anthropic
// write policy. No alias guessing, tier extrapolation, or invented zero prices.
type TokenPrices struct{ Input, Output, CacheRead, CacheWrite float64 }

func LookupPrices(v Vendor, s Surface, model string) (TokenPrices, bool) {
	if v == Anthropic && s == Messages {
		switch model {
		case "claude-sonnet-4-6":
			return TokenPrices{3, 15, .3, 3.75}, true
		case "claude-opus-4-6":
			return TokenPrices{5, 25, .5, 6.25}, true
		}
	}
	return TokenPrices{}, false
}

// CacheBreakpoints is deliberately model-and-API gated. Course-only synthetic
// declaration features do not imply support for a real provider cache API.
func CacheBreakpoints(c Config) (bool, string) {
	model := c.ResolvedModel
	if model == "" {
		model = c.Model
	}
	switch {
	case c.Vendor == OpenAI && c.Surface == Responses:
		if c.AuthMode == AuthChatGPT {
			return false, "ChatGPT plan inference omits explicit cache markers"
		}
		if model == "gpt-ch19-course" {
			return true, "explicit Responses content boundaries: system instructions and retained input; tool and opaque items are not cache markers"
		}
	case c.Vendor == Anthropic && c.Surface == Messages:
		if model == "claude-sonnet-4-6" || model == "claude-opus-4-6" {
			return true, "up to four legal boundaries: last tool, SYSTEM, retained history end, surviving handoff; unavailable or protected boundaries are omitted"
		}
	case c.Vendor == OpenAI && c.Surface == ChatCompletions:
		if model == "gpt-5.6" {
			return true, "explicit message-content boundaries: SYSTEM, retained history end, surviving handoff; tools cannot carry markers; unavailable or protected boundaries are omitted"
		}
	}
	return false, "no explicit breakpoint policy for this model/API; provider implicit behavior retained"
}

// -1 is unreported, never a measured zero. Arithmetic propagates unknown facts.
func UnknownUsage() Usage { return Usage{Input: -1, Output: -1, CacheRead: -1, CacheWrite: -1} }
func (u *Usage) Add(v Usage) {
	add := func(a *int64, b int64) {
		if *a < 0 || b < 0 || b > (1<<63-1)-*a {
			*a = -1
		} else {
			*a += b
		}
	}
	add(&u.Input, v.Input)
	add(&u.Output, v.Output)
	add(&u.CacheRead, v.CacheRead)
	add(&u.CacheWrite, v.CacheWrite)
}
func CacheHitRate(u Usage) *float64 {
	if u.Input < 0 || u.CacheRead < 0 || u.CacheWrite < 0 {
		return nil
	}
	total := float64(u.Input) + float64(u.CacheRead) + float64(u.CacheWrite)
	if total == 0 {
		return nil
	}
	rate := float64(u.CacheRead) / total
	return &rate
}

package jobs

import (
	"encoding/json"
	"example.com/ensemble/internal/common"
	"fmt"
	"math"
	"regexp"
	"time"
)

// Resolve takes pending state even for an unknown or malformed attempted call.
func (s *Service) Resolve(call common.Part) (common.Limits, string, error) {
	s.mu.Lock()
	pending := s.pending
	s.pending = nil
	s.mu.Unlock()
	limits := common.Limits{Delay: 3 * time.Second, MaxBytes: 16384}
	if pending != nil {
		overlay(s, &limits, *pending)
	}
	var err error
	if call.Name == "run_command" || call.Name == "wait_for_job" || call.Name == "send_input" || call.Name == "tool_limits" {
		var supplied common.LimitOverrides
		supplied, err = overrides(s, call.Args)
		if err == nil {
			overlay(s, &limits, supplied)
		}
	}
	note := ""
	if pending != nil {
		if err != nil {
			note = fmt.Sprintf("tool_limits consumed by %s: validation refused: %v\n", call.Name, err)
		} else {
			pattern := ""
			if limits.Pattern != nil {
				pattern = limits.Pattern.String()
			}
			note = fmt.Sprintf("tool_limits consumed by %s: ai_callback_delay=%g ai_callback_pattern=%q max_output_bytes=%d\n", call.Name, limits.Delay.Seconds(), pattern, limits.MaxBytes)
		}
	}
	return limits, note, err
}
func overlay(s *Service, limits *common.Limits, value common.LimitOverrides) {
	if value.Delay != nil {
		limits.Delay = *value.Delay
	}
	if value.MaxBytes != nil {
		limits.MaxBytes = *value.MaxBytes
	}
	if value.Pattern != nil {
		limits.Pattern = nil
		if *value.Pattern != "" {
			limits.Pattern, _ = regexp.Compile(*value.Pattern)
		}
	}
}
func overrides(s *Service, raw json.RawMessage) (common.LimitOverrides, error) {
	var fields map[string]json.RawMessage
	out := common.LimitOverrides{}
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return out, fmt.Errorf("arguments must be an object")
	}
	for key, value := range fields {
		switch key {
		case "ai_callback_delay":
			var seconds float64
			if string(value) == "null" || json.Unmarshal(value, &seconds) != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) || seconds < 0 || seconds >= float64(math.MaxInt64)/float64(time.Second) {
				return out, fmt.Errorf("ai_callback_delay must be finite nonnegative seconds within duration range")
			}
			duration := time.Duration(seconds * float64(time.Second))
			out.Delay = &duration
		case "ai_callback_pattern":
			var pattern string
			if string(value) == "null" || json.Unmarshal(value, &pattern) != nil {
				return out, fmt.Errorf("ai_callback_pattern must be a Go regular expression string")
			}
			if _, err := regexp.Compile(pattern); err != nil {
				return out, fmt.Errorf("ai_callback_pattern: invalid Go regular expression")
			}
			out.Pattern = &pattern
		case "max_output_bytes":
			var budget int
			if json.Unmarshal(value, &budget) != nil || budget <= 0 {
				return out, fmt.Errorf("max_output_bytes must be a positive integer")
			}
			out.MaxBytes = &budget
		}
	}
	return out, nil
}

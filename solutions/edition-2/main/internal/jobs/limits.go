package jobs

import (
	"example.com/ensemble/internal/common"
	"time"
)

// Resolve atomically spends pending settings on every admitted attempt. The
// tool boundary supplies only validated overrides, or none after a refusal.
func (s *Service) Resolve(explicit common.LimitOverrides) (common.Limits, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pending := s.pending
	s.pending = nil
	limits := common.Limits{Delay: 3 * time.Second, MaxBytes: 16384}
	if pending != nil {
		overlayValues(s, &limits, *pending)
	}
	overlay(s, &limits, explicit)
	return limits, pending != nil
}
func (s *Service) SetLimits(value common.LimitOverrides) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := common.LimitValues{MaxBytes: value.MaxBytes}
	if value.Delay != nil {
		v := value.Delay.Seconds()
		next.Delay = &v
	}
	if value.PatternSet {
		v := ""
		if value.Pattern != nil {
			v = value.Pattern.String()
		}
		next.Pattern = &v
	}
	s.pending = &next
}
func overlay(s *Service, limits *common.Limits, value common.LimitOverrides) {
	if value.Delay != nil {
		limits.Delay = *value.Delay
	}
	if value.MaxBytes != nil {
		limits.MaxBytes = *value.MaxBytes
	}
	if value.PatternSet {
		limits.Pattern = value.Pattern
	}
}

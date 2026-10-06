package common

// CacheLens observes outbound requests and what the provider said about them,
// and reports on prompt-cache behavior.
//
// It exists because a cache miss is silent. Every other class of defect in this
// agent announces itself: a bad tool call errors, a broken parse throws, a
// wedged job hangs where you can see it. A cache miss returns the correct
// answer, slightly later, for roughly ten times the money, and writes nothing
// anywhere. The only channel that reports it is the invoice, thirty days later,
// aggregated across every session so it cannot be attributed to a change. A
// failure mode with no natural signal needs a manufactured one.
//
// The seam is told-never-asked, like Observer: the engine hands over what it
// just did and never consults the lens for a decision. Nothing the lens
// concludes may change the request, so a broken lens cannot corrupt a
// conversation — it can only be wrong about the diagnosis.
type CacheLens interface {
	// ObserveRequest receives the exact bytes of an outbound request, before it
	// is sent. Model rather than vendor, because the cacheable prefix structure
	// and the size floor below which nothing is cached are both properties of
	// the model.
	ObserveRequest(model string, body []byte)

	// ObserveUsage receives what the provider reported for the request most
	// recently passed to ObserveRequest. Split from ObserveRequest because the
	// counts do not exist until the response comes back, and on a failed turn
	// they never arrive at all.
	ObserveUsage(u Usage)
}

// NopCacheLens is a CacheLens that discards everything.
//
// It exists so the engine never has to test for nil. A nil check at each call
// site is a branch that gets forgotten on the fifth site added later; a default
// implementation cannot be forgotten.
type NopCacheLens struct{}

func (NopCacheLens) ObserveRequest(string, []byte) {}
func (NopCacheLens) ObserveUsage(Usage)            {}

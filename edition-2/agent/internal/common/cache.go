package common

// CacheLens compares main-conversation requests, never judge/compressor prompts.
// Engine supplies sanitized wire bytes and normalized usage at the send boundary.
type CacheLens interface {
	// Engine exposes the required owner, including its Agent's data directory.
	Engine() Engine
	// ObserveRequest records a request and returns its predicted cached-byte ratio.
	ObserveRequest(name string, body []byte) float64
	// ObserveUsage compares granted tokens with that request's returned prediction.
	ObserveUsage(usage Usage, predicted float64)
}

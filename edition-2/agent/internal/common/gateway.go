package common

// GatewayOwner is the parent interface retained by an edge connection. It uses
// serialized frames so the headless core does not import WebSocket or Noise code.
// Policy and durable queue state stay at the standalone server root.
type GatewayOwner interface {
	// Logf reports connection diagnostics through the actual routing root.
	Logf(format string, args ...any)
	// NextGatewayDelivery snapshots the next queued frame after this connection's cursor.
	NextGatewayDelivery(id string, after uint64) (seq uint64, frame []byte, ok bool)
}

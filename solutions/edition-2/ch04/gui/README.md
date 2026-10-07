# Optional GUI client stub

This module uses only Ensemble's public ClientOwner interface. A Client submits
Agent-attributed requests, receives ordered observations, and unregisters when
closed. Diagnostics reach Ensemble through that same owning interface.

A future WebSocket server belongs in this module and translates browser traffic
at this boundary. No WebSocket server, browser UI, or live browser roundtrip is
implemented or claimed here. The core library does not depend on this module.

Run `go test ./... -count=1` here for the fake-backed public integration check.

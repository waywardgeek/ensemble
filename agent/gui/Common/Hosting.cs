namespace EnsembleGui;

// Value-like launch options, not an owner or another settings store. Relative
// state paths are resolved by the borrowed Agent's Sandbox. Null asset override
// selects assembly resources: deployment does not require a source checkout.
// RunAsync owns a loopback listener; ServeAsync requires the caller to bind
// loopback at Port. BasePath is a literal route prefix, not a route template.
// Account controls follow it, but registration reserves /auth/callback at root.
// The trusted GUI can select existing addresses in the borrowed Agent's Ensemble.
// Selection is per socket and does not transfer ownership. The optional agent_id
// page/socket query chooses a fixed browser relay target for that connection;
// select_agent changes only its conversation view. The caller must keep the
// Ensemble alive until ServeAsync has drained; peers may close independently.
// PublicOrigin optionally admits one HTTPS origin behind an authenticated local
// reverse proxy. It is not authentication; the proxy must gate every route,
// including WebSockets. Forwarded headers do not grant host/origin authority.
// Trusted peer POST controls use this mounted Agent as the explicit defaults
// source and cannot delete it. Pending creation does not block socket traffic.
// RendererDirectory explicitly enables shared renderer snapshots; null keeps only
// shipped rendering. Trials are never persisted in this directory.
// Null TracePath uses Agent-owned diagnostics. An explicit path is a fresh,
// operation-owned aggregate trace for embedding/tests, not the normal host mode.
public sealed record GuiOptions(int Port = 8090, string? AssetsDirectory = null,
    string SettingsPath = "settings.json", string? TracePath = null, string BasePath = "",
    string? PublicOrigin = null, string? RendererDirectory = null)
{
    // Presentation only: neither durable History nor model context is trimmed.
    public int HistoryEventLimit { get; init; } = 100;
}

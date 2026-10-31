namespace EnsembleGui;

// Value-like launch options, not an owner or another settings store. Relative
// state paths are resolved by the borrowed Agent's Sandbox. Null asset override
// selects assembly resources: deployment does not require a source checkout.
// RunAsync owns a loopback listener; ServeAsync requires the caller to bind
// loopback at Port. BasePath is a literal route prefix, not a route template.
// Account controls follow it, but registration reserves /auth/callback at root.
public sealed record GuiOptions(int Port = 8090, string? AssetsDirectory = null,
    string SettingsPath = "settings.json", string TracePath = "gui.log", string BasePath = "");

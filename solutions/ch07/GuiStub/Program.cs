using Ensemble;

// Construction-only consumer: no GUI framework, provider call, or second transcript.
// A later GUI will use these same public owners rather than a GUI-specific engine.
var application = new Ensemble.Ensemble(Console.Error);
try
{
    _ = new Agent(application, new Configuration(Vendor.Anthropic, "offline", baseUrl: "http://127.0.0.1:1", apiKey: ""));
}
finally { application.Destroy(); }

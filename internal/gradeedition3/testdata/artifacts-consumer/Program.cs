// Grader-owned outside-assembly consumer. No candidate test assembly, reflection,
// source project reference, real provider or private credentials are used.
using System.Security.Cryptography;
using System.Text.Json;
using Ensemble;
using EnsembleGui;

if (args is ["browser-host"])
{
    await ArtifactHost.RunAsync();
    return 0;
}
if (args is not ["work-product-core"])
{
    Console.Error.WriteLine("Expected work-product-core mode.");
    return 2;
}
var checks = new List<string>();
void Require(bool condition, string name)
{
    if (!condition) throw new InvalidOperationException(name);
    checks.Add(name);
}
void Refuse(Action action, string name)
{
    try { action(); }
    catch (Exception error) when (error is ArgumentException or IOException or InvalidOperationException
        or KeyNotFoundException or UnauthorizedAccessException)
    {
        checks.Add(name);
        return;
    }
    throw new InvalidOperationException(name);
}
string workspace = Path.Combine(Directory.GetCurrentDirectory(), "workspace");
Directory.CreateDirectory(workspace);
string skills = Path.Combine(workspace, "skills");
Directory.CreateDirectory(Path.Combine(skills, "root"));
Directory.CreateDirectory(Path.Combine(skills, "products"));
File.WriteAllText(Path.Combine(skills, "root", "SKILL.md"),
    "---\nname: root\ntype: primary\ndescription: Independent artifact exercise.\ntools: load_skill unload_skill wait_for_job\nloadable-skills: products\n---\nUse only explicitly granted tools.\n");
File.WriteAllText(Path.Combine(skills, "products", "SKILL.md"),
    "---\nname: products\ntype: loadable\ndescription: Deliberate work-product offering.\ntools: offer_work_product\n---\nOffer a requested file; the human decides whether to download.\n");
using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(45));
var app = new Ensemble.Ensemble(TextWriter.Null);
try
{
    var config = new Configuration(Vendor.Anthropic, "artifact-" + Guid.NewGuid().ToString("N"),
        "http://127.0.0.1:1", sandboxRoot: workspace);
    var agent = new Agent(app, config, name: "candidate", skillSpec: new SkillSpec(skills, "root"));
    int sequence = 0;
    Task<ToolResultPart> Call(string name, object values) => agent.DispatchAsync(new ToolCallPart(
        "grade-" + ++sequence, config.Target(), name, JsonSerializer.SerializeToElement(values)), deadline.Token);
    string filename = "report-" + Guid.NewGuid().ToString("N") + ".bin";
    string file = Path.Combine(workspace, filename);
    byte[] initial = RandomNumberGenerator.GetBytes(8193);
    File.WriteAllBytes(file, initial);
    Require((await Call("offer_work_product", new { path = file })).IsError, "ungranted-offer-refused");
    Require(!(await Call("load_skill", new { name = "products" })).IsError, "explicit-grant-loaded");
    var result = await Call("offer_work_product", new { path = file, ai_callback_delay = 0 });
    while (result.Job?.Status == "running")
    {
        Require(result.WorkProduct is null, "running-report-has-no-offer");
        result = await Call("wait_for_job", new { handle = result.Job.Handle, ai_callback_delay = 1 });
    }
    Require(!result.IsError && result.WorkProduct is not null, "successful-result-has-typed-offer");
    var offer = result.WorkProduct!;
    Require(offer.Name == filename && offer.OfferedBytes == initial.Length, "offer-description-matches-file");
    byte[] current = RandomNumberGenerator.GetBytes(16387);
    File.WriteAllBytes(file, current);
    using var download = new MemoryStream();
    long copied = await agent.CopyWorkProductAsync(offer.ID, download, deadline.Token);
    Require(copied == current.Length && download.ToArray().SequenceEqual(current), "exact-click-time-bytes");
    Require(download.CanWrite, "download-destination-remains-borrowed");
    var foreign = new Agent(app, config, name: "foreign");
    Refuse(() => foreign.OpenWorkProduct(offer.ID).Dispose(), "foreign-owner-refused");
    Refuse(() => agent.OpenWorkProduct(file).Dispose(), "pathname-is-not-offer-id");
    var canonical = agent.History().Events().Last(fact => fact.WorkProduct is not null);
    Require(JsonCodec.ReadEvent(JsonCodec.WriteEvent(canonical)).WorkProduct?.ID == offer.ID, "event-roundtrip-retains-description");
    var context = new Context([new Entry(1, Actor.Tool, [result])]);
    var savedConfig = JsonSerializer.SerializeToElement(new { vendor = "anthropic", model = "offline", system_prompt = "", tools = Array.Empty<object>() });
    var saved = JsonCodec.ReadSave(JsonCodec.WriteSave(new SaveDocument(context, 1, savedConfig, [])));
    var frames = Gui.Project(new HistorySnapshot([], saved.Context!, 1), "candidate").Frames;
    Require(frames.Single(frame => frame["type"]!.GetValue<string>() == "tool_finished")
        ["work_product"]?["id"]?.GetValue<string>() == offer.ID, "snapshot-only-gui-retains-description");
    using (var stream = File.OpenWrite(file)) stream.SetLength(Agent.MaximumWorkProductBytes + 1);
    Refuse(() => agent.OpenWorkProduct(offer.ID).Dispose(), "grown-file-over-limit-refused");
    File.WriteAllBytes(file, current);
    Require(!(await Call("unload_skill", new { name = "products" })).IsError, "grant-unloaded");
    Require((await Call("offer_work_product", new { path = file })).IsError, "revoked-new-offer-refused");
    agent.Destroy();
    var replacement = new Agent(app, config, name: "candidate");
    Refuse(() => replacement.OpenWorkProduct(offer.ID).Dispose(), "same-name-replacement-does-not-revive-offer");
    Console.WriteLine(JsonSerializer.Serialize(new { contract = "artifact-work-product-core-v1", passed = true, checks,
        ungraded = "HTTP/browser, timing, snippets, interaction, and chapter points remain separate." }));
    return 0;
}
catch (Exception error)
{
    Console.WriteLine(JsonSerializer.Serialize(new { contract = "artifact-work-product-core-v1", passed = false,
        checks, failure = error.Message }));
    return 1;
}
finally
{
    app.Destroy();
}

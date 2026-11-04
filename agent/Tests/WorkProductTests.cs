using Ensemble;
using System.Text.Json;
using System.Text.Json.Nodes;
using Application = Ensemble.Ensemble;

internal static class WorkProductTests
{
    private static void Check(bool value, string message)
    {
        if (!value) throw new Exception(message);
    }

    private static void Refuses(Action action)
    {
        try { action(); }
        catch (Exception error) when (error is IOException or ArgumentException
            or InvalidOperationException or KeyNotFoundException or UnauthorizedAccessException)
        {
            return;
        }
        throw new Exception("Expected work-product refusal");
    }

    // A slow external destination, not a replacement for Agent/Sandbox ownership.
    private sealed class BlockingDestination : Stream
    {
        public readonly TaskCompletionSource Entered = new(TaskCreationOptions.RunContinuationsAsynchronously);
        public readonly TaskCompletionSource Cancelled = new(TaskCreationOptions.RunContinuationsAsynchronously);
        public readonly TaskCompletionSource Release = new(TaskCreationOptions.RunContinuationsAsynchronously);
        public bool Disposed;
        public override async ValueTask WriteAsync(ReadOnlyMemory<byte> buffer, CancellationToken cancellationToken = default)
        {
            Entered.TrySetResult();
            try { await Task.Delay(Timeout.Infinite, cancellationToken); }
            catch (OperationCanceledException)
            {
                Cancelled.TrySetResult();
                await Release.Task;
                throw;
            }
        }
        protected override void Dispose(bool disposing) { Disposed = true; base.Dispose(disposing); }
        public override bool CanRead => false;
        public override bool CanSeek => false;
        public override bool CanWrite => true;
        public override long Length => throw new NotSupportedException();
        public override long Position { get => throw new NotSupportedException(); set => throw new NotSupportedException(); }
        public override void Flush() { }
        public override int Read(byte[] buffer, int offset, int count) => throw new NotSupportedException();
        public override long Seek(long offset, SeekOrigin origin) => throw new NotSupportedException();
        public override void SetLength(long value) => throw new NotSupportedException();
        public override void Write(byte[] buffer, int offset, int count) => throw new NotSupportedException();
    }

    public static async Task TransferLifetime()
    {
        string root = Path.Combine(Path.GetTempPath(), "work-product-transfer-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(root);
        var app = new Application(TextWriter.Null);
        using var caller = new CancellationTokenSource();
        using var cancelledSink = new BlockingDestination();
        using var destroyedSink = new BlockingDestination();
        try
        {
            var agent = new Agent(app, new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", sandboxRoot: root));
            string path = Path.Combine(root, "report.bin");
            byte[] bytes = Enumerable.Range(0, 4096).Select(i => (byte)i).ToArray();
            File.WriteAllBytes(path, bytes);
            var offer = agent.OfferWorkProduct(path);
            var cancelled = agent.CopyWorkProductAsync(offer.ID, cancelledSink, caller.Token);
            await cancelledSink.Entered.Task.WaitAsync(TimeSpan.FromSeconds(5));
            // A blocked destination does not hold the owner gate or another copy.
            using var complete = new MemoryStream();
            Check(await agent.CopyWorkProductAsync(offer.ID, complete) == bytes.Length
                && complete.ToArray().SequenceEqual(bytes), "Independent complete copy was blocked or corrupted");
            caller.Cancel();
            await cancelledSink.Cancelled.Task.WaitAsync(TimeSpan.FromSeconds(5));
            Check(!cancelled.IsCompleted, "Copy returned before cancellation cleanup completed");
            cancelledSink.Release.TrySetResult();
            try { await cancelled; throw new Exception("Cancelled copy reported success"); }
            catch (OperationCanceledException) { }
            Check(!cancelledSink.Disposed, "Copy disposed borrowed response stream");
            var active = agent.CopyWorkProductAsync(offer.ID, destroyedSink);
            await destroyedSink.Entered.Task.WaitAsync(TimeSpan.FromSeconds(5));
            var destroy = Task.Run(agent.Destroy);
            await destroyedSink.Cancelled.Task.WaitAsync(TimeSpan.FromSeconds(5));
            Check(!destroy.IsCompleted, "Agent destruction did not join active transfer cleanup");
            Refuses(() => agent.OpenWorkProduct(offer.ID).Dispose());
            destroyedSink.Release.TrySetResult();
            try { await active; throw new Exception("Destroyed Agent's copy reported success"); }
            catch (OperationCanceledException) { }
            await destroy.WaitAsync(TimeSpan.FromSeconds(5));
            Check(!destroyedSink.Disposed, "Agent disposed caller's destination");
        }
        finally
        {
            cancelledSink.Release.TrySetResult();
            destroyedSink.Release.TrySetResult();
            caller.Cancel();
            app.Destroy();
            Directory.Delete(root, true);
        }
    }

    public static async Task ToolAndReplay()
    {
        string root = Path.Combine(Path.GetTempPath(), "work-product-tool-" + Guid.NewGuid().ToString("N"));
        string skills = Path.Combine(root, "skills");
        Directory.CreateDirectory(Path.Combine(skills, "root"));
        Directory.CreateDirectory(Path.Combine(skills, "work-products"));
        File.WriteAllText(Path.Combine(skills, "root", "SKILL.md"),
            "---\nname: root\ntype: primary\ndescription: Work-product test.\ntools: load_skill unload_skill wait_for_job\nloadable-skills: work-products\n---\nIsolated test.\n");
        File.Copy(Path.Combine(Environment.CurrentDirectory, "agent", "skills", "work-products", "SKILL.md"),
            Path.Combine(skills, "work-products", "SKILL.md"));
        string path = Path.Combine(root, "report.bin");
        File.WriteAllBytes(path, [10, 20, 30]);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", sandboxRoot: root);
            var legacy = new Agent(app, config);
            legacy.InstallTools();
            Check(!legacy.ToolEnabled("offer_work_product"), "Legacy registry granted work-product authority");
            Check(!legacy.Engine().Render(new Context()).ToJsonString().Contains("offer_work_product"),
                "Legacy request advertised work-product authority");
            var agent = new Agent(app, config, skillSpec: new SkillSpec(skills, "root"));
            int serial = 0;
            Task<ToolResultPart> Call(string name, object args) => agent.DispatchAsync(new ToolCallPart(
                "product-" + ++serial, config.Target(), name, JsonSerializer.SerializeToElement(args)));
            Check((await Call("offer_work_product", new { path })).IsError, "Unloaded skill offered a file");
            Check(!(await Call("load_skill", new { name = "work-products" })).IsError, "Work-product skill did not load");
            Check(agent.ToolEnabled("offer_work_product"), "Explicit offer grant missing");
            var result = await Call("offer_work_product", new { path, ai_callback_delay = 0 });
            for (int attempt = 0; result.Job?.Status == "running" && attempt < 5; attempt++)
            {
                Check(result.WorkProduct is null, "Running observation published an offer");
                result = await Call("wait_for_job", new { handle = result.Job.Handle, ai_callback_delay = 1 });
            }
            Check(!result.IsError && result.WorkProduct is not null, "Completed offer lost typed metadata");
            var offer = result.WorkProduct!;
            using (var input = agent.OpenWorkProduct(offer.ID))
                Check(input.ReadByte() == 10, "Successful offer did not authorize its own file");
            var fact = agent.History().Events().Last(value => value.WorkProduct is not null);
            Check(JsonCodec.ReadEvent(JsonCodec.WriteEvent(fact)).WorkProduct?.ID == offer.ID,
                "Canonical event codec lost offer");
            Check(fact.Numbered(7, DateTimeOffset.UtcNow).WorkProduct?.ID == offer.ID, "Numbering lost offer");
            using (var partJson = JsonDocument.Parse(JsonCodec.WritePart(result).ToJsonString()))
                Check(((ToolResultPart)JsonCodec.ReadPart(partJson.RootElement)).WorkProduct?.ID == offer.ID,
                    "Part codec lost offer");
            var call = new ToolCallPart(result.CallID, config.Target(), "offer_work_product", JsonSerializer.SerializeToElement(new { path }));
            var context = new Context([new Entry(1, Actor.Agent, [call]), new Entry(2, Actor.Tool, [result])]);
            var savedConfig = JsonSerializer.SerializeToElement(new { model = "fixture", vendor = "anthropic", system_prompt = "", tools = Array.Empty<object>() });
            var restored = JsonCodec.ReadSave(JsonCodec.WriteSave(new SaveDocument(context, 2, savedConfig, []))).Context!;
            Check(((ToolResultPart)restored.Dialogue[1].Parts[0]).WorkProduct?.ID == offer.ID, "Snapshot lost offer description");
            string provider = agent.Engine().Render(restored).ToJsonString();
            static bool HasOfferMetadata(JsonElement value) => value.ValueKind switch
            {
                JsonValueKind.Object => value.EnumerateObject().Any(property =>
                    property.NameEquals("work_product") || HasOfferMetadata(property.Value)),
                JsonValueKind.Array => value.EnumerateArray().Any(HasOfferMetadata),
                _ => false
            };
            using (var request = JsonDocument.Parse(provider))
                Check(!provider.Contains(offer.ID) && !HasOfferMetadata(request.RootElement),
                    "Provider received live offer metadata");
            var replacement = new Agent(app, config);
            Refuses(() => replacement.OpenWorkProduct(offer.ID).Dispose());
            Refuses(() => new ToolResultPart("error", [], isError: true, workProduct: offer));
            Refuses(() => new Event(EventKind.ToolStarted, callID: "bad", name: "offer_work_product",
                args: JsonSerializer.SerializeToElement(new { path }), workProduct: offer));
            Refuses(() => new WorkProductInfo("not-an-id", "report.bin", 1));
            Refuses(() => new WorkProductInfo(offer.ID, "../report.bin", 1));
            Check(!(await Call("unload_skill", new { name = "work-products" })).IsError, "Skill unload failed");
            Check((await Call("offer_work_product", new { path })).IsError, "Revoked skill created another offer");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(root, true);
        }
    }

    public static void OwnershipAndBounds()
    {
        string root = Path.Combine(Path.GetTempPath(), "work-products-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(root);
        var app = new Application(TextWriter.Null);
        try
        {
            var config = new Configuration(Vendor.Anthropic, "fixture", "http://127.0.0.1:1", sandboxRoot: root);
            var agent = new Agent(app, config, name: "products");
            string path = Path.Combine(root, "report.bin");
            File.WriteAllBytes(path, [1, 2, 3]);
            Refuses(() => agent.OfferWorkProduct(Path.Combine(root, "missing.bin")));
            var info = agent.OfferWorkProduct(path);
            Check(info.Name == "report.bin" && info.OfferedBytes == 3, "Offer metadata is wrong");
            Check(!info.ID.Contains(root) && info.ID.Length == 32, "Offer identifier exposed path");
            File.WriteAllBytes(path, [4, 5, 6, 7]);
            using (var input = agent.OpenWorkProduct(info.ID))
            {
                using var copy = new MemoryStream();
                input.CopyTo(copy);
                Check(copy.ToArray().SequenceEqual(new byte[] { 4, 5, 6, 7 }), "Offer froze old bytes");
            }
            string outside = Path.Combine(Path.GetDirectoryName(root)!, "outside-product-" + Guid.NewGuid().ToString("N"));
            File.WriteAllBytes(outside, [99]);
            try
            {
                File.Delete(path);
                File.CreateSymbolicLink(path, outside);
                Refuses(() => agent.OpenWorkProduct(info.ID).Dispose());
            }
            finally
            {
                File.Delete(path);
                File.Delete(outside);
                File.WriteAllBytes(path, [4, 5, 6, 7]);
            }
            var other = new Agent(app, config, name: "other-products");
            Refuses(() => other.OpenWorkProduct(info.ID).Dispose());
            Refuses(() => agent.OpenWorkProduct(path).Dispose());
            using (var file = File.OpenWrite(path)) file.SetLength(Agent.MaximumWorkProductBytes);
            using (var input = agent.OpenWorkProduct(info.ID))
                Check(input.Length == Agent.MaximumWorkProductBytes, "Exact 64 MiB boundary refused");
            using (var file = File.OpenWrite(path)) file.SetLength(Agent.MaximumWorkProductBytes + 1);
            Refuses(() => agent.OpenWorkProduct(info.ID).Dispose());
            Refuses(() => agent.OfferWorkProduct(path));
            File.WriteAllBytes(path, [8]);
            for (int index = 1; index < Agent.MaximumWorkProductOffers; index++)
                _ = agent.OfferWorkProduct(path);
            Refuses(() => agent.OfferWorkProduct(path));
            using (var input = agent.OpenWorkProduct(info.ID))
                Check(input.ReadByte() == 8, "Offer cap silently evicted earlier grant");
            agent.Destroy();
            Refuses(() => agent.OpenWorkProduct(info.ID).Dispose());
            var replacement = new Agent(app, config, name: "products");
            Refuses(() => replacement.OpenWorkProduct(info.ID).Dispose());
            Check(replacement.OfferWorkProduct(path).ID != info.ID, "Replacement reused prior authority");
        }
        finally
        {
            app.Destroy();
            Directory.Delete(root, true);
        }
    }
}

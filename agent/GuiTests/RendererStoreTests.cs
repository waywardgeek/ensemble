using System.Collections.Immutable;
using EnsembleGui;

internal static class RendererStoreTests
{
    private static void Check(bool value, string message)
    {
        if (!value) throw new InvalidOperationException(message);
    }

    private static void Refuses(Action action, string message)
    {
        try { action(); }
        catch (Exception error) when (error is ArgumentException or IOException or InvalidDataException or InvalidOperationException) { return; }
        throw new Exception(message);
    }

    public static void Run()
    {
        string directory = Directory.CreateTempSubdirectory("ensemble-renderer-store-").FullName;
        try
        {
            var store = new RendererStore(directory);
            var files = ImmutableSortedDictionary.Create<string, string>(StringComparer.Ordinal)
                .Add("renderer.js", "import {label} from './helper.js'; export default () => label;")
                .Add("helper.js", "export const label = 'first';");
            var package = new RendererPackage("list_directory", "renderer.js", files);
            string first = store.Stage(package);
            Check(store.Stage(package) == first, "Identical snapshot must retain identity.");
            Check(store.Selection("list_directory").Accepted is null, "Staging activated code.");
            Check(store.Read(first).Files["helper.js"] == files["helper.js"], "Snapshot lost helper source.");
            var accepted = store.Accept("list_directory", first, null);
            Check(accepted.Accepted == first && accepted.Previous is null, "First acceptance incorrect.");
            string second = store.Stage(package with { Files = files.SetItem("helper.js", "export const label = 'second';") });
            Check(second != first, "Helper edits did not change reviewed version.");
            Check(new RendererStore(directory).Selection("list_directory").Accepted == first,
                "Reopened store resumed a staged candidate instead of accepted code.");
            Refuses(() => store.Accept("list_directory", second, null), "Stale acceptance succeeded.");
            Check(store.Selection("list_directory").Accepted == first, "Stale acceptance changed selection.");
            using (var lease = new FileStream(Path.Combine(directory, "selection.lock"), FileMode.Open, FileAccess.ReadWrite, FileShare.None))
                Refuses(() => store.Accept("list_directory", second, first), "Locked selection accepted.");
            Check(store.Selection("list_directory").Accepted == first, "Failed persistence changed selection.");
            accepted = store.Accept("list_directory", second, first);
            Check(accepted.Accepted == second && accepted.Previous == first, "Previous accepted version lost.");
            accepted = new RendererStore(directory).Rollback("list_directory", second);
            Check(accepted.Accepted == first && accepted.Previous == second, "Rollback failed after reopening.");
            string other = store.Stage(package with { Renderer = "read_file" });
            Refuses(() => store.Accept("list_directory", other, first), "Cross-renderer promotion succeeded.");
            Refuses(() => store.Stage(package with { Entry = "missing.js" }), "Absent entry accepted.");
            Refuses(() => store.Stage(package with { Files = files.Add("../outside.js", "") }), "Traversal accepted.");
            Refuses(() => store.Stage(package with { Files = files.SetItem("helper.js", new string('x', 1024 * 1024)) }), "Oversized package accepted.");
            File.AppendAllText(Path.Combine(directory, "versions", second + ".json"), " ");
            Refuses(() => store.Read(second), "Tampered snapshot read as approved code.");
            Refuses(() => store.Accept("list_directory", second, first), "Tampered snapshot promoted.");
            Check(store.Selection("list_directory").Accepted == first, "Tampered candidate changed accepted version.");
            File.WriteAllText(Path.Combine(directory, "read_file.json"), "{\"Accepted\":null,\"Accepted\":null}");
            Refuses(() => store.Selection("read_file"), "Duplicate selection fields accepted.");
            var fresh = new RendererStore(Path.Combine(directory, "fresh"));
            string freshVersion = fresh.Stage(package);
            fresh.Accept("list_directory", freshVersion, null);
            Check(fresh.Rollback("list_directory", freshVersion).Accepted is null, "Rollback to shipped default failed.");
            Console.WriteLine("PASS renderer store: snapshots, dependency identity, acceptance, rollback, stale/locked/tampered refusal");
        }
        finally { Directory.Delete(directory, recursive: true); }
    }
}

using System.Text.Json;
using Ensemble;
using Application = Ensemble.Ensemble;

internal static partial class ActorTests
{
    public static async Task RootedJobFiles()
    {
        string workspace = Path.Combine(Path.GetTempPath(), "ensemble-job-root-" + Guid.NewGuid().ToString("N"));
        string inside = Path.Combine(workspace, "inside");
        string outside = Path.Combine(workspace, "outside");
        Directory.CreateDirectory(inside);
        Directory.CreateDirectory(outside);
        string sentinel = Path.Combine(outside, "secret.txt");
        File.WriteAllText(sentinel, "outside-output-secret");
        File.WriteAllText(Path.Combine(inside, "input.txt"), "rooted-job-positive");
        var app = new Application(TextWriter.Null);
        try
        {
            var configuration = new Configuration(Vendor.OpenAI, "gpt-4.1", "http://127.0.0.1:1",
                sandboxRoot: inside);
            Refuse(() => _ = new Agent(app, configuration, outside));
            Check(app.Agents().Length == 0, "Refused Jobs construction retained an Agent");
            Check(Directory.GetFileSystemEntries(outside).Length == 1, "Refused construction changed outside files");

            var agent = new Agent(app, configuration);
            agent.InstallTools();
            Check(agent.Jobs().OutputDirectory.StartsWith(agent.Sandbox().Root() + Path.DirectorySeparatorChar,
                StringComparison.Ordinal), "Default output directory escaped the root");
            ToolResultPart result = await Read();
            Check(!result.IsError && JsonCodec.Text(result.Parts).Contains("rooted-job-positive"),
                "Rooted job lost its normal result");
            var job = agent.Jobs().Children().Single();
            string output = job.OutputPath;
            Check(File.ReadAllText(output).Contains("rooted-job-positive"), "Job did not retain output");

            // The first observation consumed the normal bytes. A replacement
            // pathname must never supply bytes to a later observer, even when
            // that observer would otherwise read only a suffix.
            string saved = output + ".saved";
            File.Move(output, saved);
            File.CreateSymbolicLink(output, sentinel);
            try
            {
                await job.ObserveAsync("outside-observation", delay: 0);
                throw new Exception("Job observation followed an outside output symlink");
            }
            catch (IOException) { }
            File.Delete(output);
            File.Move(saved, output);

            // The next job must not create its output through a swapped parent.
            string directory = agent.Jobs().OutputDirectory;
            Directory.Move(directory, directory + ".saved");
            Directory.CreateSymbolicLink(directory, outside);
            try
            {
                await Read();
                throw new Exception("Job creation followed an outside directory symlink");
            }
            catch (IOException) { }
            Directory.Delete(directory);
            Directory.Move(directory + ".saved", directory);
            Check(agent.Jobs().Children().Length == 1, "Refused output creation attached a Job");
            Check(File.ReadAllText(sentinel) == "outside-output-secret"
                && Directory.GetFileSystemEntries(outside).Length == 1, "Job output touched outside files");
            agent.Destroy();
            Check(File.ReadAllText(output).Contains("rooted-job-positive"), "Destruction deleted retained output");
            Check(app.Agents().Length == 0, "Job owner did not detach");

            var relative = new Agent(app, configuration, "relative-jobs");
            Check(relative.Jobs().OutputDirectory == Path.Combine(relative.Sandbox().Root(), "relative-jobs"),
                "Explicit relative Jobs directory did not resolve against the root");
            relative.Destroy();

            Task<ToolResultPart> Read() => agent.DispatchAsync(new ToolCallPart(Guid.NewGuid().ToString("N"),
                configuration.Target(), "read_file", JsonSerializer.SerializeToElement(new { path = "input.txt" })));
        }
        finally
        {
            app.Destroy();
            Directory.Delete(workspace, recursive: true);
        }

        static void Refuse(Action action)
        {
            try
            {
                action();
            }
            catch (IOException)
            {
                return;
            }
            throw new Exception("Outside output directory was accepted");
        }
    }
}

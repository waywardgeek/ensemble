using Ensemble;
using Application = Ensemble.Ensemble;

internal static class AgentTreeTests
{
    public static async Task Membership(string directory)
    {
        var app = new Application(TextWriter.Null);
        var config = new Configuration(Vendor.Anthropic, "tree-fixture", "http://127.0.0.1:1");
        string output = Path.Combine(directory, "tree-membership");
        try
        {
            Check(app.AgentTree().IsEmpty, "Empty Ensemble has a fabricated Agent.");
            var main = new Agent(app, config, outputDirectory: output, name: "main");
            var peer = new Agent(app, config, outputDirectory: output, name: "peer");
            var unnamed = new Agent(app, config, outputDirectory: output);
            var otherUnnamed = new Agent(app, config, outputDirectory: output);
            var left = new Agent(app, main, "helper", config, outputDirectory: output);
            var right = new Agent(app, peer, "helper", config, outputDirectory: output);
            var nested = new Agent(app, left, "helper", config, outputDirectory: output);
            var captured = app.AgentTree();
            Check(captured.Select(node => node.Agent).SequenceEqual([main, peer, unnamed, otherUnnamed]),
                "Standalone membership was lost or descendants were duplicated as roots.");
            Check(captured[0].Children.Single().Agent == left
                && captured[0].Children.Single().Children.Single().Agent == nested
                && captured[1].Children.Single().Agent == right,
                "Repeated local IDs collapsed distinct ownership paths.");
            Check(captured[2].Children.IsEmpty && captured[3].Children.IsEmpty,
                "Unnamed standalone Agents acquired fabricated descendants.");

            // Shutdown joins work but keeps ownership; destruction detaches it.
            // A captured value must not silently turn into a live relationship list.
            await left.ShutdownAsync();
            Check(app.AgentTree()[0].Children.Single().Agent == left,
                "Shutdown alone removed a retained child from the tree.");
            left.Destroy();
            var after = app.AgentTree();
            Check(after[0].Children.IsEmpty && after[1].Children.Single().Agent == right,
                "Targeted destruction damaged another ownership branch.");
            Check(captured[0].Children.Single().Agent == left
                && captured[0].Children.Single().Children.Single().Agent == nested,
                "A retained tree snapshot changed after detachment.");
            Check(left.AgentParent() is null && nested.AgentParent() is null,
                "Snapshot references kept a logically destroyed relationship alive.");
            app.Destroy();
            Check(app.AgentTree().IsEmpty, "Destroyed Ensemble retains tree membership.");
        }
        finally
        {
            app.Destroy();
        }
    }

    public static async Task ConcurrentMembership(string directory)
    {
        var app = new Application(TextWriter.Null);
        var config = new Configuration(Vendor.Anthropic, "tree-fixture", "http://127.0.0.1:1");
        string output = Path.Combine(directory, "tree-concurrent");
        var root = new Agent(app, config, outputDirectory: output);
        using var deadline = new CancellationTokenSource(TimeSpan.FromSeconds(15));
        var start = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        Task mutation = Task.Run(async () =>
        {
            await start.Task;
            for (int index = 0; index < 40; index++)
            {
                var child = new Agent(app, root, "helper", config, outputDirectory: output);
                _ = new Agent(app, child, "nested", config, outputDirectory: output);
                await child.ShutdownAsync();
                child.Destroy();
            }
        });
        try
        {
            start.SetResult();
            for (int index = 0; index < 400; index++)
            {
                var tree = app.AgentTree();
                Check(tree.Length == 1 && tree[0].Agent == root,
                    "A constructing or detaching child became a standalone root.");
                foreach (AgentTreeNode child in tree[0].Children)
                {
                    Check(child.Agent.ChildID() == "helper" && child.Children.Length <= 1,
                        "Captured direct membership is malformed.");
                    foreach (AgentTreeNode nested in child.Children)
                    {
                        Check(nested.Agent.ChildID() == "nested" && nested.Children.IsEmpty,
                            "Captured nested membership is malformed.");
                    }
                }
                await Task.Yield();
            }
            await mutation.WaitAsync(deadline.Token);
            Check(app.AgentTree().Single().Children.IsEmpty, "Joined mutations left a child behind.");
        }
        finally
        {
            // A failed assertion must still join the owned mutation before
            // destroying the common parent; no background cleanup escapes.
            await mutation;
            app.Destroy();
        }
    }

    private static void Check(bool condition, string message)
    {
        if (!condition)
        {
            throw new InvalidOperationException(message);
        }
    }
}

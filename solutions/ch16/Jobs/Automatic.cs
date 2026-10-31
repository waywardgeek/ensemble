using System.Text.Json;

namespace Ensemble;

public sealed partial class Jobs
{
    internal partial Job StartAutomatic(AutomaticTool automatic)
    {
        var agent = Parent();
        if (automatic.Client.Parent().Parent() != agent || automatic.Definition.Ephemeral == "")
        {
            throw new InvalidOperationException("Automatic job requires this Agent's automatic definition.");
        }
        Job job;
        ToolCallPart call;
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            int handle = agent.Parent().AllocateJobHandle();
            call = new ToolCallPart("automatic:" + handle, agent.Configuration().Target(),
                automatic.Definition.Name, JsonSerializer.SerializeToElement(new { }));
            job = new Job(this, handle, call);
        }
        try
        {
            // These are host execution facts, not invented assistant dialogue.
            // Persist permission before starting effects, just as ordinary Jobs do.
            agent.History().Append(new Event(EventKind.ToolCalled, callID: call.CallID,
                name: call.Name, args: call.Args, job: job.Snapshot()));
            agent.History().Append(new Event(EventKind.ToolStarted, callID: call.CallID,
                name: call.Name, args: call.Args, job: job.Snapshot()));
        }
        catch
        {
            job.Destroy();
            throw;
        }
        job.Start(null, call, null, automatic.Client);
        return job;
    }
}

public sealed partial class Job
{
    internal partial async Task WaitForCompletionAsync(CancellationToken cancellation)
    {
        Task work;
        lock (gate)
        {
            work = worker ?? throw new InvalidOperationException("Job has not started.");
        }
        // Cancellation detaches only this observer, never the independent job.
        await work.WaitAsync(cancellation);
    }

    internal partial async Task<ToolResultPart> CompletionAsync(CancellationToken cancellation)
    {
        await WaitForCompletionAsync(cancellation);
        lock (gate)
        {
            var info = Snapshot();
            if (structuredResult is { } completed)
            {
                return new ToolResultPart(call.CallID, completed.Parts,
                    completed.IsError, info, completed.ToolReferences);
            }
            // Expected protocol/execution errors use the normal Job output path.
            // Completion has closed the writer, and this read changes no cursor.
            return new ToolResultPart(call.CallID, [new TextPart(File.ReadAllText(OutputPath))],
                isError || status == JobStatus.Killed, info);
        }
    }
}

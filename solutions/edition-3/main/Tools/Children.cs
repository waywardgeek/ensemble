using System.Text.Json;

namespace Ensemble;

public sealed partial class Tool
{
    private partial async Task<string> ChildOperationAsync(Job job, JsonElement args,
        CancellationToken cancellation)
    {
        Agent owner = job.Parent().Parent();
        try
        {
            switch (kind)
            {
                case ToolKind.AgentsStatus:
                    // Capture membership once. Each child's actor supplies its
                    // own coherent snapshot; this is not a global time barrier.
                    var states = new List<object>();
                    foreach (Agent owned in owner.Children())
                    {
                        AgentWorkStatus status = await owned.WorkStatusAsync(cancellation);
                        states.Add(ChildStatusValue(owned, status));
                    }
                    return JsonSerializer.Serialize(new { agents = states });
                case ToolKind.SpawnSubAgent:
                    string task = String(args, "task");
                    if (string.IsNullOrWhiteSpace(task))
                    {
                        throw new ArgumentException("A child task must not be empty.");
                    }
                    string? model = args.TryGetProperty("model", out JsonElement chosenModel)
                        ? chosenModel.GetString() : null;
                    JsonElement? schema = args.TryGetProperty("submit_schema", out JsonElement chosenSchema)
                        ? chosenSchema : null;
                    Agent spawned = await owner.CreateChildAsync(String(args, "agent_id"), model, schema, cancellation);
                    // Creation owns rollback until it returns a ready child.
                    // Thereafter the parent owns the child independently of this
                    // Job. Keep its reliable task reply here; never use a detached
                    // runner or let cancellation of this observer kill the task.
                    Task<string> completion = spawned.SubmitAsync([new TextPart(task)]);
                    job.AppendAttention(JsonSerializer.Serialize(new
                    {
                        agent_id = spawned.ChildID(), task_admitted = true
                    }) + "\n");
                    await completion.WaitAsync(cancellation);
                    if (spawned.ChildFailure() is { } failure)
                    {
                        throw new IOException(failure);
                    }
                    JsonElement? submission = spawned.Submission();
                    return JsonSerializer.Serialize(new
                    {
                        agent_id = spawned.ChildID(), task_completed = true,
                        submitted = submission.HasValue, payload = submission
                    });
                case ToolKind.Submit:
                    JsonElement payload = owner.PublishSubmission(args);
                    return JsonSerializer.Serialize(new { submitted = true, payload });
                case ToolKind.SendMessageToParent:
                    QuestionReply reply = await owner.SendMessageToParentAsync(String(args, "content"), cancellation);
                    return JsonSerializer.Serialize(new { sender = reply.Sender, response = reply.Response });
                case ToolKind.RespondToParentMessage:
                    long? messageID = args.TryGetProperty("message_id", out JsonElement supplied)
                        ? supplied.GetInt64() : null;
                    owner.RespondToQuestion(String(args, "response"), String(args, "agent_id"), messageID);
                    return "Question answered.";
                case ToolKind.InterruptAgent:
                    string interruptID = String(args, "agent_id");
                    string mode = args.TryGetProperty("mode", out JsonElement requestedMode)
                        ? requestedMode.GetString()! : "now";
                    InputKind interrupt = mode switch
                    {
                        "now" => InputKind.InterruptNow,
                        "after_tool" => InputKind.InterruptAfterTool,
                        _ => throw new ArgumentException("mode must be now or after_tool.")
                    };
                    Agent interrupted = owner.FindChild(interruptID)
                        ?? throw new ArgumentException("No owned child has that ID.");
                    // The child's mailbox, not this Job, owns the admitted control.
                    // Its acknowledgment proves delivery only. Independent Jobs
                    // and queued turns retain their existing actor semantics.
                    await interrupted.PostAsync(new AgentInput(interrupt)).WaitAsync(cancellation);
                    return JsonSerializer.Serialize(new
                    {
                        agent_id = interruptID, mode, interrupt_delivered = true
                    });
                case ToolKind.ShutdownAgent:
                    string id = String(args, "agent_id");
                    Agent child = owner.FindChild(id)
                        ?? throw new ArgumentException("No owned child has that ID.");
                    // Shutdown closes admission and joins descendants. Destroy
                    // observes that same task again, collects any failure and
                    // attempts every detach. Do not let a failed shutdown skip
                    // destruction or pretend the selected child was removed.
                    await child.ShutdownAsync().ConfigureAwait(ConfigureAwaitOptions.SuppressThrowing);
                    child.Destroy();
                    return JsonSerializer.Serialize(new { agent_id = id, stopped = true });
                default:
                    throw new InvalidOperationException("Not a child workflow operation.");
            }
        }
        catch (OperationCanceledException error) when (!cancellation.IsCancellationRequested
            && kind is ToolKind.SendMessageToParent or ToolKind.SpawnSubAgent)
        {
            // The question or child task can end before this observing Job is
            // cancelled (inbox teardown deliberately precedes Jobs teardown).
            // Report that independent lifetime ending as a tool error. Only
            // cancellation of this Job belongs to Job's cancellation path.
            string reason = kind == ToolKind.SendMessageToParent
                ? "Question closed before a reply." : "Child task ended before completion.";
            throw new IOException(reason, error);
        }
        catch (Exception error) when (error is InvalidOperationException or AggregateException)
        {
            // Lifecycle refusals and collected cleanup failures are actionable
            // tool feedback. Cancellation still follows the ordinary Job path.
            throw new IOException(error.Message, error);
        }
    }

    private static partial async Task<string> JoinAgentsAsync(Job job, JsonElement args,
        CancellationToken cancellation)
    {
        string[]? ids = null;
        if (args.TryGetProperty("agent_ids", out JsonElement supplied))
        {
            if (supplied.EnumerateArray().Any(value => value.ValueKind != JsonValueKind.String))
            {
                throw new ArgumentException("agent_ids must contain only strings.");
            }
            ids = supplied.EnumerateArray().Select(value => value.GetString()!).ToArray();
        }
        try
        {
            await foreach (ChildJoinUpdate update in job.Parent().Parent().JoinChildrenAsync(ids, cancellation))
            {
                if (update.Submissions is { } submissions)
                {
                    return JsonSerializer.Serialize(new { submissions });
                }
                foreach (QuestionInfo question in update.Questions)
                {
                    // Durable actionable output, not completion. The running
                    // handle stays owned by Jobs while the parent answers.
                    job.AppendAttention(JsonSerializer.Serialize(new
                    {
                        pending_question = new
                        {
                            agent_id = question.AgentID,
                            message_id = question.MessageID,
                            content = question.Content
                        }
                    }) + "\n");
                }
            }
            throw new InvalidOperationException("Join ended without a result.");
        }
        catch (InvalidOperationException error)
        {
            // A failed/stopped selected child is normal tool feedback, not an
            // unobserved worker fault or a fabricated successful submission.
            throw new IOException(error.Message, error);
        }
    }
}

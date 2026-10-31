using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class Jobs
{
    public partial Jobs(Agent parent, string? outputDirectory)
    {
        OutputDirectory = Path.GetFullPath(outputDirectory ?? Path.Combine(Path.GetTempPath(), "ensemble-jobs-" + Guid.NewGuid().ToString("N")));
        Directory.CreateDirectory(OutputDirectory);
        parent.AddJobs(this);
    }

    public partial Agent Parent() => parent ?? throw new ObjectDisposedException(nameof(Jobs));
    public partial ImmutableArray<Job> Children()
    {
        lock (gate)
        {
            return children.ToImmutableArray();
        }
    }

    internal partial void AddJob(Job child)
    {
        lock (gate)
        {
            if (destroyed || child.destroyed || child.parent is not null || children.Any(j => j.Handle == child.Handle))
            {
                throw new InvalidOperationException("Invalid job attachment");
            }
            children.Add(child);
            child.parent = this;
        }
    }

    internal partial void RemoveJob(Job child)
    {
        lock (gate)
        {
            if (child.parent != this || !children.Contains(child))
            {
                throw new InvalidOperationException("Job belongs to another owner");
            }
            children.Remove(child);
            child.parent = null;
        }
    }

    public partial Job Find(int handle)
    {
        lock (gate)
        {
            return children.SingleOrDefault(j => j.Handle == handle)
                ?? throw new ArgumentException($"Unknown job handle: {handle}");
        }
    }

    public partial void SetLimits(JobLimits limits)
    {
        if (!double.IsFinite(limits.Delay) || limits.Delay < 0 || limits.Delay > uint.MaxValue / 1000d || limits.MaxBytes <= 0)
        {
            throw new ArgumentException("Invalid observation settings");
        }
        _ = new Regex(limits.Pattern);
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            nextLimits = limits;
        }
    }

    // Observation controls are not handler arguments. Validate names before
    // parsing into a JSON object, which would otherwise erase duplicates.
    private static JsonElement ExecutionArguments(JsonElement arguments)
    {
        var names = new HashSet<string>();
        foreach (var property in arguments.EnumerateObject())
        {
            if (!names.Add(property.Name))
            {
                throw new ArgumentException("Duplicate argument: " + property.Name);
            }
        }
        var args = JsonNode.Parse(arguments.GetRawText())!.AsObject();
        args.Remove("ai_callback_delay");
        args.Remove("ai_callback_pattern");
        args.Remove("max_output_bytes");
        return JsonSerializer.SerializeToElement(args);
    }

    internal partial async Task<ToolResultPart> DispatchAsync(ToolCallPart call, CancellationToken cancellation, string? refusal)
    {
        JobLimits selected;
        bool consumed;
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            consumed = nextLimits is not null;
            selected = nextLimits ?? new JobLimits();
            nextLimits = null;
        }
        cancellation.ThrowIfCancellationRequested();
        var tool = Parent().Tools().SingleOrDefault(t => t.Name() == call.Name);
        ToolResultPart result;
        if (!Parent().ToolEnabled(call.Name))
        {
            Parent().History().Append(new Event(EventKind.ToolCalled,
                callID: call.CallID, name: call.Name, args: call.Args));
            result = new ToolResultPart(call.CallID,
                [new TextPart("Tool is not enabled: " + call.Name)], true);
        }
        else if (tool?.Name() is "load_skill" or "unload_skill")
        {
            Parent().History().Append(new Event(EventKind.ToolCalled,
                callID: call.CallID, name: call.Name, args: call.Args));
            try
            {
                selected = Options(call.Args, selected);
                if (refusal is not null)
                {
                    throw new ArgumentException(refusal);
                }
                tool.Validate(ExecutionArguments(call.Args));
                var manager = Parent().SkillManager()
                    ?? throw new ArgumentException("No skill manager is attached");
                result = manager.Control(call);
            }
            catch (Exception error) when (error is ArgumentException or System.IO.InvalidDataException)
            {
                result = new ToolResultPart(call.CallID, [new TextPart(error.Message)], true);
            }
        }
        else if (tool?.IsControl() == true)
        {
            Parent().History().Append(new Event(EventKind.ToolCalled, callID: call.CallID, name: call.Name, args: call.Args));
            try
            {
                selected = Options(call.Args, selected);
                if (refusal is not null)
                {
                    result = new ToolResultPart(call.CallID, [new TextPart(refusal)], true);
                }
                else
                {
                    tool.Validate(ExecutionArguments(call.Args));
                    Parent().History().Append(new Event(EventKind.ToolStarted,
                        callID: call.CallID, name: call.Name, args: call.Args));
                    result = await Control(call, selected, cancellation);
                }
            }
            catch (Exception ex) when (ex is ArgumentException or JsonException or KeyNotFoundException or IOException or RegexMatchTimeoutException)
            {
                result = new ToolResultPart(call.CallID, [new TextPart(ex.Message)], true);
            }
        }
        else
        {
            Job job;
            lock (gate)
            {
                ObjectDisposedException.ThrowIf(destroyed, this);
                job = new Job(this, Parent().Parent().AllocateJobHandle(), call);
            }
            Exception? invalid = refusal is null ? null : new ArgumentException(refusal);
            var sanitized = call;
            try
            {
                selected = Options(call.Args, selected);
                sanitized = new ToolCallPart(call.CallID, call.From, call.Name,
                    ExecutionArguments(call.Args), call.Opaque);
            }
            catch (Exception ex) when (ex is ArgumentException or InvalidOperationException or JsonException)
            {
                invalid ??= ex;
            }
            try
            {
                if (invalid is null)
                {
                    if (tool is null)
                    {
                        throw new ArgumentException("Unknown tool: " + call.Name);
                    }
                    tool.Validate(sanitized.Args);
                }
            }
            catch (ArgumentException error)
            {
                invalid = error;
            }
            try
            {
                Parent().History().Append(new Event(EventKind.ToolCalled, callID: call.CallID,
                    name: call.Name, args: call.Args, job: job.Snapshot()));
                if (invalid is null)
                {
                    Parent().History().Append(new Event(EventKind.ToolStarted,
                        callID: call.CallID, name: call.Name, args: call.Args, job: job.Snapshot()));
                }
            }
            catch
            {
                // Reservation happened, but no worker has permission to start.
                job.Destroy();
                throw;
            }
            job.Start(tool, sanitized, invalid);
            try
            {
                result = await job.ObserveAsync(call.CallID, selected.Delay, selected.Pattern,
                    selected.MaxBytes, cancellation, wakeOnOutput: false);
            }
            catch (OperationCanceledException) when (cancellation.IsCancellationRequested)
            {
                // A hint/interrupt cancelled this observation, not the worker.
                // Use the same reporting path and real cursor; never invent a
                // killed result or lose the output produced before the wake.
                result = await job.ObserveAsync(call.CallID, 0, "", selected.MaxBytes,
                    CancellationToken.None, wakeOnOutput: false);
            }
            catch (RegexMatchTimeoutException)
            {
                result = new ToolResultPart(call.CallID,
                    [new TextPart("Output pattern evaluation timed out; the job is still available.")],
                    true, job.Snapshot());
            }
        }
        if (!consumed)
        {
            return result;
        }
        var pattern = selected.Pattern == "" ? "none" : selected.Pattern;
        var notice = $"[tool_limits consumed by this {call.Name} call: ai_callback_delay {selected.Delay}s, ai_callback_pattern {pattern}, max_output_bytes {selected.MaxBytes}]\n";
        return new ToolResultPart(result.CallID, [new TextPart(notice), .. result.Parts], result.IsError, result.Job, result.ToolReferences);
    }

    private async Task<ToolResultPart> Control(ToolCallPart call, JobLimits limits, CancellationToken cancellation)
    {
        var allowed = new HashSet<string>(["ai_callback_delay", "ai_callback_pattern", "max_output_bytes"]);
        if (call.Name != "tool_limits")
        {
            allowed.Add("handle");
        }
        if (call.Name == "send_input")
        {
            allowed.Add("input");
        }
        var seen = new HashSet<string>();
        foreach (var property in call.Args.EnumerateObject())
        {
            if (!seen.Add(property.Name) || !allowed.Contains(property.Name))
            {
                throw new ArgumentException("Duplicate or unknown argument: " + property.Name);
            }
        }
        if (call.Name == "tool_limits")
        {
            SetLimits(limits);
            return new ToolResultPart(call.CallID, [new TextPart("Observation limits saved for the next call, whichever tool it is. That result will report consumption. These limits never cancel work.")]);
        }
        var job = Find(Integer(call.Args, "handle", -1));
        if (call.Name == "kill_job")
        {
            await job.KillAsync();
            return await job.ObserveAsync(call.CallID, 0, "", limits.MaxBytes, cancellation);
        }
        try
        {
            if (call.Name == "send_input")
            {
                var input = String(call.Args, "input", null);
                return await job.SendInputAsync(call.CallID, input, limits.Delay, limits.Pattern, limits.MaxBytes, cancellation);
            }
            return await job.ObserveAsync(call.CallID, limits.Delay, limits.Pattern, limits.MaxBytes, cancellation);
        }
        catch (OperationCanceledException) when (cancellation.IsCancellationRequested)
        {
            return await job.ObserveAsync(call.CallID, 0, "", limits.MaxBytes,
                CancellationToken.None, wakeOnOutput: false);
        }
    }

    private static JobLimits Options(JsonElement args, JobLimits fallback)
    {
        if (args.ValueKind != JsonValueKind.Object)
        {
            throw new ArgumentException("Tool arguments must be an object");
        }
        var delay = Number(args, "ai_callback_delay", fallback.Delay);
        var pattern = String(args, "ai_callback_pattern", fallback.Pattern);
        var cap = Integer(args, "max_output_bytes", fallback.MaxBytes);
        if (cap <= 0)
        {
            throw new ArgumentException("max_output_bytes must be positive");
        }
        _ = new Regex(pattern, RegexOptions.None, TimeSpan.FromSeconds(1));
        return new JobLimits(delay, pattern, cap);
    }

    private static double Number(JsonElement args, string name, double fallback)
    {
        if (!args.TryGetProperty(name, out var value))
        {
            return fallback;
        }
        if (value.ValueKind != JsonValueKind.Number || !value.TryGetDouble(out var n)
            || !double.IsFinite(n) || n < 0 || n > uint.MaxValue / 1000d)
        {
            throw new ArgumentException(name + " must be a finite nonnegative duration");
        }
        return n;
    }

    private static int Integer(JsonElement args, string name, int fallback)
    {
        if (!args.TryGetProperty(name, out var value))
        {
            return fallback;
        }
        if (value.ValueKind != JsonValueKind.Number || !value.TryGetInt32(out var n))
        {
            throw new ArgumentException(name + " must be an integer");
        }
        return n;
    }

    private static string String(JsonElement args, string name, string? fallback)
    {
        if (!args.TryGetProperty(name, out var value) && fallback is not null)
        {
            return fallback;
        }
        if (value.ValueKind != JsonValueKind.String)
        {
            throw new ArgumentException(name + " must be a string");
        }
        return value.GetString()!;
    }

    public partial void Destroy()
    {
        lock (gate)
        {
            if (destroyed)
            {
                return;
            }
            destroyed = true;
        }
        // Start all cancellations before joining. Preserve recording failures,
        // but do not let one failure leave another owned process running.
        List<Exception> failures = [];
        var pending = Children().Select(job => job.KillAsync("shutdown")).ToArray();
        try
        {
            Task.WhenAll(pending).GetAwaiter().GetResult();
        }
        catch (Exception error)
        {
            failures.Add(error);
        }
        foreach (var job in Children())
        {
            try
            {
                job.Destroy();
            }
            catch (Exception error)
            {
                failures.Add(error);
            }
        }
        parent?.RemoveJobs(this);
        if (failures.Count != 0)
        {
            throw new AggregateException("Job shutdown could not record or close every resource", failures);
        }
    }
}

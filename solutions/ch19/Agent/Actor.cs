using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial Task<string> SubmitAsync(IEnumerable<Part> parts)
    {
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        ImmutableArray<Part> frozen;
        try
        {
            ArgumentNullException.ThrowIfNull(parts);
            frozen = parts.ToImmutableArray();
        }
        catch
        {
            // Enumeration belongs to the caller boundary. Even a failed copy
            // gets an outcome, but no mutable enumerable enters the mailbox.
            ObserveInput(new Mail(MailKind.Submit), false);
            throw;
        }
        Enqueue(new Mail(MailKind.Submit, Parts: frozen, Reply: reply));
        return reply.Task;
    }

    public partial Task PostAsync(AgentInput input)
    {
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.Input, Input: input, Reply: reply));
        return reply.Task;
    }

    public partial async IAsyncEnumerable<string> RunAsync(System.Threading.Channels.ChannelReader<string> input,
        [System.Runtime.CompilerServices.EnumeratorCancellation] CancellationToken cancellation)
    {
        await foreach (string text in input.ReadAllAsync(cancellation))
        {
            yield return await SubmitAsync([new TextPart(text)]).WaitAsync(cancellation);
        }
    }

    private partial void Enqueue(Mail message)
    {
        lock (ingress)
        {
            if (admissionClosed || destroyed)
            {
                ObserveInput(message, false);
                throw new ObjectDisposedException(nameof(Agent));
            }
            actorTask ??= Task.Run(ActorLoop);
            if (!mailbox.Writer.TryWrite(message))
            {
                ObserveInput(message, false);
                throw new ObjectDisposedException(nameof(Agent));
            }
        }
    }

    public partial Task FlushAsync(CancellationToken cancellation)
    {
        var reply = new TaskCompletionSource<string>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.Flush, Reply: reply));
        return reply.Task.WaitAsync(cancellation);
    }

    public partial Task ShutdownAsync()
    {
        lock (ingress)
        {
            if (!admissionClosed)
            {
                admissionClosed = true;
                actorTask ??= Task.Run(ActorLoop);
                mailbox.Writer.TryWrite(new Mail(MailKind.Shutdown));
            }
            // Failed construction can close admission before an actor exists.
            // There is then no asynchronous work to join.
            return actorTask ?? Task.CompletedTask;
        }
    }

    public partial async Task<Observation> WaitAsync(Func<Observation, bool> predicate,
        CancellationToken cancellation)
    {
        await foreach (var value in observations.Reader.ReadAllAsync(cancellation))
        {
            if (predicate(value))
            {
                return value;
            }
        }
        throw new ObjectDisposedException(nameof(Agent));
    }

    internal partial void Publish(Observation value)
    {
        // TryWrite with DropOldest never waits for a GUI or a forgotten waiter.
        observations.Writer.TryWrite(value);
        // Shutdown keeps observation open for admission refusals. Logical
        // destruction closes it and removes the root relationship as well.
        parent?.Observe(this, value);
    }

    private partial void ObserveInput(Mail message, bool accepted)
    {
        if (message.Kind is not (MailKind.Submit or MailKind.DirectTool or MailKind.ReverseTool or MailKind.Input))
        {
            return;
        }
        string label = message.Input?.Kind.ToString() ?? message.Call?.Name ?? message.Kind.ToString();
        Publish(new Observation(ObservationKind.Input, Text: label, Accepted: accepted));
    }

    private partial void ValidateInput(Mail message)
    {
        if (message.Kind == MailKind.Input)
        {
            ArgumentNullException.ThrowIfNull(message.Input);
            if (!Enum.IsDefined(message.Input.Kind))
            {
                throw new ArgumentException("Unknown input kind.");
            }
            ArgumentNullException.ThrowIfNull(message.Input.Text);
        }
        else if (message.Kind == MailKind.Submit)
        {
            foreach (var part in message.Parts)
            {
                if (part is not (TextPart or BlobPart))
                {
                    throw new ArgumentException("A user message contains only text or media parts.");
                }
                if (part is TextPart text)
                {
                    ArgumentNullException.ThrowIfNull(text.Text);
                }
            }
        }
        else if (message.Kind is MailKind.DirectTool or MailKind.ReverseTool)
        {
            ArgumentNullException.ThrowIfNull(message.Call);
        }
    }

    private partial void Record(Event fact, ImmutableArray<int> partIDs)
    {
        var before = History().View().Phase;
        if (fact.Kind == EventKind.ResponseEnded && !partIDs.IsDefault)
        {
            fact = fact.WithDisplay(responseID, partIDs);
        }
        var saved = History().Append(fact);
        var after = History().View().Phase;
        if (before != after)
        {
            Publish(new Observation(ObservationKind.StateChanged, saved.Seq, From: before, To: after));
        }
        if (fact.Kind is EventKind.MessageReceived or EventKind.HintReceived or EventKind.ResponseEnded)
        {
            for (int index = 0; index < fact.Parts.Length; index++)
            {
                int? id = partIDs.IsDefaultOrEmpty ? null : partIDs[index];
                Publish(new Observation(ObservationKind.PartFinal, saved.Seq, fact.Parts[index],
                    PartID: id, ResponseID: id.HasValue ? responseID : 0));
            }
        }
        if (fact.Kind is EventKind.ToolReturned or EventKind.ToolObserved)
        {
            var part = new ToolResultPart(fact.CallID, fact.Parts, fact.IsError, fact.Job, fact.ToolReferences);
            Publish(new Observation(ObservationKind.PartFinal, saved.Seq, part));
        }
    }

    private partial async Task ActorLoop()
    {
        try
        {
            while (await mailbox.Reader.WaitToReadAsync())
            {
                // Controls already waiting in the mailbox take precedence over
                // launching the next effect. No worker enters this loop.
                while (mailbox.Reader.TryRead(out var message))
                {
                    Receive(message);
                }
                if (stopping && !workerPending && controlReplies.Count == 0 && startupReplies.Count == 0)
                {
                    Finish(new ObjectDisposedException(nameof(Agent)));
                    while (requests.TryDequeue(out var queued))
                    {
                        queued.Reply?.TrySetException(new ObjectDisposedException(nameof(Agent)));
                        queued.ToolReply?.TrySetException(new ObjectDisposedException(nameof(Agent)));
                    }
                    if (mcp is not null)
                    {
                        await mcp.DestroyAsync();
                    }
                    Jobs().Destroy();
                    // Completion, not mere admission: owned work has drained,
                    // while History is still available to record that fact.
                    Record(new Event(EventKind.AgentShutdown, reason: "shutdown"));
                    if (persistence is not null && (skills is null || skills.Ready()))
                    {
                        History().Save(persistence);
                    }
                    break;
                }
                Advance();
                if (active is null && requests.Count == 0 && !workerPending && controlReplies.Count == 0 && startupReplies.Count == 0)
                {
                    foreach (var reply in flushReplies)
                    {
                        reply.TrySetResult("");
                    }
                    flushReplies.Clear();
                }
            }
        }
        catch (Exception error)
        {
            // Close admission before draining. Worker wrappers always post one
            // outcome and perform no parent access after that publication.
            lock (ingress)
            {
                admissionClosed = true;
            }
            observationCancellation?.Cancel();
            controlCancellation.Cancel();
            while (workerPending || controlReplies.Count != 0 || startupReplies.Count != 0)
            {
                var outcome = await mailbox.Reader.ReadAsync();
                if (outcome.Kind is MailKind.HttpDone or MailKind.ToolDone or MailKind.AutomaticDone)
                {
                    workerPending = false;
                }
                else if (outcome.Kind == MailKind.ControlDone)
                {
                    controlReplies.Remove(outcome.ToolReply!);
                    startupReplies.Remove(outcome.Reply!);
                }
                ObserveInput(outcome, false);
                outcome.Reply?.TrySetException(error);
                outcome.ToolReply?.TrySetException(error);
                outcome.SkillReply?.TrySetException(error);
            }
            foreach (var reply in flushReplies)
            {
                reply.TrySetException(error);
            }
            // Faults are delivered to the host, not converted into model text.
            active?.Reply?.TrySetException(error);
            active?.ToolReply?.TrySetException(error);
            foreach (var reply in startupReplies)
            {
                reply.TrySetException(error);
            }
            foreach (var reply in controlReplies)
            {
                reply.TrySetException(error);
            }
            while (requests.TryDequeue(out var queued))
            {
                queued.Reply?.TrySetException(error);
                queued.ToolReply?.TrySetException(error);
            }
            try
            {
                jobs?.Destroy();
            }
            catch (Exception cleanupError)
            {
                throw new AggregateException("Actor and cleanup both failed", error, cleanupError);
            }
            throw;
        }
        finally
        {
            lock (ingress)
            {
                admissionClosed = true;
                mailbox.Writer.TryComplete();
            }
            while (mailbox.Reader.TryRead(out var rejected))
            {
                ObserveInput(rejected, false);
                rejected.Reply?.TrySetException(new ObjectDisposedException(nameof(Agent)));
                rejected.ToolReply?.TrySetException(new ObjectDisposedException(nameof(Agent)));
                rejected.SkillReply?.TrySetException(new ObjectDisposedException(nameof(Agent)));
            }
            foreach (var reply in flushReplies)
            {
                reply.TrySetException(new ObjectDisposedException(nameof(Agent)));
            }
            flushReplies.Clear();
            observationCancellation?.Dispose();
            controlCancellation.Dispose();
        }
    }

    private partial void Receive(Mail message)
    {
        if (message.Kind == MailKind.SelectModel)
        {
            SelectModel(message);
            return;
        }
        if (message.Kind == MailKind.InitializeSkills)
        {
            try
            {
                ObjectDisposedException.ThrowIf(stopping, this);
                var work = SkillManager()!.BeginInitialization(controlCancellation.Token);
                startupReplies.Add(message.Reply!);
                _ = StartupWorker(message, work);
            }
            catch (Exception error)
            {
                message.Reply!.TrySetException(error);
            }
            return;
        }
        if (message.Kind == MailKind.CommitSkill)
        {
            try
            {
                ObjectDisposedException.ThrowIf(stopping, this);
                var activation = message.Activation!;
                activation.Cancellation.ThrowIfCancellationRequested();
                if (activation.Owner != SkillManager())
                {
                    throw new InvalidOperationException("Skill manager is no longer attached.");
                }
                message.SkillReply!.TrySetResult(activation.Owner.Commit(activation));
            }
            catch (Exception error)
            {
                message.SkillReply!.TrySetException(error);
                if (error is not (ArgumentException or InvalidOperationException or InvalidDataException
                    or OperationCanceledException))
                {
                    throw;
                }
            }
            return;
        }
        if (message.Kind == MailKind.RegisterMcpTools)
        {
            try
            {
                ObjectDisposedException.ThrowIf(stopping, this);
                if (message.Client!.Parent().Parent() != this)
                {
                    throw new InvalidOperationException("MCP client belongs to another Agent.");
                }
                message.Client.InstallTools();
                message.Reply!.TrySetResult("");
            }
            catch (Exception error) when (error is ArgumentException or InvalidOperationException)
            {
                message.Reply!.TrySetException(error);
            }
            return;
        }
        if (message.Kind == MailKind.Delta)
        {
            // The transport emits into this same ordered inbox. It never touches
            // History. Cancellation closes publication authority, not just HTTP.
            if (transportPending && !discardResponse && !stopping)
            {
                Publish(new Observation(ObservationKind.PartDelta, Delta: message.Chunk,
                    PartID: message.Chunk!.PartID, ResponseID: responseID));
            }
            return;
        }
        if (message.Kind == MailKind.Flush)
        {
            flushReplies.Add(message.Reply!);
            return;
        }
        if (message.Kind == MailKind.Wake)
        {
            return;
        }
        if (message.Kind is MailKind.Input or MailKind.Submit or MailKind.DirectTool or MailKind.ReverseTool)
        {
            Exception? refusal = null;
            try
            {
                ObjectDisposedException.ThrowIf(stopping, this);
                ValidateInput(message);
                bool needsPrimary = message.Kind is MailKind.Submit or MailKind.DirectTool
                    || (message.Kind == MailKind.Input && message.Input!.Kind is InputKind.Message or InputKind.Hint);
                if (needsPrimary && skills is not null && !skills.Ready())
                {
                    throw new InvalidOperationException("Primary skills have not finished initialization.");
                }
            }
            catch (Exception error) when (error is ArgumentException or InvalidOperationException)
            {
                refusal = error;
            }
            if (refusal is not null)
            {
                ObserveInput(message, false);
                message.Reply?.TrySetException(refusal);
                message.ToolReply?.TrySetException(refusal);
                return;
            }
        }
        if (message.Kind == MailKind.ReverseTool || (message.Kind == MailKind.DirectTool
            && Tools().Any(tool => tool.Name() == message.Call!.Name && tool.IsControl())))
        {
            ObserveInput(message, true);
            controlReplies.Add(message.ToolReply!);
            // Dispatch's synchronous prefix runs on the actor. Only observation
            // runs outside and posts its immutable outcome back to this inbox.
            var work = Jobs().DispatchAsync(message.Call!, controlCancellation.Token);
            _ = ControlWorker(message, work);
            return;
        }
        if (message.Kind == MailKind.ControlDone)
        {
            if (message.Reply is not null)
            {
                startupReplies.Remove(message.Reply);
                if (message.Error is not null)
                {
                    message.Reply.TrySetException(message.Error);
                }
                else
                {
                    message.Reply.TrySetResult("");
                }
                return;
            }
            if (message.Error is not null)
            {
                message.ToolReply!.TrySetException(message.Error);
            }
            else
            {
                var result = message.Result!;
                try
                {
                    Record(new Event(EventKind.ToolObserved, callID: result.CallID, parts: result.Parts,
                        isError: result.IsError, job: result.Job, toolReferences: result.ToolReferences));
                    StageHandoff(message.Call!, result);
                    message.ToolReply!.TrySetResult(result);
                }
                catch (Exception error)
                {
                    message.ToolReply!.TrySetException(error);
                    throw;
                }
                finally
                {
                    controlReplies.Remove(message.ToolReply!);
                }
            }
            controlReplies.Remove(message.ToolReply!);
            return;
        }
        if (message.Kind is MailKind.Submit or MailKind.DirectTool)
        {
            requests.Enqueue(message);
            ObserveInput(message, true);
            return;
        }
        if (message.Kind == MailKind.Shutdown)
        {
            stopping = true;
            interruptPending = true;
            observationCancellation?.Cancel();
            controlCancellation.Cancel();
            return;
        }
        if (message.Kind == MailKind.Input)
        {
            var input = message.Input!;
            try
            {
                switch (input.Kind)
                {
                    case InputKind.Pause:
                        paused = true;
                        Record(new Event(EventKind.AgentPaused, reason: "pause"));
                        break;
                    case InputKind.Resume:
                        paused = false;
                        Record(new Event(EventKind.AgentResumed, reason: "resume"));
                        break;
                    case InputKind.InterruptNow:
                    case InputKind.InterruptAfterTool:
                        if (active is not null)
                        {
                            Record(new Event(EventKind.TurnInterrupted, reason: input.Kind.ToString()));
                            interruptPending = true;
                            discardResponse = input.Kind == InputKind.InterruptNow;
                            if (input.Kind == InputKind.InterruptNow || automaticPending)
                            {
                                observationCancellation?.Cancel();
                            }
                        }
                        break;
                    case InputKind.ClearTools:
                        pendingToolClear = true;
                        break;
                    case InputKind.Ephemeral:
                        Record(new Event(EventKind.MessageReceived, actor: Actor.System,
                            parts: [new TextPart(input.Text)]));
                        break;
                    case InputKind.Hint:
                    case InputKind.Message:
                        if (active is null)
                        {
                            Record(new Event(EventKind.HintReceived, actor: Actor.Human,
                                parts: [new TextPart(input.Text)]));
                            // The fact is already recorded. This empty request
                            // merely triggers work once scheduling is allowed.
                            requests.Enqueue(new Mail(MailKind.Submit));
                        }
                        else
                        {
                            Record(new Event(EventKind.HintReceived, actor: Actor.Human,
                                parts: [new TextPart(input.Text)]));
                            hintPending = true;
                            if (!transportPending && nextCall < calls.Length)
                            {
                                skipTail = true;
                            }
                            if (workerPending && !transportPending && !automaticPending)
                            {
                                observationCancellation?.Cancel();
                            }
                        }
                        break;
                    default:
                        throw new ArgumentException("Unknown input kind.");
                }
                ObserveInput(message, true);
                message.Reply!.TrySetResult("");
            }
            catch (Exception error)
            {
                ObserveInput(message, false);
                message.Reply!.TrySetException(error);
            }
            return;
        }

        // There is at most one outstanding transport/observation. A new turn
        // cannot start until this completion is consumed, including cancellation.
        workerPending = false;
        observationCancellation?.Dispose();
        observationCancellation = null;
        if (message.Kind == MailKind.AutomaticDone)
        {
            ReceiveAutomatic(message);
            return;
        }
        if (message.Kind == MailKind.HttpDone)
        {
            transportPending = false;
            if ((interruptPending && discardResponse) || stopping)
            {
                return; // A cancelled/late response has no authority to dispatch.
            }
            if (message.Error is not null)
            {
                if (message.Error is PlanAuthorizationException refusal)
                {
                    // One durable error feeds replay, GUI and speech. Record
                    // before ending the turn; a failed append still faults the
                    // actor and cannot publish a success or a fabricated fact.
                    Record(new Event(EventKind.ErrorOccurred, error: refusal.Message));
                }
                Finish(message.Error);
                return;
            }
            var parsed = message.Parsed!;
            toolBeforeText = parsed.ToolBeforeText;
            // Capture before the consuming response becomes the new once-only
            // boundary. Workers never select cuts or mutate History.
            int byteTarget = ContextByteBudget();
            pendingResultStubs = configuration.StubsToolResults && byteTarget > 0
                && !parsed.Events.Any(fact => fact.Kind == EventKind.ErrorOccurred)
                ? SelectImmediateCuts(History().View(), responseID, byteTarget / 100)
                : [];
            foreach (var fact in parsed.Events)
            {
                Record(fact, fact.Kind == EventKind.ResponseEnded ? parsed.PartIDs : default);
            }
            var failure = parsed.Events.FirstOrDefault(e => e.Kind == EventKind.ErrorOccurred);
            if (failure is not null)
            {
                Finish(new InvalidOperationException(failure.Error));
                return;
            }
            var response = parsed.Events.Last(e => e.Kind == EventKind.ResponseEnded);
            // A completed response proves this request's older results were
            // presented. New calls/results from this response are not yet seen.
            presentedThrough = responseID;
            calls = response.Parts.OfType<ToolCallPart>().ToImmutableArray();
            nextCall = 0;
            responseText = JsonCodec.Text(response.Parts);
            if (interruptPending)
            {
                return; // Advance pairs the entire retained boundary without effects.
            }
            if (calls.Select(call => call.CallID).Distinct(StringComparer.Ordinal).Count() != calls.Length)
            {
                // Repeated IDs across responses are valid; ambiguous IDs within
                // this response must be refused before even its first effect.
                Finish(new InvalidDataException("Duplicate tool-call IDs in one response."));
                return;
            }
            if (calls.IsEmpty)
            {
                if (!hintPending)
                {
                    Finish();
                }
            }
            else
            {
                batches++;
                int limit = capturedToolLimit;
                if (limit > 0 && batches > limit)
                {
                    SkipRemaining("Tool batch limit reached; not executed.");
                    Finish(new InvalidOperationException("Tool batch limit reached."));
                }
            }
        }
        else if (message.Kind == MailKind.ToolDone)
        {
            if (message.Error is not null)
            {
                Finish(message.Error);
                return;
            }
            var result = message.Result!;
            var completedCall = active?.Call ?? calls[nextCall - 1];
            var isControl = Tools().Any(tool => tool.Name() == completedCall.Name && tool.IsControl());
            // A host-invoked tool has no preceding assistant call. Preserve its
            // audit outcome without inventing a call or poisoning model dialogue.
            var resultKind = active?.Kind == MailKind.DirectTool
                ? EventKind.ToolObserved : EventKind.ToolReturned;
            Record(new Event(resultKind, callID: result.CallID, parts: result.Parts,
                isError: result.IsError, job: isControl ? null : result.Job, toolReferences: result.ToolReferences));
            StageHandoff(completedCall, result);
            if (active?.Kind != MailKind.DirectTool)
            {
                StageKeepResults(completedCall, result);
            }
            if (active?.Kind == MailKind.DirectTool)
            {
                if (hintPending && !interruptPending && !stopping)
                {
                    // The hint is already a History fact. Continue it as a
                    // live turn after settling this independent direct reply.
                    requests.Enqueue(new Mail(MailKind.Submit));
                }
                var reply = active.ToolReply!;
                Finish();
                reply.TrySetResult(result);
            }
        }
    }

    private partial void Advance()
    {
        if (workerPending)
        {
            return;
        }
        if (interruptPending)
        {
            SkipRemaining("Interrupted before dispatch; not executed.");
            Finish(new OperationCanceledException("Turn interrupted."));
        }
        if (skipTail)
        {
            SkipRemaining("Hint received before dispatch; not executed.");
            skipTail = false;
        }
        if (nextCall >= calls.Length)
        {
            ApplyResultStubs();
            ApplyHandoff();
            ApplyToolClear();
        }
        if (paused || stopping)
        {
            return;
        }
        if (active is null)
        {
            if (!requests.TryDequeue(out active))
            {
                return;
            }
            batches = 0;
            int policy = MaxToolRounds();
            capturedToolLimit = policy == 0 ? 200 : policy;
            if (active.Kind == MailKind.Submit && !active.Parts.IsDefaultOrEmpty)
            {
                Record(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: active.Parts));
            }
        }
        try
        {
            observationCancellation = new CancellationTokenSource();
            if (active.Kind == MailKind.DirectTool || nextCall < calls.Length)
            {
                var call = active.Call ?? calls[nextCall++];
                string? refusal = configuration.RequireToolExplanation && active.Kind != MailKind.DirectTool
                    && (toolBeforeText || string.IsNullOrWhiteSpace(responseText))
                    ? "Tool not executed: first explain to the user what you will do and why, then retry."
                    : null;
                if (!string.IsNullOrWhiteSpace(responseText))
                {
                    Parent().Log("Agent before tools: " + responseText);
                }
                // Dispatch's synchronous prefix records ToolCalled before Start.
                // Only its observation wait continues outside the actor.
                var work = Jobs().DispatchAsync(call, observationCancellation.Token, refusal);
                workerPending = true;
                transportPending = false;
                _ = ToolWorker(work);
            }
            else
            {
                CompactTools();
                if (BeginAutomaticCollection())
                {
                    return;
                }
                // This instruction belongs to this request only. It neither changes
                // canonical History nor any opaque provider replay material.
                string budgetInstruction = capturedToolLimit > 0 && batches >= capturedToolLimit
                    ? "Tool batch limit reached. Do not call more tools; answer using the results already received."
                    : "";
                if (!turnSnapshots.IsEmpty || !roundSnapshots.IsEmpty)
                {
                    // Combine with the caller's current slot immediately before
                    // freezing the request, never with an earlier manual value.
                    var parts = History().View().Ephemeral.AddRange(turnSnapshots).AddRange(roundSnapshots);
                    Record(new Event(EventKind.MessageReceived, actor: Actor.System, parts: parts));
                }
                var requestContext = History().View();
                string body;
                try
                {
                    body = Engine().Render(requestContext, budgetInstruction).ToJsonString();
                }
                catch
                {
                    // Rendering refusal is not a transport attempt. Still clear
                    // its one-shot slot so a later turn cannot inherit snapshots.
                    if (!requestContext.Ephemeral.IsEmpty)
                    {
                        Record(new Event(EventKind.MessageReceived, actor: Actor.System));
                    }
                    throw;
                }
                Record(new Event(EventKind.RequestSent, from: configuration.Target()));
                responseID = History().Events()[^1].Seq;
                hintPending = false;
                calls = [];
                nextCall = 0;
                workerPending = true;
                transportPending = true;
                roundSnapshotsCollected = false;
                roundSnapshots = [];
                _ = TransportWorker(body, observationCancellation.Token);
            }
        }
        catch (Exception error)
        {
            Finish(error);
        }
    }

    private partial async Task ControlWorker(Mail request, Task<ToolResultPart> work)
    {
        Mail completion;
        try
        {
            completion = request with { Kind = MailKind.ControlDone, Result = await work };
        }
        catch (Exception error)
        {
            completion = request with { Kind = MailKind.ControlDone, Error = error };
        }
        mailbox.Writer.TryWrite(completion);
    }

    private partial void SkipRemaining(string reason)
    {
        while (nextCall < calls.Length)
        {
            var call = calls[nextCall++];
            Record(new Event(EventKind.ToolReturned, callID: call.CallID,
                parts: [new TextPart(reason)], isError: true));
        }
    }

    private partial void Finish(Exception? error)
    {
        if (active is not null)
        {
            if (error is not null)
            {
                SkipRemaining("Turn failed before dispatch; not executed.");
            }
            ApplyResultStubs();
            ApplyHandoff();
            ApplyToolClear();
            Record(new Event(EventKind.TurnEnded, reason: error?.Message ?? "completed"));
            Publish(new Observation(ObservationKind.TurnEnded, Text: error?.Message ?? responseText,
                Succeeded: error is null, HasReply: active.Reply is not null || active.ToolReply is not null));
            if (error is null)
            {
                active.Reply?.TrySetResult(responseText);
            }
            else
            {
                active.Reply?.TrySetException(error);
                active.ToolReply?.TrySetException(error);
            }
        }
        ApplyToolClear();
        active = null;
        automaticPending = false;
        turnSnapshotsCollected = false;
        roundSnapshotsCollected = false;
        turnSnapshots = [];
        roundSnapshots = [];
        calls = [];
        nextCall = 0;
        responseText = "";
        interruptPending = false;
        discardResponse = false;
        skipTail = false;
        hintPending = false;
        mailbox.Writer.TryWrite(new Mail(MailKind.Wake));
    }

    private partial async Task TransportWorker(string body, CancellationToken cancellation)
    {
        Mail completion;
        try
        {
            var parsed = await Engine().SendAsync(body, cancellation,
                chunk => mailbox.Writer.TryWrite(new Mail(MailKind.Delta, Chunk: chunk)));
            completion = new Mail(MailKind.HttpDone, Parsed: parsed);
        }
        catch (Exception error)
        {
            completion = new Mail(MailKind.HttpDone, Error: error);
        }
        mailbox.Writer.TryWrite(completion);
    }

    private partial async Task ToolWorker(Task<ToolResultPart> work)
    {
        Mail completion;
        try
        {
            completion = new Mail(MailKind.ToolDone, Result: await work);
        }
        catch (Exception error)
        {
            completion = new Mail(MailKind.ToolDone, Error: error);
        }
        mailbox.Writer.TryWrite(completion);
    }
}

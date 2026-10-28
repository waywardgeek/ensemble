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
            return actorTask!;
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
        if (message.Kind is not (MailKind.Submit or MailKind.DirectTool or MailKind.Input))
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
        else if (message.Kind == MailKind.DirectTool)
        {
            ArgumentNullException.ThrowIfNull(message.Call);
        }
    }

    private partial void Record(Event fact)
    {
        var before = History().View().Phase;
        History().Append(fact);
        var saved = History().Events()[^1];
        var after = History().View().Phase;
        if (before != after)
        {
            Publish(new Observation(ObservationKind.StateChanged, saved.Seq, From: before, To: after));
        }
        if (fact.Kind is EventKind.MessageReceived or EventKind.HintReceived or EventKind.ResponseEnded)
        {
            foreach (var part in fact.Parts)
            {
                Publish(new Observation(ObservationKind.PartFinal, saved.Seq, part));
            }
        }
        if (fact.Kind is EventKind.ToolReturned or EventKind.ToolObserved)
        {
            var part = new ToolResultPart(fact.CallID, fact.Parts, fact.IsError, fact.Job);
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
                if (stopping && !workerPending && controlReplies.Count == 0)
                {
                    Finish(new ObjectDisposedException(nameof(Agent)));
                    while (requests.TryDequeue(out var queued))
                    {
                        queued.Reply?.TrySetException(new ObjectDisposedException(nameof(Agent)));
                        queued.ToolReply?.TrySetException(new ObjectDisposedException(nameof(Agent)));
                    }
                    Jobs().Destroy();
                    break;
                }
                Advance();
                if (active is null && requests.Count == 0 && !workerPending && controlReplies.Count == 0)
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
            while (workerPending || controlReplies.Count != 0)
            {
                var outcome = await mailbox.Reader.ReadAsync();
                if (outcome.Kind is MailKind.HttpDone or MailKind.ToolDone)
                {
                    workerPending = false;
                }
                else if (outcome.Kind == MailKind.ControlDone)
                {
                    controlReplies.Remove(outcome.ToolReply!);
                }
                ObserveInput(outcome, false);
                outcome.Reply?.TrySetException(error);
                outcome.ToolReply?.TrySetException(error);
            }
            foreach (var reply in flushReplies)
            {
                reply.TrySetException(error);
            }
            // Faults are delivered to the host, not converted into model text.
            active?.Reply?.TrySetException(error);
            active?.ToolReply?.TrySetException(error);
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
        if (message.Kind == MailKind.Flush)
        {
            flushReplies.Add(message.Reply!);
            return;
        }
        if (message.Kind == MailKind.Wake)
        {
            return;
        }
        if (message.Kind is MailKind.Input or MailKind.Submit or MailKind.DirectTool)
        {
            Exception? refusal = null;
            try
            {
                ObjectDisposedException.ThrowIf(stopping, this);
                ValidateInput(message);
            }
            catch (Exception error) when (error is ArgumentException or ObjectDisposedException)
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
        if (message.Kind == MailKind.DirectTool
            && Tools().Any(tool => tool.Name() == message.Call!.Name && tool.IsControl()))
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
                        isError: result.IsError));
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
                        break;
                    case InputKind.Resume:
                        paused = false;
                        break;
                    case InputKind.InterruptNow:
                    case InputKind.InterruptAfterTool:
                        if (active is not null)
                        {
                            Record(new Event(EventKind.TurnInterrupted, reason: input.Kind.ToString()));
                            interruptPending = true;
                            discardResponse = input.Kind == InputKind.InterruptNow;
                            if (input.Kind == InputKind.InterruptNow)
                            {
                                observationCancellation?.Cancel();
                            }
                        }
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
                            if (workerPending && !transportPending)
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
        if (message.Kind == MailKind.HttpDone)
        {
            transportPending = false;
            if ((interruptPending && discardResponse) || stopping)
            {
                return; // A cancelled/late response has no authority to dispatch.
            }
            if (message.Error is not null)
            {
                Finish(message.Error);
                return;
            }
            foreach (var fact in message.Events)
            {
                Record(fact);
            }
            var failure = message.Events.FirstOrDefault(e => e.Kind == EventKind.ErrorOccurred);
            if (failure is not null)
            {
                Finish(new InvalidOperationException(failure.Error));
                return;
            }
            var response = message.Events.Last(e => e.Kind == EventKind.ResponseEnded);
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
                int limit = configuration.ToolRoundLimit == 0 ? 200 : configuration.ToolRoundLimit;
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
                isError: result.IsError, job: isControl ? null : result.Job));
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
            if (active.Kind == MailKind.Submit && !active.Parts.IsDefaultOrEmpty)
            {
                Record(new Event(EventKind.MessageReceived, actor: Actor.Human, parts: active.Parts));
            }
        }
        if (skipTail)
        {
            SkipRemaining("Hint received before dispatch; not executed.");
            skipTail = false;
        }
        try
        {
            observationCancellation = new CancellationTokenSource();
            if (active.Kind == MailKind.DirectTool || nextCall < calls.Length)
            {
                var call = active.Call ?? calls[nextCall++];
                string? refusal = configuration.RequireToolExplanation && active.Kind != MailKind.DirectTool
                    && string.IsNullOrWhiteSpace(responseText)
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
                string body = Engine().Render(History().View()).ToJsonString();
                Record(new Event(EventKind.RequestSent, from: configuration.Target()));
                hintPending = false;
                calls = [];
                nextCall = 0;
                workerPending = true;
                transportPending = true;
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
            Record(new Event(EventKind.TurnEnded, reason: error?.Message ?? "completed"));
            Publish(new Observation(ObservationKind.TurnEnded, Text: error?.Message ?? responseText));
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
        active = null;
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
            completion = new Mail(MailKind.HttpDone, Events: await Engine().SendAsync(body, cancellation));
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

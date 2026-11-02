using Ensemble;

namespace EnsembleGui;

internal sealed partial class VoiceLease
{
    internal partial VoiceLease(Agent target, long sequence, CancellationToken connection)
    {
        ArgumentNullException.ThrowIfNull(target);
        if (sequence <= 0) throw new ArgumentOutOfRangeException(nameof(sequence));
        Target = target;
        Sequence = sequence;
        lifetime = CancellationTokenSource.CreateLinkedTokenSource(connection);
    }

    internal partial Task<VoiceTranscript> Start(PcmVoiceAudio audio)
    {
        ArgumentNullException.ThrowIfNull(audio);
        lock (gate)
        {
            if (cancelled || destruction is not null || lifetime.IsCancellationRequested)
                throw new InvalidOperationException("Voice capture is no longer active");
            if (transcription is not null)
                throw new InvalidOperationException("Voice capture has already submitted audio");
            var transcriber = Target.Parent().Transcriber()
                ?? throw new InvalidOperationException("Separately funded transcription is not configured");
            transcription = transcriber.TranscribeAsync(audio, lifetime.Token);
            return transcription;
        }
    }

    internal partial Task<string> CommitAsync(Agent selected, long sequence)
    {
        lock (gate)
        {
            if (!ReferenceEquals(selected, Target) || sequence != Sequence)
                throw new InvalidOperationException("Voice capture does not belong to this selection");
            if (cancelled || destruction is not null || lifetime.IsCancellationRequested || consumed)
                throw new InvalidOperationException("Voice capture is no longer committable");
            if (transcription is null || !transcription.IsCompletedSuccessfully)
                throw new InvalidOperationException("Voice transcript is not ready");
            // This is the admission linearization point. Subsequent cancellation
            // cannot retract an already admitted message. Failed/uncertain sends
            // are never retried with this capture either.
            consumed = true;
            try
            {
                commit = Target.PostAsync(new AgentInput(InputKind.Message, transcription.Result.Text));
            }
            catch (Exception error)
            {
                commit = Task.FromException<string>(error);
            }
            return commit;
        }
    }

    internal partial bool IsActive()
    {
        lock (gate) return !cancelled && !consumed && destruction is null && !lifetime.IsCancellationRequested;
    }

    internal partial void Cancel()
    {
        lock (gate)
        {
            if (cancelled || destruction is not null) return;
            cancelled = true;
            lifetime.Cancel();
        }
    }

    public partial ValueTask DisposeAsync()
    {
        TaskCompletionSource completion;
        Task<VoiceTranscript>? request;
        Task<string>? admission;
        lock (gate)
        {
            if (destruction is not null) return new ValueTask(destruction);
            cancelled = true;
            completion = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            destruction = completion.Task;
            request = transcription;
            admission = commit;
        }
        // Publish the owned join before cancellation or asynchronous cleanup.
        _ = DestroyOwnedAsync(request, admission, completion);
        return new ValueTask(completion.Task);
    }

    private partial async Task DestroyOwnedAsync(Task<VoiceTranscript>? request,
        Task<string>? admission, TaskCompletionSource completed)
    {
        Exception? failure = null;
        try { lifetime.Cancel(); }
        catch (Exception error) { failure = error; }
        // Operation failures remain on their original tasks, not as destruction
        // failures. Cancellation still joins an uncooperative transport.
        if (request is not null)
        {
            try { await request.ConfigureAwait(false); }
            catch { }
        }
        if (admission is not null)
        {
            try { await admission.ConfigureAwait(false); }
            catch { }
        }
        lifetime.Dispose();
        if (failure is null) completed.TrySetResult();
        else completed.TrySetException(failure);
    }
}

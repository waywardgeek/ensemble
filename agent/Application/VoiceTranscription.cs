using System.Net.Http.Headers;

namespace Ensemble;

public sealed partial class Ensemble
{
    public partial VoiceTranscriber? Transcriber()
    {
        lock (agents) { return voiceTranscriber; }
    }

    internal partial void AddVoiceTranscriber(VoiceTranscriber child)
    {
        lock (agents)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (voiceTranscriber is not null || child.parent is not null)
                throw new InvalidOperationException("A voice transcriber is already attached.");
            voiceTranscriber = child;
            child.parent = this;
        }
    }

    internal partial void RemoveVoiceTranscriber(VoiceTranscriber child)
    {
        lock (agents)
        {
            if (!ReferenceEquals(voiceTranscriber, child) || !ReferenceEquals(child.parent, this))
                throw new InvalidOperationException("Voice transcriber is not owned here.");
            voiceTranscriber = null;
            child.parent = null;
        }
    }
}

public sealed partial class VoiceTranscriber
{
    public partial VoiceTranscriber(Ensemble parent, string apiKey)
        : this(parent, apiKey, null, TimeSpan.FromSeconds(45)) { }

    public partial VoiceTranscriber(Ensemble parent, string apiKey, HttpMessageHandler transport)
        : this(parent, apiKey, transport ?? throw new ArgumentNullException(nameof(transport)),
            TimeSpan.FromSeconds(45)) { }

    internal partial VoiceTranscriber(Ensemble parent, string apiKey,
        HttpMessageHandler? transport, TimeSpan requestTimeout)
    {
        ArgumentNullException.ThrowIfNull(parent);
        if (string.IsNullOrWhiteSpace(apiKey) || apiKey.Length > 4096
            || apiKey.Any(c => c <= ' ' || c >= 127))
            throw new ArgumentException("An explicit transcription API key is required.", nameof(apiKey));
        if (requestTimeout <= TimeSpan.Zero || requestTimeout > TimeSpan.FromSeconds(45))
            throw new ArgumentOutOfRangeException(nameof(requestTimeout));
        this.apiKey = apiKey;
        this.requestTimeout = requestTimeout;
        this.transport = transport ?? new HttpClientHandler
        {
            AllowAutoRedirect = false,
            UseCookies = false
        };
        http = new HttpClient(this.transport, disposeHandler: false)
        {
            Timeout = Timeout.InfiniteTimeSpan
        };
        try { parent.AddVoiceTranscriber(this); }
        catch
        {
            http.Dispose();
            // An injected handler remains the caller's on refused construction.
            if (transport is null) this.transport.Dispose();
            lifetime.Dispose();
            throw;
        }
    }

    public partial Ensemble Parent() => parent
        ?? throw new ObjectDisposedException(nameof(VoiceTranscriber));

    internal partial string[] DiagnosticKeys()
    {
        lock (gate) { return apiKey.Length == 0 ? [] : [apiKey]; }
    }

    public partial Task<VoiceTranscript> TranscribeAsync(PcmVoiceAudio audio,
        CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(audio);
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(stopping, this);
            cancellation.ThrowIfCancellationRequested();
            if (active is { IsCompleted: false })
                throw new InvalidOperationException("Voice transcription is already busy.");
            // Publish the owned Task before its continuation can perform HTTP.
            var admitted = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            active = RequestAsync(audio, cancellation, admitted.Task);
            admitted.SetResult();
            return active;
        }
    }

    private partial async Task<VoiceTranscript> RequestAsync(PcmVoiceAudio audio,
        CancellationToken cancellation, Task admitted)
    {
        await admitted.ConfigureAwait(false);
        using var linked = CancellationTokenSource.CreateLinkedTokenSource(lifetime.Token, cancellation);
        linked.CancelAfter(requestTimeout);
        try
        {
            linked.Token.ThrowIfCancellationRequested();
            using var request = new HttpRequestMessage(HttpMethod.Post,
                "https://api.openai.com/v1/audio/transcriptions");
            request.Headers.Authorization = new AuthenticationHeaderValue("Bearer", apiKey);
            using var form = new MultipartFormDataContent();
            form.Add(new StringContent(Model), "model");
            form.Add(new StringContent("json"), "response_format");
            form.Add(new StringContent("false"), "stream");
            var file = new ByteArrayContent(audio.WaveBytes.ToArray());
            file.Headers.ContentType = new MediaTypeHeaderValue(PcmVoiceAudio.MimeType);
            form.Add(file, "file", "voice.wav");
            request.Content = form;
            using var response = await http.SendAsync(request, HttpCompletionOption.ResponseHeadersRead,
                linked.Token).ConfigureAwait(false);
            linked.Token.ThrowIfCancellationRequested();
            // Do not expose an error body or reason phrase: either can contain
            // reflected credentials or other provider-controlled private data.
            if (!response.IsSuccessStatusCode)
                throw new HttpRequestException($"Voice transcription returned HTTP {(int)response.StatusCode}.",
                    null, response.StatusCode);
            if (response.Content.Headers.ContentLength is long length && length > MaximumResponseBytes)
                throw new InvalidDataException("Voice transcription response is too large.");
            using Stream stream = await response.Content.ReadAsStreamAsync(linked.Token).ConfigureAwait(false);
            byte[] body = new byte[MaximumResponseBytes + 1];
            int count = 0;
            while (true)
            {
                int read = await stream.ReadAsync(body.AsMemory(count), linked.Token).ConfigureAwait(false);
                linked.Token.ThrowIfCancellationRequested();
                if (read == 0) break;
                count += read;
                if (count > MaximumResponseBytes)
                    throw new InvalidDataException("Voice transcription response is too large.");
            }
            var transcript = ParseTranscript(body.AsMemory(0, count), audio.Duration, apiKey);
            linked.Token.ThrowIfCancellationRequested();
            return transcript;
        }
        catch (OperationCanceledException) when (linked.IsCancellationRequested)
        {
            throw new OperationCanceledException("Voice transcription was cancelled.", linked.Token);
        }
        catch (Exception error) when (error.ToString().Contains(apiKey, StringComparison.Ordinal))
        {
            // The public failure can outlive this credential owner. Do not keep
            // a secret-bearing inner exception and hope later logging masks it.
            throw new InvalidOperationException("Voice transcription request failed.");
        }
    }

    public partial Task DestroyAsync()
    {
        lock (gate)
        {
            if (destruction is not null) return destruction;
            stopping = true;
            var admitted = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
            destruction = DestroyOwnedAsync(active, admitted.Task);
            admitted.SetResult();
            return destruction;
        }
    }

    public partial void Destroy() => DestroyAsync().GetAwaiter().GetResult();

    private partial async Task DestroyOwnedAsync(Task<VoiceTranscript>? work, Task admitted)
    {
        await admitted.ConfigureAwait(false);
        var failures = new List<Exception>();
        try { lifetime.Cancel(); }
        catch (Exception error) { failures.Add(error); }
        if (work is not null)
        {
            // Joining a failed/cancelled request is not a cleanup failure. Its
            // original returned Task still reports that outcome to its caller.
            try { await work.ConfigureAwait(false); }
            catch (Exception) { }
        }
        try { http.Dispose(); }
        catch (Exception error) { failures.Add(error); }
        try { transport.Dispose(); }
        catch (Exception error) { failures.Add(error); }
        try { lifetime.Dispose(); }
        catch (Exception error) { failures.Add(error); }
        // Cleanup errors can outlive this owner and its diagnostic key registry.
        // Remove secret-bearing inner exceptions before clearing the only key.
        for (int index = 0; index < failures.Count; index++)
        {
            if (failures[index].ToString().Contains(apiKey, StringComparison.Ordinal))
            {
                failures[index] = new InvalidOperationException("Voice transcription resource cleanup failed.");
            }
        }
        // Clear material and detach even if one resource's disposal failed.
        lock (gate)
        {
            apiKey = "";
            active = null;
        }
        Parent().RemoveVoiceTranscriber(this);
        if (failures.Count != 0)
            throw new AggregateException("Voice transcription cleanup failed.", failures);
    }
}

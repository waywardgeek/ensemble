using System.Text;
using System.Text.Json;

namespace Ensemble;

public sealed partial class McpStreamTransport
{
    public partial McpStreamTransport(Stream input, Stream output)
    {
        ArgumentNullException.ThrowIfNull(input);
        ArgumentNullException.ThrowIfNull(output);
        if (!input.CanRead || !output.CanWrite)
        {
            throw new ArgumentException("MCP requires readable input and writable output.");
        }
        this.input = input;
        this.output = output;
        reader = new StreamReader(input, new UTF8Encoding(false, true),
            detectEncodingFromByteOrderMarks: false, leaveOpen: true);
    }

    public partial async Task SendAsync(JsonElement message, CancellationToken cancellation)
    {
        using var linked = CancellationTokenSource.CreateLinkedTokenSource(cancellation, closing.Token);
        await sending.WaitAsync(linked.Token);
        try
        {
            ObjectDisposedException.ThrowIf(Volatile.Read(ref closed) != 0, this);
            // Serialization removes formatting newlines inside the envelope.
            // JSON string newlines remain escaped, preserving line framing.
            byte[] bytes = JsonSerializer.SerializeToUtf8Bytes(message);
            await output.WriteAsync(bytes, linked.Token);
            await output.WriteAsync(new byte[] { (byte)'\n' }, linked.Token);
            await output.FlushAsync(linked.Token);
        }
        finally
        {
            sending.Release();
        }
    }

    public partial async Task<JsonElement?> ReceiveAsync(CancellationToken cancellation)
    {
        using var linked = CancellationTokenSource.CreateLinkedTokenSource(cancellation, closing.Token);
        string? line = await reader.ReadLineAsync(linked.Token);
        if (line is null)
        {
            return null;
        }
        using var document = JsonDocument.Parse(line);
        return document.RootElement.Clone();
    }

    public partial async ValueTask DisposeAsync()
    {
        if (Interlocked.Exchange(ref closed, 1) != 0)
        {
            return;
        }
        closing.Cancel();
        // Disposing the stream also releases reads on boundaries that cannot
        // implement cancellation themselves. Do not wait for the send lock first.
        try
        {
            await input.DisposeAsync();
        }
        finally
        {
            if (!ReferenceEquals(input, output))
            {
                await output.DisposeAsync();
            }
        }
        // Reader/semaphore/CTS have no OS handle allocated here. They remain
        // usable by unwinding callers; disposing them early would race Release.
    }
}

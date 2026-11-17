using System.Globalization;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Engine
{
    // Only the supplied network read is inside this boundary. Parser callbacks,
    // diagnostic writes, and usage accounting must not become provider retries.
    private static partial async Task<T> ReadProviderAsync<T>(Func<Task<T>> read)
    {
        try
        {
            return await read();
        }
        catch (IOException)
        {
            throw new HttpRequestException("Provider body read failed.");
        }
    }

    private static partial TimeSpan? HeaderRetryDelay(HttpResponseMessage response)
    {
        var advice = response.Headers.RetryAfter;
        TimeSpan? delay = advice?.Delta;
        if (delay is null && advice?.Date is DateTimeOffset date)
        {
            delay = date - DateTimeOffset.UtcNow;
        }
        return delay is TimeSpan value && value >= TimeSpan.Zero
            && value.TotalMilliseconds <= uint.MaxValue - 1 ? value : null;
    }

    private static partial TimeSpan? GeminiRetryDelay(JsonObject frame)
    {
        // Google RPC RetryInfo is duration data, never an instruction or display text.
        // Unknown details and free-form English messages remain uninterpreted.
        if (frame["error"] is not JsonObject error || error["details"] is not JsonArray details)
        {
            return null;
        }
        TimeSpan? longest = null;
        foreach (var item in details)
        {
            if (item is not JsonObject detail
                || detail["@type"] is not JsonValue type || !type.TryGetValue<string>(out var name)
                || name != "type.googleapis.com/google.rpc.RetryInfo"
                || detail["retryDelay"] is not JsonValue duration
                || !duration.TryGetValue<string>(out var text) || !text.EndsWith('s')
                || !double.TryParse(text.AsSpan(0, text.Length - 1), NumberStyles.AllowDecimalPoint,
                    CultureInfo.InvariantCulture, out double seconds)
                || !double.IsFinite(seconds) || seconds < 0
                || seconds > (uint.MaxValue - 1) / 1000d)
            {
                continue;
            }
            var delay = TimeSpan.FromSeconds(seconds);
            if (longest is null || delay > longest)
            {
                longest = delay;
            }
        }
        return longest;
    }

    private static partial async Task<TimeSpan?> HttpRetryDelayAsync(Configuration config,
        HttpResponseMessage response, CancellationToken cancellation)
    {
        TimeSpan? header = HeaderRetryDelay(response);
        if (config.Vendor != Vendor.Gemini)
        {
            return header;
        }
        // Error bodies may echo credentials or prompts. Inspect at most 64 KiB,
        // never trace or retain their contents, and ignore oversized/malformed advice.
        const int limit = 64 * 1024;
        await using var stream = await ReadProviderAsync(() => response.Content.ReadAsStreamAsync(cancellation));
        byte[] bytes = new byte[limit + 1];
        int length = 0;
        while (length < bytes.Length)
        {
            int count = await ReadProviderAsync(() => stream.ReadAsync(bytes.AsMemory(length), cancellation).AsTask());
            if (count == 0)
            {
                break;
            }
            length += count;
        }
        if (length > limit)
        {
            return header;
        }
        try
        {
            if (JsonNode.Parse(bytes.AsSpan(0, length)) is JsonObject frame
                && GeminiRetryDelay(frame) is TimeSpan body && (header is null || body > header))
            {
                return body;
            }
        }
        catch (JsonException)
        {
            // Malformed advice does not change the already known HTTP failure.
        }
        return header;
    }
}

using System.Text;
using System.Text.Json;

namespace Ensemble;

public sealed partial class Engine
{
    public partial CacheLens? Lens()
    {
        lock (lensGate)
        {
            return cacheLens;
        }
    }

    internal partial void AddLens(CacheLens child)
    {
        lock (lensGate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            if (cacheLens is not null || child.parent is not null || child.destroyed)
            {
                throw new InvalidOperationException("Invalid cache lens registration");
            }
            cacheLens = child;
            child.parent = this;
        }
    }

    internal partial void RemoveLens(CacheLens child)
    {
        lock (lensGate)
        {
            if (!ReferenceEquals(cacheLens, child) || !ReferenceEquals(child.parent, this))
            {
                throw new InvalidOperationException("Cache lens is not owned here");
            }
            cacheLens = null;
            child.parent = null;
        }
    }
}

public sealed partial class CacheLens
{
    public partial CacheLens(Engine parent, string directory)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(directory);
        this.directory = Path.GetFullPath(directory);
        parent.AddLens(this);
    }

    public partial Engine Parent() => parent ?? throw new ObjectDisposedException(nameof(CacheLens));

    private void WriteCapture(string prefix, CacheCapture capture)
    {
        File.WriteAllBytes(Path.Combine(directory, prefix + ".wire.json"), capture.Wire);
        File.WriteAllBytes(Path.Combine(directory, prefix + ".body"), capture.Prompt);
        File.WriteAllBytes(Path.Combine(directory, prefix + ".options.json"), capture.Options);
    }

    internal partial CachePrediction? Begin(string body)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            CacheCapture current;
            try
            {
                current = Capture(Parent().Parent().Configuration().Vendor, Encoding.UTF8.GetBytes(body));
            }
            catch (JsonException)
            {
                // Rendered requests are valid JSON, but the public transport also
                // accepts caller-supplied bytes. Diagnostics must not become a
                // second admission policy or echo malformed private content.
                Parent().Parent().Parent().Log("Cache lens capture skipped (invalid JSON).");
                previous = null;
                return null;
            }
            int shared = 0;
            byte[] old = previous?.Prompt ?? [];
            while (shared < Math.Min(old.Length, current.Prompt.Length)
                && old[shared] == current.Prompt[shared])
            {
                shared++;
            }
            var prediction = new CachePrediction(++sequence, old.Length, current.Prompt.Length,
                shared, current.Prompt.Length == 0 ? 0 : (double)shared / current.Prompt.Length);
            try
            {
                Directory.CreateDirectory(directory);
                // Start each lens with a deliberate empty prior, not files from
                // an older process. These are best-effort diagnostics, not saves.
                WriteCapture("request.prev", previous ?? new CacheCapture([], [], []));
                WriteCapture("request", current);
                File.WriteAllText(Path.Combine(directory, "request.prediction.json"),
                    JsonSerializer.Serialize(prediction));
            }
            catch (Exception ex) when (ex is IOException or UnauthorizedAccessException)
            {
                Parent().Parent().Parent().Log("Cache lens capture failed (" + ex.GetType().Name + ").");
            }
            previous = current;
            return prediction;
        }
    }

    internal partial void Complete(CachePrediction prediction, Usage? usage)
    {
        lock (gate)
        {
            if (destroyed)
            {
                return;
            }
            long? input = usage?.Input;
            long? write = usage?.CacheWrite;
            long? read = usage?.CacheRead;
            double? actual = input is long i && write is long w && read is long r
                ? (i == 0 && w == 0 && r == 0 ? 0 : r / ((double)i + w + r))
                : null;
            // The immutable prediction belongs to this response, not to whichever
            // request most recently overwrote the diagnostic files.
            string report = JsonSerializer.Serialize(new
            {
                cache_lens = prediction.Sequence,
                previous_bytes = prediction.PreviousBytes,
                current_bytes = prediction.CurrentBytes,
                shared_bytes = prediction.SharedBytes,
                predicted_hit_rate = prediction.PredictedHitRate,
                actual_hit_rate = actual,
                fresh_input = input, cache_write = write, cache_read = read
            });
            Parent().Parent().Parent().Log(report);
        }
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
            previous = null;
            parent?.RemoveLens(this);
        }
    }
}

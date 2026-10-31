using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Recaller
{
    public partial Recaller(Agent parent)
    {
        ArgumentNullException.ThrowIfNull(parent);
        parent.AddRecaller(this);
    }

    public partial Agent Parent() => parent ?? throw new ObjectDisposedException(nameof(Recaller));

    public partial void Destroy()
    {
        Agent? owner;
        lock (gate)
        {
            if (destroyed) return;
            destroyed = true;
            indexes = ImmutableDictionary<RecallSource, (string, RecallIndex)>.Empty;
            owner = parent;
        }
        // Never acquire Agent ingress while holding this child's operation lock.
        // Removing an index does not remove any user's source files.
        owner?.RemoveRecaller(this);
    }

    internal partial ImmutableArray<RecallHit> Search(string query, double minimum,
        CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(query);
        if (!double.IsFinite(minimum) || minimum < 0)
        {
            throw new ArgumentOutOfRangeException(nameof(minimum));
        }
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            cancellation.ThrowIfCancellationRequested();
            var owner = Parent();
            var terms = RecallText.Tokens(query, query: true);
            if (terms.Length == 0) return [];
            string? home = owner.HasDataDirectory() ? owner.DataDirectory() : null;

            var next = indexes.ToBuilder();
            foreach (var source in new[] { RecallSource.Memory, RecallSource.Docs,
                         RecallSource.Handoffs, RecallSource.Learnings, RecallSource.Skills })
            {
                cancellation.ThrowIfCancellationRequested();
                ImmutableArray<(string File, string Text, string Hash)> files;
                if (source == RecallSource.Skills)
                {
                    files = owner.SkillManager()?.RecallFiles(cancellation) ?? [];
                }
                else if (home is null)
                {
                    files = [];
                }
                else if (source == RecallSource.Memory)
                {
                    files = owner.Memory().RecallFiles(cancellation);
                }
                else
                {
                    files = RecallText.ReadCorpus(owner.Sandbox(), home, source, cancellation);
                }
                string fingerprint = RecallText.CorpusFingerprint(files);
                if (!next.TryGetValue(source, out var current) || current.Fingerprint != fingerprint)
                {
                    var snippets = files.SelectMany(file => RecallText.Chunk(source, file.File, file.Text));
                    next[source] = (fingerprint, new RecallIndex(snippets));
                }
            }
            cancellation.ThrowIfCancellationRequested();
            indexes = next.ToImmutable();
            return indexes.OrderBy(pair => pair.Key)
                .SelectMany(pair => pair.Value.Index.Search(terms, minimum))
                .OrderByDescending(hit => hit.Score)
                .ThenBy(hit => hit.Snippet.Source)
                .ThenBy(hit => hit.Snippet.File, StringComparer.Ordinal)
                .ThenBy(hit => hit.Snippet.Header, StringComparer.Ordinal)
                .ToImmutableArray();
        }
    }
}

public sealed partial class Agent
{
    public partial Recaller Recaller()
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            return recaller ?? throw new InvalidOperationException("Agent has no recaller");
        }
    }

    internal partial void AddRecaller(Recaller child)
    {
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed || admissionClosed, this);
            if (child.destroyed || child.parent is not null || recaller is not null)
            {
                throw new InvalidOperationException("Recaller ownership conflict");
            }
            recaller = child;
            child.parent = this;
        }
    }

    internal partial void RemoveRecaller(Recaller child)
    {
        lock (ingress)
        {
            if (!ReferenceEquals(recaller, child) || !ReferenceEquals(child.parent, this))
            {
                throw new InvalidOperationException("Recaller does not belong to this Agent");
            }
            recaller = null;
            child.parent = null;
        }
    }
}

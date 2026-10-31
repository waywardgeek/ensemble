namespace Ensemble;

public sealed partial class Agent
{
    public partial int ContextByteBudget()
    {
        lock (ingress)
        {
            return contextByteBudget;
        }
    }

    public partial void SetContextByteBudget(int bytes)
    {
        ArgumentOutOfRangeException.ThrowIfNegative(bytes);
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            contextByteBudget = bytes;
        }
    }

    public partial (int ResultsBytes, int CallsBytes) ToolCompactionBudgets()
    {
        int target = ContextByteBudget();
        // Bytes are indivisible. A positive tiny target still enables both
        // bands; zero alone disables automatic byte-budget reduction.
        return target == 0 ? (0, 0) : (Math.Max(1, target / 8), Math.Max(1, target / 16));
    }

    private partial void CompactTools()
    {
        // Only the actor calls this, between complete tool batches and requests.
        // Record before using the reduced view: replay needs the selected ranges,
        // not today's budget or a rerun of this selection algorithm.
        var budgets = ToolCompactionBudgets();
        var resultCut = SelectToolCut(History().View(), presentedThrough, budgets.ResultsBytes, Redaction.Result);
        if (resultCut is not null)
        {
            Record(resultCut);
        }
        var callCut = SelectToolCut(History().View(), presentedThrough, budgets.CallsBytes, Redaction.Tool);
        if (callCut is not null)
        {
            Record(callCut);
        }
    }

    internal static partial Event? SelectToolCut(Context context, long presentedThrough, int budget, Redaction level)
    {
        if (budget <= 0 || presentedThrough <= 0)
        {
            return null;
        }
        if (level is not Redaction.Result and not Redaction.Tool)
        {
            throw new ArgumentException("Tool ladder requires result or tool redaction.", nameof(level));
        }

        // These tuples are temporary selection values, not persistent batches or
        // another transcript. Repeated IDs are legal after a batch has closed.
        var batches = new List<(long From, long To, long Bytes, bool Whole)>();
        var pending = new HashSet<string>(StringComparer.Ordinal);
        long start = 0;
        long bytes = 0;
        bool whole = false;
        foreach (var entry in context.Dialogue)
        {
            if (entry.Seq > presentedThrough)
            {
                break;
            }
            bool ambiguous = false;
            foreach (var part in entry.Parts)
            {
                if (part is ToolCallPart call)
                {
                    if (start == 0)
                    {
                        start = entry.Seq;
                    }
                    if (!pending.Add(call.CallID))
                    {
                        ambiguous = true;
                        break;
                    }
                }
                else if (part is ToolResultPart result && !pending.Remove(result.CallID))
                {
                    ambiguous = true;
                    break;
                }
            }
            if (ambiguous)
            {
                // Imported history can lack a result or contain orphan traffic.
                // Keep that suffix untouched rather than guess execution pairing.
                break;
            }
            if (start == 0)
            {
                continue;
            }
            bytes += entry.ToolBytes;
            whole |= entry.HasWholeToolResults;
            if (pending.Count == 0)
            {
                batches.Add((start, entry.Seq, bytes, whole));
                start = 0;
                bytes = 0;
                whole = false;
            }
        }

        // The second band can remove only a stub-only PREFIX. It must never
        // jump across a whole batch and delete unseen/unreduced payloads.
        var eligible = level == Redaction.Tool
            ? batches.TakeWhile(batch => !batch.Whole).ToArray()
            : batches.ToArray();
        long remaining = eligible.Where(batch => level == Redaction.Tool || batch.Whole)
            .Sum(batch => batch.Bytes);
        if (remaining <= 2L * budget)
        {
            return null;
        }

        long end = 0;
        foreach (var batch in eligible)
        {
            if (level == Redaction.Tool || batch.Whole)
            {
                remaining -= batch.Bytes;
            }
            end = batch.To;
            if (remaining <= budget)
            {
                break;
            }
        }
        return new Event(EventKind.Redacted, redactFrom: eligible[0].From, redactTo: end,
            level: level, reason: level == Redaction.Result ? "tool result budget" : "tool call budget");
    }
}

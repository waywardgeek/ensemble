using System.Collections.Immutable;
using System.Runtime.CompilerServices;
using System.Text.Json;

namespace Ensemble;

public sealed partial class Agent
{
    public partial async IAsyncEnumerable<ChildJoinUpdate> JoinChildrenAsync(
        IEnumerable<string>? childIDs, [EnumeratorCancellation] CancellationToken cancellation)
    {
        cancellation.ThrowIfCancellationRequested();
        // Enumerate caller input outside ingress: arbitrary enumerators are not
        // allowed to run while our ownership/admission relationship is locked.
        string[] ids = childIDs?.ToArray() ?? [];
        Ensemble root = Parent();
        Agent[] selected;
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(admissionClosed || destroyed, this);
            if (ids.Distinct(StringComparer.Ordinal).Count() != ids.Length)
            {
                throw new ArgumentException("Duplicate child ID in join.", nameof(childIDs));
            }
            selected = ids.Length == 0 ? children.ToArray() : ids.Select(id =>
                children.FirstOrDefault(child => child.childID == id)
                ?? throw new ArgumentException("No owned child has ID: " + id, nameof(childIDs))).ToArray();
        }

        var reportedQuestions = new HashSet<long>();
        while (true)
        {
            cancellation.ThrowIfCancellationRequested();
            var changes = new List<Task>();
            ImmutableArray<QuestionInfo> questions;
            lock (root.questionGate)
            {
                ObjectDisposedException.ThrowIf(Volatile.Read(ref admissionClosed)
                    || Volatile.Read(ref destroyed), this);
                questions = incomingQuestions.Where(question => selected.Contains(question.sender)
                    && !reportedQuestions.Contains(question.id)).Select(question => question.Snapshot()).ToImmutableArray();
                changes.Add(questionsChanged.Task);
            }
            if (!questions.IsEmpty)
            {
                foreach (QuestionInfo question in questions)
                {
                    reportedQuestions.Add(question.MessageID);
                }
                yield return new ChildJoinUpdate(questions, null);
                // The caller may answer, submit work or shut down a child while
                // handling this update. Reread authoritative state afterward.
                continue;
            }

            var payloads = ImmutableDictionary.CreateBuilder<string, JsonElement>(StringComparer.Ordinal);
            foreach (Agent child in selected)
            {
                lock (child.ingress)
                {
                    if (child.destroyed)
                    {
                        throw new InvalidOperationException("Child was destroyed during join: " + child.childID);
                    }
                    if (child.submittedPayload is { } payload)
                    {
                        payloads.Add(child.childID!, payload);
                    }
                    else if (child.childFailure is { } failure)
                    {
                        throw new InvalidOperationException("Child " + child.childID + " failed: " + failure);
                    }
                    else if (child.admissionClosed)
                    {
                        throw new InvalidOperationException("Child stopped without submission: " + child.childID);
                    }
                    // Snapshot and pulse are read under the same lock. New work
                    // invalidates a submission and signals this same pulse.
                    changes.Add(child.submissionChanged.Task);
                }
            }
            // A fact changed while we inspected another child. Do not publish
            // a snapshot already known to be obsolete (including a new question).
            if (changes.Any(change => change.IsCompleted))
            {
                continue;
            }
            if (payloads.Count == selected.Length)
            {
                yield return new ChildJoinUpdate([], payloads.ToImmutable());
                yield break;
            }
            // No worker per child, timer or lossy observation queue. The pulses
            // carry no data; on every wake all selected facts are read again.
            await Task.WhenAny(changes).WaitAsync(cancellation).ConfigureAwait(false);
        }
    }
}

using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Agent
{
    public partial Task<QuestionReply> SendMessageToParentAsync(string content, CancellationToken cancellation)
    {
        ArgumentNullException.ThrowIfNull(content);
        cancellation.ThrowIfCancellationRequested();
        Agent recipient = agentParent ?? this;
        AgentQuestion question;
        QuestionInfo notice;
        lock (Parent().questionGate)
        {
            question = new AgentQuestion(this, recipient, content);
            notice = question.Snapshot();
        }
        // No ingress acquisition under the graph lock. The root's human inbox
        // is not a message to its own model and must never self-answer.
        if (!notice.ForHuman)
        {
            recipient.NotifyQuestion(notice);
        }
        return question.WaitAsync(cancellation);
    }

    private partial void NotifyQuestion(QuestionInfo question)
    {
        lock (ingress)
        {
            if (admissionClosed || destroyed)
            {
                return; // Shutdown owns the already-admitted question edge.
            }
            actorTask ??= Task.Run(ActorLoop);
            mailbox.Writer.TryWrite(new Mail(MailKind.IncomingQuestion, Question: question));
        }
    }

    public partial ImmutableArray<QuestionInfo> PendingQuestions()
    {
        lock (Parent().questionGate)
        {
            return incomingQuestions.Select(question => question.Snapshot()).ToImmutableArray();
        }
    }

    public partial async Task<ImmutableArray<QuestionInfo>> WaitForQuestionsAsync(bool humanOnly,
        CancellationToken cancellation, IEnumerable<long>? excludedIDs)
    {
        HashSet<long> excluded = excludedIDs?.ToHashSet() ?? [];
        Ensemble owner = Parent();
        while (true)
        {
            cancellation.ThrowIfCancellationRequested();
            Task changed;
            lock (owner.questionGate)
            {
                ObjectDisposedException.ThrowIf(Volatile.Read(ref admissionClosed) || Volatile.Read(ref destroyed), this);
                ImmutableArray<QuestionInfo> pending = incomingQuestions.Select(question => question.Snapshot())
                    .Where(question => (!humanOnly || question.ForHuman)
                        && !excluded.Contains(question.MessageID)).ToImmutableArray();
                if (!pending.IsEmpty)
                {
                    return pending;
                }
                changed = questionsChanged.Task;
            }
            await changed.WaitAsync(cancellation).ConfigureAwait(false);
        }
    }

    public partial void RespondToQuestion(string response, string? childID, long? messageID, string humanAddress)
    {
        ArgumentNullException.ThrowIfNull(response);
        ArgumentException.ThrowIfNullOrWhiteSpace(humanAddress);
        if (childID is not null && humanAddress != "user")
        {
            throw new ArgumentException("A human address cannot override an Agent sender.", nameof(humanAddress));
        }
        // FindChild takes ingress: resolve that actual ownership relationship
        // before entering the graph lock, never from inside it.
        Agent sender = childID is null ? this : FindChild(childID)
            ?? throw new ArgumentException("No owned child has that ID.", nameof(childID));
        lock (Parent().questionGate)
        {
            ObjectDisposedException.ThrowIf(Volatile.Read(ref admissionClosed), this);
            AgentQuestion[] matches = incomingQuestions.Where(question => ReferenceEquals(question.sender, sender)
                && (messageID is null || question.id == messageID)).ToArray();
            if (matches.Length == 0)
            {
                throw new InvalidOperationException("No matching pending question.");
            }
            if (matches.Length != 1)
            {
                throw new InvalidOperationException("Several questions are pending; provide message_id.");
            }
            matches[0].Answer(response, humanAddress);
        }
    }

    internal partial void AddQuestion(AgentQuestion question, Agent recipient)
    {
        Ensemble owner = Parent();
        lock (owner.questionGate)
        {
            // Shutdown closes admission before taking this graph lock to drain
            // edges. A concurrent registration either precedes that drain or
            // refuses; it cannot create a new edge after the drain completed.
            ObjectDisposedException.ThrowIf(Volatile.Read(ref admissionClosed) || Volatile.Read(ref destroyed), this);
            ObjectDisposedException.ThrowIf(Volatile.Read(ref recipient.admissionClosed)
                || Volatile.Read(ref recipient.destroyed), recipient);
            if (!ReferenceEquals(recipient.parent, owner) || !ReferenceEquals(agentParent ?? this, recipient)
                || question.sender is not null || question.recipient is not null)
            {
                throw new InvalidOperationException("Invalid question ownership.");
            }
            long id = checked(recipient.nextQuestionID + 1);
            // Both owners and the ID are validated before either list changes.
            question.id = id;
            outgoingQuestions.Add(question);
            question.sender = this;
            recipient.AddIncomingQuestion(question);
        }
    }

    private partial void AddIncomingQuestion(AgentQuestion question)
    {
        // The sender holds our shared graph lock and has validated both ends.
        incomingQuestions.Add(question);
        question.recipient = this;
        nextQuestionID = question.id;
        SignalQuestionsChange();
    }

    internal partial void RemoveQuestion(AgentQuestion question)
    {
        lock (Parent().questionGate)
        {
            if (!ReferenceEquals(question.sender, this))
            {
                return;
            }
            question.recipient!.RemoveIncomingQuestion(question);
            outgoingQuestions.Remove(question);
            question.sender = null;
        }
    }

    private partial void RemoveIncomingQuestion(AgentQuestion question)
    {
        incomingQuestions.Remove(question);
        question.recipient = null;
        SignalQuestionsChange();
    }

    private partial void SignalQuestionsChange()
    {
        // Called only while the Ensemble question graph is locked. This pulse
        // carries no completion data and cannot be consumed by one observer.
        TaskCompletionSource previous = questionsChanged;
        questionsChanged = new(TaskCreationOptions.RunContinuationsAsynchronously);
        previous.TrySetResult();
    }

    private partial void CloseQuestions()
    {
        lock (Parent().questionGate)
        {
            foreach (AgentQuestion question in outgoingQuestions.Concat(incomingQuestions).Distinct().ToArray())
            {
                question.Destroy();
            }
            // Empty-inbox observers also need to learn about shutdown.
            SignalQuestionsChange();
        }
    }
}

internal sealed partial class AgentQuestion
{
    internal partial AgentQuestion(Agent sender, Agent recipient, string content)
    {
        this.content = content;
        sender.AddQuestion(this, recipient);
    }

    // The recipient holds the graph lock for snapshots, replies and destruction.
    internal partial QuestionInfo Snapshot() => new(id, sender!.ChildID(), content, ReferenceEquals(sender, recipient));

    internal partial Task<QuestionReply> WaitAsync(CancellationToken cancellation) => reply.Task.WaitAsync(cancellation);

    internal partial void Answer(string response, string humanAddress)
    {
        // Derive attribution from the actual recipient before RemoveQuestion
        // clears both links. A root's unnamed Agent address is the role "root";
        // its human inbox uses the separately supplied host address instead.
        string address = ReferenceEquals(sender, recipient)
            ? humanAddress
            : recipient!.ChildID() ?? recipient.Name ?? "root";
        var answer = new QuestionReply(address, response);
        sender!.RemoveQuestion(this);
        reply.TrySetResult(answer);
    }

    internal partial void Destroy()
    {
        if (sender is null)
        {
            return;
        }
        sender.RemoveQuestion(this);
        reply.TrySetCanceled();
    }
}

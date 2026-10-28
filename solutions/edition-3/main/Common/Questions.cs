using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Ensemble
{
    // Question edges span two Agents. One graph lock makes both relationships
    // atomic without acquiring parent/child ingress locks in opposite orders.
    // Code holding this lock must never enter an Agent's ingress lock.
    internal readonly object questionGate = new();
}

/// <summary>Value-like immutable public view; it owns no Agent or pending reply.</summary>
public sealed record QuestionInfo(long MessageID, string? AgentID, string Content, bool ForHuman);

public sealed partial class Agent
{
    private readonly List<AgentQuestion> outgoingQuestions = [];
    private readonly List<AgentQuestion> incomingQuestions = [];
    private long nextQuestionID;
    private TaskCompletionSource questionsChanged = new(TaskCreationOptions.RunContinuationsAsynchronously);

    /// <summary>Ask the actual Agent parent, or this root's human inbox. The
    /// question remains admitted if this caller stops waiting. Either owner's
    /// shutdown cancels the unresolved reply and removes both relationships.</summary>
    public partial Task<string> SendMessageToParentAsync(string content, CancellationToken cancellation = default);
    public partial ImmutableArray<QuestionInfo> PendingQuestions();
    /// <summary>Observe pending questions without consuming them. humanOnly is
    /// for the terminal's human sink; ordinary parents observe child questions.</summary>
    public partial Task<ImmutableArray<QuestionInfo>> WaitForQuestionsAsync(bool humanOnly = false,
        CancellationToken cancellation = default);
    /// <summary>Answer a child's question, or this root's human question when
    /// childID is null. Omitted messageID requires exactly one matching question.
    /// Wrong IDs and ambiguity refuse without consuming any question.</summary>
    public partial void RespondToQuestion(string response, string? childID = null, long? messageID = null);
    internal partial void AddQuestion(AgentQuestion question, Agent recipient);
    internal partial void RemoveQuestion(AgentQuestion question);
    private partial void AddIncomingQuestion(AgentQuestion question);
    private partial void RemoveIncomingQuestion(AgentQuestion question);
    private partial void SignalQuestionsChange();
    private partial void CloseQuestions();
}

/// <summary>A pending reply is an edge strongly owned by its asking and receiving
/// Agents, not a value. For a human question both owners are the root Agent: that
/// root is the host's inbox, not a fabricated Agent representing the person.
/// Answering or destruction removes both links; it never destroys either Agent.</summary>
internal sealed partial class AgentQuestion
{
    internal Agent? sender;
    internal Agent? recipient;
    internal long id;
    internal readonly string content;
    private readonly TaskCompletionSource<string> reply = new(TaskCreationOptions.RunContinuationsAsynchronously);
    internal partial AgentQuestion(Agent sender, Agent recipient, string content);
    internal partial QuestionInfo Snapshot();
    internal partial Task<string> WaitAsync(CancellationToken cancellation);
    internal partial void Answer(string response);
    internal partial void Destroy();
}

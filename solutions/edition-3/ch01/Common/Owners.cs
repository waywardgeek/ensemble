using System.Collections.Immutable;

namespace Ensemble;

/// <summary>Application root and strong owner of Agents. The caller owns
/// the diagnostic writer; destroying Ensemble does not close that writer.</summary>
public sealed partial class Ensemble
{
    // Borrowed output destination; Ensemble does not own its lifetime.
    private readonly TextWriter diagnostics;
    // Strong children: destruction cascades through this collection.
    private readonly List<Agent> agents;
    private bool destroyed;

    public partial Ensemble(TextWriter diagnostics);
    // A read-only snapshot prevents callers from bypassing relationship methods.
    public partial ImmutableArray<Agent> Agents();
    // Add/remove establish/detach both ends; invalid ownership leaves both intact.
    internal partial void AddAgent(Agent child);
    internal partial void RemoveAgent(Agent child);
    // Callers supply safe diagnostics, never raw remote bodies or credentials.
    public partial void Log(string safeMessage);
    public partial void Destroy();
}

/// <summary>Owns configuration, completed conversation pairs and one Engine.
/// Only Ensemble's relationship methods write the parent back-reference.</summary>
public sealed partial class Agent
{
    internal Ensemble? parent;
    internal bool destroyed;
    // Authoritative per-conversation settings and completed pairs.
    private readonly Configuration configuration;
    private readonly List<Message> messages;
    // Strong child; only AddEngine/RemoveEngine change this membership.
    private Engine? engine;

    public partial Agent(Ensemble parent, Configuration configuration);
    public partial Ensemble Parent();
    public partial Configuration Configuration();
    public partial ImmutableArray<Message> Messages();
    public partial Engine Engine();
    // One Engine at a time. Removal detaches but does not destroy the child.
    internal partial void AddEngine(Engine child);
    internal partial void RemoveEngine(Engine child);

    // Calls are serial in this chapter. The question is committed only together
    // with a successful answer; cancellation or failure leaves history unchanged.
    public partial Task<string> AskAsync(string question, CancellationToken cancellation = default);
    public partial void Destroy();
}

/// <summary>Agent's HTTP boundary and authoritative cumulative measured usage.
/// It reaches configuration and diagnostics through its actual owning Agent.</summary>
public sealed partial class Engine
{
    private const string SystemInstructions = "You are a helpful assistant.";
    internal Agent? parent;
    internal bool destroyed;
    // Owned transport is disposed during explicit logical destruction.
    private readonly HttpClient http;
    // Measured cumulative receipt; failed answers can still consume tokens.
    private Usage usage;

    public partial Engine(Agent parent);
    public partial Agent Parent();
    public partial Usage Usage();

    // Performs exactly one POST. The supplied messages are a temporary request,
    // not another stored conversation. Caller cancellation covers headers/body.
    public partial Task<string> ExchangeAsync(
        ImmutableArray<Message> requestMessages,
        CancellationToken cancellation = default);
    public partial void Destroy();
}

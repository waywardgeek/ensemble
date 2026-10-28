using System.Collections.Immutable;

namespace Ensemble;

public sealed partial class Skills
{
    public partial ImmutableArray<McpClient> Clients()
    {
        lock (gate)
        {
            return clients.ToImmutableArray();
        }
    }

    public partial void AddClient(McpClient child)
    {
        lock (gate)
        {
            if (destroyed || child.destroyed || child.skillOwner is not null
                || child.parent?.Parent() != Parent() || child.SkillName is null
                || !definitions.TryGetValue(child.SkillName, out var definition)
                || definition.McpServers.IsDefaultOrEmpty
                || !definition.McpServers.Any(server => server.Name == child.Name)
                || clients.Any(client => client.SkillName == child.SkillName && client.Name == child.Name))
            {
                throw new InvalidOperationException("Invalid skill connection relationship.");
            }
            clients.Add(child);
            child.skillOwner = this;
        }
    }

    public partial void RemoveClient(McpClient child)
    {
        lock (gate)
        {
            if (child.skillOwner != this || !clients.Contains(child))
            {
                throw new InvalidOperationException("Connection does not belong to this skill manager.");
            }
            clients.Remove(child);
            child.skillOwner = null;
        }
    }

    internal partial bool Grants(string skill, string tool)
    {
        return Parent().Configuration().AllowsSkill(skill)
            && Parent().Configuration().AllowsTool(tool)
            && !Pending(skill) && definitions.TryGetValue(skill, out var definition)
            && definition.Tools.Contains(tool, StringComparer.Ordinal);
    }

    private partial async Task<ToolResultPart> StopClientsAsync(string name, ToolResultPart result)
    {
        // Pending manuals remain, but their external runtime resources stop now.
        // A dependency is a distinct skill: unloading a dependent does not stop it.
        var owned = Clients().Where(client => client.SkillName == name).ToArray();
        await Task.WhenAll(owned.Select(client => client.DestroyAsync()));
        return result;
    }

    private partial async Task<ImmutableArray<string>> PrepareAsync(string name,
        ImmutableArray<SkillDefinition> additions, ImmutableDictionary<string, string> values,
        Mcp manager, string workspace, CancellationToken cancellation, bool initial)
    {
        var prepared = new List<McpClient>();
        bool committed = false;
        try
        {
            foreach (var skill in additions)
            {
                if (skill.McpServers.IsDefaultOrEmpty)
                {
                    continue;
                }
                foreach (var server in skill.McpServers)
                {
                    cancellation.ThrowIfCancellationRequested();
                    McpClient client;
                    if (server.Transport == "websocket")
                    {
                        // The host offers a connected browser, but only this
                        // skill's ordinary preparation/commit can grant tools.
                        client = manager.ClaimBrowser(this, skill.Name, server.Name);
                    }
                    else
                    {
                        IMcpTransport transport;
                        if (server.Transport == "http")
                        {
                            string? token = null;
                            if (server.AuthEnvironment is { } key)
                            {
                                lock (gate)
                                {
                                    mcpCredentials.TryGetValue(key, out token);
                                }
                                if (token is null)
                                {
                                    Parent().Parent().Log("MCP authentication environment variable is missing: " + key);
                                }
                            }
                            transport = new McpHttpTransport(server.Url!, token);
                        }
                        else
                        {
                            transport = new McpProcessTransport(server, workspace);
                        }
                        try
                        {
                            client = new McpClient(manager, server.Name, transport, this, skill.Name);
                        }
                        catch
                        {
                            await transport.DisposeAsync();
                            throw;
                        }
                    }
                    prepared.Add(client);
                    try
                    {
                        await client.ConnectAsync(cancellation);
                    }
                    catch (Exception error) when (error is IOException or System.Text.Json.JsonException
                        or InvalidOperationException)
                    {
                        // Preparation errors are refused loads. Do not apply
                        // this conversion to the actor's durable commit below.
                        throw new InvalidDataException("MCP skill server initialization failed.", error);
                    }
                }
            }
            var activation = new SkillActivation(this, name, additions, values,
                prepared.ToImmutableArray(), cancellation, initial);
            var result = await Parent().CommitSkillAsync(activation);
            committed = true;
            return result;
        }
        finally
        {
            if (!committed)
            {
                // Start every cleanup before joining: one broken peer cannot
                // prevent another owned process from receiving shutdown.
                await Task.WhenAll(prepared.Select(client => client.DestroyAsync()));
            }
        }
    }

    internal partial ImmutableArray<string> Commit(SkillActivation activation)
    {
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            activation.Cancellation.ThrowIfCancellationRequested();
            var current = activation.Initial
                ? ResolveStartup()
                : Resolve(activation.Name, initial: false);
            if (!current.SequenceEqual(activation.Additions))
            {
                throw new InvalidDataException("Skill activation changed while servers were connecting; load again.");
            }
            var agent = Parent();
            var local = agent.Tools().ToDictionary(tool => tool.Name(), StringComparer.Ordinal);
            var providers = new Dictionary<string, McpClient>(StringComparer.Ordinal);
            foreach (var client in activation.Clients)
            {
                if (client.SkillOwner() != this || client.parent != agent.Mcp())
                {
                    throw new InvalidDataException("Prepared skill connection is no longer attached.");
                }
                foreach (var definition in client.Catalog())
                {
                    if (!Grants(client.SkillName!, definition.Name))
                    {
                        continue; // Discovery is not a capability grant.
                    }
                    if (local.ContainsKey(definition.Name) || !providers.TryAdd(definition.Name, client))
                    {
                        throw new InvalidDataException("MCP tool name conflicts with an existing tool.");
                    }
                }
            }
            foreach (var skill in activation.Additions)
            {
                foreach (string tool in skill.Tools)
                {
                    bool suppliedHere = providers.TryGetValue(tool, out var provider)
                        && provider.SkillName == skill.Name;
                    if (!local.ContainsKey(tool) && !suppliedHere)
                    {
                        throw new InvalidDataException("Skill tool was not discovered: " + tool);
                    }
                }
            }

            var active = Active();
            var prospectiveNames = active.Select(skill => skill.Name)
                .Concat(activation.Additions.Select(skill => skill.Name));
            var prospectiveTools = active.SelectMany(skill => skill.Tools)
                .Concat(activation.Additions.SelectMany(skill => skill.Tools))
                .Concat(activation.Initial ? UniversalControls() : ImmutableArray<string>.Empty)
                .Distinct(StringComparer.Ordinal).Order(StringComparer.Ordinal).ToArray();
            var available = Available(prospectiveNames);
            var manuals = ImmutableArray.CreateBuilder<SkillPart>();
            foreach (var skill in activation.Additions)
            {
                bool primary = activation.Initial && skill.Name == primaryName;
                var grants = primary
                    ? skill.Tools.Concat(UniversalControls()).Distinct(StringComparer.Ordinal).ToImmutableArray()
                    : skill.Tools;
                string body = Render(skill.Body, prospectiveTools, available, activation.Variables);
                manuals.Add(new SkillPart(skill.Name, body, grants, primary));
            }
            var parts = manuals.ToImmutable();
            try
            {
                foreach (var client in activation.Clients)
                {
                    client.InstallTools();
                }
                // Single durable publication. No await separates installation
                // from this fact, so the actor cannot schedule half a closure.
                agent.History().Append(new Event(EventKind.SkillLoaded, actor: Actor.System, parts: parts));
            }
            catch
            {
                foreach (var client in activation.Clients)
                {
                    foreach (var tool in client.Tools())
                    {
                        tool.Destroy();
                    }
                }
                throw;
            }
            var previous = active.SelectMany(skill => skill.Tools).ToHashSet(StringComparer.Ordinal);
            var discovered = parts.SelectMany(skill => skill.Tools).Where(tool => !previous.Contains(tool) && !agent.IsAutomaticTool(tool));
            if (activation.Initial)
            {
                initialVariables = null;
            }
            return discovered.ToImmutableArray();
        }
    }
}

public sealed partial class Agent
{
    internal partial async Task<ImmutableArray<string>> CommitSkillAsync(SkillActivation activation)
    {
        var reply = new TaskCompletionSource<ImmutableArray<string>>(TaskCreationOptions.RunContinuationsAsynchronously);
        Enqueue(new Mail(MailKind.CommitSkill, SkillReply: reply, Activation: activation));
        // Cancellation is checked by the actor. Once queued, await a definite
        // outcome before the preparer decides whether it owns rollback cleanup.
        return await reply.Task;
    }
}

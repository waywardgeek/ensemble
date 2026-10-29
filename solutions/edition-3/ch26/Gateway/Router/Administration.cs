using System.Collections.Immutable;
using System.Text.Json;

namespace Ensemble.Gateway;

public sealed partial class Router
{
    internal partial RpcMessage? Request(RouterIdentity caller, RpcMessage request)
    {
        lock (gate)
        {
            RequireLive();
            RequireIdentity(caller);
            RpcMessage Error(int code, string text) => new(ID: request.ID, Error: new(code, text));
            if (request.Jsonrpc != "2.0") return Error(-32600, "Expected JSON-RPC 2.0.");
            object result;
            switch (request.Method)
            {
                case "notifications/initialized": return null;
                case "initialize":
                    result = new
                    {
                        protocolVersion = "2024-11-05", capabilities = new { tools = new { } },
                        serverInfo = new { name = "ensemble-gateway", version = "1" }
                    };
                    break;
                case "tools/list": result = new { tools = Schemas() }; break;
                case "tools/call":
                    string name;
                    JsonElement args;
                    try
                    {
                        JsonElement call = request.Params ?? throw new JsonException("Missing tool call.");
                        Arguments(call, "name", "arguments");
                        name = Text(call, "name");
                        args = call.GetProperty("arguments");
                    }
                    catch (Exception error) when (error is JsonException or ArgumentException or KeyNotFoundException)
                    {
                        return Error(-32602, error.Message);
                    }
                    // MCP tool refusal is an ordinary tool result, not an RPC
                    // protocol error. No partially accepted grant escapes it.
                    bool refused = false;
                    string text;
                    try { text = JsonSerializer.Serialize(Call(caller, name, args)); }
                    catch (Exception error) when (error is JsonException or ArgumentException or InvalidOperationException or FormatException or OverflowException)
                    {
                        refused = true;
                        text = error.Message;
                    }
                    result = new { content = new[] { new { type = "text", text } }, isError = refused };
                    break;
                default: return Error(-32601, "Unknown MCP method.");
            }
            return new RpcMessage(ID: request.ID, Result: JsonSerializer.SerializeToElement(result));
        }
    }

    private partial object Call(RouterIdentity caller, string name, JsonElement args)
    {
        switch (name)
        {
            case "send-message":
                Arguments(args, "to_agent", "content", "conversation_id", "response_mode");
                _ = Route(caller, new(Text(args, "to_agent"), Text(args, "content"),
                    Text(args, "conversation_id", optional: true), Text(args, "response_mode", optional: true)));
                return "queued";
            case "declare-contactable":
                Arguments(args, "groups");
                if (!caller.Capable("declare-contactable")) throw new InvalidOperationException("declare-contactable capability required.");
                caller.ReplaceGroups(GroupLink.Contact, ResolveGroups(Names(args, "groups")));
                return "contactability updated";
            case "register-child":
                Arguments(args, "agent_id", "public_key", "groups");
                string id = Text(args, "agent_id");
                ValidateName(id);
                if (identities.ContainsKey(id)) throw new ArgumentException("Child identity is already registered.");
                ImmutableArray<byte> publicKey = Keys.ParsePublicKey(Text(args, "public_key"));
                RouterGroup[] proposed = ResolveGroups(Names(args, "groups"));
                RouterGroup[] allowed = caller.Memberships();
                if (proposed.Any(group => !allowed.Contains(group)))
                    throw new InvalidOperationException("Child groups exceed parent permissions.");
                RouterIdentity child = new(this, id, publicKey, "", configured: false);
                try
                {
                    child.ReplaceGroups(GroupLink.Base, proposed);
                    _ = new DeliveryQueue(child);
                }
                catch { child.Destroy(); throw; }
                return "child registered";
            case "list-groups":
                Arguments(args);
                return new
                {
                    groups = groups.ToDictionary(pair => pair.Key, pair => pair.Value.Definition()),
                    memberships = caller.Memberships().Select(group => group.name).ToArray()
                };
            case "put-group":
            case "delete-group":
            case "set-groups":
                if (!caller.Capable("group-admin")) throw new InvalidOperationException("group-admin capability required.");
                return Admin(name, args);
            default: throw new ArgumentException("Unknown Gateway tool.");
        }
    }

    private partial object Admin(string name, JsonElement args)
    {
        if (name == "set-groups")
        {
            Arguments(args, "agent_id", "groups");
            RouterIdentity target = Identity(Text(args, "agent_id"));
            RouterGroup[] replacement = ResolveGroups(Names(args, "groups"), runtimeOnly: true);
            target.ReplaceGroups(GroupLink.Grant, replacement);
            return "memberships updated";
        }
        if (name == "put-group") Arguments(args, "name", "description", "capabilities", "trust");
        else Arguments(args, "name");
        string groupName = Text(args, "name");
        ValidateName(groupName);
        if (!groupName.StartsWith("runtime:", StringComparison.Ordinal))
            throw new ArgumentException("Config groups are immutable; use the runtime: namespace.");
        if (name == "put-group")
        {
            GroupDefinition definition = new(Names(args, "capabilities"), Text(args, "trust"), Text(args, "description", optional: true));
            ValidateGroup(definition);
            if (groups.TryGetValue(groupName, out RouterGroup? group)) group.Update(definition);
            else _ = new RouterGroup(this, groupName, definition, runtime: true);
            return "group updated";
        }
        Group(groupName).Destroy();
        return "group deleted";
    }

    internal partial Envelope Route(RouterIdentity caller, OutboundMessage message)
    {
        lock (gate)
        {
            RequireLive();
            RequireIdentity(caller);
            ArgumentNullException.ThrowIfNull(message);
            ArgumentNullException.ThrowIfNull(message.Content);
            ArgumentNullException.ThrowIfNull(message.ConversationID);
            ArgumentNullException.ThrowIfNull(message.ResponseMode);
            if (!caller.Capable("send-message")) throw new InvalidOperationException("send-message capability required.");
            RouterIdentity recipient = Identity(message.ToAgent);
            RouterGroup? selected = null;
            int strongest = 0;
            foreach (RouterGroup group in caller.Memberships())
            {
                if (!group.HasMember(recipient, GroupLink.Contact)) continue;
                int rank = group.Definition().Trust switch { "owner" => 3, "trusted" => 2, _ => 1 };
                if (rank > strongest) { selected = group; strongest = rank; }
            }
            if (selected is null) throw new InvalidOperationException("Sender groups are not accepted by recipient.");
            Source source = new("agent", selected.name, caller.id, caller.displayName, selected.Definition().Trust);
            Envelope envelope = recipient.Queue().Enqueue(source, message);
            recipient.Notify();
            return envelope;
        }
    }

    private static partial void Arguments(JsonElement args, params string[] allowed)
    {
        if (args.ValueKind != JsonValueKind.Object) throw new JsonException("Expected an argument object.");
        Wire.UniqueNames(args);
        foreach (JsonProperty property in args.EnumerateObject())
        {
            if (!allowed.Contains(property.Name, StringComparer.Ordinal))
                throw new JsonException("Unknown Gateway argument: " + property.Name);
        }
    }

    private static partial string Text(JsonElement args, string name, bool optional)
    {
        if (!args.TryGetProperty(name, out JsonElement value))
        {
            if (optional) return "";
            throw new JsonException("Missing argument: " + name);
        }
        if (value.ValueKind != JsonValueKind.String) throw new JsonException("Expected string: " + name);
        return value.GetString()!;
    }

    private static partial ImmutableArray<string> Names(JsonElement args, string name)
    {
        if (!args.TryGetProperty(name, out JsonElement value) || value.ValueKind != JsonValueKind.Array)
            throw new JsonException("Expected string array: " + name);
        var result = ImmutableArray.CreateBuilder<string>();
        foreach (JsonElement item in value.EnumerateArray())
        {
            if (item.ValueKind != JsonValueKind.String) throw new JsonException("Expected string array: " + name);
            result.Add(item.GetString()!);
        }
        return result.ToImmutable();
    }

    private static partial JsonElement Schemas()
    {
        object text = new { type = "string" };
        object names = new { type = "array", items = text };
        object Schema(string name, string description, Dictionary<string, object> properties, params string[] required) =>
            new { name, description, inputSchema = new { type = "object", properties, required, additionalProperties = false } };
        return JsonSerializer.SerializeToElement(new[]
        {
            Schema("send-message", "Route a message under the recipient's group policy.", new()
            { ["to_agent"] = text, ["content"] = text, ["conversation_id"] = text, ["response_mode"] = text }, "to_agent", "content"),
            Schema("declare-contactable", "Replace this identity's accepted sender groups.", new() { ["groups"] = names }, "groups"),
            Schema("register-child", "Register an unused X25519 identity with narrow-only groups.", new()
            { ["agent_id"] = text, ["public_key"] = text, ["groups"] = names }, "agent_id", "public_key", "groups"),
            Schema("list-groups", "Inspect group definitions and your effective memberships.", new()),
            Schema("put-group", "Create or edit a runtime: group; requires group-admin.", new()
            { ["name"] = text, ["description"] = text, ["capabilities"] = names, ["trust"] = text }, "name", "capabilities", "trust"),
            Schema("delete-group", "Delete a runtime: group and remove its grants and contacts.", new() { ["name"] = text }, "name"),
            Schema("set-groups", "Replace an agent's additional runtime memberships; requires group-admin.", new()
            { ["agent_id"] = text, ["groups"] = names }, "agent_id", "groups")
        });
    }
}

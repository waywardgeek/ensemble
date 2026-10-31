using Ensemble;
using System.Collections.Immutable;
using System.Text.Json.Nodes;

namespace EnsembleGui;

public static partial class Gui
{
    public static partial string PartID(string agentID, long responseID, int partID)
    {
        return $"{agentID}:r{responseID}:p{partID}";
    }

    public static partial JsonObject PartFrame(string agentID, string id, Part part, Actor role, long sequence)
    {
        var wire = JsonCodec.WritePart(part);
        wire["role"] = role == Actor.Agent ? "assistant" : role.ToString().ToLowerInvariant();
        // Opaque provider data remains intact. Thinking text is merely a
        // display projection, not a replacement for its signed native value.
        if (part is OpaquePart opaque)
        {
            wire["type"] = "thinking";
            var value = opaque.Data;
            string text = "";
            if (value.ValueKind == System.Text.Json.JsonValueKind.Object)
            {
                foreach (var key in new[] { "thinking", "text", "reasoning_content" })
                {
                    if (value.TryGetProperty(key, out var field) && field.ValueKind == System.Text.Json.JsonValueKind.String)
                    {
                        text = field.GetString() ?? "";
                        break;
                    }
                }
            }
            wire["text"] = text;
        }
        return new JsonObject
        {
            ["type"] = "part_final", ["agent_id"] = agentID, ["part_id"] = id,
            ["seq"] = sequence, ["part"] = wire
        };
    }

    public static partial Projection Project(HistorySnapshot snapshot, string agentID)
    {
        var frames = ImmutableArray.CreateBuilder<JsonObject>();
        var ids = ImmutableHashSet.CreateBuilder<string>();
        var calls = new Dictionary<string, string>(StringComparer.Ordinal);
        string state = "idle";
        string lastAnswer = "";
        bool paused = false;
        bool shutdown = false;
        var context = new Context();
        foreach (var fact in snapshot.Events)
        {
            context = History.Reduce(context, fact);
            if (fact.Kind == EventKind.AgentPaused)
            {
                paused = true;
            }
            else if (fact.Kind == EventKind.AgentResumed)
            {
                paused = false;
            }
            if (fact.Kind == EventKind.AgentShutdown)
            {
                shutdown = true;
            }
            string next = shutdown ? "shutdown"
                : paused ? "paused" : context.Phase.ToString().ToLowerInvariant();
            if (next != state)
            {
                frames.Add(new JsonObject
                {
                    ["type"] = "state_changed", ["agent_id"] = agentID, ["seq"] = fact.Seq,
                    ["from"] = state, ["to"] = next
                });
                state = next;
            }
            if (fact.Kind is EventKind.MessageReceived or EventKind.HintReceived or EventKind.ResponseEnded)
            {
                if (fact.Kind == EventKind.ResponseEnded)
                {
                    lastAnswer = string.Concat(fact.Parts.OfType<TextPart>().Select(part => part.Text));
                }
                for (int i = 0; i < fact.Parts.Length; i++)
                {
                    // Historical logs have no parser IDs. Their event sequence
                    // and offset give a stable, disjoint replay-only namespace.
                    string id = fact.ResponseID == 0
                        ? $"{agentID}:e{fact.Seq}:p{i}"
                        : PartID(agentID, fact.ResponseID, fact.PartIDs[i]);
                    var role = fact.Kind == EventKind.ResponseEnded ? Actor.Agent : Actor.Human;
                    frames.Add(PartFrame(agentID, id, fact.Parts[i], role, fact.Seq));
                    ids.Add(id);
                    if (fact.Parts[i] is ToolCallPart call)
                    {
                        calls[call.CallID] = id;
                    }
                }
            }
            else if (fact.Kind == EventKind.ToolStarted)
            {
                string id = calls.GetValueOrDefault(fact.CallID, $"{agentID}:e{fact.Seq}:call");
                calls[fact.CallID] = id;
                frames.Add(new JsonObject
                {
                    ["type"] = "tool_dispatched", ["agent_id"] = agentID, ["seq"] = fact.Seq,
                    ["part_id"] = id, ["call_id"] = fact.CallID, ["name"] = fact.Name,
                    ["input"] = JsonNode.Parse(fact.Args!.Value.GetRawText()),
                    ["job"] = fact.Job is null ? null : JsonNode.Parse(JsonCodec.WriteEvent(fact))?["tool"]?["job"]?.DeepClone()
                });
            }
            else if (fact.Kind is EventKind.ToolReturned or EventKind.ToolObserved)
            {
                string id = $"{agentID}:e{fact.Seq}:result";
                frames.Add(new JsonObject
                {
                    ["type"] = "tool_finished", ["agent_id"] = agentID, ["seq"] = fact.Seq,
                    ["part_id"] = id, ["tool_part_id"] = calls.GetValueOrDefault(fact.CallID),
                    ["call_id"] = fact.CallID, ["result"] = JsonCodec.Text(fact.Parts),
                    ["parts"] = new JsonArray(fact.Parts.Select(p => (JsonNode)JsonCodec.WritePart(p)).ToArray()),
                    ["is_error"] = fact.IsError,
                    ["job"] = fact.Job is null ? null : JsonNode.Parse(JsonCodec.WriteEvent(fact))?["tool"]?["job"]?.DeepClone()
                });
                ids.Add(id);
            }
            else if (fact.Kind is EventKind.ErrorOccurred or EventKind.TurnInterrupted or EventKind.JobKilled or EventKind.AgentShutdown)
            {
                frames.Add(new JsonObject
                {
                    ["type"] = "lifecycle", ["agent_id"] = agentID, ["seq"] = fact.Seq,
                    ["kind"] = JsonNode.Parse(JsonCodec.WriteEvent(fact))!["type"]!.GetValue<string>(), ["message"] = fact.Error.Length == 0 ? fact.Reason : fact.Error
                });
            }
            else if (fact.Kind == EventKind.TurnEnded)
            {
                frames.Add(new JsonObject
                {
                    ["type"] = "turn_ended", ["agent_id"] = agentID, ["seq"] = fact.Seq,
                    ["reason"] = fact.Reason,
                    ["text"] = fact.Reason == "completed" ? lastAnswer : "",
                    ["error"] = fact.Reason == "completed" ? "" : fact.Reason
                });
                calls.Clear();
                lastAnswer = "";
            }
        }
        return new Projection(frames.ToImmutable(), state, ids.ToImmutable());
    }
}

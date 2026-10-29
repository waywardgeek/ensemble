using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Engine
{
    private static partial bool IncompleteOutput(Configuration config, string body)
    {
        if (config.Surface == Surface.Responses)
        {
            // The Responses parser already requires terminal completed success.
            return false;
        }
        var root = JsonNode.Parse(body)!.AsObject();
        return config.Vendor switch
        {
            Vendor.OpenAI => root["choices"]?[0]?["finish_reason"]?.GetValue<string>()
                is not ("stop" or "tool_calls"),
            Vendor.Anthropic => root["stop_reason"]?.GetValue<string>()
                is not ("end_turn" or "tool_use"),
            Vendor.Gemini => root["candidates"]?[0]?["finishReason"]?.GetValue<string>() != "STOP",
            _ => true
        };
    }
}

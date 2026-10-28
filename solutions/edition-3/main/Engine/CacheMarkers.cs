using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Engine
{
    // These operate only on the disposable, final Anthropic request. The API
    // permits four explicit breakpoints: tools, SYSTEM, handoff, and tail here.
    // Assigning twice to the same block collapses coincident handoff/tail points.
    // No TTL means the provider's default five minutes, not a retention promise.
    private static void MarkCache(JsonObject block)
    {
        block["cache_control"] = new JsonObject { ["type"] = "ephemeral" };
    }

    private static bool CacheableBlock(JsonObject block)
    {
        string? type = block["type"]?.GetValue<string>();
        if (type == "text")
        {
            return block["text"]?.GetValue<string>().Length > 0;
        }
        // Thinking/redacted-thinking and unknown opaque blocks are not explicit
        // targets. Nested tool-result content is covered by its outer block.
        return type is "tool_use" or "tool_result" or "image" or "document";
    }
}

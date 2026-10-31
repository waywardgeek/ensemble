using System.Text.Json.Nodes;
using System.Text.Json;
namespace Ensemble;

public sealed partial class Engine
{
    public partial JsonElement SavedConfiguration(Context context)
    {
        // Reuse the actual request projection, but include only skill entries.
        // No media is read, ephemera consumed, or dialogue copied into config.
        var instructions = context.Dialogue.Where(entry => entry.Parts.Any(part => part is SkillPart));
        var wire = Render(new Context(instructions));
        var config = Parent().Configuration();
        string system = config.Vendor switch
        {
            Vendor.Anthropic => wire["system"] is JsonArray systemBlocks
                ? string.Concat(systemBlocks.Select(block => block!["text"]!.GetValue<string>()))
                : wire["system"]!.GetValue<string>(),
            Vendor.OpenAI => config.Surface == Surface.Responses
                ? wire["instructions"]!.GetValue<string>()
                : wire["messages"]![0]!["content"]!.GetValue<string>(),
            _ => wire["systemInstruction"]!["parts"]![0]!["text"]!.GetValue<string>()
        };
        var tools = new JsonArray();
        if (wire["tools"] is JsonArray declarations)
        {
            IEnumerable<JsonNode?> sources = declarations;
            if (config.Vendor == Vendor.Gemini)
            {
                sources = declarations[0]?["functionDeclarations"]?.AsArray() ?? new JsonArray();
            }
            if (config.Billing is BillingMode.ChatGptPlan or BillingMode.Codex)
            {
                sources = declarations[0]!["tools"]!.AsArray();
            }
            foreach (var declaration in sources)
            {
                var source = config.Surface == Surface.ChatCompletions ? declaration!["function"]! : declaration!;
                string schema = config.Vendor == Vendor.Anthropic ? "input_schema" : "parameters";
                tools.Add(new JsonObject
                {
                    ["name"] = source["name"]!.DeepClone(),
                    ["description"] = source["description"]!.DeepClone(),
                    ["schema"] = source[schema]!.DeepClone()
                });
            }
        }
        var result = new JsonObject
        {
            ["model"] = config.Model,
            ["vendor"] = config.Vendor.ToString().ToLowerInvariant(),
            ["surface"] = config.Surface.ToString(),
            ["billing"] = config.Billing.ToString(),
            ["reasoning_effort"] = config.ReasoningEffort,
            ["reasoning_summary"] = config.ReasoningSummary,
            ["require_tool_explanation"] = config.RequireToolExplanation,
            // Documentary policy only. Startup uses current host/Agent settings.
            ["stubs_tool_results"] = config.StubsToolResults,
            ["cache_breakpoints"] = config.CacheBreakpoints,
            ["prompt_cache_key"] = config.PromptCacheKey,
            ["context_budget_bytes"] = Parent().ContextByteBudget(),
            ["system_prompt"] = system,
            ["tools"] = tools
        };
        using var parsed = JsonDocument.Parse(result.ToJsonString());
        return parsed.RootElement.Clone();
    }
}

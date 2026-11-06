using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Engine
{
    internal partial async Task<ParsedResponse> JudgeRecallAsync(string model,
        string instructions, string input, CancellationToken cancellation)
    {
        var owner = Parent();
        Configuration running = owner.BeginEngineSend();
        try
        {
            var config = running.ForRecallJudge(model, instructions);
            List<Entry> entries = [new Entry(0, Actor.Human, [new TextPart(input)])];
            // Deliberately bypass Render(Context): it projects the main Agent's
            // skills, memory, tool grants, and checkpoints. Reuse only the native
            // provider message renderers with this tool-free frozen input.
            var imageBudget = new ImageRequestBudget();
            JsonObject wire = config.Surface switch
            {
                Surface.ChatCompletions => RenderOpenAI(entries, config, imageBudget),
                Surface.Messages => RenderAnthropic(entries, config, null, null, imageBudget),
                Surface.GenerateContent => RenderGemini(entries, config, imageBudget),
                Surface.Responses => RenderResponses(entries, config, imageBudget),
                _ => throw new NotSupportedException("Unsupported recall judge surface")
            };
            AddStreamingOptions(wire, config);
            // SendCore owns authentication, same-source plan retry, disposal,
            // parsing and cumulative accounting. There is no parallel transport.
            return await SendCoreAsync(config, wire.ToJsonString(), cancellation,
                onChunk: null, auxiliary: true);
        }
        finally
        {
            owner.EndEngineSend();
        }
    }

    // The same documented capability rule governs ordinary and judge requests.
    private static partial void AddStreamingOptions(JsonObject request, Configuration config)
    {
        if (!config.DisableStreaming && config.Vendor != Vendor.Gemini
            && (config.Surface == Surface.Responses
                || config.Features.Stream))
        {
            request["stream"] = true;
            if (config.Surface == Surface.ChatCompletions)
            {
                request["stream_options"] = new JsonObject { ["include_usage"] = true };
            }
        }
    }
}

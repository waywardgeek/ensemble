using System.Text.Json.Nodes;
using Ensemble;

namespace EnsembleGui;

public static partial class Gui
{
    public static partial JsonObject MeterFrame(Agent agent)
    {
        // One coherent snapshot supplies both latest and total values. Reading
        // restored History here would bill old sessions again after restart.
        MeterSnapshot meter = agent.Engine().Meter();
        Usage latest = meter.Latest?.Counts ?? new Usage(null, null, null, null);
        Usage totals = meter.Totals;
        Configuration configuration = agent.Configuration();
        var choices = new JsonArray();
        foreach (string model in Engine.ModelChoices(configuration))
        {
            choices.Add(model);
        }
        return new JsonObject
        {
            ["type"] = "usage",
            ["funding"] = configuration.FundingDescription(),
            ["selected_model"] = configuration.Model,
            ["model_choices"] = choices,
            ["agent_id"] = agent.Address() ?? "",
            ["usage"] = new JsonObject
            {
                ["input_tokens"] = latest.Input,
                ["cache_write_tokens"] = latest.CacheWrite,
                ["cache_read_tokens"] = latest.CacheRead,
                ["output_tokens"] = latest.Output,
                ["all_input"] = latest.AllInput(),
                ["totals"] = new JsonObject
                {
                    ["input_tokens"] = totals.Input,
                    ["cache_write_tokens"] = totals.CacheWrite,
                    ["cache_read_tokens"] = totals.CacheRead,
                    ["output_tokens"] = totals.Output
                }
            },
            ["total_cost_usd"] = meter.EstimatedUsd(),
            ["hit_rate"] = meter.HitRate()
        };
    }
}

using System.Text.Json.Nodes;
using Ensemble;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial JsonObject WorkingContextFrame(Agent agent)
    {
        var status = agent.WorkingContextStatus();
        var result = new JsonObject
        {
            ["vendor"] = status.Vendor.ToString(), ["model"] = status.Model,
            ["input_tokens"] = status.InputTokens, ["working_tokens"] = status.WorkingTokens,
            ["capacity_tokens"] = status.CapacityTokens, ["percent"] = status.Percent
        };
        if (status.Offering is { } offering) result["offering"] = JsonCodec.WriteModelIdentity(offering);
        return result;
    }

    public static partial JsonObject MeterFrame(Agent agent)
    {
        // One coherent snapshot supplies both latest and total values. Reading
        // restored History here would bill old sessions again after restart.
        MeterSnapshot meter = agent.Engine().Meter();
        Usage latest = meter.Latest?.Counts ?? new Usage(null, null, null, null);
        Usage totals = meter.Totals;
        Configuration configuration = agent.Configuration();
        var compatible = Engine.ModelChoices(configuration);
        var recommendations = new JsonArray();
        foreach (var model in Engine.RecommendedModels)
        {
            recommendations.Add(new JsonObject
            {
                ["vendor"] = model.Vendor.ToString(),
                ["id"] = model.Name,
                ["label"] = model.Label,
                ["enabled"] = model.Vendor == configuration.Vendor && compatible.Contains(model.Name)
            });
        }
        var choices = new JsonArray();
        foreach (string model in compatible)
        {
            choices.Add(model);
        }
        return new JsonObject
        {
            ["type"] = "usage",
            ["working_context"] = WorkingContextFrame(agent),
            ["funding"] = configuration.FundingDescription(),
            ["selected_model"] = configuration.Model,
            ["model_choices"] = choices,
            ["model_vendor"] = configuration.Vendor.ToString(),
            ["recommended_models"] = recommendations,
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

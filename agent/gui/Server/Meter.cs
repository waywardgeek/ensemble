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
        ModelSelectionStatus selection = agent.ModelSelectionStatus();
        var compatible = selection.LegacyChoices;
        var recommendations = new JsonArray();
        foreach (var model in Engine.RecommendedModels)
        {
            recommendations.Add(new JsonObject
            {
                ["vendor"] = model.Vendor.ToString(),
                ["id"] = model.Name,
                ["label"] = model.Label,
                ["enabled"] = model.Vendor == selection.Vendor && compatible.Contains(model.Name)
            });
        }
        var choices = new JsonArray();
        foreach (string model in compatible)
        {
            choices.Add(model);
        }
        var configured = new JsonArray();
        foreach (ModelSelectionChoice choice in selection.ConfiguredChoices)
        {
            configured.Add(new JsonObject
            {
                ["offering"] = JsonCodec.WriteModelIdentity(choice.Offering),
                ["vendor"] = choice.VendorLabel, ["endpoint"] = choice.EndpointID,
                ["model"] = choice.Model, ["label"] = choice.Label,
                ["funding"] = choice.Funding,
                ["serving_context_tokens"] = choice.ServingContextTokens,
                ["advertised_context_tokens"] = choice.AdvertisedContextTokens,
                ["enabled"] = choice.Available
            });
        }
        return new JsonObject
        {
            ["type"] = "usage",
            ["selected_offering"] = selection.Offering is { } identity
                ? JsonCodec.WriteModelIdentity(identity) : null,
            ["selection_available"] = selection.Available,
            ["configured_models"] = configured,
            ["working_context"] = WorkingContextFrame(agent),
            ["funding"] = selection.Funding,
            ["selected_model"] = selection.Model,
            ["model_choices"] = choices,
            ["model_vendor"] = selection.Vendor.ToString(),
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

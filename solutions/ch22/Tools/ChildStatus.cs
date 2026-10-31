using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Tool
{
    private static partial JsonObject ChildStatusValue(Agent child, AgentWorkStatus status)
    {
        return new JsonObject
        {
            ["agent_id"] = child.ChildID(),
            ["activity"] = status.Activity.ToString().ToLowerInvariant(),
            ["work_phase"] = status.Work is null ? null
                : JsonNamingPolicy.SnakeCaseLower.ConvertName(status.Work.Phase.ToString()),
            ["tool"] = status.Work?.Tool,
            ["detail"] = status.Work?.Detail,
            ["paused"] = status.Paused,
            ["submitted"] = status.Submitted,
            ["failure"] = status.Failure
        };
    }

    private static partial async Task<string> DetailedChildStatusAsync(Job job, JsonElement args,
        CancellationToken cancellation)
    {
        Agent owner = job.Parent().Parent();
        Agent child = owner.FindChild(String(args, "agent_id"))
            ?? throw new ArgumentException("No owned child has that ID.");
        try
        {
            AgentWorkStatus status = await child.WorkStatusAsync(cancellation);
            // These are observations, not a globally atomic barrier across the
            // actor and Engine. Capture each source once; never derive counts
            // from replay or keep another mutable meter on the parent or Tool.
            string model = child.Configuration().Model;
            MeterSnapshot meter = child.Engine().Meter();
            JsonObject report = ChildStatusValue(child, status);
            report["selected_model"] = model;
            report["session_usage"] = UsageValue(meter.Totals);
            report["last_response_usage"] = meter.Latest is null ? null : UsageValue(meter.Latest.Counts);
            report["cache_hit_rate"] = meter.HitRate();
            report["estimated_cost_usd"] = meter.EstimatedUsd();
            report["cost_is_estimate"] = true;
            var rows = new JsonArray();
            foreach (MeterEntry row in meter.Rows)
            {
                rows.Add(new JsonObject
                {
                    ["vendor"] = row.Identity.Vendor.ToString(),
                    ["model"] = row.Identity.Model,
                    ["service_tier"] = row.Identity.ServiceTier,
                    ["billing"] = row.Identity.Billing.ToString(),
                    ["usage"] = UsageValue(row.Counts)
                });
            }
            report["measured_models"] = rows;
            return report.ToJsonString();
        }
        catch (InvalidOperationException error)
        {
            // A concurrently destroyed child is no longer observable. This is
            // ordinary correlated tool feedback, not successful empty telemetry.
            throw new IOException(error.Message, error);
        }
    }

    private static partial JsonObject UsageValue(Usage counts)
    {
        return new JsonObject
        {
            ["input_tokens"] = counts.Input,
            ["output_tokens"] = counts.Output,
            ["cache_read_tokens"] = counts.CacheRead,
            ["cache_write_tokens"] = counts.CacheWrite
        };
    }
}

using System.Collections.Immutable;
using System.Text.Json;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class WorkingContextLimit
{
    public partial WorkingContextLimit(Vendor vendor, string model, long workingTokens,
        long? capacityTokens)
    {
        if (!Enum.IsDefined(vendor)) throw new ArgumentOutOfRangeException(nameof(vendor));
        if (string.IsNullOrWhiteSpace(model) || model.Length > 256 || model.Any(char.IsControl))
            throw new ArgumentException("A model name must contain 1–256 non-control characters", nameof(model));
        if (workingTokens <= 0) throw new ArgumentOutOfRangeException(nameof(workingTokens));
        if (capacityTokens is <= 0 || capacityTokens is long capacity && workingTokens > capacity)
            throw new ArgumentOutOfRangeException(nameof(capacityTokens), "Working tokens must not exceed capacity tokens");
        Vendor = vendor;
        Model = model;
        WorkingTokens = workingTokens;
        CapacityTokens = capacityTokens;
    }
}

public static partial class JsonCodec
{
    public static partial JsonArray WriteWorkingContextLimits(ImmutableArray<WorkingContextLimit> limits)
    {
        var result = new JsonArray();
        foreach (var limit in limits)
        {
            result.Add(new JsonObject
            {
                ["vendor"] = limit.Vendor.ToString(),
                ["model"] = limit.Model,
                ["working_tokens"] = limit.WorkingTokens,
                ["capacity_tokens"] = limit.CapacityTokens
            });
        }
        return result;
    }

    public static partial ImmutableArray<WorkingContextLimit> ReadWorkingContextLimits(JsonElement value)
    {
        if (value.ValueKind != JsonValueKind.Array || value.GetArrayLength() > 64)
            throw new InvalidDataException("Expected at most 64 model context limits");
        var result = ImmutableArray.CreateBuilder<WorkingContextLimit>();
        var identities = new HashSet<(Vendor, string)>();
        foreach (var row in value.EnumerateArray())
        {
            if (row.ValueKind != JsonValueKind.Object || !row.EnumerateObject().Select(field => field.Name)
                    .Order(StringComparer.Ordinal).SequenceEqual(new[] { "capacity_tokens", "model", "vendor", "working_tokens" }))
                throw new InvalidDataException("Invalid model context limit fields");
            if (row.GetProperty("vendor").ValueKind != JsonValueKind.String
                || !Enum.TryParse<Vendor>(row.GetProperty("vendor").GetString(), out var vendor)
                || !Enum.IsDefined(vendor)
                || row.GetProperty("vendor").GetString() != vendor.ToString()
                || row.GetProperty("model").ValueKind != JsonValueKind.String
                || row.GetProperty("working_tokens").ValueKind != JsonValueKind.Number
                || !row.GetProperty("working_tokens").TryGetInt64(out long working))
                throw new InvalidDataException("Invalid model context limit values");
            var capacityValue = row.GetProperty("capacity_tokens");
            long? capacity = null;
            if (capacityValue.ValueKind != JsonValueKind.Null)
            {
                if (capacityValue.ValueKind != JsonValueKind.Number || !capacityValue.TryGetInt64(out long parsed))
                    throw new InvalidDataException("Capacity must be whole input tokens or null");
                capacity = parsed;
            }
            var limit = new WorkingContextLimit(vendor, row.GetProperty("model").GetString()!, working, capacity);
            if (!identities.Add((limit.Vendor, limit.Model)))
                throw new InvalidDataException("Duplicate model context limit");
            result.Add(limit);
        }
        return result.ToImmutable();
    }
}

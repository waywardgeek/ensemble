using System.Text.Json;
using System.Text.Json.Nodes;
using Ensemble;

namespace EnsembleGui;

public static partial class Gui
{
    private static partial Model ResolvePreferredModel(Agent agent, JsonElement preference)
    {
        if (preference.ValueKind != JsonValueKind.Object)
            throw new ArgumentException("Invalid vendor/model preference.");
        var fields = new HashSet<string>(StringComparer.Ordinal);
        foreach (var field in preference.EnumerateObject())
            if (field.Name is not ("vendor" or "model") || !fields.Add(field.Name)
                || field.Value.ValueKind != JsonValueKind.String
                || string.IsNullOrWhiteSpace(field.Value.GetString()))
                throw new ArgumentException("Invalid vendor/model preference.");
        if (fields.Count != 2) throw new ArgumentException("Incomplete vendor/model preference.");
        var root = agent.Parent();
        var settings = root.GlobalSettings()
            ?? throw new InvalidOperationException("Application funding preferences are unavailable.");
        string vendorID = preference.GetProperty("vendor").GetString()!;
        var vendor = root.ModelVendors().SingleOrDefault(value => value.ID == vendorID)
            ?? throw new InvalidOperationException("The configured vendor is unavailable.");
        return settings.ResolveModel(vendor, preference.GetProperty("model").GetString()!);
    }

    private static partial JsonArray PreferredModelChoices(Agent agent, ModelSelectionStatus selection)
    {
        var result = new JsonArray();
        var root = agent.Parent();
        if (root.GlobalSettings() is not { } settings) return result;
        // This is display data only. Selection resolves the saved preference
        // again, then uses the normal Agent admission/save path.
        foreach (var group in selection.ConfiguredChoices.GroupBy(value =>
            (value.Offering.VendorID, value.Model)))
        {
            var first = group.First();
            bool available = false;
            string funding = "unavailable; choose explicit funding";
            try
            {
                var vendor = root.ModelVendors().Single(value => value.ID == group.Key.VendorID);
                Model resolved = settings.ResolveModel(vendor, group.Key.Model);
                var choice = group.FirstOrDefault(value => value.Offering == resolved.Identity());
                if (choice is not null)
                {
                    available = choice.Available;
                    funding = choice.Funding;
                }
            }
            catch (Exception error) when (error is ArgumentException or InvalidOperationException
                or FormatException or IOException or InvalidDataException or UnauthorizedAccessException)
            {
                // A broken preference must not hide explicit recovery choices
                // or expose a settings path or credential in a status frame.
            }
            result.Add(new JsonObject
            {
                ["preference"] = new JsonObject
                {
                    ["vendor"] = group.Key.VendorID, ["model"] = group.Key.Model
                },
                ["vendor"] = first.VendorLabel, ["label"] = first.Label,
                ["funding"] = funding, ["enabled"] = available
            });
        }
        return result;
    }
}

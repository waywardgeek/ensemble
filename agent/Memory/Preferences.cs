using System.Text;
using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Memory
{
    internal partial AgentSettings? ReadSettings()
    {
        lock (gate)
        {
            return AgentSettings.Read(Parent().Sandbox(), HomePath("settings.json"));
        }
    }

    internal partial void WriteSettings(AgentSettings settings)
    {
        lock (gate)
        {
            var document = new JsonObject
            {
                ["version"] = 1,
                ["memory"] = JsonCodec.WriteMemorySettings(settings.Memory),
                ["recall"] = JsonCodec.WriteRecallSettings(settings.Recall),
                ["working_context"] = JsonCodec.WriteWorkingContextLimits(settings.WorkingContext.IsDefault ? [] : settings.WorkingContext)
            };
            byte[] bytes = Encoding.UTF8.GetBytes(document.ToJsonString());
            if (bytes.Length > 16 * 1024)
            {
                throw new InvalidDataException("Agent settings exceed 16 KiB");
            }
            Parent().Sandbox().ReplaceFile(HomePath("settings.json"), bytes);
        }
    }
}

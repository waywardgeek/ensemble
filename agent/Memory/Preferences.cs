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
            settings.Write(Parent().Sandbox(), HomePath("settings.json"));
        }
    }
}

using System.Collections.Immutable;
namespace Ensemble;

public sealed partial class Agent
{
    public partial bool StartPersistence(PersistenceOptions options)
    {
        ArgumentNullException.ThrowIfNull(options);
        ArgumentException.ThrowIfNullOrWhiteSpace(options.Path);
        lock (ingress)
        {
            ObjectDisposedException.ThrowIf(destroyed || destroying, this);
            if (actorTask is not null || persistence is not null)
            {
                throw new InvalidOperationException("Configure persistence once before starting the Agent");
            }
            var normalized = global::Ensemble.History.NormalizePersistence(options);
            var currentSkills = History().View().Dialogue.SelectMany(entry => entry.Parts)
                .OfType<SkillPart>().ToImmutableArray();
            bool loaded = History().LoadSave(normalized.Path);
            if (loaded)
            {
                History().ConfigureRuntime(currentSkills);
            }
            persistence = normalized;
            return loaded;
        }
    }
}

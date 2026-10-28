namespace Ensemble;

public sealed partial class Sandbox
{
    public partial (Configuration Effective, string? Refusal) DeriveChild(Configuration? requested)
    {
        // Snapshot the immutable Agent configuration before taking the file
        // gate. Persistence/publication may enter Sandbox while holding Agent
        // ingress; acquiring those locks in reverse order could deadlock.
        Configuration inherited = Parent().Configuration();
        lock (gate)
        {
            ObjectDisposedException.ThrowIf(destroyed, this);
            Configuration source = requested ?? inherited;
            bool safeMode = inherited.SafeMode || source.SafeMode;
            bool webSearch = inherited.EnableWebSearch && source.EnableWebSearch;
            var refusals = new List<string>();
            if (inherited.SafeMode && !source.SafeMode)
            {
                refusals.Add("Child cannot disable inherited safe mode.");
            }
            if (!inherited.EnableWebSearch && source.EnableWebSearch)
            {
                refusals.Add("Child cannot enable inherited disabled web search.");
            }

            // Omission inherits the physical root retained by this owner, not
            // the original (possibly relative or symlinked) launch spelling.
            string childRoot = root;
            if (requested is not null && source.SandboxRoot.Length != 0)
            {
                try
                {
                    childRoot = root.Length == 0
                        ? CanonicalPath(Path.GetFullPath(source.SandboxRoot))
                        : Resolve(source.SandboxRoot);
                }
                catch (Exception error) when (error is IOException or UnauthorizedAccessException or ArgumentException)
                {
                    // Do not return the rejected name or leak its resolved
                    // target in the diagnostic. Every refusal retains a root
                    // that is no broader than this parent's boundary.
                    childRoot = root;
                    refusals.Add("Child root is invalid or outside the inherited boundary.");
                }
            }

            var effective = new Configuration(source, safeMode, webSearch, childRoot);
            string? refusal = refusals.Count == 0 ? null : string.Join(" ", refusals);
            return (effective, refusal);
        }
    }
}

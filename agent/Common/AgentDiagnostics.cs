namespace Ensemble;

public sealed partial class Agent
{
    private readonly object diagnosticGate = new();
    private TextWriter diagnosticWriter = TextWriter.Null;
    private TextWriter? apiWriter;
    private TextWriter? guiWriter;
    private bool ownsDiagnosticWriters;
    private bool diagnosticWritersClosed;

    // Tests/embedders may supply borrowed writers; normal hosts use owned files.
    public partial void AttachDiagnostics(TextWriter diagnostics, TextWriter? apiTrace = null,
        TextWriter? guiTrace = null);
    public partial void StartDiagnostics(string? apiPath = null, string? guiPath = null);
    public partial void Log(string message);
    internal partial void Trace(string direction, string payload,
        IReadOnlyList<string>? exchangeSecrets = null);
    public partial void TraceGui(string direction, string payload);
    private partial void CloseDiagnostics();
}

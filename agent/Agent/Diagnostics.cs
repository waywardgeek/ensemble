using System.Text.Json.Nodes;

namespace Ensemble;

public sealed partial class Agent
{
    public partial void AttachDiagnostics(TextWriter diagnostics, TextWriter? apiTrace, TextWriter? guiTrace)
    {
        ArgumentNullException.ThrowIfNull(diagnostics);
        lock (diagnosticGate)
        {
            ObjectDisposedException.ThrowIf(diagnosticWritersClosed, this);
            if (ownsDiagnosticWriters)
            {
                throw new InvalidOperationException("Cannot replace owned diagnostic files.");
            }
            diagnosticWriter = diagnostics;
            apiWriter = apiTrace;
            guiWriter = guiTrace;
        }
    }

    public partial void StartDiagnostics(string? apiPath, string? guiPath)
    {
        lock (diagnosticGate)
        {
            ObjectDisposedException.ThrowIf(diagnosticWritersClosed, this);
            if (ownsDiagnosticWriters) return;
            var opened = new List<TextWriter>();
            try
            {
                string[] paths = [Path.Combine(DataDirectory(), "debug.log"),
                    apiPath ?? Path.Combine(DataDirectory(), "api.log"),
                    guiPath ?? Path.Combine(DataDirectory(), "gui.log")];
                foreach (string path in paths)
                {
                    opened.Add(new StreamWriter(Sandbox().OpenFile(
                        path, FileMode.Create, FileAccess.Write))
                        { AutoFlush = true });
                }
                diagnosticWriter = opened[0];
                apiWriter = opened[1];
                guiWriter = opened[2];
                ownsDiagnosticWriters = true;
            }
            catch (Exception openingFailure)
            {
                // A later file may fail after earlier writers opened. Attempt
                // every close without replacing the original opening failure.
                List<Exception> failures = [openingFailure];
                foreach (TextWriter writer in opened)
                {
                    try { writer.Dispose(); }
                    catch (Exception cleanupFailure) { failures.Add(cleanupFailure); }
                }
                if (failures.Count > 1)
                {
                    throw new AggregateException("Diagnostic files could not be opened and cleaned up.", failures);
                }
                throw;
            }
        }
    }

    public partial void Log(string message)
    {
        string redacted = Parent().RedactDiagnostic(message);
        lock (diagnosticGate)
        {
            if (diagnosticWritersClosed) return;
            diagnosticWriter.WriteLine($"{DateTimeOffset.UtcNow:O} {redacted}");
            diagnosticWriter.Flush();
        }
    }

    internal partial void Trace(string direction, string payload, IReadOnlyList<string>? exchangeSecrets)
    {
        var root = Parent();
        var record = new JsonObject
        {
            ["direction"] = root.RedactDiagnostic(direction, exchangeSecrets),
            ["agent"] = Name is null ? null : root.RedactDiagnostic(Name, exchangeSecrets),
            ["payload"] = root.RedactDiagnostic(payload, exchangeSecrets)
        }.ToJsonString();
        lock (diagnosticGate)
        {
            if (diagnosticWritersClosed || apiWriter is null) return;
            apiWriter.WriteLine(record);
            apiWriter.Flush();
        }
    }

    public partial void TraceGui(string direction, string payload)
    {
        string redacted = Parent().RedactDiagnostic($"{direction} {payload}");
        lock (diagnosticGate)
        {
            if (diagnosticWritersClosed || guiWriter is null) return;
            guiWriter.WriteLine($"{DateTimeOffset.UtcNow:O} {redacted}");
            guiWriter.Flush();
        }
    }

    private partial void CloseDiagnostics()
    {
        lock (diagnosticGate)
        {
            if (diagnosticWritersClosed) return;
            diagnosticWritersClosed = true;
            if (!ownsDiagnosticWriters) return;
            List<Exception> failures = [];
            foreach (TextWriter writer in new[] { diagnosticWriter, apiWriter, guiWriter }
                .OfType<TextWriter>().Distinct())
            {
                try { writer.Dispose(); }
                catch (Exception error) { failures.Add(error); }
            }
            if (failures.Count != 0) throw new AggregateException("Diagnostic writer cleanup failed.", failures);
        }
    }
}

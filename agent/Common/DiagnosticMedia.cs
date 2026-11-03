using System.Text.Json.Nodes;
using System.Text.RegularExpressions;

namespace Ensemble;

// A pure projection of a diagnostic copy. Never used to build provider requests,
// History, blobs, persistence, or transcripts.
internal static partial class DiagnosticMedia
{
    private static readonly Regex DataUrls = new(
        @"data:(?<mime>(?:image|audio)/[a-z0-9.+-]+);base64,(?<data>[a-z0-9+/=\r\n]+)",
        RegexOptions.IgnoreCase | RegexOptions.CultureInvariant | RegexOptions.NonBacktracking,
        TimeSpan.FromSeconds(2));
    internal static partial string Project(string text);
    private static partial void Visit(JsonNode? node, string? name);
    private static partial string? Text(JsonNode? node);
    private static partial string Summary(string mime, string encoded);
    private static partial string ProjectUrls(string text);
}

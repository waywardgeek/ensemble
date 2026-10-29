using System.Text;
using System.Text.Json;

namespace Ensemble;

public sealed partial class CacheLens
{
    public static partial byte[] StripMarkers(byte[] wire)
    {
        var reader = new Utf8JsonReader(wire);
        var cuts = new List<(int Start, int End)>();
        if (!reader.Read())
        {
            throw new JsonException("Empty request");
        }
        Scan(ref reader, "", cuts);
        if (reader.Read())
        {
            throw new JsonException("Multiple request roots");
        }
        // Copy every byte outside marker spans. Re-serialization would change
        // escapes, whitespace, number spelling, and thus the prefix being tested.
        using var output = new MemoryStream();
        int offset = 0;
        foreach (var cut in cuts.OrderBy(cut => cut.Start))
        {
            if (cut.Start < offset)
            {
                continue; // An outer removed marker already contains this span.
            }
            output.Write(wire.AsSpan(offset, cut.Start - offset));
            offset = cut.End;
        }
        output.Write(wire.AsSpan(offset));
        return output.ToArray();
    }

    private static bool MarkerLocation(string path)
    {
        if (path.Length == 0)
        {
            return true; // Provider-level automatic cache policy, if supplied.
        }
        var segments = path.Split('/');
        if (segments.Length == 3 && segments[1] is "system" or "tools")
        {
            return int.TryParse(segments[2], out _);
        }
        if (segments.Length < 5 || segments[1] != "messages"
            || !int.TryParse(segments[2], out _))
        {
            return false;
        }
        // Typed message blocks and nested tool-result content are marker sites;
        // input/schema/properties are not. This protects user-owned JSON keys.
        for (int i = 3; i < segments.Length; i += 2)
        {
            if (i + 1 >= segments.Length || segments[i] != "content"
                || !int.TryParse(segments[i + 1], out _))
            {
                return false;
            }
        }
        return true;
    }

    private static void Scan(ref Utf8JsonReader reader, string path, List<(int Start, int End)> cuts)
    {
        if (reader.TokenType == JsonTokenType.StartArray)
        {
            int index = 0;
            while (reader.Read() && reader.TokenType != JsonTokenType.EndArray)
            {
                Scan(ref reader, path + "/" + index++, cuts);
            }
            return;
        }
        if (reader.TokenType != JsonTokenType.StartObject)
        {
            return;
        }
        var fields = new List<(int Start, int End, bool Remove)>();
        while (reader.Read() && reader.TokenType != JsonTokenType.EndObject)
        {
            int start = checked((int)reader.TokenStartIndex);
            string name = reader.GetString()!;
            bool remove = name == "cache_control" && MarkerLocation(path);
            if (!reader.Read())
            {
                throw new JsonException("Missing property value");
            }
            if (remove)
            {
                reader.Skip();
            }
            else
            {
                Scan(ref reader, path + "/" + name, cuts);
            }
            fields.Add((start, checked((int)reader.BytesConsumed), remove));
        }
        // Remove runs together so duplicate marker fields do not leave dangling
        // commas. Keep the untouched fields' lexical bytes, including their order.
        for (int i = 0; i < fields.Count; i++)
        {
            if (!fields[i].Remove)
            {
                continue;
            }
            int first = i;
            while (i + 1 < fields.Count && fields[i + 1].Remove)
            {
                i++;
            }
            int start = first > 0 ? fields[first - 1].End : fields[first].Start;
            int end = first == 0 && i + 1 < fields.Count ? fields[i + 1].Start : fields[i].End;
            cuts.Add((start, end));
        }
    }

    public static partial CacheCapture Capture(Vendor vendor, byte[] wire)
    {
        byte[] stripped = StripMarkers(wire);
        using var document = JsonDocument.Parse(stripped);
        if (document.RootElement.ValueKind != JsonValueKind.Object)
        {
            throw new JsonException("Request must be an object");
        }
        string[] order = vendor switch
        {
            Vendor.Anthropic => ["tools", "system", "messages"],
            Vendor.OpenAI => document.RootElement.TryGetProperty("input", out _)
                ? ["tools", "instructions", "input"] : ["tools", "messages"],
            Vendor.Gemini => ["systemInstruction", "tools", "contents"],
            _ => throw new ArgumentException("Unknown provider", nameof(vendor))
        };
        // Read raw property spans instead of serializing a DOM. Only the outer
        // diagnostic envelope is new; each selected property's bytes are exact.
        var spans = new List<(string Name, int Start, int End)>();
        var reader = new Utf8JsonReader(stripped);
        reader.Read();
        while (reader.Read() && reader.TokenType != JsonTokenType.EndObject)
        {
            string name = reader.GetString()!;
            int start = checked((int)reader.TokenStartIndex);
            reader.Read();
            reader.Skip();
            spans.Add((name, start, checked((int)reader.BytesConsumed)));
        }
        byte[] Assemble(IEnumerable<(string Name, int Start, int End)> fields)
        {
            using var output = new MemoryStream();
            output.WriteByte((byte)'{');
            bool first = true;
            foreach (var field in fields)
            {
                if (!first)
                {
                    output.WriteByte((byte)',');
                }
                output.Write(stripped.AsSpan(field.Start, field.End - field.Start));
                first = false;
            }
            output.WriteByte((byte)'}');
            return output.ToArray();
        }
        return new CacheCapture((byte[])wire.Clone(),
            Assemble(order.SelectMany(name => spans.Where(field => field.Name == name))),
            Assemble(spans.Where(field => !order.Contains(field.Name, StringComparer.Ordinal))));
    }
}

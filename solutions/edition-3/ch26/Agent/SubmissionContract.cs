using System.Globalization;
using System.Numerics;
using System.Text;
using System.Text.Json;
using System.Text.RegularExpressions;

namespace Ensemble;

public sealed partial class SubmissionContract
{
    public partial SubmissionContract(JsonElement? schema)
    {
        legacy = schema is null;
        var supplied = schema ?? JsonSerializer.SerializeToElement(new
        {
            type = "object",
            properties = new { result = new { type = "string" } },
            required = new[] { "result" }
        });
        schemaBytes = Admit(supplied, 65_536, 2_048);
        CheckSchema(supplied, root: true);
        this.schema = supplied.Clone();
    }

    public partial JsonElement Schema() => schema.Clone();

    public partial JsonElement Validate(string result)
    {
        ArgumentNullException.ThrowIfNull(result);
        return Validate(JsonSerializer.SerializeToElement(new { result }));
    }

    public partial JsonElement Validate(JsonElement payload)
    {
        var payloadBytes = Admit(payload, 1_048_576, 65_536);
        if ((long)schemaBytes * payloadBytes > 67_108_864)
        {
            throw new ArgumentException("Submission schema/payload combined budget exceeded");
        }
        CheckValue(schema, payload, "$submission");
        if (legacy)
        {
            return JsonSerializer.SerializeToElement(new { result = payload.GetProperty("result").GetString() });
        }
        return payload.Clone();
    }

    private static partial int Admit(JsonElement value, int byteLimit, int nodeLimit)
    {
        if (value.ValueKind == JsonValueKind.Undefined)
        {
            throw new ArgumentException("Submission JSON is undefined");
        }
        if (Encoding.UTF8.GetByteCount(value.GetRawText()) > byteLimit)
        {
            throw new ArgumentException("Submission JSON byte limit exceeded");
        }
        var nodes = 0;
        void Visit(JsonElement item, int depth)
        {
            if (depth > 64 || ++nodes > nodeLimit)
            {
                throw new ArgumentException("Submission JSON depth/node limit exceeded");
            }
            if (item.ValueKind == JsonValueKind.Object)
            {
                var names = new HashSet<string>(StringComparer.Ordinal);
                foreach (var property in item.EnumerateObject())
                {
                    if (!names.Add(property.Name))
                    {
                        throw new ArgumentException("Submission JSON has duplicate fields");
                    }
                    Visit(property.Value, depth + 1);
                }
            }
            else if (item.ValueKind == JsonValueKind.Array)
            {
                foreach (var element in item.EnumerateArray())
                {
                    Visit(element, depth + 1);
                }
            }
            else if (item.ValueKind == JsonValueKind.Number)
            {
                _ = SubmissionNumber.Parse(item);
            }
        }
        Visit(value, 0);
        // Both the supplied representation and our owned serialization must fit.
        // Serialization preserves number lexemes; it never passes through double.
        var canonicalBytes = JsonSerializer.SerializeToUtf8Bytes(value).Length;
        if (canonicalBytes > byteLimit)
        {
            throw new ArgumentException("Submission canonical JSON byte limit exceeded");
        }
        return canonicalBytes;
    }

    private static partial void CheckSchema(JsonElement node, bool root)
    {
        if (!root && node.ValueKind is JsonValueKind.True or JsonValueKind.False)
        {
            return;
        }
        if (node.ValueKind != JsonValueKind.Object)
        {
            throw new ArgumentException("Submission schema must be an object or nested boolean");
        }
        if (root && (!node.TryGetProperty("type", out var rootType) ||
            rootType.ValueKind != JsonValueKind.String || rootType.GetString() != "object"))
        {
            throw new ArgumentException("Submission root schema must declare type object");
        }
        foreach (var field in node.EnumerateObject())
        {
            var value = field.Value;
            switch (field.Name)
            {
                case "$schema":
                    if (value.ValueKind != JsonValueKind.String ||
                        value.GetString() != "https://json-schema.org/draft/2020-12/schema")
                    {
                        throw new ArgumentException("Unsupported submission schema dialect");
                    }
                    break;
                case "type":
                    var types = value.ValueKind == JsonValueKind.Array
                        ? value.EnumerateArray().ToArray() : new[] { value };
                    var names = new HashSet<string>(StringComparer.Ordinal);
                    foreach (var type in types)
                    {
                        if (type.ValueKind != JsonValueKind.String ||
                            type.GetString() is not ("object" or "array" or "string" or "number" or "integer" or "boolean" or "null") ||
                            !names.Add(type.GetString()!))
                        {
                            throw new ArgumentException("Invalid submission schema type");
                        }
                    }
                    if (types.Length == 0)
                    {
                        throw new ArgumentException("Empty submission schema type array");
                    }
                    break;
                case "properties":
                    if (value.ValueKind != JsonValueKind.Object)
                    {
                        throw new ArgumentException("Schema properties must be an object");
                    }
                    foreach (var property in value.EnumerateObject())
                    {
                        CheckSchema(property.Value, root: false);
                    }
                    break;
                case "required":
                    if (value.ValueKind != JsonValueKind.Array)
                    {
                        throw new ArgumentException("Schema required must be an array");
                    }
                    var required = new HashSet<string>(StringComparer.Ordinal);
                    foreach (var item in value.EnumerateArray())
                    {
                        if (item.ValueKind != JsonValueKind.String || !required.Add(item.GetString()!))
                        {
                            throw new ArgumentException("Schema required must contain distinct strings");
                        }
                    }
                    break;
                case "additionalProperties":
                case "items":
                    CheckSchema(value, root: false);
                    break;
                case "minProperties": case "maxProperties":
                case "minItems": case "maxItems":
                case "minLength": case "maxLength":
                    var bound = SubmissionNumber.Parse(value);
                    if (!bound.IsInteger() || bound.Compare(default) < 0)
                    {
                        throw new ArgumentException("Schema size bound must be a nonnegative integer");
                    }
                    break;
                case "uniqueItems":
                    if (value.ValueKind is not (JsonValueKind.True or JsonValueKind.False))
                    {
                        throw new ArgumentException("Schema uniqueItems must be boolean");
                    }
                    break;
                case "minimum": case "maximum":
                case "exclusiveMinimum": case "exclusiveMaximum":
                case "multipleOf":
                    var number = SubmissionNumber.Parse(value);
                    if (field.Name == "multipleOf" && number.Compare(default) <= 0)
                    {
                        throw new ArgumentException("Schema multipleOf must be positive");
                    }
                    break;
                case "pattern":
                    if (value.ValueKind != JsonValueKind.String)
                    {
                        throw new ArgumentException("Schema pattern must be a string");
                    }
                    // C# contract: .NET regex syntax, culture-invariant, bounded
                    // evaluation. No schema keyword can fetch a resource.
                    _ = new Regex(value.GetString()!, RegexOptions.CultureInvariant, TimeSpan.FromMilliseconds(100));
                    break;
                case "enum":
                    if (value.ValueKind != JsonValueKind.Array || value.GetArrayLength() == 0)
                    {
                        throw new ArgumentException("Schema enum must be a nonempty array");
                    }
                    var values = new HashSet<string>(StringComparer.Ordinal);
                    foreach (var item in value.EnumerateArray())
                    {
                        if (!values.Add(ValueKey(item)))
                        {
                            throw new ArgumentException("Schema enum contains equal values");
                        }
                    }
                    break;
                case "title": case "description":
                    if (value.ValueKind != JsonValueKind.String)
                    {
                        throw new ArgumentException("Schema annotation must be a string");
                    }
                    break;
                case "examples":
                    if (value.ValueKind != JsonValueKind.Array)
                    {
                        throw new ArgumentException("Schema examples must be an array");
                    }
                    break;
                case "default": case "const":
                    // These are arbitrary JSON data, not nested schemas.
                    break;
                default:
                    throw new ArgumentException("Unsupported submission schema keyword: " + field.Name);
            }
        }
    }

    private static partial void CheckValue(JsonElement rule, JsonElement value, string path)
    {
        if (rule.ValueKind == JsonValueKind.True)
        {
            return;
        }
        if (rule.ValueKind == JsonValueKind.False)
        {
            Fail(path, "forbidden by schema");
        }
        if (rule.TryGetProperty("type", out var type))
        {
            var matches = type.ValueKind == JsonValueKind.Array
                ? type.EnumerateArray().Any(item => HasType(value, item.GetString()!))
                : HasType(value, type.GetString()!);
            if (!matches)
            {
                Fail(path, "wrong type");
            }
        }
        if (rule.TryGetProperty("const", out var exact) && !Equal(value, exact))
        {
            Fail(path, "does not match const");
        }
        if (rule.TryGetProperty("enum", out var allowed) && !allowed.EnumerateArray().Any(item => Equal(value, item)))
        {
            Fail(path, "not in enum");
        }
        // JSON Schema keywords constrain only their applicable types. A rule
        // without type can therefore constrain objects without rejecting null.
        switch (value.ValueKind)
        {
            case JsonValueKind.Object:
                CheckBound(rule, "minProperties", value.EnumerateObject().Count(), true, path);
                CheckBound(rule, "maxProperties", value.EnumerateObject().Count(), false, path);
                if (rule.TryGetProperty("required", out var required))
                {
                    foreach (var name in required.EnumerateArray())
                    {
                        if (!value.TryGetProperty(name.GetString()!, out _))
                        {
                            Fail(path, "missing required field " + name.GetString());
                        }
                    }
                }
                var hasProperties = rule.TryGetProperty("properties", out var properties);
                var hasAdditional = rule.TryGetProperty("additionalProperties", out var additional);
                foreach (var field in value.EnumerateObject())
                {
                    if (hasProperties && properties.TryGetProperty(field.Name, out var property))
                    {
                        CheckValue(property, field.Value, path + "." + field.Name);
                    }
                    else if (hasAdditional)
                    {
                        CheckValue(additional, field.Value, path + "." + field.Name);
                    }
                }
                break;
            case JsonValueKind.Array:
                var items = value.EnumerateArray().ToArray();
                CheckBound(rule, "minItems", items.Length, true, path);
                CheckBound(rule, "maxItems", items.Length, false, path);
                var unique = rule.TryGetProperty("uniqueItems", out var uniqueValue) && uniqueValue.GetBoolean();
                var seen = unique ? new HashSet<string>(StringComparer.Ordinal) : null;
                for (var i = 0; i < items.Length; i++)
                {
                    if (rule.TryGetProperty("items", out var itemRule))
                    {
                        CheckValue(itemRule, items[i], path + "[" + i + "]");
                    }
                    if (seen is not null && !seen.Add(ValueKey(items[i])))
                    {
                        Fail(path, "array items are not unique");
                    }
                }
                break;
            case JsonValueKind.String:
                var text = value.GetString()!;
                var length = text.EnumerateRunes().Count();
                CheckBound(rule, "minLength", length, true, path);
                CheckBound(rule, "maxLength", length, false, path);
                if (rule.TryGetProperty("pattern", out var pattern))
                {
                    try
                    {
                        if (!Regex.IsMatch(text, pattern.GetString()!, RegexOptions.CultureInvariant, TimeSpan.FromMilliseconds(100)))
                        {
                            Fail(path, "pattern did not match");
                        }
                    }
                    catch (RegexMatchTimeoutException)
                    {
                        Fail(path, "pattern evaluation limit exceeded");
                    }
                }
                break;
            case JsonValueKind.Number:
                var number = SubmissionNumber.Parse(value);
                foreach (var keyword in new[] { "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf" })
                {
                    if (!rule.TryGetProperty(keyword, out var limit))
                    {
                        continue;
                    }
                    var bound = SubmissionNumber.Parse(limit);
                    var comparison = number.Compare(bound);
                    var valid = keyword switch
                    {
                        "minimum" => comparison >= 0,
                        "maximum" => comparison <= 0,
                        "exclusiveMinimum" => comparison > 0,
                        "exclusiveMaximum" => comparison < 0,
                        _ => number.IsMultipleOf(bound)
                    };
                    if (!valid)
                    {
                        Fail(path, keyword + " constraint failed");
                    }
                }
                break;
        }
    }

    private static partial void CheckBound(JsonElement rule, string keyword, int actual, bool minimum, string path)
    {
        if (rule.TryGetProperty(keyword, out var bound))
        {
            var number = SubmissionNumber.Parse(JsonSerializer.SerializeToElement(actual));
            var comparison = number.Compare(SubmissionNumber.Parse(bound));
            if (minimum ? comparison < 0 : comparison > 0)
            {
                Fail(path, keyword + " constraint failed");
            }
        }
    }

    private static partial bool HasType(JsonElement value, string type) => type switch
    {
        "object" => value.ValueKind == JsonValueKind.Object,
        "array" => value.ValueKind == JsonValueKind.Array,
        "string" => value.ValueKind == JsonValueKind.String,
        "boolean" => value.ValueKind is JsonValueKind.True or JsonValueKind.False,
        "null" => value.ValueKind == JsonValueKind.Null,
        "number" => value.ValueKind == JsonValueKind.Number,
        "integer" => value.ValueKind == JsonValueKind.Number && SubmissionNumber.Parse(value).IsInteger(),
        _ => false
    };

    private static partial bool Equal(JsonElement left, JsonElement right) =>
        ValueKey(left) == ValueKey(right);

    // Type tags, JSON-escaped strings/field names, sorted object fields and
    // normalized rational numbers give equality keys without rounding or
    // order-sensitive object comparisons. HashSet avoids pairwise comparisons
    // across large admitted arrays; keys exist only for this validation call.
    private static partial string ValueKey(JsonElement value) => value.ValueKind switch
    {
        JsonValueKind.Number => "n" + SubmissionNumber.Parse(value).ValueKey(),
        JsonValueKind.String => "s" + JsonSerializer.Serialize(value.GetString()),
        JsonValueKind.Object => "o{" + string.Join(',', value.EnumerateObject()
            .OrderBy(field => field.Name, StringComparer.Ordinal)
            .Select(field => JsonSerializer.Serialize(field.Name) + ":" + ValueKey(field.Value))) + "}",
        JsonValueKind.Array => "a[" + string.Join(',', value.EnumerateArray().Select(ValueKey)) + "]",
        JsonValueKind.True => "true",
        JsonValueKind.False => "false",
        _ => "null"
    };

    private static partial void Fail(string path, string reason) =>
        throw new ArgumentException(path + ": " + reason);
}

internal readonly partial struct SubmissionNumber
{
    private partial SubmissionNumber(BigInteger numerator, BigInteger denominator)
    {
        this.numerator = numerator;
        this.denominator = denominator;
    }

    internal static partial SubmissionNumber Parse(JsonElement value)
    {
        if (value.ValueKind != JsonValueKind.Number)
        {
            throw new ArgumentException("Schema numeric constraint must be a JSON number");
        }
        var raw = value.GetRawText();
        if (raw.Length > 128)
        {
            throw new ArgumentException("Submission number literal limit exceeded");
        }
        var exponentIndex = raw.IndexOfAny(['e', 'E']);
        var exponent = 0;
        var mantissa = raw;
        if (exponentIndex >= 0)
        {
            if (!int.TryParse(raw.AsSpan(exponentIndex + 1), NumberStyles.AllowLeadingSign,
                CultureInfo.InvariantCulture, out exponent) || exponent is < -1_000 or > 1_000)
            {
                throw new ArgumentException("Submission number exponent limit exceeded");
            }
            mantissa = raw[..exponentIndex];
        }
        var point = mantissa.IndexOf('.');
        var scale = point < 0 ? 0 : mantissa.Length - point - 1;
        var digits = mantissa.Replace(".", "", StringComparison.Ordinal);
        var numerator = BigInteger.Parse(digits, NumberStyles.AllowLeadingSign, CultureInfo.InvariantCulture);
        var power = exponent - scale;
        return power >= 0
            ? new SubmissionNumber(numerator * BigInteger.Pow(10, power), BigInteger.One)
            : new SubmissionNumber(numerator, BigInteger.Pow(10, -power));
    }

    internal partial int Compare(SubmissionNumber other)
    {
        // The default struct is the useful exact zero, not a zero denominator.
        var leftDenominator = denominator.IsZero ? BigInteger.One : denominator;
        var rightDenominator = other.denominator.IsZero ? BigInteger.One : other.denominator;
        return (numerator * rightDenominator).CompareTo(other.numerator * leftDenominator);
    }

    internal partial string ValueKey()
    {
        var divisor = denominator.IsZero ? BigInteger.One : denominator;
        var common = BigInteger.GreatestCommonDivisor(numerator, divisor);
        return (numerator / common).ToString(CultureInfo.InvariantCulture) + "/" +
            (divisor / common).ToString(CultureInfo.InvariantCulture);
    }

    internal partial bool IsInteger() => numerator.IsZero || numerator % denominator == 0;

    internal partial bool IsMultipleOf(SubmissionNumber other) =>
        (numerator * other.denominator) % (denominator * other.numerator) == 0;
}

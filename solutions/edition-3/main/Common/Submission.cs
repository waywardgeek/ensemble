using System.Numerics;
using System.Text.Json;

namespace Ensemble;

// Value-like immutable submission contract. It has no owning Agent and no
// mutable validation cache: children can retain it without lifetime machinery.
// A null schema selects legacy text normalization; an explicit schema never does.
public sealed partial class SubmissionContract
{
    private readonly JsonElement schema;
    private readonly bool legacy;
    private readonly int schemaBytes;

    // Validate and copy before constructing a child or creating its directory.
    // Supported keywords and refusal limits are documented in Chapter 20.
    public partial SubmissionContract(JsonElement? schema = null);
    public partial JsonElement Schema();
    // Return an independent validated value. Failure has no publication effects.
    public partial JsonElement Validate(JsonElement payload);
    // The string convenience path enters the same validator, including custom
    // contracts; it is not a bypass for a structured child's answer format.
    public partial JsonElement Validate(string result);

    private static partial int Admit(JsonElement value, int byteLimit, int nodeLimit);
    private static partial void CheckSchema(JsonElement node, bool root);
    private static partial void CheckValue(JsonElement rule, JsonElement value, string path);
    private static partial bool Equal(JsonElement left, JsonElement right);
    private static partial string ValueKey(JsonElement value);
    private static partial bool HasType(JsonElement value, string type);
    private static partial void CheckBound(JsonElement rule, string keyword, int actual, bool minimum, string path);
    private static partial void Fail(string path, string reason);
}

// Value-like exact decimal number used only during schema admission/validation.
// JSON numbers are finite decimals, so rational arithmetic needs no rounding,
// guessed epsilon, or floating-point conversion. Literal/exponent limits bound
// BigInteger sizes before this representation is constructed.
internal readonly partial struct SubmissionNumber
{
    private readonly BigInteger numerator;
    private readonly BigInteger denominator;
    private partial SubmissionNumber(BigInteger numerator, BigInteger denominator);
    internal static partial SubmissionNumber Parse(JsonElement value);
    internal partial int Compare(SubmissionNumber other);
    internal partial bool IsInteger();
    internal partial string ValueKey();
    internal partial bool IsMultipleOf(SubmissionNumber other);
}

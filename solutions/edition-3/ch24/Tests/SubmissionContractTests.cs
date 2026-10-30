using System.Text.Json;
using Ensemble;

internal static class SubmissionContractTests
{
    private static JsonElement J(string text)
    {
        using var document = JsonDocument.Parse(text, new JsonDocumentOptions { MaxDepth = 128 });
        return document.RootElement.Clone();
    }

    private static void Check(bool condition, string why)
    {
        if (!condition)
        {
            throw new Exception(why);
        }
    }

    private static void Refuses(Action operation)
    {
        try
        {
            operation();
        }
        catch (ArgumentException)
        {
            return;
        }
        throw new Exception("Expected contract refusal");
    }

    private static SubmissionContract Field(string schema) => new(J(
        "{\"type\":\"object\",\"properties\":{\"value\":" + schema + "},\"required\":[\"value\"]}"));
    private static JsonElement Value(string value) => J("{\"value\":" + value + "}");

    internal static void Structured()
    {
        var legacy = new SubmissionContract();
        Check(legacy.Validate(J("{\"result\":\"ok\",\"extra\":false}")).GetRawText() == "{\"result\":\"ok\"}",
            "Legacy path did not normalize ignored fields");
        Check(legacy.Validate("ok").GetProperty("result").GetString() == "ok", "String convenience failed");
        Refuses(() => legacy.Validate(J("{\"result\":false}")));
        Refuses(() => legacy.Validate(J("{}")));
        var explicitDefault = new SubmissionContract(legacy.Schema());
        Check(explicitDefault.Validate(J("{\"result\":\"ok\",\"extra\":false}")).GetProperty("extra").GetBoolean() == false,
            "Explicit schema incorrectly entered legacy normalization");

        const string schema = """
            {"type":"object","properties":{
              "summary":{"type":"string"},
              "findings":{"type":"array","items":{"type":"object","properties":{
                "path":{"type":"string"},"fixed":{"type":"boolean"}},
                "required":["path","fixed"],"additionalProperties":false}},
              "follow_up":{"type":["string","null"]}},
             "required":["summary","findings","follow_up"],"additionalProperties":false}
            """;
        SubmissionContract contract;
        using (var document = JsonDocument.Parse(schema))
        {
            contract = new SubmissionContract(document.RootElement);
        }
        const string payload = """{"summary":"reviewed","findings":[{"path":"main.cs","fixed":false}],"follow_up":null}""";
        var accepted = contract.Validate(J(payload));
        Check(accepted.GetRawText() == payload, "Structured value was wrapped, coerced or rewritten");
        Check(contract.Schema().GetRawText() == schema, "Schema was not independently retained");
        Refuses(() => contract.Validate("bypass"));
        Refuses(() => contract.Validate(J(payload.Replace("false", "\"false\""))));
        Refuses(() => contract.Validate(J(payload.Replace(",\"fixed\":false", ""))));
        Refuses(() => contract.Validate(J(payload.Replace("\"fixed\":false", "\"fixed\":false,\"extra\":0"))));
        Refuses(() => contract.Validate(J("[]")));
    }

    internal static void SchemaAdmission()
    {
        foreach (var schema in new[]
        {
            "true", "{}", "{\"type\":[\"object\"]}", "{\"type\":\"string\"}",
            "{\"type\":\"object\",\"type\":\"object\"}",
            "{\"type\":\"object\",\"$schema\":\"http://untrusted.invalid/schema\"}"
        })
        {
            Refuses(() => new SubmissionContract(J(schema)));
        }
        foreach (var invalid in new[]
        {
            "{\"$ref\":\"https://untrusted.invalid/no-fetch\"}", "{\"format\":\"date\"}",
            "{\"allOf\":[]}", "{\"unknown\":true}", "{\"type\":[]}",
            "{\"type\":[\"null\",\"null\"]}", "{\"type\":42}", "{\"type\":\"float\"}",
            "{\"required\":false}", "{\"required\":[\"x\",\"x\"]}", "{\"required\":[0]}",
            "{\"properties\":[]}", "{\"properties\":{\"x\":42}}", "{\"items\":[]}",
            "{\"additionalProperties\":42}", "{\"uniqueItems\":1}", "{\"pattern\":false}",
            "{\"pattern\":\"[\"}", "{\"minItems\":-1}", "{\"maxLength\":1.5}",
            "{\"minimum\":\"1\"}", "{\"multipleOf\":0}", "{\"multipleOf\":-1}",
            "{\"enum\":[]}", "{\"enum\":[1,1.0]}", "{\"enum\":[{\"a\":1,\"b\":2},{\"b\":2,\"a\":1}]}",
            "{\"title\":0}", "{\"description\":false}", "{\"examples\":{}}"
        })
        {
            Refuses(() => Field(invalid));
        }
        // Annotation/enum objects are arbitrary data, never recursively compiled.
        var annotations = Field("""{"default":{"$ref":"not a schema"},"examples":[{"format":"data"}],"enum":[{"unsupported":true}]}""");
        annotations.Validate(Value("{\"unsupported\":true}"));
        Refuses(() => annotations.Validate(Value("null")));
        var noInsertion = new SubmissionContract(J("""{"type":"object","properties":{"missing":{"default":42}}}"""));
        Check(!noInsertion.Validate(J("{}")).TryGetProperty("missing", out _), "Default annotation mutated payload");
        Refuses(() => new SubmissionContract(J("""{"type":"object","default":{"x":1,"x":2}}""")));
    }

    internal static void ExactNumbers()
    {
        var exact = Field("{\"const\":9007199254740993}");
        exact.Validate(Value("9007199254740993.0"));
        Refuses(() => exact.Validate(Value("9007199254740992")));
        var decimalStep = Field("""{"type":"number","minimum":0.1,"exclusiveMaximum":0.4,"multipleOf":0.1}""");
        foreach (var value in new[] { "0.1", "0.2", "0.3", "3e-1" })
        {
            decimalStep.Validate(Value(value));
        }
        foreach (var value in new[] { "0", "-0.1", "0.25", "0.30000000000000000000000001", "0.4", "\"0.2\"" })
        {
            Refuses(() => decimalStep.Validate(Value(value)));
        }
        var integer = Field("""{"type":"integer","exclusiveMinimum":-2,"maximum":2}""");
        foreach (var value in new[] { "-1", "-0", "0.00", "10e-1", "2.000" })
        {
            integer.Validate(Value(value));
        }
        foreach (var value in new[] { "-2", "2.01", "1e-1000", "1e1000" })
        {
            Refuses(() => integer.Validate(Value(value)));
        }
        var huge = Field("{\"minimum\":1e1000,\"maximum\":1e1000}");
        huge.Validate(Value("10e999"));
        var tiny = Field("{\"const\":1e-1000}");
        tiny.Validate(Value("0.1e-999"));
        // Mathematically integral bounds need not use the integer JSON spelling.
        Field("{\"minLength\":1.0,\"maxLength\":1e1000}").Validate(Value("\"x\""));
    }

    internal static void NestedConstraints()
    {
        var arrays = Field("""{"type":"array","items":{"type":["integer","null"]},"minItems":1,"maxItems":3,"uniqueItems":true}""");
        arrays.Validate(Value("[1,2,null]"));
        foreach (var value in new[] { "[]", "[1,2,3,4]", "[1,1.0]", "[null,null]", "[false]" })
        {
            Refuses(() => arrays.Validate(Value(value)));
        }
        Refuses(() => Field("{\"uniqueItems\":true}").Validate(Value("[{\"x\":1,\"y\":2},{\"y\":2.0,\"x\":1e0}]")));
        Field("{\"uniqueItems\":true}").Validate(Value("[[1,2],[2,1],true,false]"));
        var largeUnique = Field("{\"uniqueItems\":true}");
        var many = string.Join(',', Enumerable.Range(0, 12_000));
        largeUnique.Validate(Value("[" + many + "]"));
        Refuses(() => largeUnique.Validate(Value("[" + many + ",11999.0]")));
        Field("{\"uniqueItems\":true}").Validate(Value("[1,\"1\",{},[],null,false,true,\"null\"]"));
        Field("{\"items\":false}").Validate(Value("[]"));
        Refuses(() => Field("{\"items\":false}").Validate(Value("[0]")));
        var objects = Field("""{"type":"object","properties":{"yes":true,"no":false},"additionalProperties":{"type":"integer"},"minProperties":1,"maxProperties":2}""");
        objects.Validate(Value("{\"yes\":null,\"extra\":1}"));
        foreach (var value in new[] { "{}", "{\"no\":null}", "{\"extra\":false}", "{\"a\":1,\"b\":2,\"c\":3}" })
        {
            Refuses(() => objects.Validate(Value(value)));
        }
        var strings = Field("""{"type":"string","minLength":1,"maxLength":1,"pattern":"^[^0-9]+$"}""");
        strings.Validate(Value("\"😀\""));
        foreach (var value in new[] { "\"\"", "\"ab\"", "\"5\"" })
        {
            Refuses(() => strings.Validate(Value(value)));
        }
        // A pathological expression must refuse within the regex engine's
        // bound, not monopolize a child indefinitely. This is real matching,
        // not a mocked timeout or a claim of a whole-validation CPU deadline.
        Refuses(() => Field("{\"pattern\":\"(a+)+$\"}")
            .Validate(Value("\"" + new string('a', 10_000) + "!\"")));
        Field("{\"minItems\":2,\"minimum\":5}").Validate(Value("null"));
        Refuses(() => new SubmissionContract().Validate(J("{\"result\":\"a\",\"result\":\"b\"}")));
        Refuses(() => new SubmissionContract(J("{\"type\":\"object\"}")).Validate(Value("{\"x\":1,\"x\":2}")));
    }

    internal static void Limits()
    {
        Refuses(() => new SubmissionContract(J("{\"type\":\"object\",\"description\":\"" + new string('a', 65_536) + "\"}")));
        Refuses(() => new SubmissionContract(J("{" + new string(' ', 65_536) + "\"type\":\"object\"}")));
        Refuses(() => new SubmissionContract(J("{\"type\":\"object\",\"examples\":[" + string.Join(',', Enumerable.Repeat("0", 2_048)) + "]}")));
        var plain = new SubmissionContract(J("{\"type\":\"object\"}"));
        Refuses(() => plain.Validate(Value("\"" + new string('a', 1_048_576) + "\"")));
        Refuses(() => plain.Validate(Value("[" + string.Join(',', Enumerable.Repeat("0", 65_536)) + "]")));
        foreach (var number in new[] { new string('1', 129), "1e1001", "1e-1001", "1e9999999999999999999" })
        {
            Refuses(() => plain.Validate(Value(number)));
            Refuses(() => Field("{\"const\":" + number + "}"));
        }
        var atDepth = Value(new string('[', 63) + "0" + new string(']', 63));
        plain.Validate(atDepth); // Root object depth zero, innermost number depth64.
        Refuses(() => plain.Validate(Value(new string('[', 64) + "0" + new string(']', 64))));
        var combined = new SubmissionContract(J("{\"type\":\"object\",\"description\":\"" + new string('a', 4_000) + "\"}"));
        Refuses(() => combined.Validate(Value("\"" + new string('b', 20_000) + "\"")));
        // Escaping non-ASCII code points expands the canonical owned form; both
        // representations have caps, not just the short supplied UTF-8 form.
        Refuses(() => new SubmissionContract(J("{\"type\":\"object\",\"description\":\"" + new string('é', 12_000) + "\"}")));
        Refuses(() => plain.Validate(Value("\"" + new string('é', 180_000) + "\"")));
        plain.Validate(Value(new string('1', 128)));
    }
}

package common

// The submission owner limits vocabulary and resources; jsonschema supplies
// schema checking and validation. Numbers stay exact and no resources are fetched.
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

var DefaultSubmitSchema = json.RawMessage(`{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`)

func boundedJSON(raw []byte, byteLimit, nodeLimit, depthLimit int) (any, []byte, int, error) {
	if !utf8.Valid(raw) {
		return nil, nil, 0, fmt.Errorf("JSON must be valid UTF-8")
	}
	if len(raw) > byteLimit {
		return nil, nil, 0, fmt.Errorf("JSON exceeds %d bytes", byteLimit)
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var v any
	if err := d.Decode(&v); err != nil {
		return nil, nil, 0, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return nil, nil, 0, fmt.Errorf("expected one JSON value")
	}
	nodes := 0
	var visit func(any, int) error
	visit = func(v any, depth int) error {
		nodes++
		if nodes > nodeLimit || depth > depthLimit {
			return fmt.Errorf("JSON exceeds node/depth budget")
		}
		switch x := v.(type) {
		case json.Number:
			text := string(x)
			if len(text) > 128 {
				return fmt.Errorf("numeric literal exceeds 128 bytes")
			}
			if i := strings.IndexAny(text, "eE"); i >= 0 {
				e, err := strconv.ParseInt(text[i+1:], 10, 32)
				if err != nil || e < -1000 || e > 1000 {
					return fmt.Errorf("numeric exponent exceeds 1000")
				}
			}
		case map[string]any:
			for _, v := range x {
				if err := visit(v, depth+1); err != nil {
					return err
				}
			}
		case []any:
			for _, v := range x {
				if err := visit(v, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := visit(v, 0); err != nil {
		return nil, nil, 0, err
	}
	canonical, err := json.Marshal(v)
	if err != nil {
		return nil, nil, 0, err
	}
	if len(canonical) > byteLimit {
		return nil, nil, 0, fmt.Errorf("canonical JSON exceeds byte budget")
	}
	return v, canonical, nodes, nil
}

// CanonicalSubmitSchema owns the schema bytes. Empty means the default contract;
// it stays empty so default result normalization differs from explicit custom schemas.
func CanonicalSubmitSchema(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	canonical, _, err := compileSubmissionSchema(raw)
	return canonical, err
}

func CheckSubmitSchema(raw json.RawMessage) error {
	_, err := CanonicalSubmitSchema(raw)
	return err
}

func compileSubmissionSchema(raw json.RawMessage) (json.RawMessage, *jsonschema.Schema, error) {
	value, canonical, _, err := boundedJSON(raw, 65536, 2048, 64)
	if err != nil {
		return nil, nil, err
	}
	root, ok := value.(map[string]any)
	if !ok || root["type"] != "object" {
		return nil, nil, fmt.Errorf("submission schema root must explicitly have type object")
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.UseRegexpEngine(nil)                    // The accepted pattern dialect is Go's regexp.
	compiler.UseLoader(jsonschema.SchemeURLLoader{}) // No file, network, or custom resources.

	// Draft 2020-12 permits empty/duplicate enums. Our public subset does not;
	// let the same library enforce its stricter admission, including exact equality.
	const enumURL = "urn:ensemble:submission-enum"
	if err := compiler.AddResource(enumURL, map[string]any{
		"type": "array", "minItems": json.Number("1"), "uniqueItems": true,
	}); err != nil {
		return nil, nil, err
	}
	enumMembers, err := compiler.Compile(enumURL)
	if err != nil {
		return nil, nil, err
	}
	if err := admitSubmissionVocabulary(value, "$", enumMembers); err != nil {
		return nil, nil, err
	}
	const schemaURL = "urn:ensemble:submission"
	if err := compiler.AddResource(schemaURL, value); err != nil {
		return nil, nil, err
	}
	compiled, err := compiler.Compile(schemaURL)
	if err != nil {
		return nil, nil, err
	}
	return canonical, compiled, nil
}

// Only schema locations are traversed: property names and annotation/enum/const
// values may contain arbitrary JSON. The library checks keyword value shapes.
func admitSubmissionVocabulary(value any, path string, enumMembers *jsonschema.Schema) error {
	schema, ok := value.(map[string]any)
	if !ok {
		return nil // Nested booleans are valid; other shapes fail library compilation.
	}
	for keyword, value := range schema {
		switch keyword {
		case "$schema":
			if value != "https://json-schema.org/draft/2020-12/schema" {
				return fmt.Errorf("%s: unsupported schema dialect", path)
			}
		case "properties":
			if properties, ok := value.(map[string]any); ok {
				for name, child := range properties {
					if err := admitSubmissionVocabulary(child, path+".properties."+name, enumMembers); err != nil {
						return err
					}
				}
			}
		case "items", "additionalProperties":
			if err := admitSubmissionVocabulary(value, path+"."+keyword, enumMembers); err != nil {
				return err
			}
		case "enum":
			if err := enumMembers.Validate(value); err != nil {
				return fmt.Errorf("%s.enum: %w", path, err)
			}
		case "minLength", "maxLength", "minItems", "maxItems", "minProperties", "maxProperties":
			// v6.0.3 stores these as machine ints. No admitted payload can have
			// more than 1 MiB characters, items, or properties. Saturating larger
			// nonnegative integers above that ceiling is equivalent for bounded
			// inputs and avoids integer overflow. This map is a private decode;
			// canonical schema bytes (and durable schemas) remain unchanged.
			if n, ok := value.(json.Number); ok {
				bound, valid := new(big.Rat).SetString(string(n))
				ceiling := new(big.Rat).SetInt64(1048577)
				if valid && bound.IsInt() && bound.Cmp(ceiling) > 0 {
					schema[keyword] = json.Number("1048577")
				}
			}
		case "type", "title", "description", "default", "examples", "required", "const",
			"uniqueItems", "pattern", "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf":
		default:
			return fmt.Errorf("%s: unsupported schema keyword %q", path, keyword)
		}
	}
	return nil
}

// Keep the existing concise path/keyword feedback while retaining the library's
// complete error (including unambiguous JSON-pointer locations) as the cause.
func submissionValidationError(err error) error {
	failure, ok := err.(*jsonschema.ValidationError)
	if !ok {
		return err
	}
	for len(failure.Causes) != 0 {
		failure = failure.Causes[0]
	}
	path := "$"
	if len(failure.InstanceLocation) != 0 {
		path += "." + strings.Join(failure.InstanceLocation, ".")
	}
	keyword := strings.Join(failure.ErrorKind.KeywordPath(), "/")
	if keyword == "" {
		keyword = "schema"
	}
	return fmt.Errorf("%s: violates %s: %w", path, keyword, err)
}

// PrepareSubmission validates and returns independently owned canonical payload.
func PrepareSubmission(schema, payload json.RawMessage) (json.RawMessage, error) {
	isDefault := len(schema) == 0
	if isDefault {
		schema = DefaultSubmitSchema
	}
	canonical, compiled, err := compileSubmissionSchema(schema)
	if err != nil {
		return nil, err
	}
	value, raw, _, err := boundedJSON(payload, 1048576, 65536, 64)
	if err != nil {
		return nil, err
	}
	if int64(len(raw))*int64(len(canonical)) > 67108864 {
		return nil, fmt.Errorf("submission validation exceeds canonical byte-product budget")
	}
	if err := compiled.Validate(value); err != nil {
		return nil, submissionValidationError(err)
	}
	if isDefault {
		raw, err = json.Marshal(map[string]any{"result": value.(map[string]any)["result"]})
	}
	return raw, err
}

func ValidateSubmission(schema, payload json.RawMessage) error {
	_, err := PrepareSubmission(schema, payload)
	return err
}

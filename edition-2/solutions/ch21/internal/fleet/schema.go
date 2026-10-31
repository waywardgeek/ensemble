package fleet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const defaultSubmitSchema = `{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`
const schemaDialect = "https://json-schema.org/draft/2020-12/schema"

// SubmissionContract is immutable compiled validation owned by an Agent's Fleet.
// Compile it before construction side effects; Schema returns independent bytes.
type SubmissionContract struct {
	schema *jsonschema.Schema
	raw    json.RawMessage
	custom bool
}

// CompileSubmissionSchema admits a bounded Draft 2020-12 object contract.
// Nil/empty selects the legacy result-string schema. References, combinators,
// formats/content and unknown keywords are refused, never silently ignored.
func CompileSubmissionSchema(raw json.RawMessage) (*SubmissionContract, error) {
	custom := len(raw) != 0
	if !custom {
		raw = json.RawMessage(defaultSubmitSchema)
	}
	doc, err := boundedJSON(raw, 64*1024, 2048)
	if err != nil {
		return nil, fmt.Errorf("submit_schema: %w", err)
	}
	obj, ok := doc.(map[string]any)
	if !ok || obj["type"] != "object" {
		return nil, fmt.Errorf("submit_schema must declare type object")
	}
	if err = admitSchema(doc); err != nil {
		return nil, fmt.Errorf("submit_schema: %w", err)
	}
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	c.UseLoader(noSchemaResources{})
	const location = "urn:ensemble:submission"
	if err = c.AddResource(location, doc); err != nil {
		return nil, err
	}
	schema, err := c.Compile(location)
	if err != nil {
		return nil, fmt.Errorf("submit_schema: %w", err)
	}
	owned, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	// Escaping can enlarge JSON; accepted contracts must pass the same limit
	// when their canonical bytes are recompiled from a durable row.
	if len(owned) > 64*1024 {
		return nil, fmt.Errorf("canonical submit_schema exceeds 65536 bytes")
	}
	return &SubmissionContract{schema: schema, raw: owned, custom: custom}, nil
}

// Schema returns the same canonical schema used for validation, without sharing
// its backing bytes with declarations, factories or public callers.
func (c *SubmissionContract) Schema() json.RawMessage {
	return append(json.RawMessage(nil), c.raw...)
}

type noSchemaResources struct{}

// Load refuses network and filesystem resolution even if admission changes later.
func (noSchemaResources) Load(url string) (any, error) {
	return nil, fmt.Errorf("external schema resource is unavailable: %s", url)
}

// This is an admission list, not a schema evaluator: the library's metaschema
// and validator own keyword types, numeric rules and nested instance semantics.
// Annotation/enum values are data; their keys must never enter this walk.
func admitSchema(value any) error {
	if _, ok := value.(bool); ok {
		return nil
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("nested schema must be an object or boolean")
	}
	for key, value := range obj {
		switch key {
		case "$schema":
			if value != schemaDialect {
				return fmt.Errorf("only %s is supported", schemaDialect)
			}
		case "properties":
			if properties, ok := value.(map[string]any); ok {
				for _, schema := range properties {
					if err := admitSchema(schema); err != nil {
						return err
					}
				}
			}
		case "items", "additionalProperties":
			if err := admitSchema(value); err != nil {
				return err
			}
		case "type", "required", "enum", "const", "minimum", "maximum",
			"exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minLength",
			"maxLength", "pattern", "minItems", "maxItems", "uniqueItems",
			"minProperties", "maxProperties", "title", "description", "default", "examples":
		default:
			return fmt.Errorf("unsupported schema keyword %q", key)
		}
	}
	return nil
}

// Size, nesting and numeric ceilings protect the compiler/validator's recursion
// and exact rational arithmetic. These are refusal limits, never truncation.
func boundedJSON(raw []byte, maxBytes, maxNodes int) (any, error) {
	if len(raw) > maxBytes {
		return nil, fmt.Errorf("JSON exceeds %d bytes", maxBytes)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	nodes := 0
	var visit func(any, int) error
	visit = func(value any, depth int) error {
		nodes++
		if depth > 64 || nodes > maxNodes {
			return fmt.Errorf("JSON exceeds nesting64 or node limit%d", maxNodes)
		}
		switch v := value.(type) {
		case map[string]any:
			for _, child := range v {
				if err := visit(child, depth+1); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range v {
				if err := visit(child, depth+1); err != nil {
					return err
				}
			}
		case json.Number:
			if len(v) > 128 {
				return fmt.Errorf("JSON number exceeds 128 bytes")
			}
			if i := strings.IndexAny(string(v), "eE"); i >= 0 {
				exponent, err := strconv.Atoi(string(v)[i+1:])
				if err != nil || exponent < -1000 || exponent > 1000 {
					return fmt.Errorf("JSON number exponent exceeds magnitude1000")
				}
			}
		}
		return nil
	}
	return doc, visit(doc, 0)
}

func (c *SubmissionContract) validate(raw json.RawMessage) (json.RawMessage, error) {
	doc, err := boundedJSON(raw, 1024*1024, 65536)
	if err != nil {
		return nil, fmt.Errorf("submit arguments: %w", err)
	}
	// Compact through exact-number decoding: callers cannot later mutate the
	// accepted result, and numbers do not silently round through float64.
	owned, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	if len(owned) > 1024*1024 {
		return nil, fmt.Errorf("canonical submit arguments exceed 1048576 bytes")
	}
	// Separate ceilings still permit a large required/enum schema to expand
	// work and diagnostics across many instances. Bound the combination before
	// evaluation, without duplicating schema semantics or promising CPU timing.
	// Division avoids overflow; the compiled object schema is always nonempty.
	if len(owned) > (64*1024*1024)/len(c.raw) {
		return nil, fmt.Errorf("submit schema and arguments exceed combined validation budget")
	}
	if err = c.schema.Validate(doc); err != nil {
		return nil, fmt.Errorf("submit arguments: %w", err)
	}
	return owned, nil
}

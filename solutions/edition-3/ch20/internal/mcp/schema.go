package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Normalize the transport schema at admission, rather than teaching each
// provider renderer about MCP. Nested schemas and constraints are retained.
func normalizeSchema(raw json.RawMessage) (json.RawMessage, error) {
	var schema map[string]any
	if len(raw) == 0 {
		raw = json.RawMessage(`{"type":"object"}`)
	}
	if json.Unmarshal(raw, &schema) != nil || schema == nil {
		return nil, errors.New("inputSchema must be an object")
	}
	if t, ok := schema["type"]; ok && t != "object" {
		return nil, errors.New("inputSchema must describe object arguments")
	}
	schema["type"] = "object"
	if _, ok := schema["properties"]; !ok {
		schema["properties"] = map[string]any{}
	}
	if _, ok := schema["properties"].(map[string]any); !ok {
		return nil, errors.New("inputSchema properties must be an object")
	}
	delete(schema, "$schema")
	props := schema["properties"].(map[string]any)
	// Reserved execution controls are owned by Ensemble, never by remote servers.
	for _, name := range []string{"ai_callback_delay", "ai_callback_pattern", "max_output_bytes"} {
		delete(props, name)
	}
	props["ai_callback_delay"] = map[string]any{"type": "number", "description": "Seconds before observing this job; not an execution timeout"}
	props["ai_callback_pattern"] = map[string]any{"type": "string", "description": "Observe when unseen job output matches this regular expression"}
	props["max_output_bytes"] = map[string]any{"type": "integer", "description": "Maximum bytes in this observation"}
	required := []any{}
	original, ok := schema["required"].([]any)
	if schema["required"] != nil && !ok {
		return nil, fmt.Errorf("required must be an array")
	}
	for _, entry := range original {
		name, ok := entry.(string)
		if !ok {
			return nil, fmt.Errorf("required entries must be strings")
		}
		if name == "ai_callback_delay" || name == "ai_callback_pattern" || name == "max_output_bytes" {
			continue
		}
		if _, ok := props[name]; ok {
			required = append(required, name)
		}
	}
	schema["required"] = required
	schema["additionalProperties"] = false
	return json.Marshal(schema)
}

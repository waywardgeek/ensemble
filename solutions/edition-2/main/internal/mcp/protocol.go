package mcp

import (
	"context"
	"example.com/ensemble/internal/common"
	"time"
)

func (c *connection) discover(ctx context.Context) error {
	request := func(method string, p map[string]any) (map[string]any, error) {
		d, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		return c.rpc(d, method, p)
	}
	first, err := request("server/discover", map[string]any{})
	if err != nil {
		return err
	}
	versions, ok := first["supportedVersions"].([]any)
	if !ok || len(versions) == 0 {
		return failure("mcp_protocol", "Missing supported versions.")
	}
	seen := map[string]bool{}
	for _, v := range versions {
		x, ok := v.(string)
		if !ok || seen[x] {
			return failure("mcp_protocol", "Invalid supported versions.")
		}
		seen[x] = true
	}
	caps, ok := first["capabilities"].(map[string]any)
	if !ok {
		return failure("mcp_protocol", "Missing capabilities.")
	}
	if _, ok = caps["tools"].(map[string]any); !ok || !seen[common.MCPVersion] {
		return failure("mcp_protocol", "Selected protocol/tools capability unavailable.")
	}
	defs := map[string]common.MCPDefinition{}
	cursors := map[string]bool{}
	cursor := ""
	total := 0
	for page := 0; page < 64; page++ {
		params := map[string]any{}
		if cursor != "" {
			params["cursor"] = cursor
		}
		r, err := request("tools/list", params)
		if err != nil {
			return err
		}
		tools, ok := r["tools"].([]any)
		if !ok {
			return failure("mcp_protocol", "Missing tools list.")
		}
		for _, item := range tools {
			m, ok := item.(map[string]any)
			if !ok || !fields(m, "name", "description", "inputSchema", "outputSchema", "title", "icons", "annotations", "_meta") {
				return failure("mcp_protocol", "Invalid tool descriptor.")
			}
			name, ok := m["name"].(string)
			if !ok || !validRemote(name) {
				return failure("mcp_protocol", "Invalid remote name.")
			}
			if _, ok := defs[name]; ok {
				return failure("mcp_protocol", "Duplicate remote definition.")
			}
			description := ""
			if value, exists := m["description"]; exists {
				description, ok = value.(string)
				if !ok {
					return failure("mcp_protocol", "Invalid description.")
				}
			}
			input, present := m["inputSchema"]
			if !present {
				return failure("mcp_protocol", "Missing input schema.")
			}
			in, err := c.parent.Ensemble().JSON().Encode(input, true, 256<<10)
			if err != nil {
				return failure("mcp_limit", "Schema byte limit.")
			}
			var out []byte
			if value, exists := m["outputSchema"]; exists {
				if value == nil {
					return failure("mcp_protocol", "Null output schema is unsupported.")
				}
				out, err = c.parent.Ensemble().JSON().Encode(value, true, 256<<10)
				if err != nil {
					return failure("mcp_limit", "Schema byte limit.")
				}
			}
			d := common.MCPDefinition{Name: name, Description: description, InputSchema: in, OutputSchema: out}
			if _, err = validateDefinition(c.parent, d); err != nil {
				return err
			}
			normalized := map[string]any{"name": name, "description": description, "inputSchema": input, "outputSchema": m["outputSchema"]}
			bytes, err := c.parent.Ensemble().JSON().Encode(normalized, true, 16<<20)
			if err != nil {
				return failure("mcp_limit", "Definition bytes limit.")
			}
			total += len(bytes)
			if len(defs) >= 1024 || total > 16<<20 {
				return failure("mcp_limit", "Discovery catalog limit.")
			}
			defs[name] = d
		}
		value, present := r["nextCursor"]
		if !present {
			c.mu.Lock()
			c.definitions = defs
			c.mu.Unlock()
			return nil
		}
		cursor, ok = value.(string)
		if !ok || cursor == "" || len(cursor) > 4096 || cursors[cursor] {
			return failure("mcp_protocol", "Invalid or repeated cursor.")
		}
		cursors[cursor] = true
	}
	return failure("mcp_limit", "Discovery page limit.")
}
func (s *Service) result(m map[string]any, output *schema) (string, bool, error) {
	if !fields(m, "resultType", "content", "structuredContent", "isError", "_meta") {
		return "", false, failure("mcp_invalid_result", "Unknown remote result field.")
	}
	a, ok := m["content"].([]any)
	if !ok {
		return "", false, failure("mcp_invalid_result", "Remote result requires content array.")
	}
	if len(a) > 1024 {
		return "", false, failure("mcp_limit", "Result block limit.")
	}
	blocks := []any{}
	for _, v := range a {
		x, ok := v.(map[string]any)
		if !ok {
			return "", false, failure("mcp_invalid_result", "Invalid content item.")
		}
		t, ok := x["type"].(string)
		if !ok {
			return "", false, failure("mcp_invalid_result", "Invalid content type.")
		}
		if t != "text" {
			return "", false, failure("mcp_unsupported_content", "Remote result contains unsupported content.")
		}
		text, ok := x["text"].(string)
		if !ok || !fields(x, "type", "text", "annotations", "_meta") {
			return "", false, failure("mcp_invalid_result", "Invalid text content.")
		}
		blocks = append(blocks, map[string]any{"type": "text", "text": text})
	}
	isError := false
	if v, present := m["isError"]; present {
		isError, ok = v.(bool)
		if !ok {
			return "", false, failure("mcp_invalid_result", "Invalid isError value.")
		}
	}
	result := map[string]any{"content": blocks, "isError": isError}
	structured, present := m["structuredContent"]
	if output != nil {
		if !present {
			return "", false, failure("mcp_invalid_result", "Output schema requires structuredContent.")
		}
		if err := output.validate(structured); err != nil {
			return "", false, failure("mcp_invalid_result", "Structured output failed validation.")
		}
	}
	if present {
		result["structuredContent"] = structured
	}
	bytes, err := s.json().Encode(result, true, common.MCPMessageLimit)
	if err != nil {
		return "", false, failure("mcp_limit", "Canonical result byte limit.")
	}
	return string(bytes) + "\n", isError, nil
}


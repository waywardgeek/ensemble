package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

func ToolPolicy(cfg Config, name string) error {
	if cfg.SafeMode {
		switch name {
		case "run_command", "send_input", "wait_for_job", "kill_job", "tool_limits":
			return fmt.Errorf("%s is disabled in safe mode", name)
		}
	}
	return nil
}

// Redact replaces exact values, not just provider-shaped key prefixes. Diagnostic
// text may contain Firecrawl, OAuth or future provider secrets of any spelling.
func Redact(text string, secrets ...string) string {
	for _, secret := range secrets {
		if secret != "" {
			encoded, _ := json.Marshal(secret)
			text = strings.ReplaceAll(text, string(encoded[1:len(encoded)-1]), "[REDACTED]")
			text = strings.ReplaceAll(text, secret, "[REDACTED]")
		}
	}
	return text
}

// CredentialValues is for redaction only, never subprocess propagation. Key
// names are inspected conservatively so future provider credentials are covered.
func CredentialValues(configs ...Config) []string {
	var values []string
	for _, cfg := range configs {
		values = append(values, cfg.APIKey)
		for _, server := range cfg.MCPServers {
			if server.AuthEnv != "" {
				values = append(values, os.Getenv(server.AuthEnv))
			}
		}
	}
	for _, item := range os.Environ() {
		key, value, _ := strings.Cut(item, "=")
		key = strings.ToUpper(key)
		if strings.Contains(key, "KEY") || strings.Contains(key, "TOKEN") || strings.Contains(key, "SECRET") || strings.Contains(key, "PASSWORD") {
			if value != "" {
				values = append(values, value)
			}
		}
	}
	return values
}

// RedactDiagnostic retains structured diagnostic content without exposing raw or
// JSON-escaped credentials. Truncated undecodable escapes are withheld.
func RedactDiagnostic(raw []byte, secrets ...string) []byte {
	present := false
	for _, secret := range secrets {
		present = present || secret != ""
	}
	if !present {
		return raw
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) == nil {
		var trailing any
		if decoder.Decode(&trailing) == io.EOF {
			var redact func(any) any
			redact = func(v any) any {
				switch v := v.(type) {
				case string:
					return Redact(v, secrets...)
				case []any:
					for i, item := range v {
						v[i] = redact(item)
					}
				case map[string]any:
					clean := make(map[string]any, len(v))
					for key, item := range v {
						clean[Redact(key, secrets...)] = redact(item)
					}
					return clean
				}
				return v
			}
			clean, err := json.Marshal(redact(value))
			if err == nil {
				return clean
			}
		}
	}
	// A bounded HTTP body may end mid-JSON. Remove known spellings, but do
	// not surface undecodable escape sequences: alternate Unicode/slash
	// encodings could still contain the bearer. The caller retains HTTP status.
	text := Redact(string(raw), secrets...)
	if strings.Contains(text, `\`) {
		return nil
	}
	return []byte(text)
}

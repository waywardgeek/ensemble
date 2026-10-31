package engine

import (
	"encoding/json"
	"fmt"
)

// Only documented error identifiers escape the transport. A remote error body
// can echo credentials, so arbitrary messages and parameter values stay private.
func providerError(status int, data []byte) error {
	var body struct {
		// Code supports direct Responses error events.
		Code string `json:"code"`
		// Error wraps normal API error details.
		Error json.RawMessage `json:"error"`
		// Detail wraps subscription admission failures.
		Detail json.RawMessage `json:"detail"`
		// Response wraps errors arriving after SSE delivery begins.
		Response json.RawMessage `json:"response"`
	}
	json.Unmarshal(data, &body)
	code := body.Code
	for _, nested := range []json.RawMessage{body.Error, body.Detail, body.Response} {
		var value struct {
			// Code is checked against the explicit safe vocabulary below.
			Code string `json:"code"`
			// Error may nest the code inside a failed response.
			Error struct {
				// Code identifies the provider failure independently of prose.
				Code string `json:"code"`
			} `json:"error"`
		}
		json.Unmarshal(nested, &value)
		if value.Code != "" {
			code = value.Code
		}
		if value.Error.Code != "" {
			code = value.Error.Code
		}
	}
	switch code {
	case "subscription_sharing_usage_limit_exceeded":
		return fmt.Errorf("ChatGPT app usage limit reached; no metered fallback")
	case "subscription_sharing_usage_unavailable":
		return fmt.Errorf("ChatGPT usage temporarily unavailable; retry later without changing billing")
	case "subscription_sharing_unsupported_capability":
		return fmt.Errorf("ChatGPT route rejected an unsupported capability")
	}
	if status != 0 {
		return fmt.Errorf("provider HTTP status %d", status)
	}
	return fmt.Errorf("Responses inference did not complete")
}

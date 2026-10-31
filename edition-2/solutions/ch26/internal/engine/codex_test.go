package engine

import "testing"

// A stale API override, lookalike host or URL credential must not receive a
// Codex sign-in. Local external fakes remain available without real credentials.
func TestCodexEndpointBoundary(t *testing.T) {
	for _, endpoint := range []string{"https://api.openai.com", "https://chatgpt.com.evil.test/backend-api/codex", "http://chatgpt.com/backend-api/codex", "https://chatgpt.com/backend-api/codex?redirect=1", "https://user@chatgpt.com/backend-api/codex", "https://chatgpt.com/other"} {
		if validCodexEndpoint(endpoint) {
			t.Errorf("accepted unsafe/stale endpoint %s", endpoint)
		}
	}
	for _, endpoint := range []string{"https://chatgpt.com/backend-api/codex", "https://chatgpt.com/backend-api/codex/", "http://127.0.0.1:1234", "http://[::1]:1234"} {
		if !validCodexEndpoint(endpoint) {
			t.Errorf("rejected intended endpoint %s", endpoint)
		}
	}
}

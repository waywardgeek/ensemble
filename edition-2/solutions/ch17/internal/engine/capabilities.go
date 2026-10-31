package engine

import "ensemble/internal/common"

// Route restrictions are data, separate from model capabilities. Their absence
// from this table does not prove a live account accepts a particular feature.
type responseCapabilities struct {
	namespace         string
	omit              []string
	omitExplicitCache bool
}

func capabilitiesFor(cfg common.Config) responseCapabilities {
	profiles := map[common.CredentialKind]responseCapabilities{
		common.CredentialAPIKey: {},
		// The live plan route rejects prompt_cache_options. Markers belong to that
		// same explicit mode, so omit the pair without claiming separate rejection.
		common.CredentialChatGPT: {namespace: "ensemble", omitExplicitCache: true, omit: []string{"temperature", "top_p", "max_output_tokens", "previous_response_id", "background", "conversation", "metadata", "max_tool_calls", "moderation", "multi_agent", "prompt", "prompt_cache_retention", "safety_identifier", "top_logprobs", "truncation", "user"}},
	}
	kind := cfg.BillingMode
	if cfg.Credentials != nil {
		kind = cfg.Credentials.Kind()
	}
	return profiles[kind]
}

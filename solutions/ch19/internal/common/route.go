package common

// Route constraints.
//
// A credential does not merely authenticate a request; it selects which
// deployment of an endpoint the request lands on, and those deployments do
// not accept the same parameters. The consumer-plan route rejects fifteen
// fields that a metered API key accepts without comment, and requires two
// others to hold specific values.
//
// This is published as data for the same reason model capabilities are. The
// alternative is a renderer that asks what sort of credential it has and
// branches, which puts knowledge of billing arrangements inside the code that
// formats JSON, and spreads it over as many places as there are fields. A
// table can be read, tested and pointed at when a vendor changes the rules.
//
// Sending a forbidden field is not a soft failure. The request is refused, so
// an agent that gets this wrong does not degrade — it stops.

// Route describes what a credential route permits.
type Route struct {
	// Forbidden lists request fields the route rejects outright.
	Forbidden []string

	// RequiresStateless means the route refuses server-side conversation
	// state, so store must be false and previous_response_id is unusable.
	RequiresStateless bool

	// RequiresStreaming means a non-streaming request is refused.
	RequiresStreaming bool
}

// Forbids reports whether the route rejects a named request field.
func (r Route) Forbids(field string) bool {
	for _, f := range r.Forbidden {
		if f == field {
			return true
		}
	}
	return false
}

// planRouteForbidden is the consumer-plan route's unsupported parameter list,
// transcribed from the vendor's own documentation rather than discovered one
// 400 at a time.
//
// Two entries are easy to get wrong. It is top_logprobs that is unsupported,
// not logprobs. And instructions is NOT here: it is permitted, merely
// uncacheable, which is a reason to prefer a developer message but not a
// reason to call it illegal.
var planRouteForbidden = []string{
	"background",
	"conversation",
	"max_output_tokens",
	"max_tool_calls",
	"metadata",
	"moderation",
	"multi_agent",
	"prompt",
	"prompt_cache_retention",
	"safety_identifier",
	"temperature",
	"top_logprobs",
	"top_p",
	"truncation",
	"user",
}

// RouteFor returns the constraints that apply to a credential kind.
//
// There is deliberately no permissive default. A credential kind nobody has
// characterised gets the metered route's rules, which are the ones every
// chapter before this assumed; a new restricted route must be added here
// before it can be used, rather than silently inheriting permission.
func RouteFor(k CredentialKind) Route {
	switch k {
	case CredentialChatGPTOAuth:
		return Route{
			Forbidden:         planRouteForbidden,
			RequiresStateless: true,
			RequiresStreaming: true,
		}
	default:
		return Route{}
	}
}

package common

import "testing"

// Every vendor the model table can name must have a resolved endpoint.
// Resolution and the model table are two separate lists, and a model whose
// vendor was never resolved is unreachable: selecting it is refused at
// runtime even though the model itself is perfectly legitimate. Adding a
// vendor to one list and forgetting the other is the drift this catches.
//
// fake-model is excluded on the same grounds as in the vendor invariant: it
// is driven against several vendors by env prefix and deliberately claims
// none.
func TestEveryVendorInTheModelTableHasAnEndpoint(t *testing.T) {
	eps := ResolveEndpoints(func(generic, specific, fallback string) string { return fallback })

	for id := range models() {
		v, ok := VendorFor(id)
		if !ok {
			continue
		}
		ep, ok := eps[v]
		if !ok {
			t.Errorf("model %s has vendor %v, which ResolveEndpoints never resolves", id, v)
			continue
		}
		if ep.BaseURL == "" {
			t.Errorf("vendor %v (model %s) resolved an empty base URL", v, id)
		}
	}
}

// The generic override exists so a grader can point every vendor at one fake
// server. It must beat the vendor specific variable, or a developer with a
// real key in their environment would silently reach the real API during a
// test run.
func TestResolveEndpointsPrefersTheGenericOverride(t *testing.T) {
	eps := ResolveEndpoints(func(generic, specific, fallback string) string {
		if generic == "LLM_BASE_URL" {
			return "https://fake.test"
		}
		if generic == "LLM_API_KEY" {
			return "fake-key"
		}
		return fallback
	})

	for v, ep := range eps {
		if ep.BaseURL != "https://fake.test" {
			t.Errorf("vendor %v: BaseURL = %q, want the generic override to win", v, ep.BaseURL)
		}
		if ep.APIKey != "fake-key" {
			t.Errorf("vendor %v: APIKey = %q, want the generic override to win", v, ep.APIKey)
		}
	}
}

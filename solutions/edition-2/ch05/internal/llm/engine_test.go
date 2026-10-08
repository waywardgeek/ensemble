package llm

import (
	"net/http"
	"testing"
)

func TestTransportOwnedByEachEngine(t *testing.T) {
	a, b := New(nil), New(nil)
	for _, engine := range []*Engine{a, b} {
		if engine.client.Transport == nil || engine.client.Transport == http.DefaultTransport {
			t.Fatal("Engine uses global default transport")
		}
	}
	if a.client.Transport == b.client.Transport {
		t.Fatal("Engines share transport")
	}
}

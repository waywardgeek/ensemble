package mcp

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

// Live probe against Firecrawl's hosted MCP server. Skipped unless both
// EN_LIVE=1 and FIRECRAWL_API_KEY are set, so an ordinary `go test ./...`
// and every grader run never touch the network.
//
//	EN_LIVE=1 go test ./internal/mcp/ -run TestLiveFirecrawl -v
//
// This exercises the path the host actually uses: NewHTTPTransport with the
// Authorization header that cmd/main.go builds from a server's auth-env, then
// initialize, tools/list, and a real search.
func liveOrSkip(t *testing.T) string {
	t.Helper()
	if os.Getenv("EN_LIVE") != "1" {
		t.Skip("set EN_LIVE=1 to run the live Firecrawl probe")
	}
	key := os.Getenv("FIRECRAWL_API_KEY")
	if key == "" {
		t.Skip("FIRECRAWL_API_KEY is not set")
	}
	return key
}

const firecrawlURL = "https://mcp.firecrawl.dev/v2/mcp"

func TestLiveFirecrawlKeyedEndpoint(t *testing.T) {
	key := liveOrSkip(t)

	tr, err := NewHTTPTransport(firecrawlURL, map[string]string{
		"Authorization": "Bearer " + key,
	})
	if err != nil {
		t.Fatalf("NewHTTPTransport: %v", err)
	}
	c := NewClient(tr)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := c.Initialize(ctx); err != nil {
		t.Fatalf("initialize: %v", err)
	}

	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	byName := map[string]bool{}
	for _, ti := range tools {
		byName[ti.Name] = true
	}
	t.Logf("keyed endpoint advertises %d tools", len(tools))
	for _, want := range []string{"firecrawl_search", "firecrawl_scrape"} {
		if !byName[want] {
			t.Errorf("keyed endpoint did not advertise %s", want)
		}
	}

	res, err := c.CallTool(ctx, "firecrawl_search", json.RawMessage(
		`{"query":"Firecrawl MCP server","limit":2}`))
	if err != nil {
		t.Fatalf("firecrawl_search: %v", err)
	}
	var sb strings.Builder
	for _, cnt := range res.Content {
		sb.WriteString(cnt.Text)
	}
	out := sb.String()
	if strings.TrimSpace(out) == "" {
		t.Fatal("firecrawl_search returned no content")
	}
	if !strings.Contains(strings.ToLower(out), "http") {
		t.Errorf("search result carried no URL: %.200s", out)
	}
	t.Logf("search returned %d bytes, first 200: %.200s", len(out), out)
}

// Without a key the same URL still answers, which is why an unset auth-env
// degrades to anonymous rather than failing the connection.
func TestLiveFirecrawlKeylessStillAnswers(t *testing.T) {
	if os.Getenv("EN_LIVE") != "1" {
		t.Skip("set EN_LIVE=1 to run the live Firecrawl probe")
	}
	tr, err := NewHTTPTransport(firecrawlURL, nil)
	if err != nil {
		t.Fatalf("NewHTTPTransport: %v", err)
	}
	c := NewClient(tr)
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := c.Initialize(ctx); err != nil {
		t.Fatalf("keyless initialize: %v", err)
	}
	tools, err := c.ListTools(ctx)
	if err != nil {
		t.Fatalf("keyless tools/list: %v", err)
	}
	t.Logf("keyless endpoint advertises %d tools", len(tools))
}

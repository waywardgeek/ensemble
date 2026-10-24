package mcp

import (
	"context"
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPChapter18Protocol(t *testing.T) {
	t.Setenv("CH18_MCP_TOKEN", "private-test-token")
	initialized := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("Authorization") != "Bearer private-test-token" || r.Header.Get("Accept") != "application/json, text/event-stream" {
			t.Error("missing request contract")
		}
		var m message
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			t.Error(err)
			return
		}
		if m.Method != "initialize" && (r.Header.Get("Mcp-Session-Id") != "session-test" || r.Header.Get("MCP-Protocol-Version") != "2025-03-26") {
			t.Error("session/protocol not propagated")
		}
		switch m.Method {
		case "initialize":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Mcp-Session-Id", "session-test")
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{"protocolVersion":"2025-03-26","capabilities":{"tools":{}}}}`, m.ID)
		case "notifications/initialized":
			initialized = true
			w.WriteHeader(http.StatusAccepted)
		case "tools/list":
			if !initialized {
				t.Error("discovery raced initialization")
			}
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprintf(w, ": keepalive\r\n\r\nevent: message\r\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/message\"}\r\n\r\ndata: {\"jsonrpc\":\"2.0\",\"id\":%s,\r\ndata: \"result\":{\"tools\":[]}}\r\n\r\n", m.ID)
		}
	}))
	defer srv.Close()
	tr, err := newHTTPTransport(common.MCPServer{URL: srv.URL, AuthEnv: "CH18_MCP_TOKEN"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := newPeer(tr, nil, false)
	defer p.close(io.EOF)
	manager := &manager{}
	c := &connection{peer: p}
	if err := manager.initialize(context.Background(), c); err != nil {
		t.Fatal(err)
	}
}
func TestHTTPChapter18Errors(t *testing.T) {
	for _, mode := range []string{"status", "malformed", "wrong-id", "redirect"} {
		t.Run(mode, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if mode == "status" {
					w.WriteHeader(401)
					fmt.Fprint(w, "upstream unavailable")
					return
				}
				if mode == "redirect" {
					w.Header().Set("Location", "http://127.0.0.1:1/private")
					w.WriteHeader(302)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				if mode == "malformed" {
					fmt.Fprint(w, "upstream unavailable")
				} else {
					fmt.Fprint(w, `{"jsonrpc":"2.0","id":999,"result":{}}`)
				}
			}))
			defer srv.Close()
			tr, err := newHTTPTransport(common.MCPServer{URL: srv.URL + "?private-query"}, nil)
			if err != nil {
				t.Fatal(err)
			}
			p := newPeer(tr, nil, false)
			defer p.close(io.EOF)
			err = p.call(context.Background(), "initialize", nil, nil)
			if err == nil || strings.Contains(err.Error(), "private") {
				t.Fatalf("unsafe/missing error: %v", err)
			}
		})
	}
	for _, raw := range []string{"", "file:///tmp/mcp", "https://user:secret@host/mcp", "https://host/mcp#fragment", "/relative"} {
		if _, err := newHTTPTransport(common.MCPServer{URL: raw}, nil); err == nil {
			t.Errorf("accepted invalid endpoint")
		}
	}
}

func TestHTTPChapter18R1Anonymous(t *testing.T) {
	t.Setenv("CH18_ABSENT_TOKEN", "")
	posted := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posted = true
		if r.Header.Get("Authorization") != "" {
			t.Error("anonymous attempt sent Authorization")
		}
		var m message
		_ = json.NewDecoder(r.Body).Decode(&m)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":{}}`, m.ID)
	}))
	defer srv.Close()
	var log strings.Builder
	tr, err := dial(context.Background(), common.MCPServer{Transport: "url", URL: srv.URL, AuthEnv: "CH18_ABSENT_TOKEN"}, "", &log)
	if err != nil {
		t.Fatal(err)
	}
	p := newPeer(tr, nil, false)
	defer p.close(io.EOF)
	var result json.RawMessage
	err = p.call(context.Background(), "initialize", nil, &result)
	if err != nil || !posted {
		t.Errorf("anonymous POST missing: %v", err)
	}
	if !strings.Contains(log.String(), "CH18_ABSENT_TOKEN") || !strings.Contains(strings.ToLower(log.String()), "anonymous") {
		t.Errorf("missing named diagnostic: %q", log.String())
	}
}

func TestHTTPChapter18R1RemoteErrors(t *testing.T) {
	const token = `reflected-"bearer\-secret`
	t.Setenv("CH18_REFLECTED_TOKEN", token)
	for _, mode := range []string{"401", "403", "403-json", "json-rpc", "sse-rpc", "json-tool", "sse-tool"} {
		t.Run(mode, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer "+token {
					t.Error("missing configured credential")
				}
				var m message
				_ = json.NewDecoder(r.Body).Decode(&m)
				reason := "quota denied for Bearer " + token + ": account is suspended"
				if mode == "401" || strings.HasPrefix(mode, "403") {
					code := 401
					if strings.HasPrefix(mode, "403") {
						code = 403
					}
					w.WriteHeader(code)
					if mode == "403-json" {
						_ = json.NewEncoder(w).Encode(map[string]string{"error": reason + strings.Repeat("x", 10000)})
					} else {
						fmt.Fprint(w, reason+strings.Repeat("x", 10000))
					}
					return
				}
				response := map[string]any{"jsonrpc": "2.0", "id": m.ID}
				if strings.HasSuffix(mode, "rpc") {
					response["error"] = map[string]any{"code": -32000, "message": reason}
				} else {
					response["result"] = map[string]any{"isError": true, "content": []any{map[string]any{"type": "text", "text": reason}, map[string]any{"type": "resource", "resource": map[string]any{"uri": "fixture://detail", "text": reason}}}}
				}
				raw, _ := json.Marshal(response)
				if strings.HasPrefix(mode, "sse") {
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprintf(w, "data: %s\n\n", raw)
				} else {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write(raw)
				}
			}))
			defer srv.Close()
			var log strings.Builder
			tr, err := dial(context.Background(), common.MCPServer{Transport: "url", URL: srv.URL, AuthEnv: "CH18_REFLECTED_TOKEN"}, "", &log)
			if err != nil {
				t.Fatal(err)
			}
			p := newPeer(tr, nil, false)
			defer p.close(io.EOF)
			var result common.MCPCallResult
			err = p.call(context.Background(), "tools/call", nil, &result)
			text := ""
			if strings.HasSuffix(mode, "tool") {
				if err != nil || !result.IsError {
					t.Fatalf("lost tool error semantics: %v %+v", err, result)
				}
				var out strings.Builder
				if err = writeContent(&out, result.Content); err != nil {
					t.Fatal(err)
				}
				text = out.String()
				// Decoded embedded resources must retain their content shape and safe reason.
				var block struct{ Resource struct{ URI, Text string } }
				_ = json.Unmarshal(result.Content[1], &block)
				if block.Resource.URI != "fixture://detail" || strings.Contains(block.Resource.Text, token) {
					t.Error("resource error changed or leaked")
				}
			} else {
				if err == nil {
					t.Fatal("lost remote failure")
				}
				text = err.Error()
			}
			if !strings.Contains(text, "quota denied") || !strings.Contains(text, "account is suspended") {
				t.Errorf("remote reason lost: %.150s", text)
			}
			encoded, _ := json.Marshal(token)
			if strings.Contains(text, token) || strings.Contains(text, string(encoded[1:len(encoded)-1])) || strings.Contains(log.String(), token) {
				t.Error("configured bearer leaked")
			}
			if (mode == "401" || strings.HasPrefix(mode, "403")) && len(text) > 4500 {
				t.Errorf("unbounded HTTP reason: %d", len(text))
			}
		})
	}
}

func TestHTTPChapter18R2TruncatedUnicodeBearer(t *testing.T) {
	const token = "secret"
	const escaped = `\u0073\u0065\u0063\u0072\u0065\u0074`
	t.Setenv("CH18_UNICODE_TOKEN", token)
	body := `{"error":"quota denied ` + escaped + strings.Repeat("x", 6000) + `"}`
	if !json.Valid([]byte(body)) {
		t.Fatal("fixture must send complete valid JSON")
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			t.Error("missing configured bearer")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, body)
	}))
	defer srv.Close()
	var log strings.Builder
	tr, err := dial(context.Background(), common.MCPServer{Transport: "url", URL: srv.URL, AuthEnv: "CH18_UNICODE_TOKEN"}, "", &log)
	if err != nil {
		t.Fatal(err)
	}
	p := newPeer(tr, nil, false)
	defer p.close(io.EOF)
	var result json.RawMessage
	err = p.call(context.Background(), "tools/call", nil, &result)
	if err == nil {
		t.Fatal("missing HTTP failure")
	}
	text := err.Error()
	if strings.Contains(text, token) || strings.Contains(text, escaped) || strings.Contains(log.String(), escaped) {
		t.Fatalf("recoverable configured bearer in diagnostic: %.160s", text)
	}
	if !strings.Contains(text, "403") || (!strings.Contains(text, "Forbidden") && !strings.Contains(text, "quota denied")) {
		t.Fatalf("lost safe HTTP diagnostic: %s", text)
	}
	if len(text) > 4500 {
		t.Fatalf("unbounded diagnostic: %d", len(text))
	}
}

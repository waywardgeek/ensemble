package common

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Config is everything about a request that is not conversation.
type Config struct {
	Vendor       Vendor
	Surface      Surface
	Model        string
	BaseURL      string
	APIKey       string
	SystemPrompt string
	MaxTokens    int
	Tools        []ToolDecl
	Cwd          string
}

// ToolDecl is a tool as the MODEL sees it.
type ToolDecl struct {
	Name        string
	Description string
	Schema      json.RawMessage
}

// Renderer turns a Context into one vendor's HTTP request.
type Renderer interface {
	Render(*Context, Config) (*http.Request, error)
}

// Parser turns one vendor's HTTP response back into events.
type Parser interface {
	Parse(status int, body []byte) ([]Event, error)
}

// BodyOf returns the exact bytes the renderer produced.
func BodyOf(req *http.Request) ([]byte, error) {
	if req.GetBody == nil {
		return nil, fmt.Errorf("request has no recoverable body")
	}
	rc, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

// DefaultSurface returns the default API surface for a vendor.
func DefaultSurface(v Vendor) Surface {
	switch v {
	case VendorOpenAI:
		return SurfaceChatCompletions
	case VendorGemini:
		return SurfaceGenerateContent
	default:
		return SurfaceMessages
	}
}

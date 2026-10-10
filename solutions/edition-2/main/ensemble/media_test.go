package ensemble_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ensemble/ensemble"
	"ensemble/internal/common"
)

// Renderers must include local bytes and reject unsupported capabilities or Ref
// kinds explicitly.
func TestLocalMediaIsEncodedAndUnsupportedMediaIsRefused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.bin")
	data := []byte("local bytes, not a remote locator")
	os.WriteFile(path, data, 0600)
	for _, sample := range []struct {
		vendor             ensemble.Vendor
		model, mime, field string
	}{
		{ensemble.Anthropic, "claude-haiku-5-5", "image/png", "media_type"},
		{ensemble.Anthropic, "claude-haiku-5-5", "application/pdf", "document"},
		{ensemble.OpenAI, "gpt-5.4-mini", "image/png", "image_url"},
		{ensemble.OpenAI, "gpt-5.4-mini", "application/pdf", "file_data"},
		{ensemble.Gemini, "gemini-3.8-flash", "image/png", "inlineData"},
		{ensemble.Gemini, "gemini-3.8-flash", "audio/wav", "inlineData"},
		{ensemble.Gemini, "gemini-3.8-flash", "video/mp4", "inlineData"},
		{ensemble.Gemini, "gemini-3.8-flash", "application/pdf", "inlineData"},
	} {
		t.Run(sample.model+sample.mime, func(t *testing.T) {
			a, err := ensemble.New(io.Discard).NewAgent(ensemble.Config{Vendor: sample.vendor, Model: sample.model})
			if err != nil {
				t.Fatal(err)
			}
			c := &common.Context{Dialogue: []common.Entry{{Actor: common.Human, Parts: []common.Part{{Type: "text", Text: "before"}, {Type: "blob", MIME: sample.mime, Ref: common.Ref{Kind: common.RefPath, Locator: path}}, {Type: "text", Text: "after"}}}}}
			body, err := a.Engine().Render(c)
			if err != nil || !bytes.Contains(body, []byte(base64.StdEncoding.EncodeToString(data))) || !bytes.Contains(body, []byte(sample.field)) {
				t.Fatalf("media bytes missing: %s %v", body, err)
			}
			if !bytes.Contains(body, []byte("before")) || !bytes.Contains(body, []byte("after")) {
				t.Fatal("media lost adjacent text")
			}
			if sample.vendor == ensemble.Gemini {
				// Blob requires mimeType and base64 data inside inlineData; finding
				// the bytes somewhere in the body does not validate that wire contract.
				// https://ai.google.dev/api/generate-content#Blob
				var wire struct {
					// Contents holds the rendered conversation messages.
					Contents []struct {
						// Parts preserves the text/media/text ordering of this fixture.
						Parts []struct {
							// InlineData is the Blob object whose exact field names matter.
							InlineData map[string]string `json:"inlineData"`
						} `json:"parts"`
					} `json:"contents"`
				}
				if err := json.Unmarshal(body, &wire); err != nil {
					t.Fatal(err)
				}
				if len(wire.Contents) != 1 || len(wire.Contents[0].Parts) != 3 {
					t.Fatalf("unexpected Gemini content structure: %s", body)
				}
				blob := wire.Contents[0].Parts[1].InlineData
				if blob["mimeType"] != sample.mime {
					t.Fatalf("Gemini inlineData.mimeType = %q, want %q", blob["mimeType"], sample.mime)
				}
				decoded, err := base64.StdEncoding.DecodeString(blob["data"])
				if err != nil || !bytes.Equal(decoded, data) {
					t.Fatalf("Gemini inlineData.data did not decode to the local bytes: %q, %v", decoded, err)
				}
			}
			c.Dialogue[0].Parts[1].Ref.Kind = common.RefURI
			if _, err := a.Engine().Render(c); err == nil || !strings.Contains(err.Error(), "only local paths") {
				t.Fatal("unsupported Ref did not fail loudly")
			}
		})
	}
	for _, sample := range []struct{ model, mime, want string }{{"unknown-media-model", "image/png", "image"}, {"claude-haiku-5-5", "audio/wav", "audio"}} {
		a, _ := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: sample.model})
		c := &common.Context{Dialogue: []common.Entry{{Actor: common.Human, Parts: []common.Part{{Type: "blob", MIME: sample.mime, Ref: common.Ref{Kind: common.RefPath, Locator: path}}}}}}
		if _, err := a.Engine().Render(c); err == nil || !strings.Contains(err.Error(), sample.model) || !strings.Contains(err.Error(), sample.want) {
			t.Fatalf("refusal must name model and media: %v", err)
		}
	}
	if _, known := ensemble.LookupModel("not-in-table"); known {
		t.Fatal("unknown model acquired default features")
	}
}

// Removing media bytes from projection must leave their recovery address available.
func TestMediaRedactionRetainsReference(t *testing.T) {
	a, _ := ensemble.New(io.Discard).NewAgent(ensemble.Config{Model: "claude-haiku-5-5"})
	ref := common.Ref{Kind: common.RefPath, Locator: "recoverable-image.png"}
	a.History().Append(common.Event{Type: "tool_returned", Tool: &common.ToolData{CallID: "image", Parts: []common.Part{{Type: "blob", MIME: "image/png", Ref: ref}}}})
	a.History().Append(common.Event{Type: "redacted", Redact: &common.RedactData{From: 1, To: 1, Level: "redact_result"}})
	p := a.History().Context().Dialogue[0].Parts[0].Parts[0]
	if p.Type != "redacted" || p.Ref != ref {
		t.Fatal("media redaction lost recovery Ref:", fmt.Sprint(p))
	}
}

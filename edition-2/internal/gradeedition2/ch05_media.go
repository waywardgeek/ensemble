package gradeedition2

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/waywardgeek/ensemble/internal/grade"
)

// ReadMediaModels selects declared-capability fixtures, not live model discovery.
// Defaults use existing declared rows, with older Sonnet 4.6 for Anthropic.
// A JSON map selects another implementation’s table without adding model rows.
func ReadMediaModels(path string) (map[string]string, error) {
	models := map[string]string{"anthropic": "claude-sonnet-4-6", "openai": "gpt-5.4-mini", "gemini": "gemini-3.8-flash"}
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		models = map[string]string{}
		if err := json.Unmarshal(data, &models); err != nil {
			return nil, err
		}
	}
	for vendor, model := range models {
		if (vendor != "anthropic" && vendor != "openai" && vendor != "gemini") || strings.TrimSpace(model) == "" {
			return nil, fmt.Errorf("media-models needs nonempty anthropic, openai and gemini identifiers; invalid %q", vendor)
		}
	}
	if len(models) != 3 {
		return nil, fmt.Errorf("media-models needs anthropic, openai and gemini identifiers")
	}
	return models, nil
}

func mediaCapabilities(bin string, models map[string]string) grade.Check {
	c := grade.Check{ID: "loud-refusal", Title: "Local images contain file bytes; unsupported media fails explicitly", Points: 10, Earned: 10, Passed: true}
	work, err := os.MkdirTemp("", "edition2-media-*")
	if err != nil {
		fail(&c, "%v", err)
		return c
	}
	defer os.RemoveAll(work)
	// A valid tiny file makes the positive case independent of a remote URI,
	// stale locator, or text marker that merely says an image was attached.
	var imageData bytes.Buffer
	if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		fail(&c, "%v", err)
		return c
	}
	imagePath := filepath.Join(work, "pixel.png")
	if err := os.WriteFile(imagePath, imageData.Bytes(), 0600); err != nil {
		fail(&c, "%v", err)
		return c
	}
	// A real one-sample PCM WAV avoids conflating capability refusal with a
	// MIME/file mismatch. One silent 16-bit sample at 8 kHz is enough here.
	wav := make([]byte, 46)
	copy(wav, "RIFF")
	binary.LittleEndian.PutUint32(wav[4:], 38)
	copy(wav[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(wav[16:], 16)
	binary.LittleEndian.PutUint16(wav[20:], 1)
	binary.LittleEndian.PutUint16(wav[22:], 1)
	binary.LittleEndian.PutUint32(wav[24:], 8000)
	binary.LittleEndian.PutUint32(wav[28:], 16000)
	binary.LittleEndian.PutUint16(wav[32:], 2)
	binary.LittleEndian.PutUint16(wav[34:], 16)
	copy(wav[36:], "data")
	binary.LittleEndian.PutUint32(wav[40:], 2)
	audioPath := filepath.Join(work, "sample.wav")
	if err := os.WriteFile(audioPath, wav, 0600); err != nil {
		fail(&c, "%v", err)
		return c
	}
	for _, vendor := range []string{"anthropic", "openai", "gemini"} {
		body, err := renderMedia(bin, work, vendor, models[vendor], imagePath, "image/png")
		if err != nil {
			fail(&c, "%s supported image: %v", vendor, err)
			continue
		}
		if err := imageAttachment([]byte(body), vendor, imageData.Bytes()); err != nil {
			fail(&c, "%s supported image: %v", vendor, err)
		}
	}
	// A same-model positive image prevents an implementation that refuses ALL
	// media from passing. An unknown model is tested with an actual image, not
	// the original grader's plain text prompt.
	for _, sample := range []struct{ model, mime, category, path string }{
		{models["anthropic"], "audio/wav", "audio", audioPath},
		{"unknown-model-xyz", "image/png", "image", imagePath},
	} {
		_, err := renderMedia(bin, work, "anthropic", sample.model, sample.path, sample.mime)
		if err == nil || !strings.Contains(err.Error(), sample.model) || !strings.Contains(strings.ToLower(err.Error()), sample.category) {
			fail(&c, "refusal must name model %q and %s media: %v", sample.model, sample.category, err)
		}
	}
	// Capabilities describe attachments; they do not create a closed allowlist
	// for text-only use. This is an offline rendering control, not a paid call.
	if _, err := renderMedia(bin, work, "anthropic", "unknown-model-xyz", "", ""); err != nil {
		fail(&c, "unknown model rejected ordinary text: %v", err)
	}
	return c
}

func renderMedia(bin, work, vendor, model, locator, mime string) (string, error) {
	parts := []any{map[string]any{"type": "text", "text": "before attachment"}}
	if locator != "" {
		parts = append(parts, map[string]any{"type": "blob", "mime": mime, "ref": map[string]any{"kind": 1, "locator": locator}})
	}
	parts = append(parts, map[string]any{"type": "text", "text": "after attachment"})
	event := map[string]any{"seq": 1, "type": "message_received", "time": "2026-01-01T00:00:01Z", "message": map[string]any{"actor": "human", "parts": parts}}
	data, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	path := filepath.Join(work, "media.jsonl")
	if err := os.WriteFile(path, append(data, '\n'), 0600); err != nil {
		return "", err
	}
	return runModel(bin, work, vendor, model, "http://127.0.0.1:1", "", "render", path)
}

// Inspect the provider's actual attachment fields. Substring checks previously
// let a misspelled Gemini MIME field pass while the right bytes appeared nearby.
func imageAttachment(body []byte, vendor string, want []byte) error {
	var request map[string]any
	if err := json.Unmarshal(body, &request); err != nil {
		return err
	}
	turns := list(request["messages"])
	key := "content"
	if vendor == "gemini" {
		turns, key = list(request["contents"]), "parts"
	}
	// OpenAI carries the system prompt as another message. It is not the
	// human attachment-bearing turn and must not change this fixture's shape.
	var human []any
	for _, turn := range turns {
		if object(turn)["role"] == "user" {
			human = append(human, turn)
		}
	}
	turns = human
	if len(turns) != 1 {
		return fmt.Errorf("expected one image-bearing message, got %d", len(turns))
	}
	parts := list(object(turns[0])[key])
	if len(parts) != 3 || object(parts[0])["text"] != "before attachment" || object(parts[2])["text"] != "after attachment" {
		return fmt.Errorf("attachment or adjacent text/order missing")
	}
	part := object(parts[1])
	var encoded string
	switch vendor {
	case "anthropic":
		source := object(part["source"])
		if part["type"] != "image" || source["type"] != "base64" || source["media_type"] != "image/png" {
			return fmt.Errorf("image source/type/media_type missing")
		}
		encoded, _ = source["data"].(string)
	case "openai":
		url, _ := object(part["image_url"])["url"].(string)
		const prefix = "data:image/png;base64,"
		if part["type"] != "image_url" || !strings.HasPrefix(url, prefix) {
			return fmt.Errorf("image_url lacks the local PNG data URL")
		}
		encoded = strings.TrimPrefix(url, prefix)
	case "gemini":
		inline := object(part["inlineData"])
		if inline["mimeType"] != "image/png" {
			return fmt.Errorf("inlineData.mimeType missing or incorrect")
		}
		encoded, _ = inline["data"].(string)
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || !bytes.Equal(decoded, want) {
		return fmt.Errorf("attachment bytes differ from the local file: %v", err)
	}
	return nil
}

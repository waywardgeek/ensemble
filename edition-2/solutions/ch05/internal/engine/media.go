package engine

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ensemble/internal/common"
)

// The explicit table describes supported rendering on our current API surfaces.
// Unknown models remain usable for text; guessing their media capability would
// turn an omitted attachment into a plausible but misleading answer.
func LookupModel(model string) (common.ModelFeatures, bool) {
	switch model {
	case "claude-haiku-5-5", "claude-sonnet-4-6", "claude-opus-4-6", "claude-opus-5-5":
		return common.ModelFeatures{Media: common.MediaImage | common.MediaDocument}, true
	case "claude-3-haiku-20240307":
		return common.ModelFeatures{Media: common.MediaImage}, true
	case "gpt-5.4-mini", "gpt-4o", "gpt-4o-mini":
		return common.ModelFeatures{Media: common.MediaImage | common.MediaDocument}, true
	case "gemini-3.8-flash", "gemini-2.5-flash":
		return common.ModelFeatures{Media: common.MediaImage | common.MediaAudio | common.MediaVideo | common.MediaDocument}, true
	}
	return common.ModelFeatures{}, false
}
func mediaType(mime string) (common.Media, string) {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return common.MediaImage, "image"
	case strings.HasPrefix(mime, "audio/"):
		return common.MediaAudio, "audio"
	case strings.HasPrefix(mime, "video/"):
		return common.MediaVideo, "video"
	case mime == "application/pdf":
		return common.MediaDocument, "document"
	}
	return 0, mime
}
func (e *engine) refuseBlob(p common.Part) error {
	if p.Type != "blob" {
		return nil
	}
	kind, name := mediaType(p.MIME)
	model := e.parent.Config().Model
	features, known := LookupModel(model)
	if !known || kind == 0 || features.Media&kind == 0 {
		return fmt.Errorf("model %q does not support %s media", model, name)
	}
	if p.Ref.Kind != common.RefPath {
		return fmt.Errorf("model %q cannot render %s media Ref kind %d: only local paths are supported", model, name, p.Ref.Kind)
	}
	return nil
}
func (e *engine) blob(p common.Part) (string, error) {
	if err := e.refuseBlob(p); err != nil {
		return "", err
	}
	bytes, err := os.ReadFile(p.Ref.Locator)
	if err != nil {
		return "", fmt.Errorf("model %q reading %s media: %w", e.parent.Config().Model, p.MIME, err)
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
func (e *engine) openAIBlob(p common.Part) (map[string]any, error) {
	data, err := e.blob(p)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(p.MIME, "image/") {
		return map[string]any{"type": "image_url", "image_url": map[string]string{"url": "data:" + p.MIME + ";base64," + data}}, nil
	}
	return map[string]any{"type": "file", "file": map[string]string{"filename": filepath.Base(p.Ref.Locator), "file_data": "data:" + p.MIME + ";base64," + data}}, nil
}

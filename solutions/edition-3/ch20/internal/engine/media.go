package engine

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"ensemble/internal/common"
)

// Prepare a private request copy. History retains references, never base64
// inflation, and the exact request bytes are frozen before HTTP begins.
func prepareMedia(c common.Context, cfg common.Config, owners ...common.Sandbox) (common.Context, error) {
	c.Dialogue = append([]common.Entry(nil), c.Dialogue...)
	for i := range c.Dialogue {
		parts, err := loadMedia(c.Dialogue[i].Parts, cfg, owners...)
		if err != nil {
			return common.Context{}, err
		}
		c.Dialogue[i].Parts = parts
	}
	var err error
	c.Ephemera, err = loadMedia(c.Ephemera, cfg, owners...)
	return c, err
}
func loadMedia(parts common.PartList, cfg common.Config, owners ...common.Sandbox) (common.PartList, error) {
	out := append(common.PartList(nil), parts...)
	for i, p := range out {
		if len(p.Parts) > 0 {
			nested, err := loadMedia(p.Parts, cfg, owners...)
			if err != nil {
				return nil, err
			}
			p.Parts = nested
		}
		switch p.Type {
		case "blob", "image", "audio", "video":
			locator := ""
			if p.Ref != nil {
				locator = p.Ref.Locator
			}
			mediaType := p.MIME
			if mediaType == "" {
				mediaType = mime.TypeByExtension(filepath.Ext(locator))
			}
			mediaType, _, _ = mime.ParseMediaType(mediaType)
			kind, _, _ := strings.Cut(mediaType, "/")
			if p.Type != "blob" && kind != p.Type {
				return nil, fmt.Errorf("model %q: %s MIME type is missing or inconsistent", cfg.Model, p.Type)
			}
			f, known := common.LookupModel(cfg.Model)
			supported := known && (kind == "image" && f.Media&common.MediaImage != 0 || kind == "audio" && f.Media&common.MediaAudio != 0 || kind == "video" && f.Media&common.MediaVideo != 0)
			if !supported {
				if kind == "" {
					kind = p.Type
				}
				return nil, fmt.Errorf("model %q does not support %s media", cfg.Model, kind)
			}
			// The declared renderer must also match the configured vendor.
			if (kind != "image" && cfg.Vendor != common.Gemini) || (cfg.Vendor != common.Gemini && cfg.Vendor != common.OpenAI && cfg.Vendor != common.Anthropic) {
				return nil, fmt.Errorf("model %q: %s is unsupported by %s renderer", cfg.Model, kind, cfg.Vendor)
			}
			if p.Ref == nil || p.Ref.Kind != common.RefPath {
				return nil, fmt.Errorf("model %q: %s rendering requires RefPath; URI and handle attachments are not supported", cfg.Model, kind)
			}
			if strings.HasPrefix(locator, "file://") {
				u, err := url.Parse(locator)
				if err != nil || u.Host != "" && u.Host != "localhost" {
					return nil, fmt.Errorf("model %q: %s needs a local file", cfg.Model, kind)
				}
				locator = u.Path
			}
			if locator == "" || strings.Contains(locator, "://") {
				return nil, fmt.Errorf("model %q: %s needs a local file reference", cfg.Model, kind)
			}
			if !filepath.IsAbs(locator) {
				locator = filepath.Join(cfg.BaseDir, locator)
			}
			readFile := os.ReadFile
			if len(owners) > 0 && owners[0] != nil {
				readFile = owners[0].ReadFile
			} else if cfg.SandboxRoot != "" {
				return nil, fmt.Errorf("media requires its owning sandbox")
			}
			data, err := readFile(locator)
			if err != nil {
				return nil, fmt.Errorf("model %q: reading %s media: %w", cfg.Model, kind, err)
			}
			if len(data) == 0 {
				return nil, fmt.Errorf("model %q: empty %s media", cfg.Model, kind)
			}
			p.MIME = mediaType
			p.Data = raw(base64.StdEncoding.EncodeToString(data))
		}
		out[i] = p
	}
	return out, nil
}
func mediaData(p common.Part) string { var s string; _ = json.Unmarshal(p.Data, &s); return s }

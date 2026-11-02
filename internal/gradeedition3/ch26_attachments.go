package gradeedition3

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// attachmentPNG is the payload oracle for the pending external attachment scope.
// It awards no points by itself: addressed admission, acknowledgement, stale-send
// refusal and the real candidate lifecycle still need public-interface checks.
func attachmentPNG(body []byte, caption string, challenge imageChallenge) ([]byte, error) {
	if len(body) > 16<<20 || caption == "" {
		return nil, fmt.Errorf("unbounded request or missing caption challenge")
	}
	var request struct {
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, err
	}
	type part struct {
		Type   string `json:"type"`
		Text   string `json:"text"`
		Source struct {
			Type string `json:"type"`
			MIME string `json:"media_type"`
			Data string `json:"data"`
		} `json:"source"`
	}
	var matched []part
	captions := 0
	for _, message := range request.Messages {
		if message.Role != "user" {
			continue
		}
		// A text-only copy must count too; otherwise a duplicate admitted message
		// could hide beside the correctly paired image message.
		var text string
		if json.Unmarshal(message.Content, &text) == nil {
			if text == caption {
				captions++
			}
			continue
		}
		var parts []part
		if err := json.Unmarshal(message.Content, &parts); err != nil {
			return nil, fmt.Errorf("malformed human message content")
		}
		for _, p := range parts {
			if p.Type == "text" && p.Text == caption {
				captions++
				matched = parts
			}
		}
	}
	if captions != 1 || matched == nil {
		return nil, fmt.Errorf("expected one image-bearing human caption, got %d", captions)
	}
	images := 0
	var pixels []byte
	for _, p := range matched {
		if p.Type != "image" {
			continue
		}
		images++
		source := p.Source
		if source.Type != "base64" || source.MIME != "image/png" || len(source.Data) == 0 || len(source.Data) > base64.StdEncoding.EncodedLen(8<<20) {
			return nil, fmt.Errorf("missing, unbounded, or non-PNG attachment payload")
		}
		var err error
		pixels, err = base64.StdEncoding.Strict().DecodeString(source.Data)
		if err != nil || len(pixels) > 8<<20 {
			return nil, fmt.Errorf("invalid attachment encoding")
		}
	}
	if images != 1 {
		return nil, fmt.Errorf("expected one accompanying image, got %d", images)
	}
	return validatePreparedImagePNG(pixels, challenge)
}

package gradeedition3

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"testing"
)

func TestAttachmentPayloadOracle(t *testing.T) {
	challenge, _, err := newImageChallenge()
	if err != nil {
		t.Fatal(err)
	}
	raster := image.NewNRGBA(image.Rect(0, 0, 2048, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 2048; x++ {
			raster.SetNRGBA(x, y, challenge.colors[x/512])
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, raster); err != nil {
		t.Fatal(err)
	}
	imagePart := func(mime, data string) any {
		return map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": mime, "data": data}}
	}
	caption := "fixture-caption"
	text := map[string]any{"type": "text", "text": caption}
	pixels := imagePart("image/png", base64.StdEncoding.EncodeToString(encoded.Bytes()))
	message := func(role string, content any) any { return map[string]any{"role": role, "content": content} }
	request := func(messages ...any) []byte {
		b, err := json.Marshal(map[string]any{"messages": messages})
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	valid := message("user", []any{text, pixels})
	delivered, err := attachmentPNG(request(valid), caption, challenge)
	if err != nil || !bytes.Equal(delivered, encoded.Bytes()) {
		t.Fatalf("valid pixels refused: %v", err)
	}
	cases := map[string][]byte{
		"text only":            request(message("user", caption)),
		"metadata only":        request(message("user", []any{text, map[string]any{"type": "text", "text": "image at fixture.png"}})),
		"assistant role":       request(message("assistant", []any{text, pixels})),
		"split messages":       request(message("user", []any{text}), message("user", []any{pixels})),
		"duplicate images":     request(message("user", []any{text, pixels, pixels})),
		"duplicate captions":   request(message("user", []any{text, text, pixels})),
		"duplicate messages":   request(valid, valid),
		"extra text admission": request(message("user", caption), valid),
		"missing bytes":        request(message("user", []any{text, imagePart("image/png", "")})),
		"corrupt encoding":     request(message("user", []any{text, imagePart("image/png", "!invalid")})),
		"wrong mime":           request(message("user", []any{text, imagePart("image/jpeg", base64.StdEncoding.EncodeToString(encoded.Bytes()))})),
		"malformed request":    []byte("{"),
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := attachmentPNG(body, caption, challenge); err == nil {
				t.Fatal("wrong-policy payload passed")
			}
		})
	}
	if _, err := attachmentPNG(request(valid), "different-caption", challenge); err == nil {
		t.Fatal("wrong caption passed")
	}
	changed := challenge.colors[0]
	changed.R ^= 128
	raster.SetNRGBA(10, 3, changed)
	encoded.Reset()
	if err := png.Encode(&encoded, raster); err != nil {
		t.Fatal(err)
	}
	damaged := imagePart("image/png", base64.StdEncoding.EncodeToString(encoded.Bytes()))
	if _, err := attachmentPNG(request(message("user", []any{text, damaged})), caption, challenge); err == nil {
		t.Fatal("corrupted pixels passed")
	}
}

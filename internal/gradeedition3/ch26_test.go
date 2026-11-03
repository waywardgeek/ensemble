package gradeedition3

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChapter26ImagePayloadNegativeControls(t *testing.T) {
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
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, raster); err != nil {
		t.Fatal(err)
	}
	source := map[string]any{"type": "base64", "media_type": "image/png", "data": base64.StdEncoding.EncodeToString(buffer.Bytes())}
	imagePart := map[string]any{"type": "image", "source": source}
	block := func(id string, content any) map[string]any {
		return map[string]any{"type": "tool_result", "tool_use_id": id, "content": content}
	}
	request := func(blocks ...any) []byte {
		body, _ := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": blocks}}})
		return body
	}
	valid := block("wanted", []any{imagePart})
	if _, err := imageToolPNG(request(valid), "wanted", challenge); err != nil {
		t.Fatal(err)
	}
	missing := map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": ""}}
	corrupt := map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": "not base64"}}
	cases := map[string][]byte{
		"text-only":        request(block("wanted", "Image saved at fixture.png")),
		"metadata-only":    request(block("wanted", []any{map[string]any{"type": "text", "text": "2048x32 PNG"}})),
		"wrong pairing":    request(block("another", []any{imagePart})),
		"duplicate result": request(valid, valid),
		"missing data":     request(block("wanted", []any{missing})),
		"corrupt data":     request(block("wanted", []any{corrupt})),
		"duplicate image":  request(block("wanted", []any{imagePart, imagePart})),
	}
	failed := block("wanted", []any{imagePart})
	failed["is_error"] = true
	cases["tool error"] = request(failed)
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := imageToolPNG(body, "wanted", challenge); err == nil {
				t.Fatal("wrong-policy payload passed")
			}
		})
	}
	other := challenge
	other.colors[0].R ^= 128 // Always differs by 128, beyond resampling tolerance.
	if _, err := imageToolPNG(request(valid), "wanted", other); err == nil {
		t.Fatal("wrong pixels passed")
	}
	changed := challenge.colors[0]
	changed.R ^= 128
	raster.SetNRGBA(10, 3, changed)
	buffer.Reset()
	if err := png.Encode(&buffer, raster); err != nil {
		t.Fatal(err)
	}
	oneBadPixel := map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": base64.StdEncoding.EncodeToString(buffer.Bytes())}}
	if _, err := imageToolPNG(request(block("wanted", []any{oneBadPixel})), "wanted", challenge); err == nil {
		t.Fatal("unsampled pixel corruption passed")
	}
	if _, err := imageToolPNG([]byte(`{`), "wanted", challenge); err == nil {
		t.Fatal("malformed JSON passed")
	}
}

func TestMediaOutputBarrierAndBound(t *testing.T) {
	output := &mediaOutput{ended: make(chan struct{}, 2)}
	output.Write([]byte(`{"observation":"turn_`))
	output.Write([]byte("ended\"}\n"))
	select {
	case <-output.ended:
	default:
		t.Fatal("split completion was lost")
	}
	output.Write(make([]byte, 2<<20))
	if !output.overflow {
		t.Fatal("output bound missing")
	}
	if ValidMediaScope("") || !ValidMediaScope("voice") || ValidMediaScope(" image-files") || !ValidMediaScope("image-files") {
		t.Fatal("scope is not exact")
	}
}

func TestChapter26RejectsExternalWrongPolicyCandidates(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"metadata", "lost-replay"} {
		t.Run(mode, func(t *testing.T) {
			wrapper := filepath.Join(t.TempDir(), "candidate")
			script := "#!/bin/sh\nexport EN_MEDIA_NEGATIVE=" + mode + "\nexec '" + strings.ReplaceAll(executable, "'", "'\\''") + "' -test.run='^TestMediaNegativeCLIHelper$'\n"
			if err := os.WriteFile(wrapper, []byte(script), 0700); err != nil {
				t.Fatal(err)
			}
			report := Chapter26Images(wrapper)
			if report.Passed || report.Score != 0 || report.MaxScore != 20 {
				t.Fatalf("wrong policy passed: %+v", report)
			}
			detail := strings.Join(report.Checks[0].Details, " ")
			expected := "first image delivery: expected one image"
			if mode == "lost-replay" {
				expected = "replay after source deletion"
			}
			if !strings.Contains(detail, expected) {
				t.Fatalf("failed for wrong reason: %s", detail)
			}
		})
	}
}

// This deliberately incomplete candidate is launched only by the external-grade
// negative controls. It does not call the implementation under test or acquire
// real credentials. The helper must exit before the Go test runner prints PASS.
func TestMediaNegativeCLIHelper(t *testing.T) {
	mode := os.Getenv("EN_MEDIA_NEGATIVE")
	if mode == "" {
		return
	}
	fail := func(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(3) }
	input := bufio.NewScanner(os.Stdin)
	if !input.Scan() {
		os.Exit(3)
	}
	post := func(messages []any) map[string]any {
		body, _ := json.Marshal(map[string]any{"model": "claude-3-haiku-20240307", "messages": messages, "max_tokens": 1024})
		request, err := http.NewRequest("POST", os.Getenv("LLM_BASE_URL")+"/v1/messages", bytes.NewReader(body))
		if err != nil {
			fail(err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("x-api-key", "synthetic-grade-key")
		request.Header.Set("anthropic-version", "2023-06-01")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			fail(err)
		}
		raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		response.Body.Close()
		if err != nil {
			fail(err)
		}
		var result map[string]any
		if err := json.Unmarshal(raw, &result); err != nil {
			fail(err)
		}
		return result
	}
	response := post([]any{map[string]any{"role": "user", "content": "inspect fixture"}})
	call := response["content"].([]any)[0].(map[string]any)
	id := call["id"].(string)
	content := []any{map[string]any{"type": "text", "text": "Saved 2048x32 screenshot; pixels are available at fixture.png"}}
	if mode == "lost-replay" {
		args := call["input"].(map[string]any)
		raw, err := os.ReadFile(args["path"].(string))
		if err != nil {
			fail(err)
		}
		imageSource, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			fail(err)
		}
		resized := image.NewNRGBA(image.Rect(0, 0, 2048, 32))
		for y := 0; y < 32; y++ {
			for x := 0; x < 2048; x++ {
				resized.Set(x, y, imageSource.At(x*2, y*2))
			}
		}
		var encoded bytes.Buffer
		if err := png.Encode(&encoded, resized); err != nil {
			fail(err)
		}
		content = []any{map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": base64.StdEncoding.EncodeToString(encoded.Bytes())}}}
	}
	post([]any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": id, "content": content}}}})
	fmt.Println(`{"observation":"turn_ended"}`)
	if input.Scan() {
		post([]any{map[string]any{"role": "user", "content": "The image used to be available."}})
		fmt.Println(`{"observation":"turn_ended"}`)
		input.Scan()
	}
	os.Exit(0)
}

package gradeedition3

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
	"github.com/waywardgeek/ensemble/internal/grade"
)

// Chapter26Images grades only the declared image-file scope, not all multimedia.
func Chapter26Images(bin string) grade.Report {
	check := grade.Check{ID: "image-file-delivery", Title: "Image-file pixels, preparation, and deleted-source replay", Points: 20}
	if err := runImageFileCLI(bin); err != nil {
		check.Details = []string{err.Error()}
	} else {
		check.Passed = true
		check.Earned = check.Points
		check.Details = []string{"Randomized PNG pixels returned in the matching tool result; 4096x64 prepared to 2048x32; source preserved, then deleted; identical payload replayed on the next turn; CLI joined."}
	}
	return grade.NewTitledReport("Chapter 26 — image-files scope only (20 points; 80 chapter points ungraded)", []grade.Check{check}, "")
}

type imageChallenge struct{ colors [4]color.NRGBA }

func newImageChallenge() (imageChallenge, []byte, error) {
	var challenge imageChallenge
	var seed [12]byte
	if _, err := rand.Read(seed[:]); err != nil {
		return challenge, nil, err
	}
	source := image.NewNRGBA(image.Rect(0, 0, 4096, 64))
	for band := range challenge.colors {
		challenge.colors[band] = color.NRGBA{32 + seed[band*3]%192, 32 + seed[band*3+1]%192, 32 + seed[band*3+2]%192, 255}
	}
	for y := 0; y < 64; y++ {
		for x := 0; x < 4096; x++ {
			source.SetNRGBA(x, y, challenge.colors[x/1024])
		}
	}
	var encoded bytes.Buffer
	err := png.Encode(&encoded, source)
	return challenge, encoded.Bytes(), err
}

func imageToolPNG(body []byte, callID string, challenge imageChallenge) ([]byte, error) {
	if len(body) > 16<<20 {
		return nil, fmt.Errorf("provider request exceeds grader bound")
	}
	var request struct {
		Messages []struct {
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, err
	}
	var matched json.RawMessage
	count := 0
	for _, message := range request.Messages {
		var blocks []struct {
			Type    string          `json:"type"`
			ID      string          `json:"tool_use_id"`
			Error   bool            `json:"is_error"`
			Content json.RawMessage `json:"content"`
		}
		if json.Unmarshal(message.Content, &blocks) != nil {
			continue
		}
		for _, block := range blocks {
			if block.Type == "tool_result" && block.ID == callID {
				count++
				if block.Error {
					return nil, fmt.Errorf("image tool returned an error: %s", mediaFailureText(block.Content))
				}
				matched = block.Content
			}
		}
	}
	if count != 1 {
		return nil, fmt.Errorf("expected one matching image tool result, got %d", count)
	}
	var parts []struct {
		Type   string `json:"type"`
		Source struct {
			Type string `json:"type"`
			MIME string `json:"media_type"`
			Data string `json:"data"`
		} `json:"source"`
	}
	if json.Unmarshal(matched, &parts) != nil {
		return nil, fmt.Errorf("image result is text/metadata, not typed pixels")
	}
	var pixels []byte
	images := 0
	for _, part := range parts {
		if part.Type == "image" {
			images++
			if part.Source.Type != "base64" || part.Source.MIME != "image/png" || len(part.Source.Data) == 0 || len(part.Source.Data) > base64.StdEncoding.EncodedLen(8<<20) {
				return nil, fmt.Errorf("missing, unbounded, or non-PNG inline payload")
			}
			var err error
			pixels, err = base64.StdEncoding.Strict().DecodeString(part.Source.Data)
			if err != nil || len(pixels) > 8<<20 {
				return nil, fmt.Errorf("invalid or oversized PNG encoding")
			}
		}
	}
	if images != 1 {
		return nil, fmt.Errorf("expected one image, got %d", images)
	}
	return validatePreparedImagePNG(pixels, challenge)
}

// Shared pixel oracle for tool results and human attachments. Wire ownership is
// checked separately; matching pixels alone cannot prove correct admission.
func validatePreparedImagePNG(pixels []byte, challenge imageChallenge) ([]byte, error) {
	config, err := png.DecodeConfig(bytes.NewReader(pixels))
	if err != nil || config.Width != 2048 || config.Height != 32 {
		return nil, fmt.Errorf("expected prepared 2048x32 PNG")
	}
	raster, err := png.Decode(bytes.NewReader(pixels))
	if err != nil {
		return nil, fmt.Errorf("PNG cannot be decoded")
	}
	// Check interiors, not just four sample points. The two-pixel exclusion at
	// each band boundary allows legitimate resampling between adjacent colors.
	for band, want := range challenge.colors {
		for y := 0; y < 32; y++ {
			for x := band*512 + 2; x < (band+1)*512-2; x++ {
				got := color.NRGBAModel.Convert(raster.At(x, y)).(color.NRGBA)
				for _, pair := range [][2]uint8{{got.R, want.R}, {got.G, want.G}, {got.B, want.B}, {got.A, want.A}} {
					difference := int(pair[0]) - int(pair[1])
					if difference < -3 || difference > 3 {
						return nil, fmt.Errorf("prepared pixels do not match randomized source")
					}
				}
			}
		}
	}
	return pixels, nil
}

// Extract only bounded text from an error result, never image data.
func mediaFailureText(content json.RawMessage) string {
	var text string
	if json.Unmarshal(content, &text) != nil {
		var parts []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}
		if json.Unmarshal(content, &parts) == nil {
			for _, part := range parts {
				if part.Type == "text" {
					text = part.Text
					break
				}
			}
		}
	}
	if len(text) > 512 {
		text = text[:512] + " [truncated]"
	}
	return text
}

// This observer bounds candidate diagnostics independently of the media request.
// turn_ended is a barrier, not evidence that the image was actually delivered.
type mediaOutput struct {
	pending  []byte
	ended    chan struct{}
	overflow bool
	total    int
}

func (output *mediaOutput) Write(data []byte) (int, error) {
	output.total += len(data)
	if output.total > 2<<20 {
		output.overflow = true
		return len(data), nil
	}
	output.pending = append(output.pending, data...)
	for {
		index := bytes.IndexByte(output.pending, '\n')
		if index < 0 {
			break
		}
		var event struct {
			Observation string `json:"observation"`
		}
		if json.Unmarshal(output.pending[:index], &event) == nil && event.Observation == "turn_ended" {
			select {
			case output.ended <- struct{}{}:
			default:
			}
		}
		output.pending = output.pending[index+1:]
	}
	return len(data), nil
}

func runImageFileCLI(bin string) (result error) {
	root, err := os.MkdirTemp("", "ensemble-grade26-images-")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(root); err != nil {
			if result == nil {
				result = fmt.Errorf("scratch cleanup failed: %w", err)
			} else {
				result = fmt.Errorf("%w; scratch cleanup also failed: %v", result, err)
			}
		}
	}()
	work := filepath.Join(root, "work")
	home := filepath.Join(root, "home")
	skills := filepath.Join(root, "skills")
	for _, directory := range []string{work, home, filepath.Join(skills, "primary")} {
		if err = os.MkdirAll(directory, 0700); err != nil {
			return err
		}
	}
	manual := "---\nname: primary\ntype: primary\ndescription: Isolated image-file grading authority\ntools:\n  - inspect_image\n  - wait_for_job\n---\nInspect the supplied fixture image. No other authority is granted.\n"
	if err = os.WriteFile(filepath.Join(skills, "primary", "SKILL.md"), []byte(manual), 0600); err != nil {
		return err
	}
	challenge, source, err := newImageChallenge()
	if err != nil {
		return err
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return err
	}
	callID := "image-" + hex.EncodeToString(nonce[:])
	path := filepath.Join(work, callID+".png")
	if err = os.WriteFile(path, source, 0600); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	requests := make(chan []byte, 4)
	var calls atomic.Int32
	server := fakevendor.NewWithOptions(nil, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		number := calls.Add(1)
		select {
		case requests <- append([]byte(nil), body...):
		default:
		}
		if number == 1 {
			return &fakevendor.Reply{Tools: []fakevendor.ToolCall{hintCall(callID, "inspect_image", map[string]any{"path": path, "ai_callback_delay": 10})}}
		}
		return &fakevendor.Reply{Text: "IMAGE_GRADE_TURN_COMPLETE"}
	}})
	defer server.Close()
	command := exec.CommandContext(ctx, bin)
	command.Dir = work
	command.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "TMPDIR=" + root, "DOTNET_CLI_HOME=" + home, "DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1", "DOTNET_CLI_TELEMETRY_OPTOUT=1", "LLM_VENDOR=anthropic", "LLM_MODEL=claude-3-haiku-20240307", "LLM_BASE_URL=" + server.URL(), "LLM_API_KEY=synthetic-grade-key", "LLM_REQUIRE_TOOL_EXPLANATION=false", "EN_DISABLE_STREAMING=1", "EN_DISABLE_SAVE=1", "EN_SKILLS_DIR=" + skills, "EN_PRIMARY_SKILL=primary", "EN_STUB_TOOL_RESULTS=0", "EN_CONTEXT_BUDGET_BYTES=400000"}
	input, err := command.StdinPipe()
	if err != nil {
		return err
	}
	output := &mediaOutput{ended: make(chan struct{}, 8)}
	diagnostics := &mediaOutput{ended: make(chan struct{}, 8)}
	command.Stdout = output
	command.Stderr = diagnostics
	if err = command.Start(); err != nil {
		input.Close()
		return err
	}
	exited := make(chan error, 1)
	finished := make(chan struct{})
	go func() { exited <- command.Wait(); close(finished) }()
	// Always join the candidate before reading its final diagnostics or removing files.
	defer func() {
		input.Close()
		select {
		case err = <-exited:
		case <-time.After(2 * time.Second):
			cancel()
			err = <-exited
		}
		if result == nil && err != nil {
			result = fmt.Errorf("candidate did not exit cleanly: %w", err)
		}
		if result == nil && (output.overflow || diagnostics.overflow) {
			result = fmt.Errorf("candidate exceeded bounded diagnostics")
		}
		if result == nil && calls.Load() != 3 {
			result = fmt.Errorf("expected three provider requests, got %d", calls.Load())
		}
	}()
	send := func(text string) error {
		data, _ := json.Marshal(map[string]any{"kind": "prompt", "text": text})
		_, err := fmt.Fprintln(input, string(data))
		return err
	}
	next := func() ([]byte, error) {
		select {
		case body := <-requests:
			return body, nil
		case <-finished:
			return nil, fmt.Errorf("candidate exited before the next provider request")
		case <-ctx.Done():
			return nil, fmt.Errorf("image grade timed out waiting for provider request")
		}
	}
	ended := func() error {
		select {
		case <-output.ended:
			return nil
		case <-ctx.Done():
			return fmt.Errorf("image grade timed out waiting for turn completion")
		}
	}
	if err = send("Inspect the image fixture."); err != nil {
		return err
	}
	if _, err = next(); err != nil {
		return err
	}
	body, err := next()
	if err != nil {
		return err
	}
	first, err := imageToolPNG(body, callID, challenge)
	if err != nil {
		return fmt.Errorf("first image delivery: %w", err)
	}
	if err = ended(); err != nil {
		return err
	}
	unchanged, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(unchanged, source) {
		return fmt.Errorf("inspection changed the original image")
	}
	if err = os.Remove(path); err != nil {
		return err
	}
	if err = send("Discuss the retained image without reading the file again."); err != nil {
		return err
	}
	body, err = next()
	if err != nil {
		return err
	}
	replay, err := imageToolPNG(body, callID, challenge)
	if err != nil {
		return fmt.Errorf("replay after source deletion: %w", err)
	}
	if !bytes.Equal(first, replay) {
		return fmt.Errorf("replay changed the retained PNG")
	}
	if err = ended(); err != nil {
		return err
	}
	return nil
}

// A scope name is deliberately not a permissive substring match.
func ValidMediaScope(value string) bool { return value == "image-files" || value == "attachments" }

package gradeedition3

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func hostWire(t *testing.T, values ...any) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]any{"messages": []any{map[string]any{"role": "user", "content": values}}})
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func hostResult(id string, content any) any {
	return map[string]any{"type": "tool_result", "tool_use_id": id, "content": content}
}
func hostImage(pixels []byte) any {
	return []any{map[string]any{"type": "image", "source": map[string]any{"type": "base64", "media_type": "image/png", "data": base64.StdEncoding.EncodeToString(pixels)}}}
}
func hostScriptFixture(t *testing.T) (*hostScreenScript, []byte) {
	t.Helper()
	challenge, _, err := newImageChallenge()
	if err != nil {
		t.Fatal(err)
	}
	for index := range challenge.colors {
		challenge.colors[index].R = uint8(40 + 40*index)
	} // Deterministically distinct bands.
	raster := image.NewNRGBA(image.Rect(0, 0, 64, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			raster.SetNRGBA(x, y, challenge.colors[x/16])
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, raster); err != nil {
		t.Fatal(err)
	}
	return &hostScreenScript{challenge: challenge, width: 64, height: 32}, encoded.Bytes()
}
func hostScriptStart(t *testing.T, script *hostScreenScript) {
	t.Helper()
	reply, err := script.step([]byte(`{}`))
	if err != nil || reply.Tools[0].Name != "load_skill" {
		t.Fatalf("load: %v", err)
	}
	schema := []byte(`{"tools":[{"name":"screenshot","input_schema":{"type":"object","properties":{"display":{"type":"integer"}}}},{"name":"list_displays","input_schema":{"type":"object"}}]}`)
	reply, err = script.step(schema)
	if err != nil || reply.Tools[0].Name != "list_displays" {
		t.Fatalf("inventory: %v", err)
	}
	inventory := `{"displays":[{"display":0,"width":32,"height":16,"pixel_width":64,"pixel_height":32}]}`
	reply, err = script.step(hostWire(t, hostResult("host-displays", inventory)))
	if err != nil || reply.Tools[0].Name != "screenshot" {
		t.Fatalf("default: %v", err)
	}
	var args map[string]any
	if err := json.Unmarshal([]byte(reply.Tools[0].Args), &args); err != nil {
		t.Fatal(err)
	}
	if _, present := args["display"]; present {
		t.Fatal("default capture specified a selector")
	}
}
func TestHostScreenScriptDelayedAndReplay(t *testing.T) {
	script, pixels := hostScriptFixture(t)
	hostScriptStart(t, script)
	for capture := 0; capture < 2; capture++ {
		reply, err := script.step(hostWire(t, hostResult(script.current, fmt.Sprintf("job %d, status: running", 7+capture))))
		if err != nil || reply.Tools[0].Name != "wait_for_job" {
			t.Fatalf("wait: %v", err)
		}
		reply, err = script.step(hostWire(t, hostResult(script.current, hostImage(pixels))))
		if err != nil {
			t.Fatal(err)
		}
		if capture == 0 {
			var args map[string]any
			if err := json.Unmarshal([]byte(reply.Tools[0].Args), &args); err != nil {
				t.Fatal(err)
			}
			if args["display"] != float64(0) {
				t.Fatal("explicit capture did not select zero")
			}
		} else if reply.Text != "HOST_NATIVE_CAPTURE_COMPLETE" {
			t.Fatal("capture did not finish")
		}
	}
	script.replayAllowed = true
	replay := hostWire(t, hostResult(script.imageIDs[0], hostImage(pixels)), hostResult(script.imageIDs[1], hostImage(pixels)))
	reply, err := script.step(replay)
	if err != nil || reply.Text != "HOST_NATIVE_REPLAY_COMPLETE" || len(reply.Tools) != 0 {
		t.Fatalf("replay: %v", err)
	}
	if script.stage != 6 || len(script.payloads) != 2 {
		t.Fatal("wrong final script state")
	}
}
func TestHostScreenScriptNegativeControls(t *testing.T) {
	t.Run("no-delayed-job", func(t *testing.T) {
		script, pixels := hostScriptFixture(t)
		hostScriptStart(t, script)
		if _, err := script.step(hostWire(t, hostResult(script.current, hostImage(pixels)))); err == nil {
			t.Fatal("accepted untested delayed delivery")
		}
	})
	t.Run("changed-handle", func(t *testing.T) {
		script, _ := hostScriptFixture(t)
		hostScriptStart(t, script)
		if _, err := script.step(hostWire(t, hostResult(script.current, "job 7, status: running"))); err != nil {
			t.Fatal(err)
		}
		if _, err := script.step(hostWire(t, hostResult(script.current, "job 8, status: running"))); err == nil {
			t.Fatal("accepted changed capture handle")
		}
	})
	t.Run("metadata-only", func(t *testing.T) {
		script, _ := hostScriptFixture(t)
		hostScriptStart(t, script)
		if _, err := script.step(hostWire(t, hostResult(script.current, "saved screenshot.png"))); err == nil {
			t.Fatal("accepted metadata")
		}
	})
	t.Run("replay-lost", func(t *testing.T) {
		script, pixels := hostScriptFixture(t)
		script.stage = 5
		script.replayAllowed = true
		script.payloads = [][]byte{pixels, pixels}
		script.imageIDs = []string{"a", "b"}
		if _, err := script.step(hostWire(t, hostResult("a", hostImage(pixels)))); err == nil {
			t.Fatal("accepted missing replay pixels")
		}
	})
}

func TestHostScreenReplayRequiresHumanBoundary(t *testing.T) {
	script, _ := hostScriptFixture(t)
	script.stage = 5
	if _, err := script.step([]byte(`{}`)); err == nil {
		t.Fatal("accepted replay without new human turn")
	}
}

func TestHostScreenRejectsPregrantTools(t *testing.T) {
	script, _ := hostScriptFixture(t)
	if _, err := script.step([]byte(`{"tools":[{"name":"screenshot"}]}`)); err == nil {
		t.Fatal("host tool advertised before grant")
	}
}

func TestHostRecallRequestRouting(t *testing.T) {
	judge := `{"system":[{"type":"text","text":"You select relevant stored context for the latest user query."}],"messages":[]}`
	if !hostRecallRequest([]byte(judge)) {
		t.Fatal("recall judge not identified")
	}
	for _, ordinary := range []string{
		`{"system":"ordinary","messages":[{"content":"You select relevant stored context for the latest user query."}]}`,
		`{"system":"You select relevant stored context for the latest user query.","tools":[{"name":"load_skill"}]}`,
		`{`,
	} {
		if hostRecallRequest([]byte(ordinary)) {
			t.Fatal("ordinary or invalid request diverted")
		}
	}
}

func TestHostFailureImageRetention(t *testing.T) {
	script, pixels := hostScriptFixture(t)
	script.stage, script.running, script.current = 3, true, "failed-image"
	script.failureImagePath = filepath.Join(t.TempDir(), "failure.png")
	script.challenge.colors[0].R = 0
	if _, err := script.step(hostWire(t, hostResult(script.current, hostImage(pixels)))); err == nil {
		t.Fatal("accepted altered pattern")
	}
	saved, err := os.ReadFile(script.failureImagePath)
	if err != nil || !bytes.Equal(saved, pixels) {
		t.Fatal("failed pixels were not retained exactly")
	}
	sentinel := []byte("preserve existing")
	if err := os.WriteFile(script.failureImagePath, sentinel, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := script.step(hostWire(t, hostResult(script.current, hostImage(pixels)))); err == nil {
		t.Fatal("overwriting failure not reported")
	}
	saved, err = os.ReadFile(script.failureImagePath)
	if err != nil || !bytes.Equal(saved, sentinel) {
		t.Fatal("existing evidence overwritten")
	}
}

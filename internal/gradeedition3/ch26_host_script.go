package gradeedition3

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// Pure provider script. It performs no native operation itself. The real CLI
// runner must require fresh native authorization before using its tool replies.
type hostScreenScript struct {
	failureImagePath     string
	challenge            imageChallenge
	width, height        int
	stage, waits, handle int
	current              string
	running              bool
	replayAllowed        bool
	captureHandles       []int
	payloads             [][]byte
	imageIDs             []string
}

// A fresh CLI can recall skill text before its main turn. Answer only the
// tool-free judge, without spending a step in the host capture protocol.
func hostRecallRequest(body []byte) bool {
	var request struct {
		System json.RawMessage   `json:"system"`
		Tools  []json.RawMessage `json:"tools"`
	}
	return json.Unmarshal(body, &request) == nil && len(request.Tools) == 0 &&
		bytes.Contains(request.System, []byte("You select relevant stored context for the latest user query."))
}

func (script *hostScreenScript) tool(id, name string, args map[string]any) *fakevendor.Reply {
	script.current = id
	return &fakevendor.Reply{Text: "Use only the agreed fully synthetic primary display.", Tools: []fakevendor.ToolCall{hintCall(id, name, args)}}
}

func (script *hostScreenScript) step(body []byte) (*fakevendor.Reply, error) {
	if len(body) > 16<<20 {
		return nil, fmt.Errorf("host provider request exceeds bound")
	}
	switch script.stage {
	case 0:
		var initial struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		}
		if err := json.Unmarshal(body, &initial); err != nil {
			return nil, err
		}
		for _, tool := range initial.Tools {
			if tool.Name == "screenshot" || tool.Name == "list_displays" {
				return nil, fmt.Errorf("host tools advertised before explicit skill grant")
			}
		}
		script.stage = 1
		return script.tool("host-load", "load_skill", map[string]any{"name": "host-screen", "ai_callback_delay": 10}), nil
	case 1:
		var request struct {
			Tools []struct {
				Name   string          `json:"name"`
				Schema json.RawMessage `json:"input_schema"`
			} `json:"tools"`
		}
		if err := json.Unmarshal(body, &request); err != nil {
			return nil, err
		}
		screenshots, displays := 0, 0
		for _, tool := range request.Tools {
			if tool.Name == "screenshot" {
				screenshots++
				if err := screenshotSelectionSchema(tool.Schema); err != nil {
					return nil, err
				}
			}
			if tool.Name == "list_displays" {
				displays++
			}
		}
		if screenshots != 1 || displays != 1 {
			load, loadError := hostToolText(body, script.current)
			return nil, fmt.Errorf("host tools not uniquely declared after skill load (%d screenshot, %d list_displays); load result: %s; error: %v", screenshots, displays, load, loadError)
		}
		script.stage = 2
		return script.tool("host-displays", "list_displays", map[string]any{"ai_callback_delay": 10}), nil
	case 2:
		text, err := hostToolText(body, script.current)
		if err != nil {
			return nil, err
		}
		var inventory struct {
			Displays []struct {
				Display, Width, Height int
				PixelWidth             int `json:"pixel_width"`
				PixelHeight            int `json:"pixel_height"`
			} `json:"displays"`
		}
		if err := json.Unmarshal([]byte(text), &inventory); err != nil {
			return nil, fmt.Errorf("invalid display inventory: %w", err)
		}
		found := 0
		for _, display := range inventory.Displays {
			if display.Display == 0 {
				found++
				if display.Width <= 0 || display.Height <= 0 || display.PixelWidth <= 0 || display.PixelHeight <= 0 {
					return nil, fmt.Errorf("invalid primary display dimensions")
				}
				if display.PixelWidth > 32768 || display.PixelHeight > 32768 || int64(display.PixelWidth)*int64(display.PixelHeight) > 24_000_000 {
					return nil, fmt.Errorf("primary display exceeds bounded source pixels")
				}
				// Use backing pixels and the actual preparation floor rule, not
				// logical dimensions or an assumed 1:1 display scale.
				scale := math.Min(1, 2048/math.Max(float64(display.PixelWidth), float64(display.PixelHeight)))
				width := int(math.Floor(float64(display.PixelWidth) * scale))
				height := int(math.Floor(float64(display.PixelHeight) * scale))
				if width != script.width || height != script.height {
					return nil, fmt.Errorf("primary prepared dimensions differ from agreed pattern; no capture requested")
				}
			}
		}
		if found != 1 {
			return nil, fmt.Errorf("primary display not uniquely identified")
		}
		script.stage = 3
		return script.tool("host-default", "screenshot", map[string]any{"ai_callback_delay": 0}), nil
	case 3, 4:
		pixels, imageError := imageToolPixels(body, script.current)
		if imageError != nil {
			text, err := hostToolText(body, script.current)
			if err != nil {
				return nil, err
			}
			match := hintHandle.FindStringSubmatch(text)
			if match == nil {
				return nil, fmt.Errorf("host capture supplied neither pixels nor a running Job: %w", imageError)
			}
			handle, err := strconv.Atoi(match[1])
			if err != nil || handle <= 0 {
				return nil, fmt.Errorf("invalid capture Job handle")
			}
			if script.running && handle != script.handle {
				return nil, fmt.Errorf("capture changed Job handle while waiting")
			}
			if !script.running {
				for _, previous := range script.captureHandles {
					if previous == handle {
						return nil, fmt.Errorf("distinct captures reused one Job handle")
					}
				}
				script.captureHandles = append(script.captureHandles, handle)
			}
			script.running, script.handle = true, handle
			script.waits++
			if script.waits > 16 {
				return nil, fmt.Errorf("host capture exceeded bounded Job observations")
			}
			return script.tool(fmt.Sprintf("host-wait-%d", script.waits), "wait_for_job", map[string]any{"handle": handle, "ai_callback_delay": 1}), nil
		}
		if !script.running {
			return nil, fmt.Errorf("capture did not demonstrate delayed Job delivery")
		}
		if err := validateHostScreenPNG(pixels, script.width, script.height, script.challenge); err != nil {
			if script.failureImagePath != "" && len(pixels) <= 8<<20 {
				file, saveError := os.OpenFile(script.failureImagePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
				if saveError == nil {
					_, saveError = file.Write(pixels)
					saveError = errors.Join(saveError, file.Close())
				}
				if saveError != nil {
					err = errors.Join(err, fmt.Errorf("retain failure image: %w", saveError))
				}
			}
			return nil, err
		}
		script.payloads = append(script.payloads, append([]byte(nil), pixels...))
		script.imageIDs = append(script.imageIDs, script.current)
		if script.stage == 3 {
			script.stage = 4
			script.running = false
			script.handle = 0
			return script.tool("host-explicit", "screenshot", map[string]any{"display": 0, "ai_callback_delay": 0}), nil
		}
		script.stage = 5
		return &fakevendor.Reply{Text: "HOST_NATIVE_CAPTURE_COMPLETE"}, nil
	case 5:
		if !script.replayAllowed {
			return nil, fmt.Errorf("host replay arrived before the next human turn")
		}
		for index, id := range script.imageIDs {
			pixels, err := imageToolPixels(body, id)
			if err != nil {
				return nil, fmt.Errorf("host replay: %w", err)
			}
			if !bytes.Equal(pixels, script.payloads[index]) {
				return nil, fmt.Errorf("host replay changed captured bytes")
			}
		}
		script.stage = 6
		return &fakevendor.Reply{Text: "HOST_NATIVE_REPLAY_COMPLETE"}, nil
	default:
		return nil, fmt.Errorf("unexpected extra host provider request")
	}
}

// Keep error image bytes out of diagnostics before using the shared text reader.
func hostToolText(body []byte, id string) (string, error) {
	var request struct {
		Messages []struct {
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(body, &request); err != nil {
		return "", err
	}
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
			if block.Type == "tool_result" && block.ID == id && block.Error {
				return "", fmt.Errorf("host tool refused: %s", mediaFailureText(block.Content))
			}
		}
	}
	text, err := hintToolText(body, id)
	if err != nil {
		return "", err
	}
	if len(text) > 65536 {
		return "", fmt.Errorf("host tool text exceeds bound")
	}
	return text, nil
}

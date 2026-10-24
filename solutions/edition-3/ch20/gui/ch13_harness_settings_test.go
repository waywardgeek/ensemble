package gui

import (
	"ensemble"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Read the launcher's actual payload, then cross the same settings-file load
// boundary as gui/cmd. Do not duplicate its JSON in a permissive test helper.
func TestCh13LauncherSettingsLoad(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "../examples/ch13-speech/launch.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var payloads []string
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(literal.Value)
		if err == nil && strings.HasPrefix(value, "{") && strings.Contains(value, `"tts_enabled"`) {
			payloads = append(payloads, value)
		}
		return true
	})
	if len(payloads) != 1 {
		t.Fatalf("expected one actual launcher settings payload, got %d", len(payloads))
	}
	path := filepath.Join(t.TempDir(), "settings.json")
	if err = os.WriteFile(path, []byte(payloads[0]), 0600); err != nil {
		t.Fatal(err)
	}
	app := ensemble.New(nil)
	defer app.Shutdown()
	server, err := NewServerWithSettings(app, "web/gui", path)
	if err != nil {
		t.Fatalf("actual launcher settings rejected: %v", err)
	}
	defer server.Close()
	settings := server.Hub.currentSettings()
	if !settings.TTSEnabled || !settings.PauseWhileSpeaking || settings.TTSSpeed != 1 {
		t.Fatalf("loaded settings: %+v", settings)
	}
}

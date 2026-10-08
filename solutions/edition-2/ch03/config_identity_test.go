package ensemble_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"example.com/ensemble"
)

// The public snapshot must continue to name the file receiving this Agent's
// history. Rejecting a move must also preserve the rest of the configuration.
func TestConfigurationPreservesLogIdentity(t *testing.T) {
	for _, destination := range []string{"different", "empty"} {
		t.Run(destination, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "history.log")
			a, err := ensemble.New(nil).NewAgent(ensemble.Config{
				LogPath: path, APIKey: "fixture", Model: "fixture-original",
			})
			if err != nil {
				t.Fatal(err)
			}
			defer a.Close()
			before := a.Config()
			next := a.Config()
			next.LogPath = ""
			if destination == "different" {
				next.LogPath = filepath.Join(dir, "other.log")
			}
			next.Model = "fixture-changed"
			next.System = "a rejected update must not replace this either"
			if err := a.SetConfig(next); err == nil {
				t.Fatal("accepted a different log destination")
			}
			if !reflect.DeepEqual(a.Config(), before) {
				t.Fatal("rejected update changed the configuration")
			}
			if _, err := os.Stat(filepath.Join(dir, "other.log")); !os.IsNotExist(err) {
				t.Fatalf("rejected update created another log: %v", err)
			}
		})
	}

	t.Run("same path retains a usable writer and permits model changes", func(t *testing.T) {
		models := make(chan string, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var request struct {
				Model string `json:"model"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Error(err)
			}
			models <- request.Model
			fmt.Fprint(w, `{"content":[{"type":"text","text":"log identity retained"}],"usage":{"input_tokens":1,"output_tokens":1}}`)
		}))
		defer server.Close()
		path := filepath.Join(t.TempDir(), "history.log")
		a, err := ensemble.New(nil).NewAgent(ensemble.Config{
			LogPath: path, APIKey: "fixture", Model: "fixture-original", BaseURL: server.URL,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer a.Close()
		next := a.Config()
		next.Model = "fixture-next"
		if err := a.SetConfig(next); err != nil {
			t.Fatal(err)
		}
		if _, err := a.Ask(context.Background(), "retain this history"); err != nil {
			t.Fatal(err)
		}
		if got := <-models; got != next.Model {
			t.Fatalf("request model = %q, want %q", got, next.Model)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if a.Config().LogPath != path || !strings.Contains(string(data), "log identity retained") {
			t.Fatal("configuration and persisted history no longer identify the same file")
		}
	})
}

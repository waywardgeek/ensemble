package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestGUIAuthCLI(test *testing.T) {
	binary := filepath.Join(test.TempDir(), "gui")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		test.Fatalf("build GUI: %v\n%s", err, output)
	}
	test.Run("explicit commands", func(test *testing.T) {
		directory := test.TempDir()
		password := guiAuthCommandPassword(test, binary, directory, "auth-init")
		path := filepath.Join(directory, "cr", "gui-auth.json")
		before := checkGUICredentialFile(test, path, "ensemble", password)
		output, stderr, err := runGUIAuthBinary(test, binary, directory, "auth-init")
		if err == nil || len(output) != 0 || !strings.Contains(stderr, "already exists") || strings.Contains(stderr, password) {
			test.Fatal("second init must refuse without printing a password")
		}
		if after, err := os.ReadFile(path); err != nil || !bytes.Equal(before, after) {
			test.Fatal("second init changed the existing credential file")
		}
		replacement := guiAuthCommandPassword(test, binary, directory, "auth-reset", "--username", "operator")
		if password == replacement {
			test.Fatal("reset reused the password")
		}
		checkGUICredentialFile(test, path, "operator", replacement)
		for _, name := range []string{"debug.log", "api.log", "gui.log", "save.json", "event.log", "ensemble.json"} {
			if _, err := os.Stat(filepath.Join(directory, name)); !os.IsNotExist(err) {
				test.Error("credential command touched application state:", name)
			}
		}
	})
	test.Run("config path without provider access", func(test *testing.T) {
		directory := test.TempDir()
		config := writeGUIAuthConfig(test, directory, "http://127.0.0.1:1", false)
		password := guiAuthCommandPassword(test, binary, directory, "auth-init", "--config", config)
		checkGUICredentialFile(test, filepath.Join(directory, "state", "gui-auth.json"), "ensemble", password)
		override := filepath.Join(directory, "other", "login.json")
		password = guiAuthCommandPassword(test, binary, directory, "auth-init", "--config", config, "--file", override, "--username", "operator")
		checkGUICredentialFile(test, override, "operator", password)
	})
	test.Run("invalid commands have no secret output", func(test *testing.T) {
		for _, args := range [][]string{
			{"gui-auth"}, {"gui-auth", "init"}, {"gui-auth", "reset"}, {"auth-reset"},
			{"auth-init", "--password", "sensitive-argument"},
			{"auth-init", "sensitive-argument"},
			{"auth-init", "--file", ""},
			{"auth-init", "--username", "bad:name"},
			{"auth-init", "--config", ""},
		} {
			directory := test.TempDir()
			output, stderr, err := runGUIAuthBinary(test, binary, directory, args...)
			if err == nil || len(output) != 0 || strings.Contains(stderr, "sensitive-argument") {
				test.Fatal("invalid credential command succeeded or echoed a supplied value")
			}
			if args[0] == "gui-auth" && !strings.Contains(stderr, "usage: auth-init|auth-reset") {
				test.Fatal("old credential syntax must fail before provider setup")
			}
			entries, err := os.ReadDir(directory)
			if err != nil || len(entries) != 0 {
				test.Fatal("invalid command wrote files")
			}
		}
	})
	test.Run("missing explicit credentials fail before application work", func(test *testing.T) {
		var requests atomic.Int64
		provider := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			requests.Add(1)
			response.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer provider.Close()
		directory := test.TempDir()
		config := writeGUIAuthConfig(test, directory, provider.URL, true)
		for _, name := range []string{"debug.log", "api.log", "gui.log"} {
			if err := os.WriteFile(filepath.Join(directory, name), []byte("preserve existing log"), 0600); err != nil {
				test.Fatal(err)
			}
		}
		output, stderr, err := runGUIAuthBinary(test, binary, directory, "--config", config, "--port", "0", "--gui-auth-file", filepath.Join(directory, "absent.json"))
		if err == nil || len(output) != 0 || !strings.Contains(stderr, "auth-init --file") {
			test.Fatal("missing explicit credentials did not fail with setup instructions")
		}
		for _, name := range []string{"debug.log", "api.log", "gui.log"} {
			if data, err := os.ReadFile(filepath.Join(directory, name)); err != nil || string(data) != "preserve existing log" {
				test.Error("failed authentication preflight changed log:", name)
			}
		}
		for _, name := range []string{"state", "save.json", "event.log", "ensemble.json"} {
			if _, err := os.Stat(filepath.Join(directory, name)); !os.IsNotExist(err) {
				test.Error("failed authentication preflight created application state:", name)
			}
		}
		if requests.Load() != 0 {
			test.Fatal("authentication preflight contacted the provider")
		}
	})
	test.Run("invalid existing credentials never downgrade loopback", func(test *testing.T) {
		directory := test.TempDir()
		config := writeGUIAuthConfig(test, directory, "http://127.0.0.1:1", true)
		path := filepath.Join(directory, "state", "gui-auth.json")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(`{"password":"sensitive-fixture"}`), 0600); err != nil {
			test.Fatal(err)
		}
		output, stderr, err := runGUIAuthBinary(test, binary, directory, "--config", config, "--port", "0")
		if err == nil || len(output) != 0 || !strings.Contains(stderr, "invalid GUI auth credential document") || strings.Contains(stderr, "sensitive-fixture") {
			test.Fatal("invalid existing credentials were ignored or disclosed")
		}
	})
	test.Run("credential aliases cannot be truncated by launch", func(test *testing.T) {
		for _, scenario := range []struct {
			name  string
			alias string
		}{
			{name: "debug.log"}, {name: "api.log"}, {name: "gui.log"},
			{name: "event.log"}, {name: "save.json"}, {name: "save.json.journal"},
			{name: "save.json.bak"}, {name: "ensemble.json"}, {name: "settings.json"},
			{name: "gui.log", alias: "hardlink"}, {name: "gui.log", alias: "symlink"},
			{name: "gui.log", alias: "directory"},
		} {
			test.Run(scenario.name+scenario.alias, func(test *testing.T) {
				directory := test.TempDir()
				config := writeGUIAuthConfig(test, directory, "http://127.0.0.1:1", true)
				path := filepath.Join(directory, scenario.name)
				if scenario.name != "settings.json" {
					guiAuthCommandPassword(test, binary, directory, "auth-init", "--file", path)
				}
				original, err := os.ReadFile(path)
				if err != nil {
					test.Fatal(err)
				}
				argument := scenario.name
				switch scenario.alias {
				case "hardlink", "symlink":
					argument = filepath.Join(directory, "alias.json")
					link := os.Link
					if scenario.alias == "symlink" {
						link = os.Symlink
					}
					if err := link(path, argument); err != nil {
						test.Fatal(err)
					}
				case "directory":
					alias := filepath.Join(test.TempDir(), "directory-alias")
					if err := os.Symlink(directory, alias); err != nil {
						test.Fatal(err)
					}
					argument = filepath.Join(alias, scenario.name)
				}
				_, stderr, err := runGUIAuthBinary(test, binary, directory, "--config", config, "--gui-auth-file", argument)
				if err == nil || !strings.Contains(stderr, "GUI auth file must not use an application") {
					test.Fatal("credential/output alias was not rejected before startup")
				}
				if after, err := os.ReadFile(path); err != nil || !bytes.Equal(original, after) {
					test.Fatal("launch changed the file aliased by GUI credentials")
				}
				for _, name := range []string{"debug.log", "api.log", "gui.log"} {
					if name == scenario.name {
						continue
					}
					if _, err := os.Stat(filepath.Join(directory, name)); !os.IsNotExist(err) {
						test.Error("alias rejection created application log:", name)
					}
				}
			})
		}
	})
	test.Run("legacy settings paths cannot hold credentials", func(test *testing.T) {
		for _, name := range []string{"settings.json", "settings.json.tmp"} {
			test.Run(name, func(test *testing.T) {
				directory := test.TempDir()
				path := filepath.Join("cr", name)
				password := guiAuthCommandPassword(test, binary, directory, "auth-init", "--file", path)
				original := checkGUICredentialFile(test, filepath.Join(directory, path), "ensemble", password)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				command := exec.CommandContext(ctx, binary, "--port", "0", "--gui-auth-file", path)
				command.Dir = directory
				command.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + directory, "LLM_VENDOR=openai", "LLM_MODEL=auth-fixture", "LLM_API_KEY=fixture", "LLM_BASE_URL=http://127.0.0.1:1", "EN_RECALL_DISABLED=1"}
				output, err := command.CombinedOutput()
				if ctx.Err() != nil || err == nil || !strings.Contains(string(output), "GUI auth file must not use an application") {
					test.Fatal("legacy settings credential collision was not rejected before startup")
				}
				if after := checkGUICredentialFile(test, filepath.Join(directory, path), "ensemble", password); !bytes.Equal(original, after) {
					test.Fatal("collision rejection changed credentials")
				}
				for _, name := range []string{"debug.log", "api.log", "gui.log"} {
					if _, err := os.Stat(filepath.Join(directory, name)); !os.IsNotExist(err) {
						test.Error("legacy settings collision created log:", name)
					}
				}
			})
		}
	})
	test.Run("legacy theme change persists to reserved settings path", func(test *testing.T) {
		directory := test.TempDir()
		exerciseGUIAuthLaunch(test, binary, directory, nil, false, func(origin string) {
			peer, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(origin, "http")+"/ws", nil)
			if err != nil {
				test.Fatal("cannot connect to legacy settings WebSocket")
			}
			defer peer.Close()
			peer.SetReadDeadline(time.Now().Add(3 * time.Second))
			if err := peer.WriteJSON(map[string]any{"type": "subscribe"}); err != nil {
				test.Fatal("cannot subscribe to legacy settings")
			}
			if err := peer.WriteJSON(map[string]any{"type": "update_settings", "settings": map[string]any{"theme": "dark"}}); err != nil {
				test.Fatal("cannot send theme update")
			}
			for {
				var message struct {
					Type string `json:"type"`
				}
				if err := peer.ReadJSON(&message); err != nil {
					test.Fatal("theme update was not acknowledged")
				}
				if message.Type == "settings_changed" {
					break
				}
			}
			path := filepath.Join(directory, "cr", "settings.json")
			data, err := os.ReadFile(path)
			var settings struct {
				Theme string `json:"theme"`
			}
			if err != nil || json.Unmarshal(data, &settings) != nil || settings.Theme != "dark" {
				test.Fatal("theme update did not persist to the reserved legacy settings path")
			}
			if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
				test.Fatal("theme update left its temporary settings file")
			}
		})
	})
	test.Run("service restart preserves verifier", func(test *testing.T) {
		provider := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			response.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer provider.Close()
		for _, scenario := range []struct {
			name     string
			auth     bool
			override bool
		}{
			{name: "legacy loopback"},
			{name: "loopback with credentials", auth: true},
			{name: "explicit override with config", auth: true, override: true},
		} {
			test.Run(scenario.name, func(test *testing.T) {
				directory := test.TempDir()
				config := writeGUIAuthConfig(test, directory, provider.URL, true)
				path := filepath.Join(directory, "state", "gui-auth.json")
				if scenario.override {
					path = filepath.Join(directory, "credentials", "override.json")
				}
				var password string
				var original []byte
				if scenario.auth {
					password = guiAuthCommandPassword(test, binary, directory, "auth-init", "--file", path)
					original = checkGUICredentialFile(test, path, "ensemble", password)
				}
				for attempt := 0; attempt < 2; attempt++ {
					args := []string{"--config", config}
					if scenario.override {
						args = append(args, "--gui-auth-file", path)
					}
					output := exerciseGUIAuthLaunch(test, binary, directory, args, scenario.auth)
					if scenario.auth {
						after := checkGUICredentialFile(test, path, "ensemble", password)
						if !bytes.Equal(original, after) || strings.Contains(output, password) {
							test.Fatal("service restart changed or disclosed credentials")
						}
						var document map[string]any
						if json.Unmarshal(after, &document) != nil || strings.Contains(output, document["ha1"].(string)) {
							test.Fatal("service output disclosed the verifier")
						}
					} else if _, err := os.Stat(path); !os.IsNotExist(err) {
						test.Fatal("legacy service generated credentials implicitly")
					}
				}
			})
		}
	})
}

func runGUIAuthBinary(test *testing.T, binary, directory string, args ...string) ([]byte, string, error) {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = directory
	command.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + directory, "LLM_VENDOR=invalid-to-prove-early-dispatch"}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if ctx.Err() != nil {
		test.Fatal("GUI command timed out")
	}
	return output, stderr.String(), err
}

func guiAuthCommandPassword(test *testing.T, binary, directory string, args ...string) string {
	test.Helper()
	output, stderr, err := runGUIAuthBinary(test, binary, directory, args...)
	if err != nil || stderr != "" || bytes.Count(output, []byte("\n")) != 1 {
		test.Fatal("credential command must succeed with exactly one password line and no stderr")
	}
	password := strings.TrimSuffix(string(output), "\n")
	secret, err := base64.RawURLEncoding.DecodeString(password)
	if err != nil || len(secret) != 16 {
		test.Fatal("CLI password is not a 128-bit base64url value")
	}
	return password
}

func checkGUICredentialFile(test *testing.T, path, username, password string) []byte {
	test.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		test.Fatal("credential file is missing")
	}
	var document map[string]any
	digest := sha256.Sum256([]byte(username + ":Ensemble:" + password))
	if json.Unmarshal(data, &document) != nil || len(document) != 4 || document["version"] != float64(1) || document["username"] != username || document["realm"] != "Ensemble" || document["ha1"] != hex.EncodeToString(digest[:]) || bytes.Contains(data, []byte(password)) {
		test.Fatal("persisted credential file does not contain only the expected verifier")
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode() != 0600 {
		test.Fatal("credential file is not private")
	}
	return data
}

func writeGUIAuthConfig(test *testing.T, directory, providerURL string, credential bool) string {
	test.Helper()
	config := map[string]any{
		"version": 1, "data_dir": "state", "enable_web_search": false, "recall": map[string]any{"disabled": true},
		"provider": map[string]any{"vendor": "openai", "model": "auth-fixture", "billing_mode": "api", "base_url": providerURL, "credential_file": "provider-key"},
	}
	data, err := json.Marshal(config)
	if err != nil {
		test.Fatal(err)
	}
	path := filepath.Join(directory, "settings.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		test.Fatal(err)
	}
	if credential {
		if err := os.WriteFile(filepath.Join(directory, "provider-key"), []byte("fixture"), 0600); err != nil {
			test.Fatal(err)
		}
	}
	return path
}

func exerciseGUIAuthLaunch(test *testing.T, binary, directory string, args []string, authenticated bool, inspect ...func(string)) string {
	test.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		test.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	args = append(args, "--port", fmt.Sprint(port))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = directory
	command.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + directory, "LLM_VENDOR=openai", "LLM_MODEL=auth-fixture", "LLM_API_KEY=fixture", "LLM_BASE_URL=http://127.0.0.1:1", "EN_RECALL_DISABLED=1"}
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	input, err := command.StdinPipe()
	if err != nil {
		test.Fatal(err)
	}
	defer input.Close()
	if err := command.Start(); err != nil {
		test.Fatal(err)
	}
	defer func() {
		if command.ProcessState == nil {
			command.Process.Kill()
			command.Wait()
		}
	}()
	client := &http.Client{Timeout: 250 * time.Millisecond}
	defer client.CloseIdleConnections()
	ready := false
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		response, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/", port))
		if err == nil {
			io.Copy(io.Discard, response.Body)
			response.Body.Close()
			want := http.StatusOK
			if authenticated {
				want = http.StatusUnauthorized
				if !strings.HasPrefix(response.Header.Get("WWW-Authenticate"), "Digest ") {
					test.Fatal("authenticated service omitted its Digest challenge")
				}
			}
			if response.StatusCode != want {
				test.Fatalf("service status %d, want %d", response.StatusCode, want)
			}
			ready = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !ready {
		test.Fatal("GUI listener did not become ready")
	}
	for _, check := range inspect {
		check(fmt.Sprintf("http://127.0.0.1:%d", port))
	}
	input.Close()
	if err := command.Wait(); err != nil {
		test.Fatal("GUI did not exit cleanly")
	}
	if strings.Contains(output.String(), "GUI authentication enabled (SHA-256 Digest).") != authenticated {
		test.Fatal("service authentication status was incorrect")
	}
	return output.String()
}

func TestGUIAuthPathPolicy(test *testing.T) {
	directory := test.TempDir()
	for _, host := range []string{"127.0.0.1", "127.0.0.2", "::1", "::ffff:127.0.0.1", "localhost"} {
		if path, err := guiAuthPath(host, directory, "", false); err != nil || path != "" {
			test.Error("missing default credentials must preserve legacy loopback:", host)
		}
		if _, err := guiAuthPath(host, directory, "", true); err == nil {
			test.Error("an explicit empty credential path must fail:", host)
		}
	}
	for _, host := range []string{"0.0.0.0", "::", "192.0.2.1", "host.invalid", "localhost.", ""} {
		if _, err := guiAuthPath(host, directory, "", false); err == nil {
			test.Error("missing credentials must fail closed for:", host)
		}
	}
}

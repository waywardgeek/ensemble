package gui

import (
	"bytes"
	"encoding/json"
	"ensemble/ensemble"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type downloadSecurityInbox chan ensemble.FleetMessage

func (inbox downloadSecurityInbox) Deliver(message ensemble.FleetMessage) error {
	inbox <- message
	return nil
}

type downloadSecurityFixture struct {
	root      *ensemble.Ensemble
	agent     ensemble.Agent
	hub       *Server
	server    *httptest.Server
	workspace string
	logPath   string
}

func newDownloadSecurityFixture(test *testing.T) *downloadSecurityFixture {
	test.Helper()
	fixture := &downloadSecurityFixture{workspace: test.TempDir(), root: ensemble.New(io.Discard)}
	var err error
	fixture.agent, err = fixture.root.Start(test.TempDir(), ensemble.Config{
		Model: "offline", Workspace: fixture.workspace, Recall: ensemble.RecallSettings{Disabled: true},
	}, ensemble.AgentSpec{}, make(downloadSecurityInbox, 1))
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() {
		if err := fixture.root.Close(); err != nil {
			test.Error(err)
		}
	})
	fixture.logPath = filepath.Join(test.TempDir(), "gui.log")
	log, err := os.Create(fixture.logPath)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(func() { log.Close() })
	fixture.hub, err = New(fixture.root, fixture.agent, log)
	if err != nil {
		test.Fatal(err)
	}
	test.Cleanup(fixture.hub.Close)
	fixture.server = httptest.NewServer(fixture.hub.Handler(""))
	fixture.server.Client().Timeout = 5 * time.Second
	test.Cleanup(fixture.server.Close)
	return fixture
}

func (fixture *downloadSecurityFixture) offer(test *testing.T, agent ensemble.Agent, path string) string {
	test.Helper()
	for _, eventType := range []string{"tool_called", "tool_returned"} {
		data, err := json.Marshal(map[string]any{
			"type": eventType,
			"tool": map[string]any{
				"call_id": "download-security-offer", "name": "offer_download",
				"args":  map[string]string{"path": path},
				"parts": []ensemble.Part{{Type: "text", Text: "Download offered: " + filepath.Base(path)}},
			},
		})
		if err != nil {
			test.Fatal(err)
		}
		var event ensemble.Event
		if err := json.Unmarshal(data, &event); err != nil {
			test.Fatal(err)
		}
		if err := agent.Record(event); err != nil {
			test.Fatal(err)
		}
	}
	headers := http.Header{}
	connection, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(fixture.server.URL, "http")+"/ws", headers)
	if err != nil {
		test.Fatal("connect to download GUI:", err)
	}
	defer connection.Close()
	if err := connection.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		test.Fatal(err)
	}
	if err := connection.WriteJSON(map[string]string{"type": "select_agent", "key": ensemble.AgentKey(agent)}); err != nil {
		test.Fatal(err)
	}
	for {
		var message struct {
			Type     string `json:"type"`
			CallID   string `json:"call_id"`
			Token    string `json:"token"`
			Filename string `json:"filename"`
		}
		if err := connection.ReadJSON(&message); err != nil {
			test.Fatal("successful recorded offer did not reach the browser:", err)
		}
		if message.Type == "download_offered" {
			if message.Token == "" || message.CallID != "download-security-offer" || message.Filename != filepath.Base(path) {
				test.Fatal("download offer is missing its capability, call identity or filename")
			}
			return "/download/" + url.PathEscape(message.Token)
		}
	}
}

func downloadSecurityWrite(test *testing.T, path string, content []byte) {
	test.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		test.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0600); err != nil {
		test.Fatal(err)
	}
}

func (fixture *downloadSecurityFixture) request(test *testing.T, method, uri, authorization string, want int) (http.Header, []byte) {
	test.Helper()
	request, err := http.NewRequest(method, fixture.server.URL+uri, nil)
	if err != nil {
		test.Fatal(err)
	}
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	response, err := fixture.server.Client().Do(request)
	if err != nil {
		test.Fatal("download request failed:", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		test.Fatal(err)
	}
	if response.StatusCode != want {
		test.Fatalf("%s download status = %d, want %d", method, response.StatusCode, want)
	}
	if want != http.StatusOK && response.Header.Get("Content-Disposition") != "" {
		test.Error("rejected request returned an attachment")
	}
	return response.Header, body
}

func TestDownloadSecurityHTTPContract(test *testing.T) {
	fixture := newDownloadSecurityFixture(test)
	filename := "report \"final\".bin"
	content := []byte{0, 255, 254, '\r', '\n', '<', '>', 128, 1}
	path := filepath.Join(fixture.workspace, filename)
	downloadSecurityWrite(test, path, content)
	uri := fixture.offer(test, fixture.agent, filename)
	token := strings.TrimPrefix(uri, "/download/")
	guess := "x" + token[1:]
	if guess == token {
		guess = "y" + token[1:]
	}
	for name, target := range map[string]string{
		"unknown token": "/download/not-issued",
		"guessed token": "/download/" + guess,
		"relative path": "/download/" + url.PathEscape(filename),
		"absolute path": "/download/" + url.PathEscape(path),
		"query path":    "/download/?path=" + url.QueryEscape(path),
		"query token":   "/download/?token=" + url.QueryEscape(token),
	} {
		test.Run(name, func(test *testing.T) {
			fixture.request(test, http.MethodGet, target, "", http.StatusNotFound)
		})
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions} {
		test.Run(method, func(test *testing.T) {
			headers, _ := fixture.request(test, method, uri, "", http.StatusMethodNotAllowed)
			allowed := strings.ReplaceAll(headers.Get("Allow"), " ", "")
			if allowed != "GET,HEAD" && allowed != "HEAD,GET" {
				test.Errorf("Allow = %q, want GET and HEAD", headers.Get("Allow"))
			}
		})
	}
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		test.Run(method, func(test *testing.T) {
			headers, body := fixture.request(test, method, uri, "", http.StatusOK)
			for name, want := range map[string]string{
				"Content-Type": "application/octet-stream", "X-Content-Type-Options": "nosniff",
				"Cache-Control": "no-store", "Referrer-Policy": "no-referrer",
				"Content-Length": strconv.Itoa(len(content)),
			} {
				if got := headers.Get(name); got != want {
					test.Errorf("%s = %q, want %q", name, got, want)
				}
			}
			disposition, parameters, err := mime.ParseMediaType(headers.Get("Content-Disposition"))
			if err != nil || disposition != "attachment" || parameters["filename"] != filename {
				test.Errorf("incorrect binary attachment disposition: %q", headers.Get("Content-Disposition"))
			}
			if method == http.MethodGet && !bytes.Equal(body, content) {
				test.Error("download changed binary bytes")
			}
			if method == http.MethodHead && len(body) != 0 {
				test.Error("HEAD returned a body")
			}
		})
	}
	fixture.hub.Close()
	trace, err := os.ReadFile(fixture.logPath)
	if err != nil {
		test.Fatal(err)
	}
	decodedToken, err := url.PathUnescape(token)
	if err != nil {
		test.Fatal(err)
	}
	if bytes.Contains(trace, []byte(decodedToken)) || !bytes.Contains(trace, []byte(`"type":"download_offered"`)) || !bytes.Contains(trace, []byte(`"token":"[redacted]"`)) {
		test.Error("GUI trace must record the offer while redacting its capability")
	}
}

func TestDownloadSecurityRechecksWorkspaceAtClick(test *testing.T) {
	for _, change := range []string{"missing file", "file symlink escape", "directory symlink escape"} {
		test.Run(change, func(test *testing.T) {
			fixture := newDownloadSecurityFixture(test)
			relative := filepath.Join("reports", "report.bin")
			path := filepath.Join(fixture.workspace, relative)
			content := []byte("allowed workspace file")
			downloadSecurityWrite(test, path, content)
			uri := fixture.offer(test, fixture.agent, relative)
			_, body := fixture.request(test, http.MethodGet, uri, "", http.StatusOK)
			if !bytes.Equal(body, content) {
				test.Fatal("offered workspace file was not downloadable before replacement")
			}
			outside := test.TempDir()
			downloadSecurityWrite(test, filepath.Join(outside, "report.bin"), []byte("outside secret"))
			if change == "directory symlink escape" {
				if err := os.Rename(filepath.Dir(path), filepath.Join(fixture.workspace, "original")); err != nil {
					test.Fatal(err)
				}
				if err := os.Symlink(outside, filepath.Dir(path)); err != nil {
					test.Fatal(err)
				}
			} else {
				if err := os.Remove(path); err != nil {
					test.Fatal(err)
				}
				if change == "file symlink escape" {
					if err := os.Symlink(filepath.Join(outside, "report.bin"), path); err != nil {
						test.Fatal(err)
					}
				}
			}
			for _, method := range []string{http.MethodGet, http.MethodHead} {
				_, body := fixture.request(test, method, uri, "", http.StatusNotFound)
				if bytes.Contains(body, content) || bytes.Contains(body, []byte("outside secret")) {
					test.Error("rejected download disclosed file bytes")
				}
			}
		})
	}
}

func TestDownloadSecurityAgentDeletionAndReplacement(test *testing.T) {
	fixture := newDownloadSecurityFixture(test)
	content := []byte("named agent download")
	downloadSecurityWrite(test, filepath.Join(fixture.workspace, "report.bin"), content)
	peer, err := fixture.root.CreateAgent("reporter", "offline", "")
	if err != nil {
		test.Fatal(err)
	}
	uri := fixture.offer(test, peer, "report.bin")
	fixture.request(test, http.MethodGet, uri, "", http.StatusOK)
	if err := fixture.root.DeleteAgent("reporter"); err != nil {
		test.Fatal(err)
	}
	fixture.request(test, http.MethodGet, uri, "", http.StatusNotFound)
	replacement, err := fixture.root.CreateAgent("reporter", "offline", "")
	if err != nil {
		test.Fatal(err)
	}
	fixture.request(test, http.MethodGet, uri, "", http.StatusNotFound)
	freshURI := fixture.offer(test, replacement, "report.bin")
	_, body := fixture.request(test, http.MethodGet, freshURI, "", http.StatusOK)
	if !bytes.Equal(body, content) {
		test.Error("replacement agent could not offer its own download")
	}
	fixture.request(test, http.MethodGet, uri, "", http.StatusNotFound)
}

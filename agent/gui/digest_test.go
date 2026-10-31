package gui

import (
	"crypto/sha256"
	"encoding/hex"
	"ensemble/ensemble"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func digestTestServer(t *testing.T) (*Server, *httptest.Server, string) {
	t.Helper()
	root := ensemble.New(io.Discard)
	agent, err := root.NewAgent(ensemble.Config{Model: "offline", DataDir: t.TempDir(), Workspace: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	hub, err := New(root, agent, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "gui-auth.json")
	password, err := CreateDigestCredentials(path, "ensemble", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := hub.ConfigureDigestAuth(path); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(hub.Handler(""))
	t.Cleanup(func() { server.Close(); hub.Close(); agent.Shutdown() })
	return hub, server, password
}

func digestTestHeader(password, nonce, method, uri, cnonce, count string) string {
	hash := func(text string) string {
		digest := sha256.Sum256([]byte(text))
		return hex.EncodeToString(digest[:])
	}
	response := hash(hash("ensemble:Ensemble:"+password) + ":" + nonce + ":" + count + ":" + cnonce + ":auth:" + hash(method+":"+uri))
	return fmt.Sprintf("Digest username=\"ensemble\", realm=\"Ensemble\", algorithm=SHA-256, qop=auth, nonce=%q, nc=%s, cnonce=%q, uri=%q, response=%q", nonce, count, cnonce, uri, response)
}

func digestTestRequest(t *testing.T, server *httptest.Server, method, uri, header string) (int, http.Header) {
	t.Helper()
	request, err := http.NewRequest(method, server.URL+uri, nil)
	if err != nil {
		t.Fatal(err)
	}
	if header != "" {
		request.Header.Set("Authorization", header)
	}
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	return response.StatusCode, response.Header
}

func digestTestNonce(t *testing.T, server *httptest.Server, uri string) string {
	t.Helper()
	status, headers := digestTestRequest(t, server, "GET", uri, "")
	if status != http.StatusUnauthorized || !strings.Contains(headers.Get("WWW-Authenticate"), "algorithm=SHA-256") {
		t.Fatal("missing SHA-256 authentication challenge")
	}
	match := regexp.MustCompile(`nonce="([^"]+)"`).FindStringSubmatch(headers.Get("WWW-Authenticate"))
	if len(match) != 2 {
		t.Fatal("missing challenge nonce")
	}
	return match[1]
}

func TestDigestRFC7616SHA256(t *testing.T) {
	passwordHash := sha256.Sum256([]byte("Mufasa:http-auth@example.org:Circle of Life"))
	guard := digestAuthentication{credential: digestCredentials{Username: "Mufasa", Realm: "http-auth@example.org", HA1: hex.EncodeToString(passwordHash[:])}}
	fields := map[string]string{
		"username": "Mufasa", "realm": "http-auth@example.org", "algorithm": "SHA-256", "qop": "auth",
		"nonce": "7ypf/xlj9XXwfDPEoM4URrv/xwf94BcCAzFZH4GiTo0v", "nc": "00000001",
		"cnonce": "f2/wE4q74E6zIJEtWaHKaf5wv/H5QzzpXusqGemxURZJ", "uri": "/dir/index.html",
		"response": "753927fa0e85d155564e2e272a28d1802ca10daf4496794697cf8db5856cb6c1",
	}
	request := httptest.NewRequest("GET", "http://example.org/dir/index.html", nil)
	if !guard.validResponse(request, fields) {
		t.Fatal("RFC 7616 section 3.9.1 known-answer response rejected")
	}
	fields["response"] = strings.Repeat("0", 64)
	if guard.validResponse(request, fields) {
		t.Fatal("incorrect response accepted")
	}
}

func TestDigestAssetsWebSocketAndReplay(t *testing.T) {
	_, server, password := digestTestServer(t)
	for _, uri := range []string{"/", "/gui/app.js", "/ws"} {
		digestTestNonce(t, server, uri)
	}
	nonce := digestTestNonce(t, server, "/")
	for _, count := range []string{"00000002", "00000001"} {
		header := digestTestHeader(password, nonce, "GET", "/", "parallel", count)
		if status, _ := digestTestRequest(t, server, "GET", "/", header); status != 200 {
			t.Fatalf("out-of-order legitimate request failed: %d", status)
		}
	}
	header := digestTestHeader(password, nonce, "GET", "/", "parallel", "00000001")
	if status, _ := digestTestRequest(t, server, "GET", "/", header); status != 401 {
		t.Fatal("replayed response admitted")
	}
	nonce = digestTestNonce(t, server, "/ws")
	requestHeaders := http.Header{"Authorization": {digestTestHeader(password, nonce, "GET", "/ws", "socket", "00000001")}, "Origin": {server.URL}}
	conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", requestHeaders)
	if err != nil {
		t.Fatalf("authenticated websocket failed: %v", err)
	}
	conn.Close()
	response.Body.Close()
	requestHeaders.Set("Authorization", digestTestHeader(password, nonce, "GET", "/ws", "foreign-origin", "00000001"))
	requestHeaders.Set("Origin", "http://untrusted.invalid")
	if conn, response, err = websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws", requestHeaders); err == nil {
		conn.Close()
		t.Fatal("authenticated cross-origin websocket admitted")
	} else if response == nil || response.StatusCode != 403 {
		t.Fatal("cross-origin policy did not survive authentication")
	}
	response.Body.Close()
}

func TestDigestRejectsAlteredProofsWithoutBurningNonce(t *testing.T) {
	for _, defect := range []string{"password", "username", "realm", "algorithm", "qop", "uri", "method", "nc", "duplicate", "username alias", "forged nonce"} {
		t.Run(defect, func(t *testing.T) {
			_, server, password := digestTestServer(t)
			nonce := digestTestNonce(t, server, "/")
			header := digestTestHeader(password, nonce, "GET", "/", "same-challenge", "00000001")
			bad := header
			switch defect {
			case "password":
				bad = digestTestHeader(password+"wrong", nonce, "GET", "/", "same-challenge", "00000001")
			case "username":
				bad = strings.Replace(header, `username="ensemble"`, `username="unknown"`, 1)
			case "realm":
				bad = strings.Replace(header, `realm="Ensemble"`, `realm="different"`, 1)
			case "algorithm":
				bad = strings.Replace(header, "algorithm=SHA-256", "algorithm=MD5", 1)
			case "qop":
				bad = strings.Replace(header, "qop=auth", "qop=auth-int", 1)
			case "uri":
				bad = digestTestHeader(password, nonce, "GET", "/ws", "same-challenge", "00000001")
			case "method":
				bad = digestTestHeader(password, nonce, "POST", "/", "same-challenge", "00000001")
			case "nc":
				bad = digestTestHeader(password, nonce, "GET", "/", "same-challenge", "00000000")
			case "duplicate":
				bad += `, username="ensemble"`
			case "username alias":
				bad += `, username*=UTF-8''ensemble`
			case "forged nonce":
				bad = digestTestHeader(password, nonce+"x", "GET", "/", "same-challenge", "00000001")
			}
			if status, _ := digestTestRequest(t, server, "GET", "/", bad); status != 401 {
				t.Fatalf("invalid proof returned %d", status)
			}
			if status, _ := digestTestRequest(t, server, "GET", "/", header); status != 200 {
				t.Fatal("failed proof consumed legitimate challenge")
			}
		})
	}
}

func TestDigestNonceExpiryAndInstanceIsolation(t *testing.T) {
	hub, server, password := digestTestServer(t)
	for _, offset := range []time.Duration{-6 * time.Minute, time.Minute} {
		nonce, err := hub.digest.nonce(time.Now().Add(offset))
		if err != nil {
			t.Fatal(err)
		}
		status, headers := digestTestRequest(t, server, "GET", "/", digestTestHeader(password, nonce, "GET", "/", "expiry", "00000001"))
		if status != 401 || strings.Contains(headers.Get("WWW-Authenticate"), "stale=true") != (offset < 0) {
			t.Fatal("expired or future challenge handled incorrectly")
		}
	}
	foreign := digestAuthentication{}
	nonce, err := foreign.nonce(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if status, _ := digestTestRequest(t, server, "GET", "/", digestTestHeader(password, nonce, "GET", "/", "foreign", "00000001")); status != 401 {
		t.Fatal("another server's challenge accepted")
	}
}

func TestDigestConcurrentReplay(t *testing.T) {
	_, server, password := digestTestServer(t)
	nonce := digestTestNonce(t, server, "/")
	header := digestTestHeader(password, nonce, "GET", "/", "concurrent", "00000001")
	statuses := make(chan int, 8)
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			request, _ := http.NewRequest("GET", server.URL+"/", nil)
			request.Header.Set("Authorization", header)
			response, err := server.Client().Do(request)
			if err != nil {
				statuses <- 0
				return
			}
			response.Body.Close()
			statuses <- response.StatusCode
		}()
	}
	workers.Wait()
	close(statuses)
	admitted := 0
	for status := range statuses {
		if status == 200 {
			admitted++
		} else if status != 401 && status != 429 {
			t.Fatalf("unexpected replay status %d", status)
		}
	}
	if admitted != 1 {
		t.Fatalf("concurrent proof admitted %d times", admitted)
	}
}

func TestDigestThrottleAndBoundedState(t *testing.T) {
	hub, server, password := digestTestServer(t)
	nonce := digestTestNonce(t, server, "/")
	for count := 1; count <= 3; count++ {
		header := digestTestHeader(password+"wrong", nonce, "GET", "/", "guesses", fmt.Sprintf("%08x", count))
		if status, _ := digestTestRequest(t, server, "GET", "/", header); status != 401 {
			t.Fatal("initial guessing attempt did not receive a challenge")
		}
	}
	status, headers := digestTestRequest(t, server, "GET", "/", "")
	if seconds, err := strconv.Atoi(headers.Get("Retry-After")); status != 429 || err != nil || seconds < 1 {
		t.Fatal("guessing burst not throttled with Retry-After")
	}
	request := httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "192.0.2.8:1234"
	response := httptest.NewRecorder()
	hub.Handler("").ServeHTTP(response, request)
	if response.Code != 401 {
		t.Fatal("one source locked out another source")
	}
	now := time.Now()
	for count := range digestFailureLimit + 1 {
		hub.digest.failed(fmt.Sprint("source-", count), now)
	}
	if len(hub.digest.failures) > digestFailureLimit {
		t.Fatal("failure tracking grew without bound")
	}
	for count := range digestReplayLimit {
		hub.digest.replays[fmt.Sprint(count)] = now.Add(time.Minute)
	}
	header := digestTestHeader(password, nonce, "GET", "/", "capacity", "00000001")
	request = httptest.NewRequest("GET", "/", nil)
	request.RemoteAddr = "192.0.2.9:1234"
	request.Header.Set("Authorization", header)
	response = httptest.NewRecorder()
	hub.Handler("").ServeHTTP(response, request)
	if response.Code != 503 || len(hub.digest.replays) != digestReplayLimit {
		t.Fatal("full replay tracking must refuse, not forget valid replay records")
	}
}

func TestDigestHeaderParsing(t *testing.T) {
	for _, header := range []string{
		"Basic dXNlcjpwYXNz", "Digest", `Digest username="unterminated`, `Digest username="x", USERNAME="y"`,
		`Digest username="x" realm="y"`, `Digest username="x",`, `Digest username=`, "Digest username=\"bad\x01\"",
		"Digest username=" + strings.Repeat("x", 8192),
	} {
		if _, valid := digestFields(header); valid {
			t.Fatal("malformed authentication parameters accepted")
		}
	}
	if fields, valid := digestFields(`Digest username="a\"b", algorithm=SHA-256`); !valid || fields["username"] != `a"b` {
		t.Fatal("quoted parameter escape mishandled")
	}
}

func TestDigestConfigurationOrder(t *testing.T) {
	root := ensemble.New(io.Discard)
	agent, err := root.NewAgent(ensemble.Config{Model: "offline", DataDir: t.TempDir(), Workspace: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer agent.Shutdown()
	path := filepath.Join(t.TempDir(), "gui-auth.json")
	if _, err := CreateDigestCredentials(path, "ensemble", false); err != nil {
		t.Fatal(err)
	}
	for _, before := range []bool{true, false} {
		hub, err := New(root, agent, io.Discard)
		if err != nil {
			t.Fatal(err)
		}
		defer hub.Close()
		handler := hub.Handler("")
		if before {
			if err := hub.ConfigureDigestAuth(path); err != nil {
				t.Fatal(err)
			}
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", "/", nil))
		if before && response.Code != 401 {
			t.Fatal("handler created before configuration bypasses authentication")
		}
		if !before && hub.ConfigureDigestAuth(path) == nil {
			t.Fatal("configuration after serving must fail explicitly")
		}
	}
}

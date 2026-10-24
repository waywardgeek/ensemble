package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// The normal entry starts the existing GUI, not an Engine in this client.
// All browser capabilities stay here, outside the listener's five tools.
type launchConfig struct {
	root, workspace, baseURL, logPath, typeDuring string
	streaming                                     bool
}

func normalConfig(root string, args []string) (launchConfig, string, error) {
	c := launchConfig{root: root, workspace: os.Getenv("WORKSPACE"), baseURL: os.Getenv("LLM_BASE_URL"), logPath: os.Getenv("TTS_LOG"), streaming: os.Getenv("STREAMING") != "off"}
	if len(args) != 1 && !(len(args) == 3 && args[1] == "--type-during-turn") {
		return c, "", errors.New("usage: tts-harness.sh PROMPT [--type-during-turn TEXT]")
	}
	if strings.TrimSpace(args[0]) == "" {
		return c, "", errors.New("prompt must not be empty")
	}
	if len(args) == 3 {
		c.typeDuring = args[2]
		if c.typeDuring == "" {
			return c, "", errors.New("typing text must not be empty (spaces are allowed)")
		}
	}
	for name, value := range map[string]string{"WORKSPACE": c.workspace, "LLM_BASE_URL": c.baseURL, "TTS_LOG": c.logPath} {
		if value == "" {
			return c, "", fmt.Errorf("%s is required", name)
		}
	}
	u, err := url.Parse(c.baseURL)
	if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return c, "", errors.New("LLM_BASE_URL must be a local HTTP provider URL")
	}
	ip := net.ParseIP(u.Hostname())
	if u.Hostname() != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return c, "", errors.New("normal harness requires a loopback external provider")
	}
	c.root, err = filepath.Abs(c.root)
	if err != nil {
		return c, "", err
	}
	c.workspace, err = filepath.Abs(c.workspace)
	if err != nil {
		return c, "", err
	}
	c.logPath, err = filepath.Abs(c.logPath)
	if err != nil {
		return c, "", err
	}
	info, err := os.Stat(c.workspace)
	if err != nil || !info.IsDir() {
		return c, "", errors.New("WORKSPACE must be an existing directory")
	}
	if _, err = os.Stat(filepath.Join(c.root, "gui", "web", "gui", "app.js")); err != nil {
		return c, "", errors.New("source root does not contain GUI assets")
	}
	return c, args[0], nil
}
func browserBinary() (string, error) {
	if value := os.Getenv("CHROME_BIN"); value != "" {
		if _, err := os.Stat(value); err != nil {
			return "", err
		}
		return value, nil
	}
	for _, name := range []string{"chromium", "chromium-browser", "google-chrome"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	path := "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	return "", errors.New("no installed Chromium browser; set CHROME_BIN (the harness never installs one)")
}
func guiEnvironment(c launchConfig) []string {
	var env []string
	for _, name := range []string{"PATH", "HOME", "TMPDIR"} {
		if value := os.Getenv(name); value != "" {
			env = append(env, name+"="+value)
		}
	}
	vendor, model := os.Getenv("LLM_VENDOR"), os.Getenv("LLM_MODEL")
	if vendor == "" {
		vendor = "openai"
	}
	if model == "" {
		model = "chapter13-local"
	}
	capability := "sse"
	if !c.streaming {
		capability = "plain"
	}
	return append(env, "LLM_VENDOR="+vendor, "LLM_MODEL="+model, "LLM_API_KEY=chapter13-local", "LLM_BASE_URL="+c.baseURL, "LLM_STREAM_CAPABILITY="+capability, "LLM_DISABLE_STREAMING="+strconv.FormatBool(!c.streaming), "LLM_SPEAKING=true")
}

type ownedProcess struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
	log   *os.File
	done  chan error
}

func startOwned(cmd *exec.Cmd, logPath string) (*ownedProcess, error) {
	log, err := os.OpenFile(logPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Close()
		return nil, err
	}
	cmd.Stdout = log
	cmd.Stderr = log
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err = cmd.Start(); err != nil {
		stdin.Close()
		log.Close()
		return nil, err
	}
	p := &ownedProcess{cmd: cmd, stdin: stdin, log: log, done: make(chan error, 1)}
	go func() { p.done <- cmd.Wait() }()
	return p, nil
}
func (p *ownedProcess) stop() error {
	_ = p.stdin.Close()
	// Only the group created by startOwned is signalled, never an attached GUI.
	_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGTERM)
	select {
	case <-p.done:
	case <-time.After(2 * time.Second):
		_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
		<-p.done
	}
	_ = syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL) // owned browser helpers, if any
	return p.log.Close()
}
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	return port, l.Close()
}
func awaitHTTP(ctx context.Context, address string, process *ownedProcess) error {
	client := http.Client{Timeout: time.Second}
	for {
		select {
		case err := <-process.done:
			process.done <- err
			return fmt.Errorf("owned GUI exited before readiness: %v", err)
		default:
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
		if err != nil {
			return err
		}
		r, err := client.Do(req)
		if err == nil {
			r.Body.Close()
			if r.StatusCode == 200 {
				return nil
			}
		}
		if err = sleep(ctx, 30*time.Millisecond); err != nil {
			return err
		}
	}
}

// Asynchronous synth fake only; text is still produced by the real GUI pipeline.
const asynchronousSynth = `(()=>{let generation=0;window.SpeechSynthesisUtterance=class{constructor(text){this.text=text;}};Object.defineProperty(window,'speechSynthesis',{configurable:true,value:{speak(u){const id=++generation;setTimeout(()=>{if(id!==generation)return;u.onstart?.();setTimeout(()=>{if(id===generation)u.onend?.();},25);},0);},cancel(){generation++;},getVoices(){return[];},addEventListener(){}}});})()`

func launchPage(ctx context.Context, c launchConfig, runDir string) (p *page, stop func() error, err error) {
	browser, err := browserBinary()
	if err != nil {
		return nil, nil, err
	}
	binary := filepath.Join(runDir, "gui")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd")
	build.Dir = filepath.Join(c.root, "gui")
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err = build.Run(); err != nil {
		return nil, nil, fmt.Errorf("build existing GUI: %w", err)
	}
	port, err := freePort()
	if err != nil {
		return nil, nil, err
	}
	settings := filepath.Join(runDir, "settings.json")
	if err = os.WriteFile(settings, []byte(`{"tts_enabled":true,"pause_while_speaking":true}`), 0600); err != nil {
		return nil, nil, err
	}
	cmd := exec.Command(binary, "-port", strconv.Itoa(port), "-gui-dir", filepath.Join(c.root, "gui", "web", "gui"), "-settings", settings, "-save", filepath.Join(runDir, "save.json"))
	cmd.Dir = runDir
	cmd.Env = guiEnvironment(c)
	gui, err := startOwned(cmd, filepath.Join(runDir, "gui-process.log"))
	if err != nil {
		return nil, nil, err
	}
	var chrome *ownedProcess
	stop = func() error {
		var e error
		if chrome != nil {
			e = chrome.stop()
		}
		return errors.Join(e, gui.stop())
	}
	cleanup := stop
	defer func() {
		if err != nil {
			err = errors.Join(err, cleanup())
		}
	}()
	address := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err = awaitHTTP(ctx, address, gui); err != nil {
		return nil, nil, err
	}
	profile := filepath.Join(runDir, "browser")
	cmd = exec.Command(browser, "--headless=new", "--use-mock-keychain", "--no-first-run", "--no-default-browser-check", "--remote-debugging-port=0", "--user-data-dir="+profile, "about:blank")
	chrome, err = startOwned(cmd, filepath.Join(runDir, "browser-process.log"))
	if err != nil {
		return nil, nil, err
	}
	var debugPort string
	for {
		data, e := os.ReadFile(filepath.Join(profile, "DevToolsActivePort"))
		if e == nil {
			debugPort = strings.Split(string(data), "\n")[0]
			if _, e = strconv.Atoi(debugPort); e != nil {
				return nil, nil, errors.New("invalid private browser debugging port")
			}
			break
		}
		select {
		case e := <-chrome.done:
			chrome.done <- e
			return nil, nil, fmt.Errorf("owned browser exited before readiness: %v", e)
		default:
		}
		if err = sleep(ctx, 30*time.Millisecond); err != nil {
			return nil, nil, err
		}
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+debugPort+"/json/list", nil)
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	var targets []struct {
		Type string `json:"type"`
		URL  string `json:"webSocketDebuggerUrl"`
	}
	err = json.NewDecoder(response.Body).Decode(&targets)
	response.Body.Close()
	if err != nil {
		return nil, nil, err
	}
	target := ""
	for _, t := range targets {
		if t.Type == "page" {
			target = t.URL
			break
		}
	}
	if target == "" {
		return nil, nil, errors.New("owned browser has no page")
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, target, nil)
	if err != nil {
		return nil, nil, err
	}
	p = &page{conn: conn}
	if err = p.call(ctx, "Page.enable", nil, nil); err == nil {
		err = p.call(ctx, "Page.addScriptToEvaluateOnNewDocument", map[string]any{"source": asynchronousSynth}, nil)
	}
	if err == nil {
		err = p.call(ctx, "Page.navigate", map[string]any{"url": address}, nil)
	}
	if err != nil {
		conn.Close()
		return nil, nil, err
	}
	return p, stop, nil
}

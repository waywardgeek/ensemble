package gradeedition3

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// probeOutput retains bounded diagnostics for precise negative-control assertions.
// stdout/stderr may arrive concurrently; a snapshot is read only after joining.
type probeOutput struct {
	gate     sync.Mutex
	text     bytes.Buffer
	overflow bool
}

func (output *probeOutput) Write(data []byte) (int, error) {
	output.gate.Lock()
	defer output.gate.Unlock()
	count := len(data)
	remaining := (64 << 10) - output.text.Len()
	if count > remaining {
		output.overflow = true
		data = data[:remaining]
	}
	output.text.Write(data)
	return count, nil
}

// AttachmentProbe configures a partial, unscored public-GUI probe. Explicit
// executable paths avoid installing browser tools or borrowing live services.
type AttachmentProbe struct {
	Candidate        string
	Node             string
	BrowserDriver    string
	PlaywrightModule string
	BrowserMutation  string
}

// RunAttachmentPickerProbe proves only picker-to-provider delivery. It must not
// award the attachment scope until recipient and adverse-lifecycle gates exist.
func RunAttachmentPickerProbe(config AttachmentProbe) (result error) {
	for _, path := range []string{config.Candidate, config.Node, config.BrowserDriver, config.PlaywrightModule} {
		if !filepath.IsAbs(path) {
			return fmt.Errorf("probe paths must be absolute")
		}
	}
	root, err := os.MkdirTemp("", "ensemble-attachment-probe-")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.RemoveAll(root); err != nil && result == nil {
			result = err
		}
	}()
	home := filepath.Join(root, "home")
	if err := os.Mkdir(home, 0700); err != nil {
		return err
	}
	challenge, source, err := newImageChallenge()
	if err != nil {
		return err
	}
	imagePath := filepath.Join(root, "challenge.png")
	if err := os.WriteFile(imagePath, source, 0600); err != nil {
		return err
	}
	caption := "attachment-" + filepath.Base(root)
	requests := make(chan []byte, 8)
	var calls atomic.Int32
	provider := fakevendor.NewWithOptions(nil, fakevendor.Options{Route: func(body []byte) *fakevendor.Reply {
		calls.Add(1)
		select {
		case requests <- append([]byte(nil), body...):
		default:
		}
		return &fakevendor.Reply{Text: "ATTACHMENT_PROBE_COMPLETE"}
	}})
	defer provider.Close()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		return err
	}
	// The normal executable owns binding. A competing bind fails startup; never
	// kill a process merely because it occupies the proposed port.
	base := "http://127.0.0.1:" + strconv.Itoa(port) + "/"
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	env := []string{"PATH=" + os.Getenv("PATH"), "HOME=" + home, "TMPDIR=" + root,
		"DOTNET_CLI_HOME=" + home, "DOTNET_SKIP_FIRST_TIME_EXPERIENCE=1",
		"DOTNET_CLI_TELEMETRY_OPTOUT=1", "DOTNET_GENERATE_ASPNET_CERTIFICATE=false",
		"LLM_API_KEY=synthetic-attachment-key"}
	command := exec.CommandContext(ctx, config.Candidate, "--port", strconv.Itoa(port),
		"--workspace", root, "--data-dir", filepath.Join(root, "data"),
		"--agent-name", "AttachmentProbe", "--vendor", "anthropic",
		"--model", "claude-3-haiku-20240307", "--base-url", provider.URL(),
		"--billing", "api", "--transcription", "off", "--no-save")
	command.Dir, command.Env = root, env
	diagnostics := &mediaOutput{ended: make(chan struct{}, 8)}
	command.Stdout, command.Stderr = diagnostics, diagnostics
	if err := command.Start(); err != nil {
		return err
	}
	exited := make(chan error, 1)
	go func() { exited <- command.Wait() }()
	stopped := false
	defer func() {
		if !stopped {
			_ = command.Process.Signal(os.Interrupt)
			select {
			case err := <-exited:
				if result == nil && err != nil {
					result = fmt.Errorf("candidate shutdown: %w", err)
				}
			case <-time.After(5 * time.Second):
				cancel()
				<-exited
				if result == nil {
					result = fmt.Errorf("candidate did not join graceful shutdown")
				}
			}
		}
		if result == nil && calls.Load() != 1 {
			result = fmt.Errorf("expected one provider request, got %d", calls.Load())
		}
		if result == nil && diagnostics.overflow {
			result = fmt.Errorf("candidate diagnostic bound exceeded")
		}
	}()
	client := &http.Client{Timeout: time.Second}
	for {
		select {
		case err := <-exited:
			stopped = true
			return fmt.Errorf("candidate exited during startup: %v", err)
		case <-ctx.Done():
			return fmt.Errorf("candidate startup timed out")
		default:
		}
		response, err := client.Get(base)
		if err == nil {
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	browser := exec.CommandContext(ctx, config.Node, config.BrowserDriver)
	browser.Dir = root
	browser.Env = append(append([]string{}, env...), "GUI_URL="+base,
		"PLAYWRIGHT_MODULE="+config.PlaywrightModule, "ATTACHMENT_MUTATION="+config.BrowserMutation, "ATTACHMENT_CAPTION="+caption,
		"ATTACHMENT_PNG="+imagePath, "ATTACHMENT_RESULT="+filepath.Join(root, "browser-result.json"))
	browserOutput := &probeOutput{}
	browser.Stdout, browser.Stderr = browserOutput, browserOutput
	if err := browser.Run(); err != nil {
		return fmt.Errorf("browser probe: %w; %s", err, browserOutput.text.String())
	}
	if browserOutput.overflow {
		return fmt.Errorf("browser diagnostic bound exceeded")
	}
	select {
	case body := <-requests:
		if _, err := attachmentPNG(body, caption, challenge); err != nil {
			return err
		}
	case <-ctx.Done():
		return fmt.Errorf("provider received no attachment")
	}
	return nil
}

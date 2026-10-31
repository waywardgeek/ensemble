package gradeedition2

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/waywardgeek/ensemble/internal/grade"
)

//go:embed ch26_probe.go.txt
var gatewayProbe string

// Chapter26 retains the seven original Gateway weights. The explicitly selected
// historical welcome field changes only outside fixture source, never a runtime.
// Ownership, dynamic policy and actual actor/model integration retain their
// separately reviewed source and live evidence; these rows are not a security audit.
func Chapter26(bin, core, welcomeField string) (grade.Report, error) {
	if welcomeField != "QueueDepth" && welcomeField != "QueuedCount" {
		return grade.Report{}, fmt.Errorf("gateway-welcome must be QueueDepth or historical QueuedCount")
	}
	result := grade.NewCh27Result()
	core, _ = filepath.Abs(core)
	work, err := os.MkdirTemp("", "gateway-grade26-")
	if err != nil {
		return grade.Report{}, err
	}
	defer os.RemoveAll(work)
	runErr := gatewayExercise(bin, core, work, welcomeField, &result)
	if runErr != nil {
		result.Fatal = runErr.Error()
	}
	checks := grade.Ch27Checks(result)
	checks[2].Title = "Actual Gateway binary: authenticated identities and welcome queue counts"
	checks[4].Title = "Handler-return receipt, ordered offline/resume and clean server restart"
	checks[5].Title = "Actual outside Client/Handler and resolved public dependency boundary"
	return grade.NewTitledReport("Edition 2, Chapter 26 — The Gateway", checks, ""), nil
}

func gatewayExercise(bin, core, work, field string, result *grade.Ch27Result) error {
	module := grade.DiscoverBase(core)
	if module == "" {
		return fmt.Errorf("missing submitted module identity")
	}
	source := strings.NewReplacer("GATEWAY_PACKAGE", module+"/gateway", "WELCOME_FIELD", field).Replace(gatewayProbe)
	mod := fmt.Sprintf("module outside.example/gateway26\n\ngo 1.25.0\nrequire %s v0.0.0\nreplace %s => %s\n", module, module, strconv.Quote(core))
	if err := recallWrite(work, map[string]string{"go.mod": mod, "main.go": source}); err != nil {
		return err
	}
	probe := filepath.Join(work, "probe")
	if _, err := reuseGo(work, "build", "-o", probe, "."); err != nil {
		return err
	}
	deps, err := reuseGo(work, "list", "-deps", ".")
	if err != nil {
		return err
	}
	found := false
	for _, p := range strings.Fields(deps) {
		if p == module+"/gateway" {
			found = true
		}
		if strings.HasPrefix(p, module+"/internal/") {
			result.Errs["C6"] = "public Gateway imports private runtime: " + p
		}
	}
	if !found {
		return fmt.Errorf("resolved outside consumer lacks submitted Gateway package")
	}
	keys, err := gatewayRun(probe, work, "-dir", work, "-phase", "keys")
	if err != nil {
		return err
	}
	agents := map[string]any{}
	for _, line := range strings.Split(keys, "\n") {
		f := strings.Fields(line)
		if len(f) != 2 {
			continue
		}
		groups := []string{"agents"}
		name := strings.ToUpper(f[0][:1]) + f[0][1:]
		if f[0] == "parent" {
			groups = append(groups, "admin")
		}
		if f[0] == "outsider" {
			groups = []string{"strangers"}
		}
		agents[f[0]] = map[string]any{"public_key": f[1], "display_name": name, "groups": groups}
	}
	if len(agents) != 3 {
		return fmt.Errorf("outside key creation returned %d identities", len(agents))
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	address := listener.Addr().String()
	listener.Close()
	cfg := map[string]any{"gateway": map[string]any{"listen": address, "data_dir": filepath.Join(work, "state")}, "agents": agents, "groups": map[string]any{
		"agents":    map[string]any{"trust": "trusted", "capabilities": []string{"send-message", "declare-contactable"}},
		"strangers": map[string]any{"trust": "external", "capabilities": []string{"send-message", "declare-contactable"}},
		"admin":     map[string]any{"trust": "owner", "capabilities": []string{"group-admin"}}}}
	data, _ := json.Marshal(cfg)
	config := filepath.Join(work, "gateway.json")
	if err = os.WriteFile(config, data, 0600); err != nil {
		return err
	}
	key, err := gatewayRun(bin, work, "-config", config, "-print-key")
	if err != nil {
		return err
	}
	key = strings.TrimSpace(key)
	args := []string{"-dir", work, "-url", "ws://" + address + "/v1/ws", "-public", key}
	phase := func(name string) error {
		output, err := gatewayRun(probe, work, append(args, "-phase", name)...)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		var observed map[string]string
		if err = json.Unmarshal([]byte(output), &observed); err != nil {
			return fmt.Errorf("%s observations: %w", name, err)
		}
		for id, msg := range observed {
			if result.Errs[id] == "" {
				result.Errs[id] = msg
			}
		}
		return nil
	}
	server, err := gatewayStart(bin, work, config, address)
	if err != nil {
		return err
	}
	defer server.stop()
	if err = phase("scenario"); err != nil {
		return err
	}
	if err = phase("seed"); err != nil {
		return err
	}
	// A new process and the same private directory distinguish persisted queues
	// from reconnecting to a still-running router. Exit status is part of success.
	if err = server.stop(); err != nil {
		return err
	}
	server, err = gatewayStart(bin, work, config, address)
	if err != nil {
		return err
	}
	defer server.stop()
	if err = phase("restore"); err != nil {
		return err
	}
	return server.stop()
}

func gatewayRun(bin, dir string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s %v: %w\n%s", filepath.Base(bin), args, err, out)
	}
	return string(out), nil
}

type gatewayProcess struct {
	cmd     *exec.Cmd
	done    chan error
	log     *os.File
	stopped bool
}

func gatewayStart(bin, work, config, address string) (*gatewayProcess, error) {
	log, err := os.CreateTemp(work, "gateway-log-")
	if err != nil {
		return nil, err
	}
	p := &gatewayProcess{cmd: exec.Command(bin, "-config", config), done: make(chan error, 1), log: log}
	p.cmd.Dir = work
	p.cmd.Stdout = log
	p.cmd.Stderr = log
	if err = p.cmd.Start(); err != nil {
		log.Close()
		return nil, err
	}
	go func() { p.done <- p.cmd.Wait() }()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err = <-p.done:
			p.stopped = true
			log.Close()
			b, _ := os.ReadFile(log.Name())
			return nil, fmt.Errorf("Gateway child exited before readiness: %v\n%s", err, b)
		default:
		}
		conn, e := net.DialTimeout("tcp", address, 30*time.Millisecond)
		if e == nil {
			conn.Close()
			if e = p.cmd.Process.Signal(syscall.Signal(0)); e == nil {
				return p, nil
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	p.stop()
	return nil, fmt.Errorf("new Gateway child did not bind %s", address)
}
func (p *gatewayProcess) stop() error {
	if p.stopped {
		return nil
	}
	p.stopped = true
	p.cmd.Process.Signal(os.Interrupt)
	var err error
	select {
	case err = <-p.done:
	case <-time.After(3 * time.Second):
		p.cmd.Process.Kill()
		<-p.done
		err = fmt.Errorf("Gateway did not exit after interrupt")
	}
	p.log.Close()
	if err != nil {
		b, _ := os.ReadFile(p.log.Name())
		return fmt.Errorf("Gateway shutdown: %w\n%s", err, b)
	}
	return nil
}

package grade

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Ch27Run grades the gateway chapter by driving the student's real gateway
// binary with a probe program compiled OUTSIDE the student's module.
//
// Compiling the probe in its own module, with only a replace directive
// pointing at the student's agent module, is the only honest way to check
// that the client library is reusable by a third party. A test living
// inside the module would compile against internal packages and prove
// nothing about what an outside consumer can reach.
func Ch27Run(agentDir string) Ch27Result {
	r := NewCh27Result()
	if err := runCh27(agentDir, &r); err != nil {
		r.Fatal = err.Error()
	}
	return r
}

func runCh27(agentDir string, r *Ch27Result) error {
	if _, err := os.Stat(filepath.Join(agentDir, "go.mod")); err != nil {
		return fmt.Errorf("no agent module at %s: %w", agentDir, err)
	}

	work, err := os.MkdirTemp("", "ch27-grade-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)

	// 1. Build the student's gateway binary.
	gatewayBin := filepath.Join(work, "gateway")
	if out, err := ch27Cmd(agentDir, nil, "go", "build", "-o", gatewayBin, "./cmd/gateway"); err != nil {
		return fmt.Errorf("building ./cmd/gateway failed: %v\n%s", err, out)
	}

	// 2. The public client package must not reach into internal/.
	deps, err := ch27Cmd(agentDir, nil, "go", "list", "-deps", "./gateway")
	if err != nil {
		r.fail("C6", "go list -deps ./gateway failed: %v\n%s", err, deps)
	} else if strings.Contains(deps, "/agent/internal") {
		var leaked []string
		for _, line := range strings.Split(deps, "\n") {
			if strings.Contains(line, "/agent/internal") {
				leaked = append(leaked, strings.TrimSpace(line))
			}
		}
		r.fail("C6", "agent/gateway imports internal packages, so it is not usable outside the module: %s",
			strings.Join(leaked, ", "))
	} else {
		r.pass("C6")
	}

	// 3. Build the probe in its own module.
	probeDir := filepath.Join(work, "probe")
	if err := os.MkdirAll(probeDir, 0o755); err != nil {
		return err
	}
	absAgent, err := filepath.Abs(agentDir)
	if err != nil {
		return err
	}
	goMod := "module ch27probe\n\ngo 1.24\n\nrequire github.com/waywardgeek/ensemble/agent v0.0.0\n\nreplace github.com/waywardgeek/ensemble/agent => " + absAgent + "\n"
	if err := os.WriteFile(filepath.Join(probeDir, "go.mod"), []byte(goMod), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(probeDir, "main.go"), []byte(ch27ProbeSource), 0o644); err != nil {
		return err
	}

	// Offline: every dependency is already in the module cache because the
	// student's module uses it. A grader that reached the network would
	// grade the network.
	env := append(os.Environ(), "GOFLAGS=-mod=mod", "GOPROXY=off", "GONOSUMCHECK=1", "GOWORK=off")
	if out, err := ch27Cmd(probeDir, env, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("probe go mod tidy failed (is the pinned API present?): %v\n%s", err, out)
	}
	probeBin := filepath.Join(work, "probe-bin")
	if out, err := ch27Cmd(probeDir, env, "go", "build", "-o", probeBin, "."); err != nil {
		return fmt.Errorf("the probe does not compile against agent/gateway, so a pinned name is missing or has a different shape: %v\n%s", err, out)
	}

	// 4. Generate agent keys with the student's own key helpers.
	dataDir := filepath.Join(work, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	keyOut, err := ch27Cmd(work, nil, probeBin, "-mode", "genkeys", "-dir", dataDir)
	if err != nil {
		return fmt.Errorf("probe key generation failed: %v\n%s", err, keyOut)
	}
	pubKeys := map[string]string{}
	for _, line := range strings.Split(keyOut, "\n") {
		if fields := strings.Fields(line); len(fields) == 3 && fields[0] == "KEY" {
			pubKeys[fields[1]] = fields[2]
		}
	}
	for _, id := range []string{"parent", "peer", "outsider"} {
		if pubKeys[id] == "" {
			return fmt.Errorf("probe did not emit a public key for %q; got:\n%s", id, keyOut)
		}
	}

	// 5. Write the gateway config.
	port, err := ch27FreePort()
	if err != nil {
		return err
	}
	listen := fmt.Sprintf("127.0.0.1:%d", port)
	cfgPath := filepath.Join(work, "gateway.json")
	if err := os.WriteFile(cfgPath, []byte(ch27Config(listen, dataDir, pubKeys)), 0o644); err != nil {
		return err
	}

	// 6. Start the gateway and learn its public key.
	gwPub, err := ch27Cmd(work, nil, gatewayBin, "-config", cfgPath, "-print-key")
	if err != nil {
		return fmt.Errorf("gateway -print-key failed: %v\n%s", err, gwPub)
	}
	gwPub = strings.TrimSpace(ch27LastLine(gwPub))
	if gwPub == "" {
		return fmt.Errorf("gateway -print-key printed nothing")
	}

	gw := exec.Command(gatewayBin, "-config", cfgPath)
	gw.Dir = work
	var gwLog strings.Builder
	gw.Stdout = &gwLog
	gw.Stderr = &gwLog
	if err := gw.Start(); err != nil {
		return fmt.Errorf("starting the gateway failed: %w", err)
	}
	defer func() {
		_ = gw.Process.Kill()
		_, _ = gw.Process.Wait()
	}()

	if err := ch27WaitForPort(listen, 15*time.Second); err != nil {
		return fmt.Errorf("the gateway never listened on %s: %v\ngateway log:\n%s", listen, err, gwLog.String())
	}

	// 7. Run the scenario.
	out, runErr := ch27Cmd(work, nil, probeBin,
		"-mode", "scenario",
		"-dir", dataDir,
		"-url", "ws://"+listen+"/v1/ws",
		"-gateway-key", gwPub,
	)
	obs, fails := parseCh27Obs(out)
	if runErr != nil && len(obs) == 0 && len(fails) == 0 {
		return fmt.Errorf("the probe produced no observations: %v\n%s\ngateway log:\n%s", runErr, out, gwLog.String())
	}
	for key, reason := range fails {
		r.fail(ch27Owner(key), "%s: %s", key, reason)
	}
	scoreCh27(r, obs)
	return nil
}

// ch27Owner maps an observation key to the check it belongs to.
func ch27Owner(key string) string {
	if len(key) >= 2 && key[0] == 'c' {
		return "C" + string(key[1])
	}
	return "C1"
}

// scoreCh27 turns the probe's observations into check outcomes.
//
// Each expectation names the observation it needs. A missing observation is
// a failure rather than a silent pass, because the probe emits one for every
// step it completes and an absent key means the step did not happen.
func scoreCh27(r *Ch27Result, obs map[string]string) {
	expect := []struct {
		check string
		key   string
		want  string
		why   string
	}{
		{"C1", "c1_roundtrip", "ok", "two peers did not complete a KK handshake and exchange an encrypted message"},
		{"C1", "c1_wrongkey", "rejected", "a handshake with the wrong static key was not rejected"},

		{"C3", "c3_parent_welcome", "ok", "the gateway did not welcome the first agent"},
		{"C3", "c3_peer_welcome", "ok", "the gateway did not welcome a second simultaneous agent"},
		{"C3", "c3_impostor", "rejected", "an agent holding the wrong private key was allowed to connect"},

		{"C2", "c2_provenance", "ok", "the delivered content did not begin with the provenance header"},
		{"C2", "c2_body", "ok", "the delivered content did not preserve the original body"},
		{"C2", "c2_trust", "trusted", "trust on the envelope did not come from the sender's group"},
		{"C2", "c2_convid", "ok", "the conversation id did not survive routing"},
		{"C2", "c2_mode", "async", "the response mode did not survive routing"},

		{"C4", "c4_delivered", "ok", "a message from an accepted group was not delivered"},
		{"C4", "c4_refused", "ok", "a message from a group the recipient does not accept was not refused"},
		{"C4", "c4_not_delivered", "ok", "the refused message was delivered anyway"},

		{"C5", "c5_queued", "ok", "messages sent to an offline agent were not queued and replayed on reconnect"},
		{"C5", "c5_acked", "ok", "the queue did not drain after the client acknowledged delivery"},
		{"C5", "c5_resume", "ok", "resuming from a last-acked sequence replayed the wrong set of messages"},

		{"C6", "c6_handler", "ok", "the Handler interface was not called by a receive goroutine"},

		{"C7", "c7_registered", "ok", "a parent could not register a child"},
		{"C7", "c7_child_connected", "ok", "the registered child could not connect on its own key"},
		{"C7", "c7_child_routed", "ok", "the child could not use the group it was granted"},
		{"C7", "c7_narrow_ok", "ok", "a parent could not grant a group it does hold"},
		{"C7", "c7_no_widen", "ok", "a parent was allowed to grant a group it does not hold"},
	}

	for _, e := range expect {
		got, ok := obs[e.key]
		if !ok {
			r.fail(e.check, "%s (observation %s missing)", e.why, e.key)
			continue
		}
		if got != e.want {
			r.fail(e.check, "%s (%s = %q, want %q)", e.why, e.key, got, e.want)
			continue
		}
		r.pass(e.check)
	}
}

func parseCh27Obs(out string) (map[string]string, map[string]string) {
	obs := map[string]string{}
	fails := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "OBS "):
			if k, v, ok := strings.Cut(strings.TrimPrefix(line, "OBS "), "="); ok {
				obs[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		case strings.HasPrefix(line, "PROBEFAIL "):
			if k, v, ok := strings.Cut(strings.TrimPrefix(line, "PROBEFAIL "), "="); ok {
				fails[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
	}
	return obs, fails
}

func ch27Config(listen, dataDir string, keys map[string]string) string {
	return `{
  "gateway": { "listen": "` + listen + `", "data_dir": "` + dataDir + `" },
  "agents": {
    "parent":   { "display_name": "Parent",   "public_key": "` + keys["parent"] + `", "groups": ["admin", "agents"] },
    "peer":     { "display_name": "Peer",     "public_key": "` + keys["peer"] + `", "groups": ["agents"] },
    "outsider": { "display_name": "Outsider", "public_key": "` + keys["outsider"] + `", "groups": ["strangers"] }
  },
  "groups": {
    "admin":     { "description": "administration", "capabilities": ["group-admin", "send-message", "declare-contactable"], "trust": "owner" },
    "agents":    { "description": "standard agents", "capabilities": ["send-message", "declare-contactable"], "trust": "trusted" },
    "strangers": { "description": "untrusted", "capabilities": ["send-message", "declare-contactable"], "trust": "external" }
  }
}
`
}

// --- small helpers ---

func ch27Cmd(dir string, env []string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if env != nil {
		cmd.Env = env
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ch27FreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func ch27WaitForPort(addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		last = err
		time.Sleep(50 * time.Millisecond)
	}
	return last
}

func ch27LastLine(s string) string {
	scanner := bufio.NewScanner(strings.NewReader(s))
	out := ""
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			out = line
		}
	}
	return out
}

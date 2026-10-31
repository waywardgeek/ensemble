package gradeedition2

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Current fixtures use the accepted configuration owner. No inherited sandbox,
// credentials, diagnostic destinations or settings may select another session.
func chapter20Env(work, skills, endpoint string) []string {
	primary := ""
	if skills != "" {
		primary = "base"
	}
	env := webStatusEnv(os.Environ(), work, skills, endpoint, "claude-sonnet-4-6", primary)
	keys := map[string]bool{"HOME": true, "TMPDIR": true, "EN_SANDBOX_ROOT": true, "EN_SAFE_MODE": true, "EN_WEB_SEARCH": true, "EN_RECALL_DISABLED": true, "EN_SKILL_DIR": true}
	clean := env[:0]
	for _, item := range env {
		k, _, _ := strings.Cut(item, "=")
		if !keys[k] {
			clean = append(clean, item)
		}
	}
	return append(clean, "HOME="+work, "TMPDIR="+work, "EN_SANDBOX_ROOT=", "EN_SAFE_MODE=false", "EN_WEB_SEARCH=false", "EN_RECALL_DISABLED=1", "EN_SKILL_DIR=")
}

func chapter20GUIStart(bin, work, base string, current bool) (*mcpCommand, *guiPeer, error) {
	if !current {
		return managementStart(bin, work, base)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	port := fmt.Sprint(l.Addr().(*net.TCPAddr).Port)
	l.Close()
	assets := filepath.Join(work, "web")
	if err = contextStaticFixture(assets); err != nil {
		return nil, nil, err
	}
	p, err := startMCPCommandEnv(bin, work, []string{"--port", port, "--gui-dir", assets}, chapter20Env(work, filepath.Join(work, "skills"), base))
	if err != nil {
		return nil, nil, err
	}
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		peer, e := openGUIPeer("ws://127.0.0.1:" + port + "/ws")
		select {
		case exit := <-p.done:
			p.done <- exit
			if peer != nil {
				peer.close()
			}
			p.close()
			return nil, nil, fmt.Errorf("GUI child exited before readiness: %v", exit)
		default:
		}
		if e == nil {
			return p, peer, nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	p.close()
	return nil, nil, fmt.Errorf("GUI readiness deadline: %s", p.stderr.String())
}

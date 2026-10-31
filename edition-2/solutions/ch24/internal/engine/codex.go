package engine

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/url"
	"strings"

	"ensemble/internal/common"
)

// Codex's HTTP session-id header supplies cache affinity. Keep it equal to the
// body key, stable for this Agent/model, without exposing its data directory.
func codexCacheKey(cfg common.Config) string {
	return fmt.Sprintf("ensemble-%x", sha256.Sum256([]byte(cfg.DataDir+"\x00"+cfg.Model)))
}

// A selected credential file must not silently accompany an old public API
// override. Loopback peers support deterministic transport tests without keys.
func validCodexEndpoint(base string) bool {
	u, err := url.Parse(base)
	if err != nil || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	if u.Scheme == "https" && u.Host == "chatgpt.com" && strings.TrimRight(u.Path, "/") == "/backend-api/codex" {
		return true
	}
	ip := net.ParseIP(u.Hostname())
	return ip != nil && ip.IsLoopback() && (u.Scheme == "http" || u.Scheme == "https")
}

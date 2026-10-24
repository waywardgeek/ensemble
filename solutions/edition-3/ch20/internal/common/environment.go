package common

import (
	"os"
	"strings"
)

// ProcessEnvironment is shared by local command and authorized MCP launches.
// Unconfined commands retain ordinary build settings, not transport credentials.
// A SandboxRoot selects the stricter Chapter 19 allowlist; Sandbox preparation
// then adds its own rooted HOME/TMP and Git configuration.
func ProcessEnvironment(agents ...Agent) []string {
	private := map[string]bool{"LLM_API_KEY": true, "ANTHROPIC_API_KEY": true, "OPENAI_API_KEY": true, "GEMINI_API_KEY": true, "FIRECRAWL_API_KEY": true, "TERM": true}
	confined := false
	hide := func(servers []MCPServer) {
		for _, server := range servers {
			if server.AuthEnv != "" {
				private[server.AuthEnv] = true
			}
		}
	}
	for _, a := range agents {
		if a == nil {
			continue
		}
		cfg := a.Config()
		confined = confined || cfg.SandboxRoot != ""
		hide(cfg.MCPServers)
		if skills := a.Skills(); skills != nil {
			for _, name := range skills.MCPAuthEnvNames() {
				private[name] = true
			}
		}
	}
	locale := map[string]bool{"LANG": true, "LC_ALL": true, "LC_CTYPE": true, "LC_COLLATE": true, "LC_MESSAGES": true, "LC_MONETARY": true, "LC_NUMERIC": true, "LC_TIME": true, "TZ": true}
	env := []string{}
	for _, item := range os.Environ() {
		key, _, _ := strings.Cut(item, "=")
		if private[key] || confined && !locale[key] {
			continue
		}
		env = append(env, item)
	}
	if confined {
		env = append(env, "PATH=/usr/bin:/bin:/usr/sbin:/sbin")
	}
	return append(env, "TERM=dumb")
}

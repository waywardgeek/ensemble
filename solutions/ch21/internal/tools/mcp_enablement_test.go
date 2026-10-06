package tools

import (
	"encoding/json"
	"testing"

	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/internal/skills"
)

// A bridged MCP server advertises whatever it likes, and the host used to
// enable all of it: every advertised name was appended to the skill's tool
// list on connect. That looked correct for as long as keyless Firecrawl
// advertised exactly the three tools its SKILL.md documented -- the inventory
// matched the documentation by luck, not by enforcement. Adding an API key
// turns the same endpoint into twenty-seven tools, among them ones that
// delete monitors and spend credits, and every one of them would have been
// declared to the model and callable by it.
//
// The rule is that the skill file decides. A server may offer more; the tools:
// header is what brings a tool into existence for the agent.
func TestMCPToolsAbsentFromSkillAreNeitherDeclaredNorCallable(t *testing.T) {
	const md = `---
name: web-search
description: test skill
type: loadable
tools: firecrawl_search
mcp_servers:
  - name: firecrawl
    transport: url
    url: https://example.invalid/mcp
---

body
`
	props, err := skills.ParseSkillMD(md)
	if err != nil {
		t.Fatalf("ParseSkillMD: %v", err)
	}
	if len(props.MCPServers) != 1 {
		t.Fatalf("want 1 mcp server, got %d", len(props.MCPServers))
	}

	sr := skills.NewSkillRegistry()
	sr.Register(props)

	// A skill is only loadable if a loaded skill says so.
	base, err := skills.ParseSkillMD(`---
name: base
description: base skill
type: primary
loadable-skills: web-search
---

body
`)
	if err != nil {
		t.Fatalf("ParseSkillMD base: %v", err)
	}
	sr.Register(base)
	if err := sr.LoadInitial("base", nil); err != nil {
		t.Fatalf("LoadInitial base: %v", err)
	}

	reg := NewRegistry()
	reg.WireSkills(sr, nil, nil)

	// Stand in for main.go's connect callback: the server advertises two
	// tools, only one of which the skill lists. Both get registered, exactly
	// as the real bridge registers everything it is told about.
	advertised := []string{"firecrawl_search", "firecrawl_monitor_delete"}
	reg.SetOnSkillMCPConnect(func(skill string, servers []common.MCPServerConfig) ([]string, error) {
		for _, name := range advertised {
			reg.RegisterTool(common.Tool{
				Name:        name,
				Description: "bridged",
				Schema:      json.RawMessage(`{"type":"object"}`),
				Run: func(c *common.Call, args json.RawMessage) (string, error) {
					return "ran", nil
				},
			})
		}
		return advertised, nil
	})

	loadSkill, err := reg.Lookup("load_skill")
	if err != nil {
		t.Fatalf("lookup load_skill: %v", err)
	}
	if _, err := loadSkill.Run(&common.Call{}, json.RawMessage(`{"name":"web-search"}`)); err != nil {
		t.Fatalf("load_skill: %v", err)
	}

	// Both exist -- registration is not the gate.
	for _, name := range advertised {
		if _, err := reg.Lookup(name); err != nil {
			t.Fatalf("%s should be registered: %v", name, err)
		}
	}

	if !reg.IsEnabled("firecrawl_search") {
		t.Error("firecrawl_search is listed in tools: and must be enabled")
	}
	if reg.IsEnabled("firecrawl_monitor_delete") {
		t.Error("firecrawl_monitor_delete is not listed in tools: and must not be callable")
	}

	declared := map[string]bool{}
	for _, d := range reg.Declarations() {
		declared[d.Name] = true
	}
	if !declared["firecrawl_search"] {
		t.Error("firecrawl_search should be declared to the model")
	}
	if declared["firecrawl_monitor_delete"] {
		t.Error("firecrawl_monitor_delete must never be declared to the model")
	}
}

// auth-env names an environment variable. The key itself must never appear in
// a SKILL.md, which is a file committed to the repository.
func TestParseMCPAuthEnvNamesAVariableNotASecret(t *testing.T) {
	const md = `---
name: web-search
description: test skill
type: loadable
tools: firecrawl_search
mcp_servers:
  - name: firecrawl
    transport: url
    url: https://mcp.firecrawl.dev/v2/mcp
    auth-env: FIRECRAWL_API_KEY
---

body
`
	props, err := skills.ParseSkillMD(md)
	if err != nil {
		t.Fatalf("ParseSkillMD: %v", err)
	}
	srv := props.MCPServers[0]
	if srv.AuthEnv != "FIRECRAWL_API_KEY" {
		t.Errorf("AuthEnv = %q, want FIRECRAWL_API_KEY", srv.AuthEnv)
	}
	if srv.URL != "https://mcp.firecrawl.dev/v2/mcp" {
		t.Errorf("URL = %q", srv.URL)
	}
}

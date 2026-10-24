package common

import (
	"strings"
	"testing"
)

type ch19EnvironmentAgent struct {
	Agent
	cfg    Config
	skills SkillRegistry
}

func (a ch19EnvironmentAgent) Config() Config        { return a.cfg }
func (a ch19EnvironmentAgent) Skills() SkillRegistry { return a.skills }

type ch19EnvironmentSkills struct{ SkillRegistry }

func (ch19EnvironmentSkills) MCPAuthEnvNames() []string { return []string{"E3_DECLARED_AUTH"} }

func (ch19EnvironmentSkills) ListSkills() []SkillInfo {
	return []SkillInfo{{Skill: Skill{MCPServers: []MCPServer{{Name: "declared", AuthEnv: "E3_DECLARED_AUTH"}}}}}
}
func TestCh19R5ProcessEnvironmentContexts(t *testing.T) {
	for key, value := range map[string]string{"GOFLAGS": "-mod=readonly -p=1", "E3_BUILD_SETTING": "retained", "LLM_API_KEY": "dummy", "OPENAI_API_KEY": "dummy", "ANTHROPIC_API_KEY": "dummy", "GEMINI_API_KEY": "dummy", "FIRECRAWL_API_KEY": "dummy", "E3_CONFIGURED_AUTH": "dummy", "E3_DECLARED_AUTH": "dummy", "PATH": "/untrusted/bin", "HOME": "/untrusted/home", "TMPDIR": "/untrusted/tmp", "GIT_CONFIG_GLOBAL": "/untrusted/git", "SSH_AUTH_SOCK": "/untrusted/ssh", "BASH_ENV": "/untrusted/startup", "LANG": "C", "LC_TIME": "C", "TZ": "UTC", "TERM": "untrusted"} {
		t.Setenv(key, value)
	}
	for _, confined := range []bool{false, true} {
		t.Run(map[bool]string{false: "unconfined", true: "confined"}[confined], func(t *testing.T) {
			cfg := Config{MCPServers: []MCPServer{{Name: "configured", AuthEnv: "E3_CONFIGURED_AUTH"}}}
			if confined {
				cfg.SandboxRoot = "/owned-fixture"
			}
			env := map[string]string{}
			for _, item := range ProcessEnvironment(ch19EnvironmentAgent{cfg: cfg, skills: ch19EnvironmentSkills{}}) {
				key, value, _ := strings.Cut(item, "=")
				env[key] = value
			}
			for _, key := range []string{"LLM_API_KEY", "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "GEMINI_API_KEY", "FIRECRAWL_API_KEY", "E3_CONFIGURED_AUTH", "E3_DECLARED_AUTH"} {
				if _, ok := env[key]; ok {
					t.Errorf("credential variable inherited: %s", key)
				}
			}
			if env["TERM"] != "dumb" {
				t.Errorf("TERM=%q", env["TERM"])
			}
			if !confined {
				for key, want := range map[string]string{"GOFLAGS": "-mod=readonly -p=1", "E3_BUILD_SETTING": "retained", "PATH": "/untrusted/bin", "HOME": "/untrusted/home"} {
					if env[key] != want {
						t.Errorf("ordinary %s=%q want %q", key, env[key], want)
					}
				}
				return
			}
			for _, key := range []string{"GOFLAGS", "E3_BUILD_SETTING", "HOME", "TMPDIR", "GIT_CONFIG_GLOBAL", "SSH_AUTH_SOCK", "BASH_ENV"} {
				if _, ok := env[key]; ok {
					t.Errorf("confined arbitrary/inherited variable: %s", key)
				}
			}
			if env["PATH"] != "/usr/bin:/bin:/usr/sbin:/sbin" || env["LANG"] != "C" || env["LC_TIME"] != "C" || env["TZ"] != "UTC" {
				t.Errorf("confined system/locale settings: %+v", env)
			}
		})
	}
}

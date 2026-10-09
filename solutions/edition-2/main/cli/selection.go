package cli

import (
	"example.com/ensemble"
 "context"
	"fmt"
	"os"
	"strings"
)

// SessionSelector preserves explicit presence independently of defaults. GUI
// uses the same selection operation after its own flag parsing.
type SessionSelector struct {
	Path    string
	Present bool
}

func SelectAgent(app *ensemble.Ensemble, config ensemble.Config, human bool, selection SessionSelector) (*ensemble.Agent, error) {
	if selection.Present && strings.TrimSpace(selection.Path) == "" {
		return nil, &ensemble.SessionError{Code: "session_conflict", Detail: "blank session selector"}
	}
	explicitLog := os.Getenv("CH02_LOG") != ""
	if selection.Present && explicitLog {
		return nil, &ensemble.SessionError{Code: "session_conflict", Detail: "CH02_LOG conflicts with --session-dir"}
	}
	if !selection.Present && (!human || explicitLog) {
		if err:=app.PrepareMCPSelections(context.Background(),config.MCPBindings);err!=nil{return nil,err}
 return app.NewAgent(config)
	}
	path := selection.Path
	if !selection.Present {
		path = ".ensemble/session"
	}
	config.DataDir = path
	config.LogPath = ""
	config.System = ""
	var system *string
	if value, ok := os.LookupEnv("LLM_SYSTEM"); ok {
		system = &value
	}
	return app.OpenSession(ensemble.SessionOptions{Config: config, System: system})
}
func sessionArguments(args []string) ([]string, SessionSelector, error) {
	out := []string{}
	selector := SessionSelector{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--session-dir" || strings.HasPrefix(arg, "--session-dir=") {
			if selector.Present {
				return nil, selector, fmt.Errorf("session_conflict: duplicate session selector")
			}
			selector.Present = true
			if arg == "--session-dir" {
				i++
				if i == len(args) {
					return nil, selector, fmt.Errorf("session_conflict: missing session directory")
				}
				selector.Path = args[i]
			} else {
				selector.Path = strings.TrimPrefix(arg, "--session-dir=")
			}
		} else {
			out = append(out, arg)
		}
	}
	if selector.Present && strings.TrimSpace(selector.Path) == "" {
		return nil, selector, fmt.Errorf("session_conflict: blank session directory")
	}
	return out, selector, nil
}

func mcpArguments(args []string)([]string,string,error){
 out:=[]string{};path:="";seen:=false
 for i:=0;i<len(args);i++{a:=args[i];if a!="--mcp-config"&&!strings.HasPrefix(a,"--mcp-config="){out=append(out,a);continue};if seen{return nil,"",fmt.Errorf("duplicate MCP configuration")};seen=true
 if a=="--mcp-config"{i++;if i==len(args){return nil,"",fmt.Errorf("missing MCP configuration")};path=args[i]}else{path=strings.TrimPrefix(a,"--mcp-config=")};if strings.TrimSpace(path)==""{return nil,"",fmt.Errorf("blank MCP configuration")}}
 return out,path,nil
}

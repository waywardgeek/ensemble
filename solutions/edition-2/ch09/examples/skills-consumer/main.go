// Public, headless two-Agent embedding. Defaults to local capability controls;
// --ask explicitly enables two bounded model turns after the live-plan gate.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"example.com/ensemble"
	"example.com/ensemble/cli"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	ask := flag.Bool("ask", false, "send one bounded read-and-report turn to each Agent")
	flag.Parse()
	app := ensemble.New(os.Stderr)
	defer app.Close()
	base := cli.Configuration(app)
	base.MaxTokens = 1024
	// Constructor-owned catalog inputs use the same parser as directory sources.
	catalog := map[string][]byte{
		"base":    []byte("---\nname: base\ndescription: Read notes\ntype: primary\ntools: read_file\nloadable-skills: edit inspect\n---\nRead the requested notes and report actual findings. Project ${PROJECT}.\n$TOOLS\n$SKILLS\n"),
		"edit":    []byte("---\nname: edit\ndescription: Edit notes\ntype: loadable\ntools: write_file\n---\nProject ${PROJECT}. Write only when asked.\n$TOOLS\n"),
		"inspect": []byte("---\nname: inspect\ndescription: Inspect notes\ntype: loadable\n---\nProject ${PROJECT}. Explain the actual note contents.\n$TOOLS\n"),
	}
	agents := []*ensemble.Agent{}
	for i, name := range []string{"alpha", "beta"} {
		workspace, err := filepath.Abs(name)
		if err != nil {
			return err
		}
		if err = os.MkdirAll(workspace, 0700); err != nil {
			return err
		}
		c := base
		c.Workspace = workspace
		c.LogPath = filepath.Join(workspace, "events.jsonl")
		c.PolicyPath = filepath.Join(workspace, "policy.json")
		c.Builtins = []string{"read_file", "load_skill", "unload_skill"}
		if i == 0 {
			c.Builtins = append(c.Builtins, "write_file")
		}
		c.Skills = &ensemble.SkillConfig{Catalog: catalog, Primary: "base", Variables: map[string]string{"PROJECT": name + "-$TOOLS"}}
		a, err := app.NewAgent(c)
		if err != nil {
			return err
		}
		agents = append(agents, a)
		if _, err = a.UpdatePolicy(a.ExecutionPolicy().Revision, json.RawMessage(`{"max_model_requests":2}`)); err != nil {
			return err
		}
	}
	if _, err := app.LoadSkill(agents[0].ID(), "edit"); err != nil {
		return err
	}
	second, err := app.SkillState(agents[1].ID())
	if err != nil || second.Revision != 0 {
		return fmt.Errorf("second Agent changed with first: %v", err)
	}
	if _, err = app.LoadSkill(agents[1].ID(), "edit"); err == nil {
		return fmt.Errorf("second ceiling unexpectedly allowed writing")
	}
	if _, err = app.LoadSkill(agents[1].ID(), "inspect"); err != nil {
		return err
	}
	var failures []error
	for _, a := range agents {
		inspection, err := app.InspectSkills(a.ID())
		if err != nil {
			return err
		}
		if err = json.NewEncoder(os.Stdout).Encode(struct {
			Agent  string
			Skills ensemble.SkillInspection
		}{a.ID(), inspection}); err != nil {
			return err
		}
		if *ask {
			completion, attemptErr := demonstrate(a)
			record := terminalRecord{Agent: a.ID(), Completion: completion}
			if attemptErr != nil {
				record.Error = attemptErr.Error()
				failures = append(failures, fmt.Errorf("%s: %w", a.ID(), attemptErr))
			}
			if err = json.NewEncoder(os.Stdout).Encode(record); err != nil {
				return err
			}
		}
	}
	return errors.Join(failures...)
}

// Completion is an owned terminal receipt, including partial text/parts and
// accepted usage when the request ends at its limit or with another failure.
type terminalRecord struct {
	Agent      string
	Completion ensemble.Completion
	Error      string `json:",omitempty"`
}

func demonstrate(a *ensemble.Agent) (ensemble.Completion, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	h, err := a.Submit("Use read_file exactly once to read notes.txt. Then report the note marker and your project marker exactly as written in your active manual. Do not write any files or call another tool.")
	if err != nil {
		return ensemble.Completion{AgentID: a.ID(), Outcome: "error"}, err
	}
	completion, waitErr := h.Wait(ctx)
	if waitErr != nil {
		// Waiting's deadline is not a request outcome. Cancel this attempt and
		// collect its actual terminal receipt before starting the next Agent.
		cancelErr := h.Cancel()
		completion, err = h.Wait(context.Background())
		waitErr = errors.Join(waitErr, cancelErr, err)
	}
	if completion.Error != nil {
		err = fmt.Errorf("%s: %s", completion.Error.Code, completion.Error.Message)
	} else if completion.Outcome != "success" {
		err = fmt.Errorf("request %s ended with %s", h.ID(), completion.Outcome)
	}
	return completion, errors.Join(waitErr, err)
}

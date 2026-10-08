// The workflow is a public consumer. Role-specific tools and instructions are
// ordinary Agent configuration; the library has no author/editor special case.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"example.com/ensemble"
)

type application struct {
	owner     *ensemble.Ensemble
	workspace string
	directory string
}

func (a *application) agent(role string, tools []string) (*ensemble.Agent, error) {
	return a.owner.NewAgent(ensemble.Config{Vendor: os.Getenv("LLM_VENDOR"), Model: os.Getenv("LLM_MODEL"), ResolvedModel: os.Getenv("LLM_RESOLVED_MODEL"), APIKey: os.Getenv("LLM_API_KEY"), BaseURL: os.Getenv("LLM_BASE_URL"), Workspace: a.workspace, LogPath: filepath.Join(a.directory, role+".log"), Builtins: tools, MaxTokens: 1024, System: "You are the " + role + " in a bounded writing workflow. Follow the user's exact file and role instructions. Be concise."})
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	directory := os.Getenv("ENSEMBLE_RUN_DIRECTORY")
	if directory == "" {
		return fmt.Errorf("ENSEMBLE_RUN_DIRECTORY is required")
	}
	directory, err := filepath.Abs(directory)
	if err != nil {
		return err
	}
	a := &application{owner: ensemble.New(os.Stderr), directory: directory, workspace: filepath.Join(directory, "workspace")}
	defer a.owner.Close()
	if err = os.MkdirAll(a.workspace, 0700); err != nil {
		return err
	}
	mode := "workflow"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	if mode == "collection" {
		return a.collection()
	}
	if mode != "workflow" {
		return fmt.Errorf("usage: workflow [workflow | collection]")
	}
	roles := []struct {
		name   string
		tools  []string
		prompt string
	}{
		{"author", []string{"read_file", "write_file"}, "Write draft.txt using write_file: a four-sentence public library announcement describing a Saturday repair cafe at 10 a.m. Include the exact phrase 'bring a broken lamp'. Then return a short handoff describing the draft."},
		{"editor", []string{"read_file", "write_file", "edit_file"}, "Read draft.txt and edit it using edit_file. Make the opening sentence more welcoming and add the exact phrase 'repairs are free'. Preserve Saturday, 10 a.m., and 'bring a broken lamp'. Return a short account of your actual edit."},
		{"reviewer", []string{"read_file"}, "Read draft.txt, verify Saturday, 10 a.m., 'bring a broken lamp', and 'repairs are free'. Give your actual review in one short paragraph; do not modify files."},
	}
	previous := ""
	for _, role := range roles {
		agent, err := a.agent(role.name, role.tools)
		if err != nil {
			return err
		}
		h, err := agent.Submit(role.prompt + "\nPrevious completed handoff: " + previous)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		c, err := h.Wait(ctx)
		cancel()
		if err != nil {
			return err
		}
		if c.Error != nil {
			return fmt.Errorf("%s: %s", role.name, c.Error.Message)
		}
		fmt.Printf("%s [%s/%s]: %s\n", role.name, c.AgentID, c.RequestID, c.Text)
		if err = a.save(role.name, c); err != nil {
			return err
		}
		previous = c.Text
	}
	draft, err := os.ReadFile(filepath.Join(a.workspace, "draft.txt"))
	if err != nil {
		return err
	}
	fmt.Printf("Final draft:\n%s\n", draft)
	return a.owner.Close()
}
func (a *application) save(name string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(a.directory, name+"-completion.json"), append(data, '\n'), 0600)
}
func (a *application) collection() error {
	agents := make([]*ensemble.Agent, 2)
	handles := make([]ensemble.RequestHandle, 2)
	for i := range agents {
		var err error
		agents[i], err = a.agent(fmt.Sprintf("collection-%d", i+1), nil)
		if err != nil {
			return err
		}
		handles[i], err = agents[i].Submit(fmt.Sprintf("Reply with exactly COLLECTION-%d and no tools.", i+1))
		if err != nil {
			return err
		}
	}
	// A queued cancellation has its own handle and does not target the active turn.
	canceled, err := agents[0].Submit("This queued request must be canceled before activation.")
	if err != nil {
		return err
	}
	if err = canceled.Cancel(); err != nil {
		return err
	}
	cc, err := canceled.Wait(context.Background())
	if err != nil {
		return err
	}
	if cc.Outcome != "canceled" {
		return fmt.Errorf("queued request activated unexpectedly")
	}
	fmt.Printf("Queued %s/%s: canceled\n", cc.AgentID, cc.RequestID)
	collection := a.owner.Collect(handles)
	for _, h := range handles {
		if _, err = h.Wait(context.Background()); err != nil {
			return err
		}
	}
	// The both-ready boundary is explicit. One wait must drain both in declared order.
	results, exhausted, err := collection.Wait(context.Background())
	if err != nil {
		return err
	}
	if len(results) != 2 || !exhausted {
		return fmt.Errorf("collection did not coalesce both ready handles")
	}
	for i, c := range results {
		fmt.Printf("Collection %s/%s: %s\n", c.AgentID, c.RequestID, c.Text)
		if c.Error != nil {
			return fmt.Errorf("request failed: %s", c.Error.Message)
		}
		if _, err = handles[i].Wait(context.Background()); err != nil {
			return err
		}
	}
	repeated, done, err := collection.Wait(context.Background())
	if err != nil || !done || len(repeated) != 0 {
		return fmt.Errorf("collection returned a duplicate")
	}
	independent, done, err := a.owner.Collect(handles).Wait(context.Background())
	if err != nil || !done || len(independent) != 2 {
		return fmt.Errorf("independent collection lost completion")
	}
	if err = a.save("collection", map[string]any{"results": results, "canceled": cc, "exhausted": exhausted, "independent_results": independent}); err != nil {
		return err
	}
	return a.owner.Close()
}

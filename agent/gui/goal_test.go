package gui_test

import (
	"encoding/json"
	gui "ensemble-gui-stub"
	"ensemble/ensemble"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestGoalStatusReplayAndSelection(t *testing.T) {
	dir := t.TempDir()
	for name, text := range map[string]string{
		"base":  "---\nname: base\ndescription: Goal test.\ntype: primary\nloadable-skills: goals\ntools: load_skill unload_skill\n---\nBase instructions.\n",
		"goals": "---\nname: goals\ntype: loadable\ndescription: Goal test.\ntools: goal_control\n---\nGoal instructions.\n",
	} {
		path := filepath.Join(dir, "skills", name, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	app := ensemble.New(io.Discard)
	a, err := app.Start(filepath.Join(dir, "state"), ensemble.Config{Model: "fake", BuiltinTools: true, Workspace: dir, Recall: ensemble.RecallSettings{Disabled: true}}, ensemble.AgentSpec{SkillsDir: filepath.Join(dir, "skills"), PrimarySkill: "base"}, make(selectionInbox, 1))
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()
	peer, err := app.CreateAgent("peer", "fake", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Skills().LoadDynamic("goals", 0); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "GOAL.md")
	if err := os.WriteFile(file, []byte("socket goal"), 0600); err != nil {
		t.Fatal(err)
	}
	control := a.(ensemble.GoalController)
	want, err := control.ConfigureGoal(ensemble.GoalConfig{File: file, Enabled: true, IdleSeconds: 60})
	if err != nil {
		t.Fatal(err)
	}
	hub, err := gui.New(app, a, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer hub.Close()
	server := httptest.NewServer(hub.Handler(""))
	defer server.Close()
	c := dial(t, server.URL)
	check := func(key, file, phase string) {
		t.Helper()
		frames := until(t, c, func(m map[string]any) bool { return m["type"] == "goal_status" && m["agent_key"] == key })
		b, _ := json.Marshal(frames[len(frames)-1]["goal"])
		var status ensemble.GoalStatus
		if err := json.Unmarshal(b, &status); err != nil {
			t.Fatal(err)
		}
		if status.File != file || status.Phase != phase {
			t.Fatalf("wrong selected goal: %+v", status)
		}
		if phase == "waiting" && !status.NextReminder.Equal(want.NextReminder) {
			t.Fatalf("replay changed due time: %+v", status)
		}
	}
	check("main", file, "waiting")
	c.Close()
	c = dial(t, server.URL)
	check("main", file, "waiting")
	if err := c.WriteJSON(map[string]string{"type": "select_agent", "key": "peer"}); err != nil {
		t.Fatal(err)
	}
	check("peer", "", "none")
	if s := peer.(ensemble.GoalController).GoalStatus(); s.File != "" || s.Enabled {
		t.Fatalf("child inherited goal: %+v", s)
	}
	if err := c.WriteJSON(map[string]string{"type": "select_agent", "key": "main"}); err != nil {
		t.Fatal(err)
	}
	check("main", file, "waiting")
	a.Pause()
	check("main", file, "paused")
	if _, err := control.ConfigureGoal(ensemble.GoalConfig{File: file}); err != nil {
		t.Fatal(err)
	}
	check("main", file, "paused")
	a.Unpause()
	check("main", file, "off")
	if err := a.Skills().MarkUnload("goals"); err != nil {
		t.Fatal(err)
	}
	check("main", "", "none")
}

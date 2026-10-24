package agent

import (
	"context"
	"encoding/json"
	"ensemble/internal/common"
	"ensemble/internal/history"
	"strings"
	"testing"
)

type policySkills struct{ common.SkillRegistry }

func (*policySkills) Snapshot() common.CapabilitySnapshot { return common.CapabilitySnapshot{} }
func TestChapter14ActorPolicy(t *testing.T) {
	a, err := New(&seamParent{}, common.Config{Model: "claude-opus-5-course", ContextTarget: 1, BaseDir: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	a.config.SavePath = ""
	a.history = history.New(a)
	a.skills = &policySkills{}
	if a.Config().ContextTarget != 20000 {
		t.Fatal("positive target not clamped")
	}
	appendEvent := func(e common.Event) {
		t.Helper()
		if err := a.history.Append(e); err != nil {
			t.Fatal(err)
		}
	}
	batch := func(id, name, text string) {
		appendEvent(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{From: common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, Parts: common.PartList{{Type: "tool_call", From: &common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, CallID: id, Name: name, Args: json.RawMessage(`{}`)}}}})
		appendEvent(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{CallID: id, Name: name, Parts: common.PartList{common.Text(text)}}})
	}
	presented := func() {
		appendEvent(common.Event{Type: common.RequestSent, Request: &common.RequestData{To: common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}}})
	}
	batch("old", "read", strings.Repeat("x", 300))
	presented()
	appendEvent(common.Event{Type: common.MessageReceived, Message: &common.MessageData{Actor: common.Human, Parts: common.PartList{common.Text("visible next message")}}})
	batch("keep", "keep_tool_results", "kept")
	c, err := a.prepareContext()
	if err != nil {
		t.Fatal(err)
	}
	if c.Dialogue[1].Parts[0].Stub != "" {
		t.Fatal("keep lost prior batch")
	}
	presented()
	batch("next", "read", strings.Repeat("y", 300))
	c, err = a.prepareContext()
	if err != nil {
		t.Fatal(err)
	}
	if c.Dialogue[1].Parts[0].Stub != "" {
		t.Fatal("old kept result was reclassified below the byte bands")
	}
	if c.Dialogue[len(c.Dialogue)-1].Parts[0].Stub != "" {
		t.Fatal("unseen batch cleared")
	}
	appendEvent(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{From: common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, Parts: common.PartList{{Type: "tool_call", From: &common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, CallID: "h", Name: "micro_handoff", Args: json.RawMessage(`{}`)}}}})
	reply := make(chan contextResult, 1)
	a.contextRequest(contextRequest{ctx: context.Background(), operation: "micro_handoff", text: "checkpoint", reply: reply})
	if r := <-reply; r.err != nil {
		t.Fatal(r.err)
	}
	if a.history.Context().Handoff != "" {
		t.Fatal("checkpoint before receipt")
	}
	if err := a.returned(1, common.ToolData{CallID: "h", Name: "micro_handoff", Parts: common.PartList{common.Text("receipt")}}); err != nil {
		t.Fatal(err)
	}
	c = a.history.Context()
	if c.Handoff != "checkpoint" {
		t.Fatalf("missing handoff: %+v", c)
	}
	events := a.history.Events()
	if events[len(events)-2].Type != common.ToolReturned || events[len(events)-1].Type != common.MicroHandoff {
		t.Fatal("checkpoint/receipt order")
	}
	if err := a.SetContextTarget(-1); err == nil {
		t.Fatal("negative target accepted")
	}
	if err := a.SetContextTarget(0); err != nil || a.Config().ContextTarget != 0 {
		t.Fatal("zero default rejected")
	}
}
func TestChapter14BandsProtectUnseenAndLive(t *testing.T) {
	a, _ := New(&seamParent{}, common.Config{Model: "claude-sonnet-5-course", ContextTarget: 20000, BaseDir: t.TempDir(), DataDir: t.TempDir()})
	a.config.SavePath = ""
	a.history = history.New(a)
	a.skills = &policySkills{}
	add := func(e common.Event) {
		t.Helper()
		if err := a.history.Append(e); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"old", "new"} {
		add(common.Event{Type: common.ResponseEnded, Response: &common.ResponseData{From: common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, Parts: common.PartList{{Type: "tool_call", From: &common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}, CallID: id, Name: "read", Args: json.RawMessage(`{}`)}}}})
		add(common.Event{Type: common.ToolReturned, Tool: &common.ToolData{CallID: id, Name: "read", Parts: common.PartList{common.Text(strings.Repeat("x", 6000))}}})
		if id == "old" {
			add(common.Event{Type: common.RequestSent, Request: &common.RequestData{To: common.Provenance{Vendor: common.OpenAI, Surface: common.ChatCompletions, Model: "offline"}}})
		}
	}
	c, err := a.prepareContext()
	if err != nil {
		t.Fatal(err)
	}
	if c.Dialogue[1].Parts[0].Stub == "" || c.Dialogue[3].Parts[0].Stub != "" {
		t.Fatalf("band cut included unseen batch; seen=%d entries=%d", c.SentThrough, len(c.Dialogue))
	}
}

func TestChapter14LiveCallsContainsOnlyRunningJobs(t *testing.T) {
	a, err := New(&seamParent{}, common.Config{Model: "offline", BaseDir: t.TempDir(), DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	a.config.SavePath = ""
	a.history = history.New(a)
	for i, id := range []string{"finished", "killed"} {
		job := common.JobData{Handle: i + 1, CallID: id, Name: "work", Status: common.JobRunning}
		if err := a.recordJob(common.JobReport{Job: job}); err != nil {
			t.Fatal(err)
		}
		if !a.liveCalls[id] {
			t.Fatal("running call not protected")
		}
		job.Status = common.JobDone
		if id == "killed" {
			job.Status = common.JobKilledStatus
		}
		if err := a.recordJob(common.JobReport{Job: job}); err != nil {
			t.Fatal(err)
		}
		if _, ok := a.liveCalls[id]; ok {
			t.Errorf("inactive %s retained in new membership map", id)
		}
	}
}

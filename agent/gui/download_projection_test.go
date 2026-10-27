package gui

import (
	"encoding/json"
	"ensemble/ensemble"
	"io"
	"testing"
)

func TestDownloadProjectionWaitsForSuccessfulJob(t *testing.T) {
	for _, sample := range []struct {
		name   string
		events []string
		want   int
	}{
		{"running then success", []string{
			`{"seq":1,"type":"tool_called","tool":{"call_id":"offer","name":"offer_download","args":{"path":"report.pdf"},"job":{"handle":1,"status":"running"}}}`,
			`{"seq":2,"type":"tool_returned","tool":{"call_id":"offer","job":{"handle":1,"status":"running"}}}`,
			`{"seq":3,"type":"job_finished","tool":{"job":{"handle":1,"status":"done"}}}`,
		}, 1},
		{"running then failure", []string{
			`{"seq":1,"type":"tool_called","tool":{"call_id":"offer","name":"offer_download","args":{"path":"report.pdf"},"job":{"handle":1,"status":"running"}}}`,
			`{"seq":2,"type":"tool_returned","tool":{"call_id":"offer","job":{"handle":1,"status":"running"}}}`,
			`{"seq":3,"type":"job_finished","tool":{"job":{"handle":1,"status":"done","failed":true}}}`,
		}, 0},
		{"finish before return", []string{
			`{"seq":1,"type":"tool_called","tool":{"call_id":"offer","name":"offer_download","args":{"path":"report.pdf"},"job":{"handle":1,"status":"running"}}}`,
			`{"seq":2,"type":"job_finished","tool":{"job":{"handle":1,"status":"done"}}}`,
			`{"seq":3,"type":"tool_returned","tool":{"call_id":"offer","job":{"handle":1,"status":"done"}}}`,
		}, 1},
		{"return before finish", []string{
			`{"seq":1,"type":"tool_called","tool":{"call_id":"offer","name":"offer_download","args":{"path":"report.pdf"},"job":{"handle":1,"status":"running"}}}`,
			`{"seq":2,"type":"tool_returned","tool":{"call_id":"offer","job":{"handle":1,"status":"done"}}}`,
			`{"seq":3,"type":"job_finished","tool":{"job":{"handle":1,"status":"done"}}}`,
		}, 1},
		{"report error before successful finish", []string{
			`{"seq":1,"type":"tool_called","tool":{"call_id":"offer","name":"offer_download","args":{"path":"report.pdf"},"job":{"handle":1,"status":"running"}}}`,
			`{"seq":2,"type":"tool_returned","tool":{"call_id":"offer","is_error":true,"job":{"handle":1,"status":"done"}}}`,
			`{"seq":3,"type":"job_finished","tool":{"job":{"handle":1,"status":"done"}}}`,
		}, 1},
		{"successful finish before report error", []string{
			`{"seq":1,"type":"tool_called","tool":{"call_id":"offer","name":"offer_download","args":{"path":"report.pdf"},"job":{"handle":1,"status":"running"}}}`,
			`{"seq":2,"type":"job_finished","tool":{"job":{"handle":1,"status":"done"}}}`,
			`{"seq":3,"type":"tool_returned","tool":{"call_id":"offer","is_error":true,"job":{"handle":1,"status":"done"}}}`,
		}, 1},
		{"killed job", []string{
			`{"seq":1,"type":"tool_called","tool":{"call_id":"offer","name":"offer_download","args":{"path":"report.pdf"},"job":{"handle":1,"status":"running"}}}`,
			`{"seq":2,"type":"job_finished","tool":{"job":{"handle":1,"status":"killed"}}}`,
		}, 0},
	} {
		t.Run(sample.name, func(t *testing.T) {
			root := ensemble.New(io.Discard)
			agent, err := root.NewAgent(ensemble.Config{Model: "offline", DataDir: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			defer agent.Shutdown()
			server, err := New(root, agent, io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			defer server.Close()
			server.mu.Lock()
			defer server.mu.Unlock()
			view := server.view(agent)
			count := 0
			for _, raw := range sample.events {
				var event ensemble.Event
				if err := json.Unmarshal([]byte(raw), &event); err != nil {
					t.Fatal(err)
				}
				if event.Type != "tool_called" && event.Tool.Job != nil {
					event.Tool.CalledSeq = 1
				}
				if link := view.downloadEvent(event); link != nil {
					count++
					if link.callID != "offer" || view.downloads[event.Seq] != link {
						t.Fatal("download lost call or replay identity")
					}
				}
			}
			if count != sample.want {
				t.Fatalf("issued %d capabilities, want %d", count, sample.want)
			}
		})
	}
}

func TestDownloadProjectionRejectsReusedJobHandles(t *testing.T) {
	root := ensemble.New(io.Discard)
	agent, err := root.NewAgent(ensemble.Config{Model: "offline", DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	defer agent.Shutdown()
	server, err := New(root, agent, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	server.mu.Lock()
	defer server.mu.Unlock()
	view := server.view(agent)
	for _, raw := range []string{
		`{"seq":1,"type":"tool_called","tool":{"call_id":"old","name":"offer_download","args":{"path":"report.pdf"},"job":{"handle":1,"status":"running"}}}`,
		`{"seq":2,"type":"tool_returned","tool":{"call_id":"old","called_seq":1,"job":{"handle":1,"status":"running"}}}`,
		`{"seq":3,"type":"tool_called","tool":{"call_id":"new","name":"run_command","job":{"handle":1,"status":"running"}}}`,
		`{"seq":4,"type":"job_finished","tool":{"called_seq":3,"job":{"handle":1,"status":"done"}}}`,
		`{"seq":5,"type":"tool_returned","tool":{"call_id":"new","called_seq":3,"job":{"handle":1,"status":"done"}}}`,
		`{"seq":6,"type":"tool_returned","tool":{"call_id":"old"}}`,
	} {
		var event ensemble.Event
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			t.Fatal(err)
		}
		if link := view.downloadEvent(event); link != nil {
			t.Fatalf("reused job handle issued old download: %+v", link)
		}
	}
}

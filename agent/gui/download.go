package gui

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"ensemble/ensemble"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

type downloadOffer struct {
	agent    ensemble.Agent
	path     string
	filename string
}

type downloadLink struct {
	token    string
	filename string
	callID   string
}

type downloadRequest struct {
	path   string
	callID string
	job    uint64
}

// A capability is issued only for a successfully recorded offer, never for a
// path supplied directly by an HTTP client. Replay reuses the view's link.
func (v *agentView) downloadEvent(event ensemble.Event) *downloadLink {
	if event.Tool == nil {
		return nil
	}
	switch event.Type {
	case "tool_called":
		if event.Tool.Name == "offer_download" {
			var args struct {
				Path string `json:"path"`
			}
			if json.Unmarshal(event.Tool.Args, &args) == nil && args.Path != "" {
				request := &downloadRequest{path: args.Path, callID: event.Tool.CallID}
				if event.Tool.Job != nil {
					request.job = event.Seq
					v.downloadJobs[request.job] = request
				}
				v.pendingDownloads[request.callID] = request
			}
		}
	case "tool_returned":
		request := v.pendingDownloads[event.Tool.CallID]
		if event.Tool.Job != nil {
			request = v.downloadJobs[event.Tool.CalledSeq]
			if event.Tool.Job.Status == "running" {
				return nil
			}
			if event.Tool.IsError && event.Tool.Job.Status == "done" && !event.Tool.Job.Failed {
				return nil
			}
		} else if request != nil && request.job != 0 {
			return nil
		}
		return v.finishDownload(event.Seq, request, !event.Tool.IsError &&
			(event.Tool.Job == nil || event.Tool.Job.Status == "done" && !event.Tool.Job.Failed))
	case "job_finished":
		if event.Tool.Job == nil {
			return nil
		}
		request := v.downloadJobs[event.Tool.CalledSeq]
		return v.finishDownload(event.Seq, request, event.Tool.Job.Status == "done" && !event.Tool.Job.Failed)
	}
	return nil
}

func (v *agentView) finishDownload(seq uint64, request *downloadRequest, success bool) *downloadLink {
	if request == nil {
		return nil
	}
	if v.pendingDownloads[request.callID] == request {
		delete(v.pendingDownloads, request.callID)
	}
	if v.downloadJobs[request.job] == request {
		delete(v.downloadJobs, request.job)
	}
	if !success {
		return nil
	}
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return nil
	}
	link := &downloadLink{token: hex.EncodeToString(random[:]), filename: filepath.Base(request.path), callID: request.callID}
	v.parent.downloads[link.token] = downloadOffer{agent: v.agent, path: request.path, filename: link.filename}
	v.downloads[seq] = link
	return link
}

func (link *downloadLink) message() map[string]any {
	return map[string]any{"type": "download_offered", "call_id": link.callID, "token": link.token, "filename": link.filename}
}

func (s *Server) serveDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	token := strings.TrimPrefix(r.URL.Path, "/download/")
	s.mu.Lock()
	offer, found := s.downloads[token]
	s.mu.Unlock()
	if !found {
		http.NotFound(w, r)
		return
	}
	current, err := s.root.FindAgent(ensemble.AgentKey(offer.agent), s.agent)
	if err != nil || current != offer.agent {
		http.NotFound(w, r)
		return
	}
	file, err := offer.agent.Sandbox().OpenWorkspaceFile(offer.path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": offer.filename}))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	http.ServeContent(w, r, offer.filename, info.ModTime(), file)
}

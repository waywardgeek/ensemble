package gradeedition2

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/internal/fakevendor"
)

// File-mode is the current public startup contract. All values and credentials
// here belong to the local fake provider; inherited application settings cannot
// select an endpoint or persistent state directory.
func chapter21GUIStart(bin, work, base string) (*mcpCommand, *guiPeer, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, nil, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	// Use the chapter's published key-file reference and explicit port override.
	// Optional streaming/sandbox/preference spellings are not launch prerequisites.
	keyPath := filepath.Join(work, "fixture-key")
	if err := os.WriteFile(keyPath, []byte("local-fixture-only"), 0600); err != nil {
		return nil, nil, err
	}
	config := map[string]any{
		"version": 1, "provider": "anthropic", "model": "claude-sonnet-4-6",
		"base_url": base, "api_key_file": keyPath, "auth_mode": "api",
		"base_dir": work, "data_dir": filepath.Join(work, "state"),
		"save_path":  filepath.Join(work, "state", "save.json"),
		"skills_dir": filepath.Join(work, "skills"), "skill": "base",
		"default_agent_id": "main", "enable_web_search": false,
	}
	raw, err := json.Marshal(config)
	if err != nil {
		return nil, nil, err
	}
	path := filepath.Join(work, "application.json")
	if err = os.WriteFile(path, raw, 0600); err != nil {
		return nil, nil, err
	}
	env := chapter20Env(work, "", base)
	p, err := startMCPCommandEnv(bin, work, []string{"--config", path, "--port", fmt.Sprint(port)}, env)
	if err != nil {
		return nil, nil, err
	}
	// This scenario consumes WebSocket frames, not terminal lines. Keep stdout
	// flowing so the scanner can reach EOF and reap the child normally.
	go func() {
		for range p.lines {
		}
	}()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		peer, e := openGUIPeer(fmt.Sprintf("ws://127.0.0.1:%d/ws", port))
		select {
		case exit := <-p.done:
			p.done <- exit
			if peer != nil {
				peer.close()
			}
			p.close()
			return nil, nil, fmt.Errorf("GUI child exited before readiness: %v; %s", exit, p.stderr.String())
		default:
		}
		if e == nil {
			return p, peer, nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	p.close()
	return nil, nil, fmt.Errorf("GUI readiness deadline: %s", p.stderr.String())
}

func chapter21ToolReply(body []byte, reply *fakevendor.Reply) *fakevendor.Reply {
	if reply != nil && reply.ToolName == "wait_for_agent_change" && len(reply.Tools) == 0 {
		var args map[string]json.RawMessage
		if json.Unmarshal([]byte(reply.ToolArgs), &args) == nil {
			// The current synchronous wait has its explicit five-second timeout. A
			// legacy callback argument is not part of this tool's public descriptor.
			delete(args, "ai_callback_delay")
			raw, _ := json.Marshal(args)
			copy := *reply
			copy.ToolArgs = string(raw)
			return &copy
		}
	}
	return chapter20ToolLimits(body, reply)
}

func managementTree(frame map[string]any, current bool) bool {
	if !current {
		return frame["type"] == "agent_tree" || frame["type"] == "agents_tree"
	}
	// A selected-conversation replay is not proof of a complete fleet roster.
	if frame["type"] == "replay" {
		if _, full := object(frame["data"])["agent_tree"]; !full {
			return false
		}
	}
	return len(managementRoster(frame)) != 0
}

// The roster may be the data itself or an explicit field in the envelope.
// Replay may carry both a full fleet roster and selected-conversation snapshots;
// the latter must not hide other parent-linked nodes.
func managementRoster(frame map[string]any) []any {
	var rows any
	switch frame["type"] {
	case "agent_tree", "agents_update":
		rows = frame["data"]
		if _, exists := frame["data"]; !exists {
			rows = frame["agents"]
		}
		if data, ok := rows.(map[string]any); ok {
			rows = data["agents"]
			if tree, exists := data["agent_tree"]; exists {
				rows = tree
			}
		}
	case "replay":
		data := object(frame["data"])
		rows = data["agents"]
		if tree, exists := data["agent_tree"]; exists {
			rows = tree
		}
	default:
		return nil
	}
	return managementRows(rows)
}

// Flatten a public nested tree without losing its ownership edges. A flat
// roster remains valid; duplicate identities or contradictory edges do not.
func managementRows(value any) []any {
	rows, ok := value.([]any)
	if !ok {
		return nil
	}
	seen := make(map[string]bool)
	var out []any
	var visit func([]any, string) bool
	visit = func(rows []any, parent string) bool {
		for _, value := range rows {
			raw := object(value)
			identity, hasID := raw["id"]
			if !hasID {
				identity = raw["agent_id"]
			}
			id, ok := identity.(string)
			if !ok || id == "" || seen[id] {
				return false
			}
			if alias, exists := raw["agent_id"]; exists && alias != id {
				return false
			}
			// Normalize a copy; never rewrite a received public frame.
			row := make(map[string]any, len(raw)+1)
			for k, v := range raw {
				row[k] = v
			}
			row["id"] = id
			if edge, exists := row["parent_id"]; exists {
				p, ok := edge.(string)
				if !ok || (parent != "" && p != parent) {
					return false
				}
			} else if parent != "" {
				// Nesting itself names the parent.
				row["parent_id"] = parent
			}
			seen[id] = true
			out = append(out, row)
			if children, exists := row["children"]; exists && children != nil {
				nodes, ok := children.([]any)
				if !ok || !visit(nodes, id) {
					return false
				}
			}
		}
		return true
	}
	if !visit(rows, "") {
		return nil
	}
	return out
}

// The model addresses a child locally; the browser uses its advertised identity.
// Resolve only the two demonstrated ID forms, with an explicit ownership edge.
func managementOwnedChildID(frame map[string]any, local, parent string) string {
	if frame["type"] == "replay" || frame["replay"] == true {
		return ""
	}
	found := ""
	for _, value := range managementRoster(frame) {
		row := object(value)
		id, _ := row["id"].(string)
		if row["parent_id"] != parent || (id != local && id != parent+"/"+local) {
			continue
		}
		if found != "" {
			return "" // Ambiguous local/canonical identities.
		}
		found = id
	}
	return found
}

func managementSelected(frame map[string]any) string {
	if frame["type"] != "replay" {
		return ""
	}
	data, selected := object(frame["data"]), ""
	for _, field := range []string{"selected_agent_id", "selected_agent"} {
		if value, exists := data[field]; exists {
			id, ok := value.(string)
			if !ok || id == "" || (selected != "" && selected != id) {
				return ""
			}
			selected = id
		}
	}
	if selected == "" {
		// A replay may contain just the selected conversation, rather than
		// repeating its identity in an extra envelope field.
		rows := managementRows(data["agents"])
		if len(rows) == 1 {
			row := object(rows[0])
			if turn, ok := object(row["state"])["turn"].(string); ok && turn != "" {
				selected, _ = row["id"].(string)
			}
		}
	}
	return selected
}

func managementReplayAgent(frame map[string]any, id string) map[string]any {
	if frame["type"] != "replay" || id == "" {
		return nil
	}
	for _, value := range managementRows(object(frame["data"])["agents"]) {
		if row := object(value); row["id"] == id {
			return row
		}
	}
	return nil
}

func managementTreeNode(frame map[string]any, id string, current bool) map[string]any {
	if !current {
		return managementNode(frame["root"], id)
	}
	for _, value := range managementRoster(frame) {
		if row := object(value); row["id"] == id {
			return row
		}
	}
	return nil
}

// Decode semantic phases only on identified activity frames or roster rows.
// Extra phases remain extra observations, never substitutes for the required ones.
func managementActivityPhase(frame map[string]any, id, parent string) string {
	if frame["replay"] == true || frame["type"] == "replay" {
		return ""
	}
	var activity any
	switch frame["type"] {
	case "agent_status":
		if frame["agent_id"] != id {
			return ""
		}
		activity = frame["activity"]
		if _, nested := frame["activity"]; !nested {
			activity = frame
		}
	case "agent_activity":
		key := id
		if parent != "" && !strings.HasPrefix(id, parent+"/") {
			key = parent + "/" + id
		}
		if frame["key"] != key {
			return ""
		}
		activity = frame["activity"]
	default:
		node := managementTreeNode(frame, id, true)
		actualParent, _ := node["parent_id"].(string)
		if node == nil || actualParent != parent {
			return ""
		}
		activity = node["activity"]
		if _, nested := node["activity"]; !nested {
			activity = node
		}
	}
	if a, ok := activity.(map[string]any); ok {
		if actual, exists := a["agent_id"]; exists && actual != id {
			return ""
		}
		if actual, exists := a["parent_id"]; exists && actual != parent {
			return ""
		}
	}
	phase := managementPhase(activity)
	switch phase {
	case "waiting_for_model", "in_flight":
		return "thinking"
	case "running_tools", "executing_tools", "toolcall":
		return "tool_call"
	case "<nil>":
		return ""
	default:
		return phase
	}
}

// Only live typed tool cards establish invocation/result correlation. Arbitrary
// text or tool arguments mentioning a tool name must not release a held child.
func managementLiveToolPart(frame map[string]any, agent, call, kind string) map[string]any {
	if frame["type"] != "part_final" || frame["replay"] == true || frame["agent_id"] != agent {
		return nil
	}
	part := object(frame["part"])
	if part["type"] != kind || part["call_id"] != call {
		return nil
	}
	return part
}

func managementStateWith(frames []map[string]any, current bool) string {
	if !current {
		return managementState(frames)
	}
	selected, state := "", ""
	for _, frame := range frames {
		if frame["type"] == "replay" {
			selected, state = managementSelected(frame), ""
			// Turn state belongs to the selected conversation snapshot, not the
			// fleet roster's lifecycle string (which may truthfully say "running").
			if node := managementReplayAgent(frame, selected); node != nil {
				state, _ = object(node["state"])["turn"].(string)
			}
		} else if selected != "" && frame["type"] == "state_changed" && frame["agent_id"] == selected {
			state, _ = frame["to"].(string)
		}
	}
	return state
}

// Reuse the accepted strict wrapper validation, including identity/state
// agreement. Only the current route maps the resulting typed submission.
func managementStatus(text string, current bool) (map[string]any, error) {
	rows, err := fleetStatuses("["+text+"]", false, current)
	if err != nil || len(rows) != 1 || rows[0] == nil {
		return nil, fmt.Errorf("invalid management status: %v", err)
	}
	row := rows[0]
	if current {
		row["payload"] = row["submission"]
	}
	return row, nil
}

// Keep selected history and live state only. Foreign replay events must not
// supply text or lifecycle evidence for the requested conversation.
func managementSelectionFrames(frames []map[string]any, id string) []map[string]any {
	var selected []map[string]any
	active := false
	for _, frame := range frames {
		if frame["type"] == "replay" {
			active = managementSelected(frame) == id && managementReplayAgent(frame, id) != nil
			if active {
				selected = nil
				for _, event := range list(object(frame["data"])["events"]) {
					if e := object(event); e["agent_id"] == id {
						selected = append(selected, e)
					}
				}
				selected = append(selected, frame)
			}
		} else if active && frame["agent_id"] == id {
			selected = append(selected, frame)
		}
	}
	return selected
}

func managementSubmitted(row map[string]any, payloadKey string, current bool) bool {
	return (!current || (row["agent_id"] == "worker23" && row["state"] == "submitted")) &&
		reflect.DeepEqual(row[payloadKey], fleetJSON(fleetReviewPayload))
}

// Named peers use the same identity field as current conversation selection.
func managementNamedCommand(kind, id string, current bool) map[string]string {
	field := "name"
	if current {
		field = "agent_id"
	}
	return map[string]string{"type": kind, field: id}
}

// Current GUI input uses the browser command; the historical route is unchanged.
func managementInputCommand(text string, current bool) map[string]string {
	kind := "prompt"
	if current {
		kind = "input"
	}
	return map[string]string{"type": kind, "text": text}
}

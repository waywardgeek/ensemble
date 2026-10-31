package tools

import (
	"encoding/json"
	"ensemble/internal/common"
)

// FleetTools returns the seven workflow declarations. They share ordinary job supervision. A question or join may
// outlive its first wait, while a later wait_for_job recovers its real answer.
func FleetTools() []common.ToolDefinition {
	return []common.ToolDefinition{
		{ToolDeclaration: common.ToolDeclaration{Name: "spawn_sub_agent", Description: "Start a full child agent with an isolated data directory. It inherits your model and restrictions; optional model selects another supported model. submit_schema optionally defines the entire submit argument object. Children must submit results.", Schema: json.RawMessage(`{"type":"object","properties":{"agent_id":{"type":"string"},"task":{"type":"string"},"model":{"type":"string"},"submit_schema":{"type":"object"}},"required":["agent_id","task"]}`)}, Run: spawnChild},
		{ToolDeclaration: common.ToolDeclaration{Name: "join_agents", Description: "Wait for the listed children to submit; an empty list means all children. Returns early with pending questions: answer using respond_to_parent_message, then join again. A running job report is not completion.", Schema: waitSchema(`"agent_ids":{"type":"array","items":{"type":"string"}}`, "")}, Run: joinChildren},
		{ToolDeclaration: common.ToolDeclaration{Name: "agents_status", Description: "List only your owned children and their explicit submission or failure state.", Schema: json.RawMessage(`{"type":"object","properties":{}}`)}, Run: childrenStatus},
		{ToolDeclaration: common.ToolDeclaration{Name: "shutdown_agent", Description: "Stop and remove one owned child recursively. Its saved conversation remains for a later spawn with the same ID.", Schema: json.RawMessage(`{"type":"object","properties":{"agent_id":{"type":"string"}},"required":["agent_id"]}`)}, Run: shutdownChild},
		{ToolDeclaration: common.ToolDeclaration{Name: "send_message_to_parent", Description: "Ask your parent (or the human for a root agent) a question and wait for a reply. This does not submit your work. If backgrounded, wait_for_job retrieves the reply.", Schema: waitSchema(`"content":{"type":"string"}`, `"content"`)}, Run: askParent},
		{ToolDeclaration: common.ToolDeclaration{Name: "respond_to_parent_message", Description: "Answer a question sent by one of your children. Omit message_id only when that child has one pending question.", Schema: json.RawMessage(`{"type":"object","properties":{"agent_id":{"type":"string"},"message_id":{"type":"string"},"response":{"type":"string"}},"required":["agent_id","response"]}`)}, Run: replyChild},
		{ToolDeclaration: common.ToolDeclaration{Name: "submit", Description: "Finish your task with a result. Rejected while children remain unfinished: join them or shut them down first.", Schema: json.RawMessage(`{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`)}, Run: submitFleet},
	}
}
func fleetJSON(value any) (string, error) { data, err := json.Marshal(value); return string(data), err }
func spawnChild(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// ID names the child under this Agent.
		ID string `json:"agent_id"`
		// Task supplies the child's first human turn.
		Task string `json:"task"`
		// Model optionally overrides the inherited provider model.
		Model string `json:"model"`
		// SubmitSchema is the whole child submit argument schema, not a result wrapper.
		SubmitSchema json.RawMessage `json:"submit_schema"`
	}
	if err := decode(raw, &args, "agent_id", "task"); err != nil {
		return "", err
	}
	status, err := owner.Agent().Fleet().SpawnWithSchema(args.ID, args.Task, args.Model, args.SubmitSchema)
	if err != nil {
		return "", err
	}
	return fleetJSON(status)
}
func joinChildren(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// IDs selects owned children; omission joins the current child set.
		IDs []string `json:"agent_ids"`
	}
	if err := decode(raw, &args); err != nil {
		return "", err
	}
	return owner.Agent().Fleet().Join(owner.Job().Context(), args.IDs)
}
func childrenStatus(owner common.ToolContext, _ json.RawMessage) (string, error) {
	return fleetJSON(owner.Agent().Fleet().List())
}
func shutdownChild(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// ID identifies an owned child, never an arbitrary fleet member.
		ID string `json:"agent_id"`
	}
	if err := decode(raw, &args, "agent_id"); err != nil {
		return "", err
	}
	if err := owner.Agent().Fleet().ShutdownChild(args.ID); err != nil {
		return "", err
	}
	return "child stopped; saved work retained", nil
}
func askParent(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Content is the question, distinct from a submitted result.
		Content string `json:"content"`
	}
	if err := decode(raw, &args, "content"); err != nil {
		return "", err
	}
	message, err := owner.Agent().Fleet().AskParent(owner.Job().Context(), args.Content)
	if err != nil {
		return "", err
	}
	return fleetJSON(message)
}
func replyChild(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// ID identifies the questioning child.
		ID string `json:"agent_id"`
		// MessageID disambiguates multiple questions from that child.
		MessageID string `json:"message_id"`
		// Response is delivered together with the parent's sender address.
		Response string `json:"response"`
	}
	if err := decode(raw, &args, "agent_id", "response"); err != nil {
		return "", err
	}
	if err := owner.Agent().Fleet().Reply(args.ID, args.MessageID, args.Response); err != nil {
		return "", err
	}
	return "reply delivered", nil
}
func submitFleet(owner common.ToolContext, raw json.RawMessage) (string, error) {
	if err := owner.Agent().Fleet().SubmitJSON(raw); err != nil {
		return "", err
	}
	status := owner.Agent().Fleet().Status()
	if len(status.Payload) != 0 {
		return "submitted: " + string(status.Payload), nil
	}
	return "submitted: " + status.Result, nil
}

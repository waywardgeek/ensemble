package tools

import (
	"encoding/json"
	"ensemble/internal/common"
)

// SkillTools supplies lifecycle verbs only when the Agent has a skill registry.
// Handlers follow their actual owner to the registry; no captured sibling state.
func SkillTools() []common.ToolDefinition {
	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`)
	return []common.ToolDefinition{
		{ToolDeclaration: common.ToolDeclaration{Name: "load_skill", Description: "Load an available skill and its dependencies. Returns its instructions and enables its tools.", Schema: schema}, Run: loadSkill},
		{ToolDeclaration: common.ToolDeclaration{Name: "unload_skill", Description: "Mark a loaded skill for removal at the next compaction. Its MCP servers stop now; compiled tools and instructions remain until then.", Schema: schema}, Run: unloadSkill},
	}
}

func loadSkill(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Name identifies a reachable loadable skill, not a path supplied by the model.
		Name string `json:"name"`
	}
	if err := decode(raw, &args, "name"); err != nil {
		return "", err
	}
	return owner.Agent().Skills().LoadDynamic(args.Name, owner.EventSeq())
}

func unloadSkill(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		// Name identifies the loaded skill to mark pending removal.
		Name string `json:"name"`
	}
	if err := decode(raw, &args, "name"); err != nil {
		return "", err
	}
	if err := owner.Agent().Skills().MarkUnload(args.Name); err != nil {
		return "", err
	}
	return "Skill " + args.Name + " is pending unload at the next compaction.", nil
}

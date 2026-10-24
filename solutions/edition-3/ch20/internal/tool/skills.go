package tool

import (
	"encoding/json"
	"ensemble/internal/common"
	"fmt"
	"io"
)

func (r *registry) skillTools() []entry {
	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"],"additionalProperties":false}`)
	return []entry{
		{definition: common.ToolDefinition{Name: "load_skill", Description: "Enable a named skill and its dependencies atomically. Offers only advertise; tools and instructions become available after a successful load.", Parameters: schema}},
		{definition: common.ToolDefinition{Name: "unload_skill", Description: "Mark an enabled skill pending unload. Instruction survivors remain. Compiled and external grants are removed immediately unless independently owned; peers with no remaining grants are closed.", Parameters: schema}},
	}
}
func (r *registry) skillCall(owner common.ToolCall, tool string, args arguments, seq common.Seq) error {
	if err := args.only("name"); err != nil {
		return err
	}
	name, err := args.text("name", true)
	if err != nil {
		return err
	}
	if tool == "unload_skill" {
		if err = r.parent.Skills().UnloadSkill(name); err != nil {
			return err
		}
		_, err = fmt.Fprintf(owner, "Skill %s is pending unload. Its instruction survivors remain; compiled and external grants are removed immediately unless independently owned, and peers with no remaining grants are closed.", name)
		return err
	}
	if seq == 0 {
		return fmt.Errorf("load_skill requires recorded tool admission")
	}
	err = r.parent.Skills().LoadDynamic(name, seq)
	if err != nil {
		return err
	}
	_, err = io.WriteString(owner, "Enabled skills: "+name+". Instructions are supplied through projected context; this result is status, not an instruction channel.")
	return err
}

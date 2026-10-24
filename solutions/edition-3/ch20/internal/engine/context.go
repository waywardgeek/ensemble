package engine

import "ensemble/internal/common"

// Every model receives surviving skill manuals in SYSTEM. InlineTools describes
// synthetic declaration placement only, not a separate instruction channel.
func projectPrompt(c common.Context, cfg common.Config, defs []common.ToolDefinition) (common.Context, common.Config, []common.ToolDefinition) {
	model := cfg.ResolvedModel
	if model == "" {
		model = cfg.Model
	}
	features, _ := common.LookupModel(model)
	base := cfg.SystemPrompt
	if base == "" {
		base = c.System
	}
	for _, skill := range c.ActiveSkills {
		base += "\n\n" + skill.Body
	}
	cfg.SystemPrompt = base
	if c.ToolsInitialized && features.InlineTools {
		decls := c.BaseTools
		defs = nil
		for _, d := range decls {
			defs = append(defs, common.ToolDefinition{Name: d.Name, Description: d.Description, Parameters: d.Schema})
		}
	}
	out := make([]common.Entry, 0, len(c.Dialogue))
	for _, entry := range c.Dialogue {
		switch entry.Kind {
		case common.KindSkill:
			// Keep the survivor in History, not as an extra instruction message.
			continue
		case common.KindTools:
			if !features.InlineTools || entry.Seq <= c.BaseThrough {
				continue
			}
			entry.Actor = common.System
		case common.KindHandoff:
			continue
		}
		out = append(out, entry)
	}
	c.Dialogue = out
	if c.Handoff != "" {
		c.Ephemera = append(common.PartList{common.Text(c.Handoff)}, c.Ephemera...)
	}
	return c, cfg, defs
}
func inlineFeatures(cfg common.Config) bool {
	model := cfg.ResolvedModel
	if model == "" {
		model = cfg.Model
	}
	f, _ := common.LookupModel(model)
	return f.InlineTools
}

func offers(skills []common.SkillInfo) string {
	text := ""
	for _, skill := range skills {
		if !skill.Enabled {
			text += "\n- " + skill.Name + ": " + skill.Description
		}
	}
	if text != "" {
		return "\n\nAvailable skills:" + text
	}
	return ""
}

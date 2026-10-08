package tools

import (
	"bytes"
	"encoding/json"
	"example.com/ensemble/internal/common"
	"io"
)

func (r *Registry) Management(name string) bool {
	return name == "load_skill" || name == "unload_skill"
}
func (r *Registry) SkillOperation(p common.Part, revision uint64) (common.SkillOperation, error) {
	name := ""
	fail := func() (common.SkillOperation, error) {
		err := &common.SkillError{Code: "invalid_skill_arguments", Name: name, Revision: revision, Detail: "management call requires exactly one valid name"}
		r.parent.Ensemble().Logf("%s", err.Detail)
		return common.SkillOperation{}, err
	}
	d := json.NewDecoder(bytes.NewReader(p.Args))
	token, err := d.Token()
	if err != nil || token != json.Delim('{') {
		return fail()
	}
	count := 0
	invalid := false
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return fail()
		}
		var raw json.RawMessage
		if d.Decode(&raw) != nil {
			return fail()
		}
		count++
		if key != "name" {
			invalid = true
			continue
		}
		var value string
		if json.Unmarshal(raw, &value) != nil {
			invalid = true
			continue
		}
		valid := len(value) > 0 && len(value) <= 64
		for i, c := range []byte(value) {
			if !(c >= 'a' && c <= 'z' || i > 0 && (c >= '0' && c <= '9' || c == '-')) {
				valid = false
			}
		}
		if !valid {
			invalid = true
		} else {
			name = value
		}
	}
	if _, err = d.Token(); err != nil {
		return fail()
	}
	if _, err = d.Token(); err != io.EOF || invalid || count != 1 || name == "" {
		return fail()
	}
	action := "load"
	if p.Name == "unload_skill" {
		action = "unload"
	}
	return common.SkillOperation{Action: action, Name: name}, nil
}
func (r *Registry) SkillAcknowledgement(p common.Part, result common.SkillResult, err error, note string) common.ToolEvent {
	var value any = result
	if err != nil {
		failure := err.(*common.SkillError)
		value = struct {
			Error    string `json:"error"`
			Name     string `json:"name"`
			Revision uint64 `json:"revision"`
		}{failure.Code, failure.Name, failure.Revision}
	}
	raw, _ := json.Marshal(value)
	text := note + string(raw)
	return common.ToolEvent{CallID: p.CallID, IsError: err != nil, Parts: []common.Part{{Type: "text", Text: &text}}}
}

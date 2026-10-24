package agent

import (
	"ensemble/internal/common"
	"errors"
	"strings"
	"unicode"
)

type modelRequest struct {
	model string
	reply chan error
}

func (a *agent) SetModel(model string) error {
	reply := make(chan error, 1)
	if err := a.box.post(modelRequest{model, reply}, true); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-a.done:
		return common.ErrStopped
	}
}

// The actor serializes model changes against admission. No active request,
// pending turn or history fact is rewritten, and Engine keeps its usage buckets.
func (a *agent) changeModel(model string) error {
	if a.stopping {
		return common.ErrStopped
	}
	if a.active != nil || len(a.pending) > 0 {
		return errors.New("cannot change model during an active turn")
	}
	if model == "" || strings.TrimSpace(model) != model || strings.ContainsFunc(model, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
		return errors.New("model must be a non-empty model identifier")
	}
	c := a.Config()
	// Known foreign identifiers are mistakes; unknown same-provider names remain
	// usable so a stale catalog cannot prevent selecting a newly released model.
	foreign := (strings.HasPrefix(model, "claude-") && c.Vendor != common.Anthropic) ||
		(strings.HasPrefix(model, "gemini-") && c.Vendor != common.Gemini) ||
		((strings.HasPrefix(model, "gpt-") || strings.HasPrefix(model, "o1") || strings.HasPrefix(model, "o3") || strings.HasPrefix(model, "o4")) && c.Vendor != common.OpenAI)
	if foreign {
		return errors.New("model change cannot change provider")
	}
	if c.Model == model {
		return nil
	}
	c.Model = model
	c.ResolvedModel = ""
	if c.Vendor == common.OpenAI {
		c.RequireToolExplanation = false
	}
	if err := common.NormalizeOpenAI(&c); err != nil {
		return err
	}
	a.configMu.Lock()
	a.config.Model = c.Model
	a.config.ResolvedModel = c.ResolvedModel
	a.config.RequireToolExplanation = c.RequireToolExplanation
	a.configMu.Unlock()
	a.publish(common.ModelChanged{Model: model})
	return nil
}

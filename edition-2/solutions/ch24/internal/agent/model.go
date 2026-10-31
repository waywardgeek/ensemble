package agent

import (
	"fmt"
	"strings"

	"ensemble/internal/common"
)

type modelSelection struct {
	model string
	reply chan error
}

// SetModel changes only this Agent's model within its configured provider and
// billing route. The actor rejects busy changes, so render/send cannot disagree
// about the request model. Existing counts and conversation remain untouched.
func (a *agent) SetModel(model string) error {
	model = strings.TrimSpace(model)
	if model == "" {
		return fmt.Errorf("model must not be empty")
	}
	reply := make(chan error, 1)
	if err := a.post(modelSelection{model, reply}); err != nil {
		return err
	}
	select {
	case err := <-reply:
		return err
	case <-a.done:
		return common.ErrStopped
	}
}

func (r *runtime) selectModel(selection modelSelection) {
	if r.active != nil {
		selection.reply <- fmt.Errorf("change model between turns")
		return
	}
	r.owner.configMu.Lock()
	r.owner.config.Model = selection.model
	// A resolved identity belongs to the previous selection, not the new alias.
	r.owner.config.ResolvedModel = ""
	r.owner.configMu.Unlock()
	selection.reply <- nil
}

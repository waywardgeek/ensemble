package cli

import (
	"context"
	"encoding/json"
	"ensemble"
	"fmt"
	"path/filepath"
)

// installAgentFactory keeps application defaults out of the fleet owner. Root
// and child construction share this seam; NewAgent remains the library's
// pre-activation validation and registration boundary.
func installAgentFactory(root *ensemble.Ensemble) error {
	root.SetAgentFactory(func(ctx context.Context, request ensemble.Config) (ensemble.Agent, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		config := request
		if request.ParentID != "" {
			parent, ok := root.Agent(request.ParentID)
			if !ok {
				return nil, fmt.Errorf("unknown parent %q", request.ParentID)
			}
			config = parent.Config().Child(filepath.Join(parent.Config().DataDir, "children", request.ID))
			config.ID, config.ParentID, config.Prompt = request.ID, request.ParentID, request.Prompt
			config.SubmitSchema = append(json.RawMessage(nil), request.SubmitSchema...)
			if request.Model != "" {
				config.Model = request.Model
				config.ResolvedModel = ""
				config.StreamCapability = ensemble.StreamUnknown
				config.RequireToolExplanation = false
			}
			if request.ReasoningEffort != "" {
				config.ReasoningEffort = request.ReasoningEffort
			}
		}
		return root.NewAgent(config)
	})
	return nil
}

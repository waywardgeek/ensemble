package agent

import (
	"fmt"

	"ensemble/internal/common"
)

// Save leaves runtime configuration and execution untouched. The registry
// supplies one effective checkpoint (including pending unload), not the launch
// recipe. No API key, endpoint, handler, or variable callback is serialized.
func (a *agent) Save() error {
	a.saveMu.Lock()
	defer a.saveMu.Unlock()
	cfg := a.Config()
	if cfg.SavePath == "" {
		return nil
	}
	capabilities := a.skills.Snapshot()
	prompt, err := a.engine.EffectiveSystemPrompt(a.history.Context())
	if err != nil {
		return err
	}
	declarations := []common.ToolDeclaration{}
	for _, t := range capabilities.Definitions {
		declarations = append(declarations, common.ToolDeclaration{Name: t.Name, Description: t.Description, Schema: t.Parameters})
	}
	limit := cfg.MaxToolRounds
	if limit == 0 {
		limit = common.DefaultMaxToolRounds
	}
	return a.history.SaveFile(cfg.SavePath, common.SavedConfig{
		Vendor: cfg.Vendor, Surface: cfg.Surface, Model: cfg.Model,
		AuthMode: cfg.AuthMode, ReasoningEffort: cfg.ReasoningEffort, ReasoningSummary: cfg.ReasoningSummary,
		ResolvedModel: cfg.ResolvedModel, SystemPrompt: prompt,
		ContextTarget: cfg.ContextTarget, MaxToolRounds: limit, DisableStreaming: cfg.DisableStreaming,
		StreamCapability:       cfg.StreamCapability,
		RequireToolExplanation: cfg.RequireToolExplanation,
		Tools:                  declarations, Skills: capabilities.Skills,
	}, cfg.LogMode)
}

// Load is construction-only root wiring. Neither a failed archive nor a
// backup's descriptive configuration can alter the Agent supplied by callers.
func (a *agent) Load() error {
	if a.config.SavePath == "" {
		return nil
	}
	r, err := a.history.LoadFile(a.config.SavePath)
	if err != nil {
		return err
	}
	a.recovery = r
	if r.Path == "" {
		return nil
	}
	parts := func(ps common.PartList) {}
	parts = func(ps common.PartList) {
		for _, p := range ps {
			a.partID = max(a.partID, p.DisplayID)
			parts(p.Parts)
		}
	}
	context := a.history.Context()
	for _, e := range context.Dialogue {
		parts(e.Parts)
	}
	parts(context.Ephemera)
	for _, e := range a.history.Events() {
		a.nextTurn = max(a.nextTurn, e.TurnID)
		if e.Message != nil {
			parts(e.Message.Parts)
		}
		if e.Response != nil {
			parts(e.Response.Parts)
		}
		if e.Tool != nil {
			a.partID = max(a.partID, e.Tool.DisplayID, e.Tool.CallPartID)
			parts(e.Tool.Parts)
		}
		if e.Job != nil {
			a.partID = max(a.partID, e.Job.CallPartID)
			a.nextTurn = max(a.nextTurn, e.Job.TurnID)
		}
		if e.Redact != nil {
			parts(e.Redact.Replacement)
		}
	}
	// Legacy/context-only saves may have no turn IDs. Sequence is also a safe
	// high-water mark; do not manufacture or rewrite canonical historical IDs.
	a.nextTurn = max(a.nextTurn, uint64(r.AsOf))
	if a.partID == ^uint64(0) || a.nextTurn == ^uint64(0) {
		return fmt.Errorf("history identity space exhausted")
	}
	a.viewState.Through = r.AsOf
	a.viewValue.Store(a.viewState)
	if r.Backup {
		a.parent.Logger().Printf("recovered backup %s through sequence %d", r.Path, r.AsOf)
	}
	return nil
}

// SetPersistencePolicy publishes validated next-save policy without changing a turn.
func (a *agent) SetPersistencePolicy(mode common.LogMode, records int) error {
	if mode == "" {
		mode = common.LogFull
	}
	if mode != common.LogFull && mode != common.LogDialogueOnly {
		return fmt.Errorf("invalid log_mode")
	}
	if records < 0 {
		return fmt.Errorf("negative diagnostic_max_records")
	}
	a.configMu.Lock()
	defer a.configMu.Unlock()
	a.config.LogMode, a.config.DiagnosticMaxRecords = mode, &records
	return nil
}

// SetSecurityPolicy changes capability gates, never the root of an open Sandbox.
func (a *agent) SetSecurityPolicy(safe, web bool) {
	a.configMu.Lock()
	changed := a.config.SafeMode != safe || a.config.EnableWebSearch != web
	a.config.SafeMode, a.config.EnableWebSearch = safe, web
	a.configMu.Unlock()
	if changed && a.skills != nil {
		if err := a.skills.RefreshPolicy(); err != nil {
			a.parent.Logger().Print("security policy: " + common.Redact(err.Error(), append(common.CredentialValues(), a.Config().APIKey)...))
		}
	}
}

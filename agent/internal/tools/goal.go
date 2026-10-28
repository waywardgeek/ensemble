package tools

import (
	"encoding/json"
	"errors"
	"math"

	"ensemble/internal/common"
)

// GoalTools installs a dormant control alongside the skill lifecycle tools.
func GoalTools() []common.ToolDefinition {
	return []common.ToolDefinition{{ToolDeclaration: common.ToolDeclaration{
		Name:        "goal_control",
		Description: "Configure this Agent's goal file and idle reminders, or read status with no fields. Loading the goals skill does not enable reminders. Disabling retains the goal in SYSTEM; an empty file clears it.",
		Schema:      json.RawMessage(`{"type":"object","properties":{"file":{"type":"string"},"enabled":{"type":"boolean"},"idle_seconds":{"type":"number","exclusiveMinimum":0}}}`),
	}, Run: controlGoal}}
}
func controlGoal(owner common.ToolContext, raw json.RawMessage) (string, error) {
	var args struct {
		File        *string  `json:"file"`
		Enabled     *bool    `json:"enabled"`
		IdleSeconds *float64 `json:"idle_seconds"`
	}
	if err := decode(raw, &args); err != nil {
		return "", err
	}
	controller, ok := owner.Agent().(common.GoalController)
	if !ok {
		return "", errors.New("Agent does not support goal reminders")
	}
	status := controller.GoalStatus()
	if args.File != nil || args.Enabled != nil || args.IdleSeconds != nil {
		cfg := status.GoalConfig
		if args.File != nil {
			cfg.File = *args.File
		}
		if args.Enabled != nil {
			cfg.Enabled = *args.Enabled
		}
		if args.IdleSeconds != nil {
			if *args.IdleSeconds <= 0 || math.IsNaN(*args.IdleSeconds) || math.IsInf(*args.IdleSeconds, 0) {
				return "", errors.New("idle_seconds must be positive and finite")
			}
			cfg.IdleSeconds = *args.IdleSeconds
		}
		var err error
		status, err = controller.ConfigureGoal(cfg)
		if err != nil {
			return "", err
		}
	}
	body, err := json.Marshal(status)
	return string(body), err
}

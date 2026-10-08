package common

import (
	"context"
	"encoding/json"
	"example.com/ensemble"
)

type Preferences struct {
	Theme        string  `json:"theme"`
	FontSize     int     `json:"font_size"`
	SidebarWidth int     `json:"sidebar_width"`
	ActionsWidth int     `json:"actions_width"`
	Autoplay     bool    `json:"autoplay"`
	SpeechRate   float64 `json:"speech_rate"`
}
type PreferencesSnapshot struct {
	Revision    uint64      `json:"revision"`
	Preferences Preferences `json:"preferences"`
}
type PreferencesWatch interface {
	Service() PreferencesService
	Next(context.Context) (PreferencesSnapshot, error)
	Close()
	Done() <-chan struct{}
}
type PreferencesService interface {
	CloseWatch(PreferencesWatch)
	Server() Server
	Snapshot() PreferencesSnapshot
	Subscribe() (PreferencesSnapshot, PreferencesWatch, error)
	Update(uint64, json.RawMessage) (PreferencesSnapshot, error)
	Close()
}
type Server interface {
	Ensemble() ensemble.ClientOwner
	AgentID() string
	Origin() string
	Trace(string, string, string, any)
	Preferences() PreferencesService
}

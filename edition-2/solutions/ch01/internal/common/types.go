// Package common supplies vocabulary; implementation packages own the behavior.
package common

import (
	"context"
	"log"
)

// Config belongs to one Agent. Credentials are transport inputs, never messages.
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

// Message is one retained turn. Completed conversations contain user/assistant pairs.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage is measured by the provider; clients must not estimate it from text.
type Usage struct {
	Input  int `json:"input"`
	Output int `json:"output"`
}

// Parent interfaces expose the actual owners, rather than copied sibling services.
type Ensemble interface{ Logger() *log.Logger }

// An Agent exposes its Engine so clients read accounting from the owner.
// These interfaces contain only the services needed by the current text loop.
type Agent interface {
	Ensemble() Ensemble
	Config() Config
	Engine() Engine
	Ask(context.Context, string) (string, error)
}
type Engine interface {
	Agent() Agent
	Send(context.Context, []Message) (string, error)
	Usage() Usage
}

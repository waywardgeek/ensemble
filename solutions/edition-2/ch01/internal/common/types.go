// Package common declares shared vocabulary; implementations live in their spokes.
package common

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Conversation []Message

type Config struct {
	APIKey  string
	Model   string
	BaseURL string
}

type Usage struct {
	Input  int64 `json:"input"`
	Output int64 `json:"output"`
}

type Ensemble interface {
	Logf(string, ...any)
}

type Agent interface {
	Ensemble() Ensemble
	Config() Config
	History() Conversation
	Commit(Message, Message)
}

type Engine interface {
	Agent() Agent
}

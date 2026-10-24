package ensemble

import (
	"ensemble/internal/auth"
	"ensemble/internal/common"
	"fmt"
)

type Auth = common.Auth
type AuthMode = common.AuthMode
type AuthStatus = common.AuthStatus
type AuthOptions = auth.Options

const (
	AuthChatGPT = common.AuthChatGPT
	AuthAPI     = common.AuthAPI
)

func NewAuth(options AuthOptions) Auth { return auth.New(options) }
func (e *Ensemble) Auth() Auth         { return e.auth }

// SetAuth wires application authentication before any Agent is constructed.
func (e *Ensemble) SetAuth(a Auth) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if a == nil || e.closed || e.nextAgent != 0 {
		return fmt.Errorf("auth must be configured before Agents")
	}
	e.auth = a
	return nil
}

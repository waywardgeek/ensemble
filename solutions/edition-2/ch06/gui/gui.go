// Package gui is the optional GUI boundary stub. Browser transport comes later.
package gui

import (
	"context"
	"ensemble/ensemble"
)

// Submit accepts the application's existing Agent, so a future GUI cannot
// accidentally start a separate conversation merely by choosing another view.
func Submit(ctx context.Context, agent ensemble.Agent, input string) (string, error) {
	return agent.Ask(ctx, input)
}

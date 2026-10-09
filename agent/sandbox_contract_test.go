package agent

import (
	"github.com/waywardgeek/ensemble/agent/internal/common"
	"github.com/waywardgeek/ensemble/agent/sandbox"
)

// *sandbox.Sandbox must satisfy common.Sandbox. The public implementation
// and the hub interface deliberately never import each other, so this
// assertion is the only thing keeping them in step. Without it the two
// drift apart and the failure shows up as an inscrutable type error at a
// distant call site.
//
// It lives in a test file because the codebase holds no mutable
// package-level variables in production code. A compile-time assertion is
// still a compile-time assertion here: this file has to build for any test
// in the package to run, so a drift between the two types fails the build
// exactly as it did before.
var _ common.Sandbox = (*sandbox.Sandbox)(nil)

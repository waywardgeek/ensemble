package ensemble

import (
	"ensemble/internal/common"
	"ensemble/internal/diagnostics"
	"ensemble/internal/sandbox"
)

// DiagnosticLog retains complete API or GUI trace records independently of
// recovery history. Positive limits keep newest records; zero is unlimited.
type DiagnosticLog = diagnostics.Log

// OpenDiagnosticLog wires an existing trace destination to its owner's writer.
// The callback is sampled by Retain at a safe startup or shutdown boundary.
func OpenDiagnosticLog(path string, limit func() int) (*DiagnosticLog, error) {
	return diagnostics.Open(path, limit)
}

// OpenAgentDiagnosticLog confines diagnostics to the Agent root. It owns a
// separate rooted handle so retention can run after the Agent has shut down.
func OpenAgentDiagnosticLog(a Agent, path string, limit func() int) (*DiagnosticLog, error) {
	files, err := sandbox.New(a)
	if err != nil {
		return nil, err
	}
	log, err := diagnostics.Open(path, limit, files)
	if err != nil {
		_ = files.Close()
	}
	return log, err
}

// RedactCredentials protects human-facing adapter errors and diagnostic text.
func RedactCredentials(text string, configs ...Config) string {
	secrets := common.CredentialValues(configs...)
	return common.Redact(text, secrets...)
}

package fleet

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ensemble/internal/common"
)

// Custom records are separate from the legacy result-string row, so a later
// same-ID task can retrieve old work without guessing its type from JSON keys.
type submissionRecord struct {
	Version int             `json:"version"`       // Identifies this envelope, not the user's payload.
	Schema  json.RawMessage `json:"submit_schema"` // Retains the contract that accepted these bytes.
	Payload json.RawMessage `json:"payload"`       // The direct submit arguments, not a quoted string.
}

// SubmitSchema returns owned declaration bytes from the runtime validator.
func (m *manager) SubmitSchema() json.RawMessage { return m.contract.Schema() }

// Submit keeps the original string API without bypassing a custom contract.
func (m *manager) Submit(result string) error {
	raw, err := json.Marshal(map[string]string{"result": result})
	if err != nil {
		return err
	}
	return m.SubmitJSON(raw)
}

// SubmitJSON validates before writing or notifying a join. The caller's byte
// slice is never retained, and invalid data leaves the correction path available.
func (m *manager) SubmitJSON(raw json.RawMessage) error {
	owned, err := m.contract.validate(raw)
	if err != nil {
		return err
	}
	var result common.Submission
	var record any
	if m.contract.custom {
		result.Payload = owned
		record = submissionRecord{Version: 1, Schema: m.contract.Schema(), Payload: owned}
	} else {
		// Legacy decoding ignored extra keys. Keep that behavior and its disk
		// representation rather than silently making old callers strict.
		var legacy struct {
			Result string `json:"result"`
		}
		if err := json.Unmarshal(owned, &legacy); err != nil {
			return err
		}
		result.Result = legacy.Result
		record = map[string]string{"result": result.Result}
	}
	m.mu.Lock()
	if ids := m.active(); len(ids) > 0 {
		m.mu.Unlock()
		return fmt.Errorf("active children %v: join_agents or shutdown_agent before submit", ids)
	}
	file, err := m.parent.Sandbox().OpenFile(filepath.Join(m.parent.Config().DataDir, "submit.md"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	err = errors.Join(json.NewEncoder(file).Encode(record), file.Close())
	if err != nil {
		m.mu.Unlock()
		return err
	}
	m.submitted = true
	m.result = result
	m.failure = ""
	m.wakes = 0
	m.signal()
	m.mu.Unlock()
	if parent := m.parent.ParentAgent(); parent != nil {
		parent.Fleet().Changed()
	}
	return nil
}

// Submission preserves default text retrieval; custom results require the typed
// API rather than pretending their JSON is ordinary submitted text.
func (m *manager) Submission(id string) (string, error) {
	value, err := m.SubmissionData(id)
	if err != nil {
		return "", err
	}
	if len(value.Payload) != 0 {
		return "", errors.New("structured submission: use SubmissionData")
	}
	return value.Result, nil
}

// SubmissionData resolves the owned child before storage and never treats a
// historical fallback as completion of its new task. Each custom row retains
// its own schema; a respawn may have a different current contract.
func (m *manager) SubmissionData(id string) (common.Submission, error) {
	a, err := m.findChild(id)
	if err != nil {
		return common.Submission{}, err
	}
	if s := a.Fleet().Status(); s.State == "submitted" {
		return common.Submission{Result: s.Result, Payload: s.Payload}, nil
	}
	data, err := a.Sandbox().ReadFile(filepath.Join(a.Config().DataDir, "submit.md"))
	if err != nil {
		return common.Submission{}, err
	}
	var last common.Submission
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		last, err = readSubmission([]byte(line))
		if err != nil {
			return common.Submission{}, fmt.Errorf("read submission: %w", err)
		}
	}
	return last, nil
}

func readSubmission(raw []byte) (common.Submission, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return common.Submission{}, err
	}
	// Only the original exact row shape is legacy; malformed custom envelopes
	// cannot be mistaken for empty successful text or skipped for an older row.
	if len(fields) == 1 && fields["result"] != nil {
		var text string
		if string(fields["result"]) == "null" {
			return common.Submission{}, errors.New("null legacy result")
		}
		if err := json.Unmarshal(fields["result"], &text); err != nil {
			return common.Submission{}, err
		}
		return common.Submission{Result: text}, nil
	}
	var record submissionRecord
	if err := json.Unmarshal(raw, &record); err != nil {
		return common.Submission{}, err
	}
	if len(fields) != 3 || record.Version != 1 || len(record.Schema) == 0 || len(record.Payload) == 0 {
		return common.Submission{}, errors.New("invalid or unsupported submission envelope")
	}
	contract, err := CompileSubmissionSchema(record.Schema)
	if err != nil {
		return common.Submission{}, err
	}
	payload, err := contract.validate(record.Payload)
	if err != nil {
		return common.Submission{}, err
	}
	return common.Submission{Payload: payload}, nil
}

// Resolve ownership before disk fallback; a guessed foreign ID reveals nothing.
func (m *manager) findChild(id string) (common.Agent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c := m.children[id]; c != nil {
		return c.agent, nil
	}
	return nil, fmt.Errorf("unknown agent %q", id)
}

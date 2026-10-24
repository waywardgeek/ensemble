package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"time"
)

type submissionRow struct {
	Version   int             `json:"version"`
	Timestamp time.Time       `json:"timestamp"`
	AgentID   string          `json:"agent_id"`
	ParentID  string          `json:"parent_id"`
	Schema    json.RawMessage `json:"submit_schema,omitempty"`
	Result    *string         `json:"result,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// ReadSubmission validates all durable rows before returning the last accepted
// result. The caller must not activate a recovered terminal conversation.
func ReadSubmission(sb Sandbox, c Config) (json.RawMessage, error) {
	raw, err := sb.ReadFile(filepath.Join(c.DataDir, "submit.md"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return parseSubmissions(raw, c)
}

// Parse the exact bytes read by either recovery or append. Every row carries
// its own contract, and every row must validate before any result is returned.
func parseSubmissions(raw []byte, c Config) (json.RawMessage, error) {
	configuredSchema, err := CanonicalSubmitSchema(c.SubmitSchema)
	if err != nil {
		return nil, err
	}
	scan := bufio.NewScanner(bytes.NewReader(raw))
	scan.Buffer(make([]byte, 4096), 1100000)
	var last json.RawMessage
	line := 0
	for scan.Scan() {
		line++
		data := scan.Bytes()
		if len(bytes.TrimSpace(data)) == 0 {
			return nil, fmt.Errorf("submit.md row %d is empty", line)
		}
		var row submissionRow
		d := json.NewDecoder(bytes.NewReader(data))
		d.DisallowUnknownFields()
		if err := d.Decode(&row); err != nil {
			return nil, fmt.Errorf("submit.md row %d: %w", line, err)
		}
		var extra any
		if err := d.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("submit.md row %d has trailing data", line)
		}
		if row.Version != 1 || row.Timestamp.IsZero() || row.AgentID != c.ID || row.ParentID != c.ParentID {
			return nil, fmt.Errorf("submit.md row %d has invalid version/identity/timestamp", line)
		}
		if len(row.Schema) != 0 {
			if row.Result != nil || len(row.Payload) == 0 {
				return nil, fmt.Errorf("submit.md row %d has invalid custom payload", line)
			}
			storedSchema, schemaErr := CanonicalSubmitSchema(row.Schema)
			if schemaErr != nil {
				return nil, fmt.Errorf("submit.md row %d schema: %w", line, schemaErr)
			}
			// Validate the recorded contract, not just the caller's current one.
			last, err = PrepareSubmission(storedSchema, row.Payload)
			if err == nil && !bytes.Equal(storedSchema, configuredSchema) {
				err = errors.New("stored submit schema differs from configured contract")
			}
		} else {
			if row.Result == nil || len(configuredSchema) != 0 || len(row.Payload) != 0 {
				return nil, fmt.Errorf("submit.md row %d has invalid default result", line)
			}
			payload, _ := json.Marshal(map[string]string{"result": *row.Result})
			last, err = PrepareSubmission(nil, payload)
		}
		if err != nil {
			return nil, fmt.Errorf("submit.md row %d: %w", line, err)
		}
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	if line == 0 {
		return nil, errors.New("submit.md is empty")
	}
	return last, nil
}

// SaveSubmission appends exactly one versioned row by atomic replacement, using
// the Agent's rooted IO owner. It runs only after the final conversation save.
func SaveSubmission(sb Sandbox, c Config, payload json.RawMessage) error {
	schema, err := CanonicalSubmitSchema(c.SubmitSchema)
	if err != nil {
		return err
	}
	canonical, err := PrepareSubmission(schema, payload)
	if err != nil {
		return err
	}
	path := filepath.Join(c.DataDir, "submit.md")
	prior, err := sb.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err == nil {
		if _, err := parseSubmissions(prior, c); err != nil {
			return err
		}
	}
	row := submissionRow{Version: 1, Timestamp: time.Now().UTC(), AgentID: c.ID, ParentID: c.ParentID, Schema: schema}
	if len(schema) != 0 {
		row.Payload = canonical
	} else {
		var v struct {
			Result string `json:"result"`
		}
		_ = json.Unmarshal(canonical, &v)
		row.Result = &v.Result
	}
	raw, err := json.Marshal(row)
	if err != nil {
		return err
	}
	if len(prior) > 0 && prior[len(prior)-1] != '\n' {
		prior = append(prior, '\n')
	}
	raw = append(append(prior, raw...), '\n')
	f, err := sb.CreateTemp(c.DataDir, ".submit-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer sb.Remove(name)
	if _, err = f.Write(raw); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return sb.Rename(name, path)
}

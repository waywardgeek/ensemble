package history

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"ensemble/internal/common"
)

// Called only under mu: publication follows a durable journal append. A failed
// write rolls back the owned partial record so later appends cannot hide it.
func (h *history) appendJournal(e common.Event) error {
	if h.journal == "" {
		return nil
	}
	if err := h.parent.Sandbox().MkdirAll(filepath.Dir(h.journal), 0700); err != nil {
		return err
	}
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	f, err := h.parent.Sandbox().OpenFile(h.journal, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	offset, err := f.Seek(0, io.SeekEnd)
	if err == nil {
		_, err = f.Write(append(data, '\n'))
	}
	if err == nil {
		err = f.Sync()
	}
	if err != nil {
		err = errors.Join(err, f.Truncate(offset), f.Sync())
	}
	return errors.Join(err, f.Close())
}

func readJournal(path string, owners ...common.Sandbox) ([]common.Event, error) {
	readFile := os.ReadFile
	if len(owners) > 0 && owners[0] != nil {
		readFile = owners[0].ReadFile
	}
	data, err := readFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, []byte{'\n'})
	var events []common.Event
	var last common.Seq
	for i, line := range lines {
		if i == len(lines)-1 {
			break
		} // a torn final record was never admitted.
		var e common.Event
		dec := json.NewDecoder(bytes.NewReader(line))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&e); err != nil {
			return nil, fmt.Errorf("journal line %d: %w", i+1, err)
		}
		if dec.Decode(new(any)) != io.EOF || e.Seq == 0 || e.Seq <= last || e.Time.IsZero() {
			return nil, fmt.Errorf("invalid journal envelope at line %d", i+1)
		}
		if err := validate(e); err != nil {
			return nil, err
		}
		events = append(events, e)
		last = e.Seq
	}
	return events, nil
}
func replayJournal(events []common.Event, c *common.Context, through common.Seq) ([]common.Event, common.Seq, error) {
	var tail []common.Event
	for _, e := range events {
		if e.Seq <= through {
			continue
		}
		if e.Seq != through+1 {
			return nil, through, fmt.Errorf("journal gap after %d", through)
		}
		Apply(c, e)
		through = e.Seq
		tail = append(tail, e)
	}
	return tail, through, nil
}
func (h *history) trimJournal(boundary common.Seq) error {
	events, err := readJournal(h.journal, h.parent.Sandbox())
	if err != nil {
		return err
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	for _, e := range events {
		if e.Seq > boundary {
			if err := enc.Encode(e); err != nil {
				return err
			}
		}
	}
	return replaceFile(h.journal, out.Bytes(), h.parent.Sandbox())
}

func repairJournal(path string, events []common.Event, owners ...common.Sandbox) error {
	readFile := os.ReadFile
	if len(owners) > 0 && owners[0] != nil {
		readFile = owners[0].ReadFile
	}
	data, err := readFile(path)
	if errors.Is(err, os.ErrNotExist) || (err == nil && (len(data) == 0 || data[len(data)-1] == '\n')) {
		return nil
	}
	if err != nil {
		return err
	}
	var out bytes.Buffer
	for _, e := range events {
		if err := json.NewEncoder(&out).Encode(e); err != nil {
			return err
		}
	}
	return replaceFile(path, out.Bytes(), owners...)
}

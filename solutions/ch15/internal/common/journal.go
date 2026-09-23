package common

// The journal: events reach disk as they happen (Chapter 15 rule 9).
//
// The save file is a snapshot, written at a normal shutdown. Everything the
// agent records between snapshots is appended here, one JSON event per line,
// the moment it is recorded. A crash — SIGKILL, a panic, the laptop lid —
// therefore leaves a tail after the save file's as_of anchor, and recovery is
// Chapter 11's load unchanged: install the snapshot, apply the tail.
//
// The layout is this agent's choice, and the chapter leaves it to the
// student: the journal sits beside the save as <save>.journal, so the two can
// never describe different conversations. Each event is one write(2) on a
// file opened O_APPEND. That survives the process dying, because the bytes are
// in the kernel the moment write returns; it does not survive the machine
// losing power, which would need an fsync per event. That trade is written
// down rather than hidden.

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

// JournalPath is where the journal for a save file lives.
func JournalPath(savePath string) string { return savePath + ".journal" }

// Journal is the append side.
type Journal struct {
	f *os.File
}

// OpenJournal opens (creating if needed) the journal for appending.
func OpenJournal(path string) (*Journal, error) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("journal: %w", err)
	}
	return &Journal{f: f}, nil
}

// Append writes one event as one line, in one write.
func (j *Journal) Append(e Event) error {
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("journal: marshal event %d: %w", e.Seq, err)
	}
	b = append(b, '\n')
	if _, err := j.f.Write(b); err != nil {
		return fmt.Errorf("journal: write event %d: %w", e.Seq, err)
	}
	return nil
}

// Reset empties the journal. Call it only AFTER a snapshot that folds in
// every journaled event has been written: the snapshot first, then the
// truncation, never the other way round, or a crash between the two loses
// the tail.
func (j *Journal) Reset() error {
	if err := j.f.Truncate(0); err != nil {
		return fmt.Errorf("journal: truncate: %w", err)
	}
	return nil
}

func (j *Journal) Close() error { return j.f.Close() }

// ReadJournal returns the journaled events in file order. A missing journal
// is an empty one. A line that does not parse is skipped and reported to
// diag: the last line of a journal written by a process that died mid-write
// is the expected case, and one bad line must not cost the lines around it.
func ReadJournal(path string, diag func(error)) ([]Event, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("journal: %w", err)
	}
	var out []Event
	r := bufio.NewReader(bytes.NewReader(data))
	for n := 1; ; n++ {
		line, err := r.ReadBytes('\n')
		if len(bytes.TrimSpace(line)) > 0 {
			var e Event
			if uerr := json.Unmarshal(line, &e); uerr != nil {
				if diag != nil {
					diag(fmt.Errorf("journal %s line %d: skipped: %v", path, n, uerr))
				}
			} else {
				out = append(out, e)
			}
		}
		if err == io.EOF {
			return out, nil
		}
		if err != nil {
			return nil, fmt.Errorf("journal: %w", err)
		}
	}
}

// Recover is the load: the save file if there is one, plus every journaled
// event past it. It returns the SaveFile it assembled — the snapshot, its
// anchor, and a log extended by the journal's tail — ready for Restore and
// NextSeq.
//
// A save file that exists and does not parse is an error: ch11 rule 2 still
// holds, and the caller must refuse to start rather than overwrite it. No
// save file is an ordinary fresh start, and a journal with no save file is a
// session that crashed before its first snapshot: all of it is tail.
func Recover(savePath string, diag func(error)) (*SaveFile, error) {
	sf := &SaveFile{}
	if _, err := os.Stat(savePath); err == nil {
		loaded, err := Load(savePath)
		if err != nil {
			return nil, err
		}
		sf = loaded
	}
	tail, err := ReadJournal(JournalPath(savePath), diag)
	if err != nil {
		return nil, err
	}
	last := sf.NextSeq() - 1
	for _, e := range tail {
		// Events already in the snapshot's log are in the journal too when
		// the process died after writing the snapshot and before resetting
		// the journal. Seq says which is which.
		if e.Seq <= last {
			continue
		}
		sf.Log = append(sf.Log, e)
		last = e.Seq
	}
	return sf, nil
}

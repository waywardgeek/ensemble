package memory

// The memory store: the one component that touches the memory directory.
//
// Everything above it is pure. The reducer has no filesystem, no renderer
// dereferences anything, and the compaction policy decides in terms of bytes
// it was handed. This package is where that stops being true, and it is
// deliberately the only place.
//
// Its whole job is dumb: read files, hand back bytes. It performs no
// substitution, resolves no references, and makes no decisions about what
// should be compacted. It reads a directory and assembles events that carry
// what it found, so that replaying those events tomorrow says exactly what it
// said today even if the directory has changed underneath.
//
// Layout:
//
//	SOUL.md                              identity, read once at startup
//	MEMORY.md                            curated long-term memory
//	memory/2026-09-26-1.md               session memories, never retired
//	memory/bucket-0/2026-09-20-1_2026-09-26-3.md    8x, named by range
//	memory/bucket-1/2026-08-01-1_2026-09-19-2.md    64x, named by range
//
// A bucket file's name carries the range it folded, so "Sept 20's first
// memory through Sept 26's third" is readable with no log in hand at all.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

// File is one memory file on disk, with its bytes already read.
type File struct {
	ID   common.MemoryFileID // this file's own identity; ordering key
	Thru common.MemoryFileID // for a bucket file, the newest source it folded
	Path string
	Text string
}

// Store reads and writes the memory directory under one workspace root.
type Store struct {
	root string
	now  func() time.Time
}

func New(root string) *Store {
	return &Store{root: root, now: time.Now}
}

// SetClock exists so a test can produce predictable filenames. Production
// never calls it.
func (s *Store) SetClock(f func() time.Time) { s.now = f }

// dirFor returns the directory a band's files live in, and whether the band
// is a single file rather than a directory of them.
func (s *Store) dirFor(b common.Band) (path string, single bool) {
	switch b {
	case common.BandSoul:
		return filepath.Join(s.root, "SOUL.md"), true
	case common.BandMemory:
		return filepath.Join(s.root, "MEMORY.md"), true
	case common.Band64x:
		return filepath.Join(s.root, "memory", "bucket-1"), false
	case common.Band8x:
		return filepath.Join(s.root, "memory", "bucket-0"), false
	case common.BandSession:
		return filepath.Join(s.root, "memory"), false
	}
	return "", false
}

var (
	// 2026-09-26-1.md
	plainName = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(\d+)\.md$`)
	// 2026-09-20-1_2026-09-26-3.md
	rangeName = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2})-(\d+)_(\d{4}-\d{2}-\d{2})-(\d+)\.md$`)
)

// Files lists one band's LIVE memory, oldest first, with contents read.
//
// Live means "still in the band", which for session memories is not the same
// as "still on disk". Every uncompressed memory stays in one directory
// forever, because a later chapter searches that directory and a corpus with
// holes in it is worth much less. So a session memory leaves the band by
// being covered by a bucket file, not by being deleted, and this function is
// where that distinction is applied.
//
// Ordering is by (date, number) and comes from the FILENAMES, never from
// directory iteration order or modification time. That is what makes a
// restore reproducible: the same files always describe the same memory.
func (s *Store) Files(b common.Band) ([]File, error) {
	all, err := s.allFiles(b)
	if err != nil || b != common.BandSession {
		return all, err
	}

	// A session memory is in the session band exactly when no bucket file
	// has folded it. The bucket names carry their ranges, so this is
	// derivable from the directory alone, with no index to keep in step and
	// nothing to go stale. That is what makes the files on disk a complete
	// description of the memory that loads from them.
	folded, err := s.allFiles(common.Band8x)
	if err != nil {
		return nil, err
	}
	var live []File
	for _, f := range all {
		covered := false
		for _, b := range folded {
			if !f.ID.Before(b.ID) && !b.Thru.Before(f.ID) {
				covered = true
				break
			}
		}
		if !covered {
			live = append(live, f)
		}
	}
	return live, nil
}

// AllSessions returns every uncompressed memory ever written, folded or not.
//
// Nothing in this chapter calls it. It is here because it is the reason the
// session directory is append-only, and a reader who wonders why folded
// memories are kept should be able to find the answer in the code rather
// than only in a commit message.
func (s *Store) AllSessions() ([]File, error) {
	return s.allFiles(common.BandSession)
}

func (s *Store) allFiles(b common.Band) ([]File, error) {
	path, single := s.dirFor(b)
	if path == "" {
		return nil, fmt.Errorf("no directory for band %s", b)
	}
	if single {
		text, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			return nil, nil // a band with no file is simply empty
		}
		if err != nil {
			return nil, err
		}
		if len(strings.TrimSpace(string(text))) == 0 {
			// An empty file is not memory. Emitting a populate for it would
			// hand the reducer an event it must panic on.
			return nil, nil
		}
		return []File{{Path: path, Text: string(text)}}, nil
	}

	entries, err := os.ReadDir(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []File
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		f, ok := parseName(e.Name())
		if !ok {
			continue // not ours; leave it alone
		}
		f.Path = filepath.Join(path, e.Name())
		text, err := os.ReadFile(f.Path)
		if err != nil {
			return nil, err
		}
		if len(strings.TrimSpace(string(text))) == 0 {
			continue
		}
		f.Text = string(text)
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID.Before(out[j].ID) })
	return out, nil
}

func parseName(name string) (File, bool) {
	if m := rangeName.FindStringSubmatch(name); m != nil {
		n1, _ := strconv.Atoi(m[2])
		n2, _ := strconv.Atoi(m[4])
		return File{
			ID:   common.MemoryFileID{Date: m[1], Num: n1},
			Thru: common.MemoryFileID{Date: m[3], Num: n2},
		}, true
	}
	if m := plainName.FindStringSubmatch(name); m != nil {
		n, _ := strconv.Atoi(m[2])
		id := common.MemoryFileID{Date: m[1], Num: n}
		return File{ID: id, Thru: id}, true
	}
	return File{}, false
}

// PopulateEvents assembles one event per file in a band, each carrying that
// file's bytes.
//
// This is the whole reason the store exists. The events it returns are
// self-contained: replaying them needs no disk, so a context reconstructed
// next year is the context the model actually saw, not whatever the files
// happen to say by then.
func (s *Store) PopulateEvents(b common.Band, source string) ([]common.BandPopulatedData, error) {
	files, err := s.Files(b)
	if err != nil {
		return nil, err
	}
	out := make([]common.BandPopulatedData, 0, len(files))
	for _, f := range files {
		out = append(out, common.BandPopulatedData{
			Band:   b,
			File:   f.ID,
			Text:   f.Text,
			Thru:   f.Thru,
			Source: source,
		})
	}
	return out, nil
}

// NextSessionID picks the next same-day number in the session directory,
// continuing the numbering that is already there rather than inventing a
// second scheme over the same files.
func (s *Store) NextSessionID() (common.MemoryFileID, error) {
	// allFiles, not Files: numbering must not collide with a memory that
	// has been folded into a bucket. Those memories are still on disk and
	// still own their numbers, even though they have left the band.
	files, err := s.allFiles(common.BandSession)
	if err != nil {
		return common.MemoryFileID{}, err
	}
	today := s.now().Format("2006-01-02")
	max := 0
	for _, f := range files {
		if f.ID.Date == today && f.ID.Num > max {
			max = f.ID.Num
		}
	}
	return common.MemoryFileID{Date: today, Num: max + 1}, nil
}

// WriteSession stores one new session memory and returns it.
func (s *Store) WriteSession(text string) (File, error) {
	id, err := s.NextSessionID()
	if err != nil {
		return File{}, err
	}
	dir := filepath.Join(s.root, "memory")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return File{}, err
	}
	path := filepath.Join(dir, fmt.Sprintf("%s-%d.md", id.Date, id.Num))
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return File{}, err
	}
	return File{ID: id, Thru: id, Path: path, Text: text}, nil
}

// WriteBucket stores one compressed range file in an 8x or 64x band.
//
// The name carries both ends of the range, so the file says what it folded
// without anything else having to remember.
func (s *Store) WriteBucket(b common.Band, from, thru common.MemoryFileID, text string) (File, error) {
	dir, single := s.dirFor(b)
	if single || dir == "" {
		return File{}, fmt.Errorf("band %s is not a bucket", b)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return File{}, err
	}
	name := fmt.Sprintf("%s-%d_%s-%d.md", from.Date, from.Num, thru.Date, thru.Num)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return File{}, err
	}
	return File{ID: from, Thru: thru, Path: path, Text: text}, nil
}

// WriteCurated replaces MEMORY.md. The one band with a real writer: the
// curator edits it in place, where every other band only ever gains new
// immutable files.
func (s *Store) WriteCurated(text string) error {
	return os.WriteFile(filepath.Join(s.root, "MEMORY.md"), []byte(text), 0o644)
}

// Retire deletes files that have been folded upward.
//
// Bucket files are deleted rather than archived once their content lives in a
// coarser band, matching the cascade this one is a translation of. Session
// memories are NEVER retired: they are the uncompressed record, and the whole
// point of compressing upward is that the original survives on disk even
// after it has left the context.
func (s *Store) Retire(files []File) error {
	sessionDir := filepath.Join(s.root, "memory")
	for _, f := range files {
		if filepath.Dir(f.Path) == sessionDir {
			return fmt.Errorf("refusing to retire session memory %s: session files are permanent", f.Path)
		}
		if err := os.Remove(f.Path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

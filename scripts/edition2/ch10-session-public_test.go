package ch10_test

// Public-only fixtures derived from Chapter 10 and session-api.md. Runtime
// positives are required; compilation/predecessor absence earns no acceptance.
import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"example.com/ensemble"
)

func options(t *testing.T, name string) ensemble.SessionOptions {
	t.Helper()
	workspace := t.TempDir()
	base := "CH10 public base <&> with exact final newline.\n"
	return ensemble.SessionOptions{Config: ensemble.Config{
		Vendor: "openai", Model: "fixture-public", APIKey: "CH10_PRIVATE_CONFIG_ONLY",
		BaseURL: "http://127.0.0.1:1", Workspace: workspace,
		DataDir: filepath.Join(workspace, name), Builtins: []string{"read_file"},
	}, System: &base}
}
func application(t *testing.T) *ensemble.Ensemble {
	t.Helper()
	e := ensemble.New(io.Discard)
	t.Cleanup(func() { _ = e.Close() })
	return e
}
func opened(t *testing.T, e *ensemble.Ensemble, o ensemble.SessionOptions) *ensemble.Agent {
	t.Helper()
	a, err := e.OpenSession(o)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Close() })
	return a
}
func code(t *testing.T, err error, expected string) {
	t.Helper()
	var failure *ensemble.SessionError
	if !errors.As(err, &failure) || failure.Code != expected {
		t.Fatalf("wanted %s, got %v", expected, err)
	}
}
func state(t *testing.T, a *ensemble.Agent) ensemble.SessionState {
	t.Helper()
	s, err := a.Session()
	if err != nil || s == nil {
		t.Fatalf("session missing: %v", err)
	}
	return *s
}
func export(t *testing.T, a *ensemble.Agent) ensemble.CheckpointExport {
	t.Helper()
	x, err := a.ExportCheckpoint()
	if err != nil || len(x.Bytes) == 0 || x.AsOf == 0 {
		t.Fatalf("export failed: %v", err)
	}
	return x
}
func closed(t *testing.T, a *ensemble.Agent) {
	t.Helper()
	if err := a.Close(); err != nil {
		t.Fatal(err)
	}
}
func copies(t *testing.T, source, destination string) {
	t.Helper()
	if err := os.MkdirAll(destination, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"events.log", "checkpoint.json", "origin.json"} {
		b, err := os.ReadFile(filepath.Join(source, name))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(destination, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCh10PublicSelectorsAndConfig(t *testing.T) {
	e := application(t)
	o := options(t, "selected")
	_, err := e.NewAgent(o.Config)
	code(t, err, "session_conflict")
	bad := o
	bad.Config.LogPath = filepath.Join(o.Config.Workspace, "standalone.log")
	_, err = e.OpenSession(bad)
	code(t, err, "session_conflict")
	if _, err = os.Stat(o.Config.DataDir); !os.IsNotExist(err) {
		t.Fatal("selector refusal opened state")
	}
	a := opened(t, e, o)
	s := state(t, a)
	if s.Resumed || !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(s.ID) {
		t.Fatalf("invalid fresh identity: %+v", s)
	}
	c := a.Config()
	if c.DataDir != o.Config.DataDir || c.LogPath != filepath.Join(o.Config.DataDir, "events.log") || c.System != *o.System {
		t.Fatal("returned creation configuration differs")
	}
	if err = a.SetConfig(c); err != nil {
		t.Fatalf("unchanged configuration rejected: %v", err)
	}
	for _, which := range []string{"directory", "log", "base"} {
		changed := a.Config()
		switch which {
		case "directory":
			changed.DataDir += "-other"
		case "log":
			changed.LogPath += "-other"
		case "base":
			changed.System = "different base"
		}
		if err = a.SetConfig(changed); err == nil {
			t.Fatalf("changed %s accepted", which)
		}
		if !reflect.DeepEqual(a.Config(), c) {
			t.Fatalf("changed %s partially applied", which)
		}
	}
	plain := o.Config
	plain.DataDir = ""
	plain.LogPath = filepath.Join(o.Config.Workspace, "fresh.log")
	fresh, err := e.NewAgent(plain)
	if err != nil {
		t.Fatal(err)
	}
	defer fresh.Close()
	if s, err := fresh.Session(); err != nil || s != nil {
		t.Fatal("standalone acquired a session")
	}
	_, err = fresh.Checkpoint()
	code(t, err, "session_conflict")
	_, err = fresh.ExportCheckpoint()
	code(t, err, "session_conflict")
}

func TestCh10PublicOwnedCaptureAndResume(t *testing.T) {
	e := application(t)
	o := options(t, "owned")
	a := opened(t, e, o)
	if err := a.Ephemeral("pending one-request guidance"); err != nil {
		t.Fatal(err)
	}
	before := state(t, a)
	x := export(t, a)
	if !reflect.DeepEqual(before, state(t, a)) {
		t.Fatal("export claimed file commit")
	}
	inspection, err := e.InspectCheckpoint(x.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	want := inspection.Snapshot()
	for i := range x.Bytes {
		x.Bytes[i] = 'X'
	}
	x = export(t, a)
	again, err := e.InspectCheckpoint(x.Bytes)
	if err != nil || !reflect.DeepEqual(want, again.Snapshot()) {
		t.Fatal("export buffer aliases owner")
	}
	ack, err := a.Checkpoint()
	if err != nil || ack.AsOf != x.AsOf || ack.WatchRevision == 0 {
		t.Fatalf("wrong checkpoint acknowledgement: %+v %v", ack, err)
	}
	current := state(t, a)
	if current.CheckpointSeq == nil || *current.CheckpointSeq != ack.AsOf {
		t.Fatal("commit not reflected in public session")
	}
	borrowed, _ := a.Session()
	borrowed.ID = "caller-mutation"
	*borrowed.CheckpointSeq = 0
	if !reflect.DeepEqual(current, state(t, a)) {
		t.Fatal("session state aliases owner")
	}
	id := a.ID()
	closed(t, a)
	b := opened(t, e, o)
	resumed := state(t, b)
	if !resumed.Resumed || resumed.ID != before.ID || b.ID() == id {
		t.Fatal("durable/runtime identity conflated")
	}
	saved := export(t, b)
	restored, err := e.InspectCheckpoint(saved.Bytes)
	if err != nil || !reflect.DeepEqual(want, restored.Snapshot()) {
		t.Fatal("settled pending guidance did not resume exactly")
	}
	for _, name := range []string{"events.log", "checkpoint.json"} {
		data, err := os.ReadFile(filepath.Join(o.Config.DataDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Contains(data, []byte(o.Config.APIKey)) || bytes.Contains(data, []byte(o.Config.BaseURL)) {
			t.Fatal("transport configuration leaked into store")
		}
	}
}

func TestCh10PublicMountReservations(t *testing.T) {
	e := application(t)
	o := options(t, "original")
	a := opened(t, e, o)
	if _, err := a.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	_, err := e.OpenSession(o)
	code(t, err, "session_in_use")
	clone := o
	clone.Config.DataDir = filepath.Join(o.Config.Workspace, "copied")
	copies(t, o.Config.DataDir, clone.Config.DataDir)
	_, err = e.OpenSession(clone)
	code(t, err, "session_in_use")
	other := application(t)
	_, err = other.OpenSession(o)
	code(t, err, "session_in_use")
	_, err = other.InspectSession(o.Config.DataDir)
	code(t, err, "session_in_use")
	// The copied store itself is valid; after releasing the same-root identity
	// reservation its own independent lock permits a genuine successful mount.
	closed(t, a)
	b := opened(t, e, clone)
	if !state(t, b).Resumed {
		t.Fatal("copied existing store presented as fresh")
	}
}

func TestCh10PublicSnapshotOnlyOrigin(t *testing.T) {
	e := application(t)
	o := options(t, "source")
	a := opened(t, e, o)
	if err := a.Ephemeral("origin pending guidance"); err != nil {
		t.Fatal(err)
	}
	x := export(t, a)
	sessionID := state(t, a).ID
	closed(t, a)
	destination := o
	destination.Config.DataDir = filepath.Join(o.Config.Workspace, "imported")
	b, err := e.ImportSession(x.Bytes, destination)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	s := state(t, b)
	if s.ID != sessionID || s.Resumed {
		t.Fatal("import changed identity or first-mount flag")
	}
	origin, err := os.ReadFile(filepath.Join(destination.Config.DataDir, "origin.json"))
	if err != nil || !bytes.Equal(origin, x.Bytes) {
		t.Fatal("origin bytes not preserved")
	}
	events := b.Events()
	if len(events) != 1 || events[0].Type != "session_anchor" || events[0].Seq != x.AsOf+1 {
		t.Fatal("import invented prefix history")
	}
	dump, err := b.Dump()
	if err != nil || len(bytes.Split(bytes.TrimSpace(dump), []byte{'\n'})) != 2 {
		t.Fatal("raw dump contains invented prefix")
	}
	// A pre-origin request refusal needs a fixture with an actual captured send;
	// this no-HTTP group deliberately does not invent one at the ephemeral seq.
	if _, err = b.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	closed(t, b)
	c := opened(t, e, destination)
	if !state(t, c).Resumed || state(t, c).ID != sessionID {
		t.Fatal("imported origin did not survive later checkpoint")
	}
	again, _ := os.ReadFile(filepath.Join(destination.Config.DataDir, "origin.json"))
	if !bytes.Equal(origin, again) {
		t.Fatal("later checkpoint replaced origin")
	}
	closed(t, c)
	i, err := e.InspectSession(destination.Config.DataDir)
	if err != nil || i.Boundary().CompleteHistory || i.Boundary().OriginAsOf != x.AsOf || !i.Boundary().Settled {
		t.Fatalf("wrong imported boundary: %v", err)
	}
}

func TestCh10PublicPlainSystemPresence(t *testing.T) {
	e := application(t)
	o := options(t, "system")
	a := opened(t, e, o)
	closed(t, a)
	for _, replacement := range []string{"", "different"} {
		bad := o
		bad.System = &replacement
		_, err := e.OpenSession(bad)
		code(t, err, "session_incompatible")
	}
	omitted := o
	omitted.System = nil
	b := opened(t, e, omitted)
	if b.Config().System != *o.System {
		t.Fatal("omission did not adopt recorded base")
	}
	closed(t, b)
	c := opened(t, e, o)
	closed(t, c)
}

func TestCh10PublicInspectionUnregisteredAndOwned(t *testing.T) {
	e := application(t)
	o := options(t, "inspection")
	a := opened(t, e, o)
	if err := a.Ephemeral("owned inspection text"); err != nil {
		t.Fatal(err)
	}
	x := export(t, a)
	closed(t, a)
	i, err := e.InspectCheckpoint(x.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := json.Marshal(i.Snapshot())
	v := i.Snapshot()
	if len(v.Ephemera) == 0 || len(v.Ephemera[0].Parts) == 0 || v.Ephemera[0].Parts[0].Text == nil {
		t.Fatal("positive pending guidance absent")
	}
	*v.Ephemera[0].Parts[0].Text = "mutated"
	second, _ := json.Marshal(i.Snapshot())
	if !bytes.Equal(first, second) {
		t.Fatal("inspection snapshot shares text pointer")
	}
	// Inspection carries the SessionID but must not reserve a live identity.
	b := opened(t, e, o)
	if !strings.Contains(b.Config().System, "CH10 public base") {
		t.Fatal("inspection changed live identity")
	}
}

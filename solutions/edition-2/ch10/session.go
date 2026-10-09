package ensemble

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"example.com/ensemble/internal/common"
	"example.com/ensemble/internal/eventlog"
	"example.com/ensemble/internal/llm"
	"example.com/ensemble/internal/persistence"
	"example.com/ensemble/internal/policy"
	"example.com/ensemble/internal/skills"
)

type SessionOptions = common.SessionOptions
type SessionState = common.SessionState
type SessionError = common.SessionError
type CheckpointAck = common.CheckpointAck
type CheckpointExport = common.CheckpointExport
type SessionBoundary = common.SessionBoundary

func sessionError(code, detail string) error { return &SessionError{Code: code, Detail: detail} }
func (e *Ensemble) reserveSession(key string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.closed {
		return sessionError("session_conflict", "application closed")
	}
	if e.sessionReservations == nil {
		e.sessionReservations = map[string]bool{}
	}
	if e.sessionReservations[key] {
		return sessionError("session_in_use", "session identity is already mounted")
	}
	if strings.HasPrefix(key, "path:") {
		dir := strings.TrimPrefix(key, "path:")
		for _, leaf := range []string{"owner.lock", "events.log", "checkpoint.json", "origin.json"} {
			if e.settingsPaths[filepath.Join(dir, leaf)] {
				return sessionError("session_conflict", "store leaf already belongs to settings")
			}
		}
	}
	e.sessionReservations[key] = true
	return nil
}
func (e *Ensemble) ReleaseSession(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.sessionReservations, key)
}
func (a *Agent) sessionState() *SessionState {
	if a.store == nil {
		return nil
	}
	return &SessionState{ID: a.sessionID, Resumed: a.resumed, CheckpointSeq: a.store.CheckpointSequence()}
}
func (a turnAgent) SessionState() *SessionState { return a.sessionState() }
func (a *Agent) Session() (*SessionState, error) {
	if a.actor != nil {
		return a.actor.Session()
	}
	return a.sessionState(), nil
}
func (a *Agent) Checkpoint() (CheckpointAck, error) {
	return a.CheckpointContext(context.Background())
}
func (a *Agent) CheckpointContext(ctx context.Context) (CheckpointAck, error) {
	if a.actor == nil {
		return CheckpointAck{}, sessionError("session_conflict", "Agent is not a mounted session")
	}
	r, err := a.actor.CheckpointContext(ctx, true)
	return r.CheckpointAck, err
}
func (e *Ensemble) CheckpointContext(ctx context.Context, id string) (CheckpointAck, error) {
	a, err := e.Agent(id)
	if err != nil {
		return CheckpointAck{}, err
	}
	return a.CheckpointContext(ctx)
}
func (e *Ensemble) Checkpoint(id string) (CheckpointAck, error) {
	a, err := e.Agent(id)
	if err != nil {
		return CheckpointAck{}, err
	}
	return a.Checkpoint()
}
func (a *Agent) ExportCheckpoint() (CheckpointExport, error) {
	if a.actor == nil {
		return CheckpointExport{}, sessionError("session_conflict", "Agent is not a mounted session")
	}
	r, err := a.actor.Checkpoint(false)
	return r.CheckpointExport, err
}
func (a turnAgent) BeginCheckpoint(cp common.Checkpoint, save bool) (<-chan common.CheckpointResult, error) {
	return a.store.Begin(cp, save)
}
func (a turnAgent) ApplyCheckpoint(r common.CheckpointResult) { a.store.Applied(r) }
func (a turnAgent) CaptureSession(cursor uint64) (common.Checkpoint, error) {
	return a.captureSession(cursor)
}
func (a *Agent) watermarks(cursor uint64) common.Watermarks {
	w := common.Watermarks{Event: a.context.LastSeq, Request: cursor}
	if w.Request < a.context.RequestCursor {
		w.Request = a.context.RequestCursor
	}
	if a.skills != nil {
		w.Activation = a.skills.LastActivation()
	}
	for h := range a.context.Jobs {
		if h > w.Job {
			w.Job = h
		}
	}
	return w
}
func (a *Agent) captureSession(cursor uint64) (common.Checkpoint, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.context.Session == nil {
		return common.Checkpoint{}, sessionError("session_conflict", "no session identity")
	}
	state, err := llm.Clone(a.engine, a.context)
	if err != nil {
		return common.Checkpoint{}, err
	}
	w := a.watermarks(cursor)
	state.RequestCursor = w.Request
	identity := *state.Session.Identity
	cp := common.Checkpoint{Version: 1, StateVersion: 1, SessionID: a.sessionID, Identity: identity, AsOf: state.LastSeq, HighWatermarks: w}
	snap := &common.SemanticState{Session: common.SnapshotSession{ID: a.sessionID, Identity: identity, AsOf: cp.AsOf, HighWatermarks: w}, Context: state, Usage: []common.UsageAccount{}, Limits: a.jobs.PendingLimits(), Window: common.SnapshotWindow{Events: []common.WindowEvent{}, RenderableCount: a.renderableCount, EventCount: a.eventCount}}
	if a.skills != nil {
		snap.Skills = a.skills.Snapshot()
	}
	for from, usage := range a.engine.UsageByModel() {
		snap.Usage = append(snap.Usage, common.UsageAccount{From: from, Usage: usage})
	}
	sort.Slice(snap.Usage, func(i, j int) bool {
		x, y := snap.Usage[i].From, snap.Usage[j].From
		return x.Vendor+"/"+x.Model+"/"+x.Surface < y.Vendor+"/"+y.Model+"/"+y.Surface
	})
	// The material dictionary alone owns each body in the exported value.
	snap.Context.SkillPrimary = ""
	strip := func(entries []common.Entry) {
		for i := range entries {
			if entries[i].Purpose == "skill" {
				entries[i].Parts = []common.Part{}
			}
		}
	}
	strip(snap.Context.Entries)
	strip(snap.Context.DeferredSkills)
	strip(snap.Context.PendingSkills)
	for _, e := range a.recent {
		owned, err := llm.Clone(a.engine, e)
		if err != nil {
			return cp, err
		}
		item := common.WindowEvent{Event: owned, Activations: []uint64{}}
		if owned.Skills != nil {
			for _, m := range owned.Skills.Activated {
				item.Activations = append(item.Activations, m.Activation)
			}
			owned.Skills.Activated = []common.SkillActivation{}
		}
		snap.Window.Events = append(snap.Window.Events, item)
	}
	cp.State = snap
	return cp, nil
}
func (a *Agent) installedIdentity() (common.SessionIdentity, error) {
	identity := common.SessionIdentity{Mode: "plain", Handlers: []common.HandlerIdentity{}}
	base := a.config.System
	identity.System = &base
	for _, d := range a.registry.Installed() {
		identity.Handlers = append(identity.Handlers, common.HandlerIdentity{Name: d.Name, Description: d.Description, Schema: d.Schema})
	}
	if a.config.Skills != nil {
		identity.Mode = "skills"
		identity.System = nil
		s, err := a.skills.Identity()
		if err != nil {
			return identity, err
		}
		identity.Skills = &s
	}
	return identity, a.codec.Identity(identity)
}
func (a *Agent) installState(cp common.Checkpoint, live bool) error {
	bad := func(detail string) error { return sessionError("session_corrupt", detail) }
	if cp.State == nil {
		return bad("semantic state required")
	}
	snap, err := llm.Clone(a.engine, *cp.State)
	if err != nil {
		return err
	}
	c := snap.Context
	if c.Session == nil || c.Session.Identity == nil || c.Session.SessionID != cp.SessionID || c.LastSeq != cp.AsOf || c.RequestCursor != cp.HighWatermarks.Request {
		return bad("semantic boundary mismatch")
	}
	a.sessionID = cp.SessionID
	a.identity = cp.Identity
	if !sameJSON(a.codec, *c.Session.Identity, cp.Identity) {
		return bad("semantic creation identity mismatch")
	}
	if cp.Identity.Mode == "skills" {
		if snap.Skills == nil || !c.SkillMode || c.SkillPrimary != "" {
			return bad("missing skill state or duplicated primary")
		}
		if a.skills == nil {
			a.skills = skills.Historical(skillAgent{a})
		}
		if err = a.skills.Restore(snap.Skills, live); err != nil {
			return err
		}
		if snap.Skills.State.Primary != cp.Identity.Skills.Primary || len(snap.Skills.Ceiling) != len(cp.Identity.Handlers) {
			return bad("skill identity ceiling mismatch")
		}
		for i, n := range snap.Skills.Ceiling {
			if n != cp.Identity.Handlers[i].Name {
				return bad("skill ceiling differs from installed identity")
			}
		}
		for _, tr := range snap.Skills.Transitions {
			if tr.Seq > cp.AsOf {
				return bad("skill transition exceeds boundary")
			}
		}
		records := map[uint64]common.SkillActivation{}
		materialSeq := map[uint64]uint64{}
		seenMaterial := map[uint64]bool{}
		for _, m := range snap.Skills.Material {
			records[m.Record.Activation] = m.Record
			materialSeq[m.Record.Activation] = m.EventSeq
			if m.Record.Type == "primary" {
				c.SkillPrimary = m.Record.Body
			}
		}
		restore := func(entries []common.Entry) error {
			for i := range entries {
				e := &entries[i]
				if e.Purpose != "skill" {
					continue
				}
				m, ok := records[e.Activation]
				if !ok || m.Name != e.SkillName || m.Type == "primary" || len(e.Parts) != 0 || e.Seq != materialSeq[e.Activation] || seenMaterial[e.Activation] {
					return bad("invalid material reference")
				}
				seenMaterial[e.Activation] = true
				e.Parts = []common.Part{}
				if m.Body != "" {
					e.Parts = append(e.Parts, llm.Text(fmt.Sprintf("[skill %s activation %d]\n%s\n[/skill]", m.Name, m.Activation, m.Body)))
				}
			}
			return nil
		}
		for _, entries := range [][]common.Entry{c.Entries, c.DeferredSkills, c.PendingSkills} {
			if err = restore(entries); err != nil {
				return err
			}
		}
		for id, m := range records {
			if m.Type != "primary" && !seenMaterial[id] {
				return bad("missing retained material entry")
			}
		}
		for i := range snap.Window.Events {
			item := &snap.Window.Events[i]
			if item.Event.Skills != nil {
				if len(item.Event.Skills.Activated) != 0 {
					return bad("duplicated window material")
				}
				for _, id := range item.Activations {
					m, ok := records[id]
					if !ok {
						return bad("window material reference missing")
					}
					item.Event.Skills.Activated = append(item.Event.Skills.Activated, m)
				}
				found := false
				for _, tr := range snap.Skills.Transitions {
					if tr.Seq == item.Event.Seq {
						t := item.Event.Skills
						found = tr.Action == t.Action && tr.Name == t.Name && reflect.DeepEqual(tr.State, t.State) && reflect.DeepEqual(tr.Activated, item.Activations)
					}
				}
				if !found {
					return bad("watch skill transition witness mismatch")
				}
			} else if len(item.Activations) != 0 {
				return bad("misplaced material reference")
			}
		}
	} else if snap.Skills != nil || c.SkillMode || c.SkillPrimary != "" {
		return bad("plain session contains skill state")
	}
	llm.ReindexContext(a.engine, &c)
	if err = llm.ValidateSemantic(a.engine, c); err != nil {
		return err
	}
	if err = llm.ValidateWindow(a.engine, c, snap.Window.Events); err != nil {
		return err
	}
	if err = a.jobs.RestoreLimits(c, snap.Limits); err != nil {
		return err
	}
	if err = a.engine.RestoreUsage(snap.Usage, c.Responses); err != nil {
		return err
	}
	if snap.Window.EventCount == 0 || snap.Window.EventCount > 1000000 || snap.Window.RenderableCount > snap.Window.EventCount || len(snap.Window.Events) > 100 || uint64(len(snap.Window.Events)) != min(snap.Window.RenderableCount, 100) {
		return bad("invalid watch/count boundary")
	}
	var previous uint64
	a.recent = nil
	for _, item := range snap.Window.Events {
		if item.Event.Seq <= previous || item.Event.Seq > cp.AsOf || !llm.RenderableEvent(a.engine, item.Event.Type) {
			return bad("invalid saved watch window")
		}
		previous = item.Event.Seq
		a.recent = append(a.recent, item.Event)
	}
	a.context = c
	a.eventCount = snap.Window.EventCount
	a.renderableCount = snap.Window.RenderableCount
	w := a.watermarks(c.RequestCursor)
	if w.Event != cp.HighWatermarks.Event || w.Request != cp.HighWatermarks.Request || w.Activation != cp.HighWatermarks.Activation || w.Job != cp.HighWatermarks.Job {
		return bad("semantic watermarks disagree with durable maxima")
	}
	return nil
}
func sameJSON(codec common.SessionCodec, a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return codec.EqualJSON(x, y)
}
func (e *Ensemble) OpenSession(options SessionOptions) (*Agent, error) {
	return e.openSession(nil, options)
}
func (e *Ensemble) ImportSession(raw []byte, options SessionOptions) (*Agent, error) {
	if raw == nil {
		return nil, sessionError("session_corrupt", "import requires complete checkpoint")
	}
	return e.openSession(append([]byte{}, raw...), options)
}
func (e *Ensemble) openSession(imported []byte, options SessionOptions) (_ *Agent, err error) {
	config := options.Config
	if config.DataDir == "" || config.LogPath != "" || config.System != "" {
		return nil, sessionError("session_conflict", "session requires DataDir, no LogPath, and System through SessionOptions")
	}
	if options.System != nil {
		config.System = *options.System
	}
	if config.Skills != nil && config.System != "" {
		return nil, sessionError("session_incompatible", "System conflicts with primary skill")
	}
	a, err := e.construct(config)
	if err != nil {
		return nil, err
	}
	good := false
	defer func() {
		if !good {
			a.discardSession()
		}
	}()
	directory, err := persistence.Resolve(a, config.DataDir)
	if err != nil {
		return nil, err
	}
	if err = e.reserveSession("path:" + directory); err != nil {
		return nil, err
	}
	a.reservedPath = directory
	a.config.DataDir = directory
	a.config.LogPath = filepath.Join(directory, "events.log")
	a.store, err = persistence.Open(a, directory, false)
	if err != nil {
		return nil, err
	}
	if err = llm.ValidateConfig(a.engine, a.config, true); err != nil {
		return nil, err
	}
	if !a.registry.Match(a.config.Tools) {
		return nil, sessionError("session_incompatible", "installed handlers differ")
	}
	a.config.Tools = a.registry.Declarations()
	if a.config.Skills != nil {
		a.skills, err = skills.New(skillAgent{a})
		if err != nil {
			return nil, sessionError("session_incompatible", "current skill catalog is invalid")
		}
	}
	// Read current policy before any session append, without a write or Actor.
	a.policy, err = policy.New(a, a.config.PolicyPath)
	if err != nil {
		return nil, err
	}
	names, err := a.store.Entries()
	if err != nil {
		return nil, err
	}
	hasLog := false
	for _, n := range names {
		if n == "events.log" {
			hasLog = true
		}
	}
	if imported != nil {
		for _, n := range names {
			if n != "owner.lock" {
				return nil, sessionError("session_conflict", "import destination is occupied")
			}
		}
		cp, err := a.codec.Decode(imported)
		if err != nil {
			return nil, err
		}
		if cp.State == nil {
			return nil, sessionError("session_corrupt", "import requires non-null semantic state")
		}
		if err = a.compatibility(cp.Identity, options.System); err != nil {
			return nil, err
		}
		if err = a.installState(cp, true); err != nil {
			return nil, err
		}
		if !llm.Settled(a.engine, a.context) {
			return nil, sessionError("session_unfinished", "import boundary is unfinished")
		}
		if cp.AsOf == ^uint64(0) {
			return nil, sessionError("session_limit", "anchor sequence exhausted")
		}
		if cp.State.Window.EventCount >= 1000000 {
			return nil, sessionError("session_limit", "import anchor exceeds total event bound")
		}
		if err = e.reserveSession("id:" + a.sessionID); err != nil {
			return nil, err
		}
		a.reservedID = true
		a.origin = cp.State
		if err = a.store.WriteOrigin(imported); err != nil {
			return nil, err
		}
		a.log, err = eventlog.New(a, a.config.LogPath)
		if err != nil {
			return nil, err
		}
		anchor := common.SessionFact{SessionID: a.sessionID, OriginAsOf: cp.AsOf, OriginSHA256: fmt.Sprintf("%x", sha256.Sum256(imported)), HighWatermarks: &cp.HighWatermarks}
		if err = a.append(Event{Type: "session_anchor", Session: &anchor}, true, false); err != nil {
			return nil, err
		}
	} else if hasLog {
		if err = a.loadStore(options.System, true); err != nil {
			return nil, err
		}
		a.resumed = true
		if !llm.Settled(a.engine, a.context) {
			return nil, sessionError("session_unfinished", "conversation has unfinished accepted work")
		}
		if err = e.reserveSession("id:" + a.sessionID); err != nil {
			return nil, err
		}
		a.reservedID = true
		a.log, err = eventlog.Resume(a, a.config.LogPath, a.eventCount)
		if err != nil {
			return nil, sessionError("session_io", "cannot reopen validated append log")
		}
	} else {
		for _, n := range names {
			if n != "owner.lock" {
				return nil, sessionError("session_corrupt", "incomplete or ambiguous session store")
			}
		}
		identity, err := a.installedIdentity()
		if err != nil {
			return nil, err
		}
		a.identity = identity
		var id [16]byte
		if _, err = rand.Read(id[:]); err != nil {
			return nil, sessionError("session_io", "cannot allocate session identity")
		}
		a.sessionID = fmt.Sprintf("%x", id)
		if err = e.reserveSession("id:" + a.sessionID); err != nil {
			return nil, err
		}
		a.reservedID = true
		var initial common.SkillCandidate
		if a.skills != nil {
			initial, err = a.skills.Prepare(common.SkillOperation{Action: "initialize", Name: a.config.Skills.Primary})
			if err != nil {
				return nil, err
			}
		}
		a.log, err = eventlog.New(a, a.config.LogPath)
		if err != nil {
			return nil, err
		}
		if err = a.append(Event{Type: "session_initialized", Session: &common.SessionFact{SessionID: a.sessionID, Identity: &a.identity}}, true, false); err != nil {
			return nil, err
		}
		if initial != nil {
			t := initial.Transition()
			if err = a.appendPrepared(Event{Type: "skills_initialized", Skills: &t}, true, false, nil, initial); err != nil {
				return nil, err
			}
		}
	}
	if a.identity.Mode == "skills" && a.skills.LastActivation() == 0 {
		return nil, sessionError("session_unfinished", "skill initialization incomplete")
	}
	// The shared allocator changes only after all candidate validation succeeded.
	floor := a.watermarks(a.context.RequestCursor).Job
	e.mu.Lock()
	if e.closed {
		e.mu.Unlock()
		return nil, sessionError("session_conflict", "application closed")
	}
	if floor > e.nextHandle {
		e.nextHandle = floor
	}
	e.nextAgent++
	a.id = fmt.Sprintf("agent-%d", e.nextAgent)
	// Startup performs no I/O and all validation is complete. Publish only
	// after the actor pointer is installed so readers cannot see a partial Agent.
	a.actor = llm.NewActor(turnAgent{a})
	e.agents[a.id] = a
	e.mu.Unlock()
	good = true
	return a, nil
}
func (a *Agent) compatibility(identity common.SessionIdentity, system *string) error {
	if identity.Mode == "plain" {
		if a.config.Skills != nil || identity.System == nil || system != nil && *system != *identity.System {
			return sessionError("session_incompatible", "base System or skill mode differs")
		}
		a.config.System = *identity.System
	} else if a.config.Skills == nil {
		return sessionError("session_incompatible", "current skill configuration is required")
	}
	current, err := a.installedIdentity()
	if err != nil {
		return err
	}
	if !sameJSON(a.codec, current, identity) {
		return sessionError("session_incompatible", "creation identity differs from current selection")
	}
	return nil
}
func (a *Agent) discardSession() {
	if a.log != nil {
		_ = a.log.Close()
		a.log = nil
	}
	if a.policy != nil {
		a.policy.Close()
	}
	a.engine.Close()
	if a.store != nil {
		_ = a.store.Close()
	}
	if a.reservedID {
		a.parent.ReleaseSession("id:" + a.sessionID)
		a.reservedID = false
	}
	if a.reservedPath != "" {
		a.parent.ReleaseSession("path:" + a.reservedPath)
		a.reservedPath = ""
	}
}

// sessionReader is an Agent-owned validation read; no observer or actor is live.
type sessionReader struct {
	common.SessionReadAgent
	state common.SessionReadState
}

func (r *sessionReader) AcceptReadEvent(event common.Event, line int) error {
	return r.AcceptSessionRecord(event, line, &r.state)
}
func (a *Agent) AcceptSessionRecord(event common.Event, line int, r *common.SessionReadState) error {
	bad := func(detail string) error {
		return sessionError("session_corrupt", fmt.Sprintf("record %d: %s", line, detail))
	}
	if !r.First {
		r.First = true
		if a.origin != nil {
			if event.Type != "session_anchor" || event.Session == nil || event.Session.OriginSHA256 != r.OriginHash || event.Session.HighWatermarks == nil || *event.Session.HighWatermarks != a.origin.Session.HighWatermarks {
				return bad("origin anchor mismatch")
			}
		} else {
			if event.Type == "session_anchor" {
				return sessionError("session_origin_required", "anchor log requires its immutable origin")
			}
			if event.Type != "session_initialized" || event.Session == nil || event.Session.Identity == nil {
				return bad("missing session initializer")
			}
			if err := a.codec.Identity(*event.Session.Identity); err != nil {
				return err
			}
			if r.Live {
				if err := a.compatibility(*event.Session.Identity, r.System); err != nil {
					return err
				}
			}
			a.sessionID = event.Session.SessionID
			a.identity = *event.Session.Identity
		}
	} else if event.Type == "session_initialized" || event.Type == "session_anchor" {
		return bad("interior session construction fact")
	}
	if err := a.append(event, false, false); err != nil {
		return bad("invalid event transition")
	}
	if r.Checkpoint != nil && event.Seq == r.Checkpoint.AsOf {
		cp := r.Checkpoint
		reduced, err := a.captureSession(cp.HighWatermarks.Request)
		if err != nil {
			return err
		}
		if reduced.HighWatermarks.Event != cp.HighWatermarks.Event || reduced.HighWatermarks.Request != cp.HighWatermarks.Request || reduced.HighWatermarks.Activation != cp.HighWatermarks.Activation || reduced.HighWatermarks.Job != cp.HighWatermarks.Job {
			return bad("checkpoint watermarks mismatch")
		}
		if cp.State != nil {
			// Compare strict wire values: raw JSON is stored as strings, so canonical
			// semantic numbers never erase its exact recorded replay spelling.
			x, err := a.codec.Encode(reduced)
			if err != nil {
				return err
			}
			y, err := a.codec.Encode(*cp)
			if err != nil {
				return err
			}
			if !a.codec.EqualJSON(x, y) {
				return bad("checkpoint differs from reduced prefix")
			}
			if err = a.installState(*cp, r.Live); err != nil {
				return err
			}
		} else {
			a.context.RequestCursor = cp.HighWatermarks.Request
		}
		r.Checkpoint = nil
	}
	return nil
}
func (a *Agent) loadStore(system *string, live bool) error {
	origin, hasOrigin, err := a.store.Read("origin.json")
	if err != nil {
		return err
	}
	raw, hasCheckpoint, err := a.store.Read("checkpoint.json")
	if err != nil {
		return err
	}
	reader := &sessionReader{SessionReadAgent: a, state: common.SessionReadState{System: system, Live: live}}
	if hasOrigin {
		cp, err := a.codec.Decode(origin)
		if err != nil {
			return err
		}
		if cp.State == nil {
			return sessionError("session_corrupt", "origin must contain state")
		}
		if live {
			if err = a.compatibility(cp.Identity, system); err != nil {
				return err
			}
		}
		if err = a.installState(cp, live); err != nil {
			return err
		}
		a.origin = cp.State
		reader.state.OriginHash = fmt.Sprintf("%x", sha256.Sum256(origin))
	}
	if hasCheckpoint {
		cp, err := a.codec.Decode(raw)
		if err != nil {
			return err
		}
		reader.state.Checkpoint = &cp
		a.store.LoadedCheckpoint(cp.AsOf)
		if hasOrigin && cp.AsOf <= a.context.LastSeq {
			return sessionError("session_corrupt", "checkpoint precedes imported anchor")
		}
	}
	f, err := os.Open(a.config.LogPath)
	if err != nil {
		if os.IsNotExist(err) {
			return sessionError("session_corrupt", "session event log is missing")
		}
		return sessionError("session_io", "cannot read event log")
	}
	defer f.Close()
	if err = eventlog.Stream(reader, f, true); err != nil {
		var problem *common.SessionError
		if errors.As(err, &problem) {
			return err
		}
		return sessionError("session_corrupt", err.Error())
	}
	if !reader.state.First || reader.state.Checkpoint != nil {
		return sessionError("session_corrupt", "checkpoint boundary absent from log")
	}
	if live && a.identity.Mode == "skills" {
		snapshot := a.skills.Snapshot()
		if snapshot == nil {
			return sessionError("session_unfinished", "skill initializer missing")
		}
		if err = a.skills.Restore(snapshot, true); err != nil {
			return err
		}
	}
	return nil
}

// SessionInspection owns an inert, unregistered projection under its real root.
type SessionInspection struct{ agent *Agent }

func (s *SessionInspection) Session() SessionState {
	a := s.agent
	var seq *uint64
	if a.inspectedCheckpoint != nil {
		value := *a.inspectedCheckpoint
		seq = &value
	}
	return SessionState{ID: a.sessionID, Resumed: false, CheckpointSeq: seq}
}
func (s *SessionInspection) Boundary() SessionBoundary {
	a := s.agent
	origin := uint64(0)
	if a.origin != nil {
		origin = a.origin.Session.AsOf
	}
	return SessionBoundary{LogSeq: a.context.LastSeq, OriginAsOf: origin, CompleteHistory: origin == 0, Settled: llm.Settled(a.engine, a.context)}
}
func (s *SessionInspection) Snapshot() Context                  { return s.agent.Snapshot() }
func (s *SessionInspection) Events() []Event                    { return s.agent.Events() }
func (s *SessionInspection) UsageByModel() map[Provenance]Usage { return s.agent.UsageByModel() }
func (s *SessionInspection) InspectSkills() SkillInspection {
	v, _ := s.agent.InspectSkills()
	return v
}
func (s *SessionInspection) Render(c Config) ([]byte, error) { return s.agent.Render(c) }
func (s *SessionInspection) ReconstructRequest(seq uint64) ([]byte, error) {
	return s.agent.ReconstructRequest(seq)
}
func (e *Ensemble) InspectSession(directory string) (*SessionInspection, error) {
	a, err := e.construct(Config{})
	if err != nil {
		return nil, err
	}
	defer a.engine.Close()
	a.skills = skills.Historical(skillAgent{a})
	path, err := persistence.Resolve(a, directory)
	if err != nil {
		return nil, err
	}
	a.config.DataDir = path
	a.config.LogPath = filepath.Join(path, "events.log")
	a.store, err = persistence.Open(a, path, true)
	if err != nil {
		return nil, err
	}
	defer a.store.Close()
	if err = a.loadStore(nil, false); err != nil {
		return nil, err
	}
	a.inspectedCheckpoint = a.store.CheckpointSequence()
	a.store = nil
	return &SessionInspection{agent: a}, nil
}
func (e *Ensemble) InspectCheckpoint(raw []byte) (*SessionInspection, error) {
	a, err := e.construct(Config{})
	if err != nil {
		return nil, err
	}
	defer a.engine.Close()
	cp, err := a.codec.Decode(raw)
	if err != nil {
		return nil, err
	}
	if cp.State == nil {
		return nil, sessionError("session_origin_required", "null checkpoint needs history")
	}
	if err = a.installState(cp, false); err != nil {
		return nil, err
	}
	a.origin = cp.State
	seq := cp.AsOf
	a.inspectedCheckpoint = &seq
	return &SessionInspection{agent: a}, nil
}

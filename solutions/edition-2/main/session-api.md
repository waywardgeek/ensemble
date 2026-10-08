# Public session API

Published signatures for implementation; see persistence-format.md for version 1.
All types are public aliases to common declarations. Returned data owns buffers.

```go
type SessionOptions struct {
    Config Config // DataDir required; LogPath forbidden; System must be empty
    System *string // nil omitted; explicit empty differs from omission
}
type SessionState struct {
    ID string `json:"id"`
    Resumed bool `json:"resumed"`
    CheckpointSeq *uint64 `json:"checkpoint_seq"`
}
type CheckpointAck struct {
    AsOf uint64 `json:"as_of"`
    WatchRevision uint64 `json:"watch_revision"`
}
type CheckpointExport struct { AsOf uint64; Bytes []byte }
type SessionError struct { Code, Detail string }
func (*SessionError) Error() string
func (e *Ensemble) OpenSession(SessionOptions) (*Agent, error)
func (e *Ensemble) ImportSession([]byte, SessionOptions) (*Agent, error)
func (e *Ensemble) InspectSession(string) (*SessionInspection, error)
func (e *Ensemble) InspectCheckpoint([]byte) (*SessionInspection, error)
func (a *Agent) Session() (*SessionState, error)
func (a *Agent) ExportCheckpoint() (CheckpointExport, error)
func (a *Agent) Checkpoint() (CheckpointAck, error)
func (e *Ensemble) Checkpoint(agentID string) (CheckpointAck, error)
func (s *SessionInspection) Session() SessionState
func (s *SessionInspection) Boundary() SessionBoundary
func (s *SessionInspection) Snapshot() Context
func (s *SessionInspection) Events() []Event
func (s *SessionInspection) UsageByModel() map[Provenance]Usage
func (s *SessionInspection) InspectSkills() SkillInspection
func (s *SessionInspection) Render(Config) ([]byte, error)
func (s *SessionInspection) ReconstructRequest(uint64) ([]byte, error)
type SessionBoundary struct {
    LogSeq uint64 `json:"log_seq"`
    OriginAsOf uint64 `json:"origin_as_of"`
    CompleteHistory bool `json:"complete_history"`
    Settled bool `json:"settled"`
}
```

Config gains creation-only DataDir. Options.System is sole session-construction
presence input; Config.System must be empty to avoid two competing selectors.
Fresh constructor retains original Config semantics and refuses DataDir.
Session() returns null for standalone; export/checkpoint refuse session_conflict.
Inspection is inert/unregistered, releases shared nonblocking lock before return,
needs no credentials/catalog and cannot admit tools. No caller-close obligation
for inspections. Invalid state fails all-or-nothing; unfinished valid history is
inspectable with Settled=false. File-backed current policy remains independent.

Open reserves canonical store and SessionID, then mounts only validated settled
state; first creation/import has Resumed=false. Import preserves SessionID and
requires non-null state/unoccupied destination. All source files remain unchanged
on validation refusal. Export and save share one gate; busy is immediate. Export
returns anchored exact versioned bytes without claiming checkpoint.json commit.
Save returns only after replace and actor application/session_changed publication.
Caller/connection disappearance does not roll back a write. Close joins work,
writes final settled state, releases lock and preserves error on repeated close.

Stable codes: session_in_use, session_unsupported, session_conflict,
session_corrupt, session_incompatible, session_unfinished,
session_origin_required, session_busy, session_io, session_limit,
history_unavailable. Safe details never echo content/credentials. Config() reports
actual events.log and DataDir; unchanged SetConfig round trips remain valid.
Browser state/command shapes are exactly Chapter 10 §10.9, including job_access
and delivery-before-command_ack with connection-bound cancellation.

Connection-lifetime checkpoint waiting also has these public operations:

```go
func (a *Agent) CheckpointContext(context.Context) (CheckpointAck, error)
func (e *Ensemble) CheckpointContext(context.Context, string) (CheckpointAck, error)
```

Cancellation stops only the caller's wait after admission; Actor/SessionStore
finish the accepted save and publish its result. It never rolls back a commit.
The optional GUI uses the connection context for this wait and its watch-delivery
barrier, so disconnect or overflow releases both without parking the Actor.

### Exact provider argument strings

`Part.ArgumentsText *string` (`arguments_text,omitempty`) optionally retains the
decoded Chat Completions function.arguments string separately from `Args` object
JSON. Engine checks correspondence and uses it for Chat replay, including changed
current model aliases. It is not target-bound opaque material. Existing records
without the field use their accepted Args bytes. See persistence-format.md for
strict optional semantic encoding and duplicate-argument comparison.

Public Context snapshots normalize absent structural collections to empty
collections in the owned copy, so full-log and checkpoint/tail inspection have
identical Go representations. Raw JSON nil/bytes remain exact. Events/Dump and
stored history retain their original accepted representation.

package common

// Band orders memory from stable identity to recent completed work.
type Band uint8

const (
	// BandSoul never graduates; humans maintain identity on disk.
	BandSoul Band = iota + 1
	// BandMemory is the terminal curated-memory band.
	BandMemory
	// Band64x receives older 8x memories.
	Band64x
	// Band8x receives older session memories.
	Band8x
	// BandSession receives compressed conversation segments.
	BandSession
)

// MemoryFileID identifies corpus files independently of a particular event log.
type MemoryFileID struct {
	// Date is the UTC creation date, formatted YYYY-MM-DD.
	Date string `json:"date"`
	// Num distinguishes memories created on the same date.
	Num int `json:"num"`
}

// BandSettings measures bytes, not estimated tokens. Zero selects defaults.
type BandSettings struct {
	// Budget is the desired retained text size.
	Budget int `json:"budget"`
	// HighWatermark starts graduation; zero selects twice Budget.
	HighWatermark int `json:"high_watermark"`
	// LowWatermark is the retained size after graduation; zero selects Budget.
	LowWatermark int `json:"low_watermark"`
	// Disabled removes content while preserving the files for re-enable.
	Disabled bool `json:"disabled"`
}

// MemorySettings belongs to each Agent, including independent corpus switches.
type MemorySettings struct {
	// Conversation controls compression of finished dialogue segments.
	Conversation BandSettings `json:"conversation"`
	// Soul controls human-maintained identity data.
	Soul BandSettings `json:"soul"`
	// Memory controls terminal curated memory.
	Memory BandSettings `json:"memory"`
	// Tier64 controls the older compression tier.
	Tier64 BandSettings `json:"64x"`
	// Tier8 controls the newer compression tier.
	Tier8 BandSettings `json:"8x"`
	// Session controls newly compressed conversation memories.
	Session BandSettings `json:"session"`
}

// BandPopulatedData carries the actual nondeterministic bytes for pure replay.
type BandPopulatedData struct {
	// TaskID correlates managed replacement; empty preserves legacy disk/replay facts.
	TaskID string `json:"task_id,omitempty"`
	// Band selects the entry lifetime and canonical position.
	Band Band `json:"band"`
	// File identifies this memory's first source file or creation identity.
	File MemoryFileID `json:"file"`
	// Text is memory data, never a system instruction.
	Text string `json:"text"`
	// Thru marks the last file included by a graduation.
	Thru MemoryFileID `json:"thru,omitzero"`
	// Source explains whether compression or disk scanning supplied the bytes.
	Source string `json:"source,omitempty"`
}

// BandDepopulatedData names removals without resolving anything from disk.
type BandDepopulatedData struct {
	// TaskID holds managed retirement until this exact replacement commits.
	TaskID string `json:"task_id,omitempty"`
	// Band selects the source band to retire.
	Band Band `json:"band"`
	// Thru retires only the oldest range; nil retires the entire band.
	Thru *MemoryFileID `json:"thru,omitempty"`
	// Reason distinguishes graduation from a user's disabled switch.
	Reason string `json:"reason"`
}

// CompactorLaunchedData is observational: replay never launches external work.
type CompactorLaunchedData struct {
	// TaskID identifies one managed attempt; empty retains legacy replay behavior.
	TaskID string `json:"task_id,omitempty"`
	// SelectionHash binds the durable candidate to the captured source fingerprints.
	SelectionHash string `json:"selection_hash,omitempty"`
	// Source is zero for conversation, otherwise the graduating band.
	Source Band `json:"source,omitempty"`
	// File names the first selected corpus identity or new session identity.
	File MemoryFileID `json:"file,omitzero"`
	// From and To identify a managed conversation cut's inclusive boundaries.
	From uint64 `json:"from,omitempty"`
	// To is the last selected completed checkpoint boundary.
	To uint64 `json:"to,omitempty"`
	// TracePath is the owner-relative child trace directory, retained after dismissal.
	TracePath string `json:"trace_path,omitempty"`
	// Band is the destination, matching the successful population event.
	Band Band `json:"band"`
	// Thru correlates this attempt with its eventual populated memory.
	Thru MemoryFileID `json:"thru"`
}

// PendingCompression holds proposed removals until their replacement arrives.
// A crash between the two commit events therefore leaves old information live.
type PendingCompression struct {
	// Launch correlates the eventual replacement with this attempt.
	Launch CompactorLaunchedData `json:"launch"`
	// Cut is a proposed conversation cut, applied only on matching population.
	Cut *RedactData `json:"cut,omitempty"`
	// Retire is a proposed source-band retirement, applied with its replacement.
	Retire *BandDepopulatedData `json:"retire,omitempty"`
	// Committed retains the applied pair until disk promotion and archive finish.
	Committed *BandPopulatedData `json:"committed,omitempty"`
}

// MemoryFingerprint identifies exact owner-relative bytes or their absence.
// Only Memory derives these paths; model output never supplies corpus addresses.
type MemoryFingerprint struct {
	// Path is relative to the owning Agent's data directory.
	Path string `json:"path"`
	// Hash is the SHA-256 of present bytes, including an empty file.
	Hash string `json:"hash,omitempty"`
	// Exists distinguishes an absent destination from a present empty file.
	Exists bool `json:"exists"`
}

// CompressionSelection freezes only selected inputs, leaving new dialogue free
// to continue while a worker runs. The actor separately owns policy generations.
type CompressionSelection struct {
	// EntriesHash covers selected entry identities/parts and task boundaries.
	EntriesHash string `json:"entries_hash"`
	// SourcePolicy is the effective policy at selection, not a mutable config copy.
	SourcePolicy BandSettings `json:"source_policy"`
	// DestinationPolicy records the other policy relevant to this attempt.
	DestinationPolicy BandSettings `json:"destination_policy"`
	// Sources contains fingerprints of selected source files in retirement order.
	Sources []MemoryFingerprint `json:"sources,omitempty"`
	// Destination requires the same prior file or absence at commit time.
	Destination MemoryFingerprint `json:"destination"`
}

// Compression is one finite proposed cut. The original context stays live until success.
// It is work data passed to a worker, not a new Agent or runtime child.
type Compression struct {
	// Source is zero for conversation, otherwise the graduating band.
	Source Band
	// Destination determines the resulting entry kind and disk directory.
	Destination Band
	// File and Thru identify the output and the source range it covers.
	File MemoryFileID
	// Thru correlates launch and completion independently of log sequence.
	Thru MemoryFileID
	// From and To bound the original dialogue events removed after success.
	From uint64
	// To is the inclusive last completed segment boundary.
	To uint64
	// Text is the exact input selected for this bounded call.
	Text string
	// Ratio is the requested reduction, not a claimed measured result.
	Ratio int
	// Retired identifies the source files archived after the commit.
	Retired []Entry
	// Selection is captured by Memory; callers must treat it as immutable.
	Selection *CompressionSelection
	// SelectionHash binds the staged record to the parent-authored launch fact.
	SelectionHash string
}

// Memory owns corpus policy and disk operations through its required Agent parent.
type Memory interface {
	// Agent returns the required owner of configuration, history and logging.
	Agent() Agent
	// Settings returns effective byte budgets for a band (zero means conversation).
	Settings(Band) BandSettings
	// Validate checks a settings value without mutating or persisting it.
	Validate(MemorySettings) error
	// Scan rereads one band's disk files, including human edits.
	Scan(Band) ([]BandPopulatedData, error)
	// Conversation selects the oldest complete checkpoint segments.
	Conversation() (*Compression, error)
	// Graduation selects the next oldest-first waterfall step.
	Graduation() (*Compression, error)
	// ValidateSelection refuses changed selected entries, files, destination or policy.
	// New unrelated dialogue and files do not invalidate the selected prefix.
	ValidateSelection(*Compression) error
	// Stage writes a hidden candidate for the matching launched attempt. It does
	// not change active files or History; callers serialize these commit operations.
	Stage(string, *Compression, string) (*BandPopulatedData, error)
	// Finish promotes and archives a committed receipt idempotently, preserving
	// conflicting files and the hidden record on any error.
	Finish(string) error
	// Recover returns a terminal fact for the actor to append, then a second call
	// cleans its hidden record. It never launches work or mutates History itself.
	Recover() (*CompactionOutcomeData, error)
	// Write persists replacement bytes before any source retirement.
	Write(*Compression, string) error
	// Archive preserves retired files outside the active scan pattern.
	Archive(*Compression) error
}

// CompactionOutcomeData records a terminal managed attempt without changing the
// main turn or its response usage. Only the latest outcome remains in Context.
type CompactionOutcomeData struct {
	// TaskID identifies the attempt whose source/result and trace remain auditable.
	TaskID string `json:"task_id"`
	// Usage retains drained attempt measurements for audit; replay never imports spend.
	Usage []UsageBucket `json:"usage,omitempty"`
	// State is committed, failed, canceled, invalidated, or abandoned.
	State string `json:"state"`
	// Reason explains refusal or cleanup failure without provider credentials.
	Reason string `json:"reason,omitempty"`
	// TracePath preserves the dismissed child's owner-relative trace location.
	TracePath string `json:"trace_path,omitempty"`
}

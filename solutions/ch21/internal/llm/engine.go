package llm

// The engine: vendor-independent single-request and execution primitives.
// The actor owns turn setup and multi-round orchestration.
//
// Notice how little of it there is, and that none of it mentions a vendor.
// Everything vendor-shaped was pushed into the renderer and the parser, which
// is the entire point of Chapter 2.

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/waywardgeek/ensemble/agent/internal/common"
)

type Engine struct {
	Log   *common.Log
	Ctx   *common.Context
	Cfg   common.Config
	HTTP  *http.Client
	Path  string // where the log is persisted, so `dump` can find it
	Jobs  common.JobManager
	Tools common.ToolRegistry
	Host  common.Host

	// Creds resolves the bearer credential for each outbound request.
	//
	// Nil means "use Cfg.APIKey directly", which is the behaviour of every
	// chapter before this one and the state of every test that builds an
	// Engine as a struct literal. That is modelled absence rather than a
	// fallback: a nil provider is a specific, documented configuration —
	// a static key living in Config — not an unresolved failure being
	// papered over with a plausible default.
	Creds common.CredentialProvider

	// lastCredDesc remembers what noteCredential last reported, so the log
	// records every CHANGE of credential without repeating itself every turn.
	lastCredDesc string

	// CredsByVendor holds the provider to install when the agent switches to
	// a given vendor at runtime. The host resolves it for every vendor up
	// front, exactly as it resolves endpoints, so a model switch is a lookup
	// rather than a rediscovery. Credential discovery is host policy — it
	// reads the user's home directory — and must not happen inside the
	// library, let alone mid-turn.
	//
	// A vendor absent from the map means "no provider", i.e. use Cfg.APIKey.
	// Creds has to move with Vendor for the same reason BaseURL and APIKey
	// do: a provider left behind presents one vendor's bearer token to
	// another vendor's host. The observed symptom is quieter than a failure
	// and worse — a metered API key is billed while the user believes they
	// are spending a subscription they already paid for.
	CredsByVendor map[common.Vendor]common.CredentialProvider

	// Journal, when set, receives every event as it is recorded (ch15).
	Journal *Journal
	// Target is the context-size target in bytes (ch15 rule 10); zero
	// means DefaultContextTarget. Read on every request, so a settings
	// change takes effect at the next cut and never rewrites a past one.
	Target func() int

	// ToolRoundLimit reads runtime settings once per human turn.
	// Nil uses Config.MaxToolRounds; zero selects DefaultToolRounds.
	ToolRoundLimit func() int

	// Memory is the store band events are built from: the one component
	// that reads the memory directory. Nil disables the memory system
	// entirely, which is what every chapter before this one wants.
	Memory *Store

	// Bands is the per-band settings, read fresh rather than captured, so
	// switching a band off takes effect at the next sync instead of the
	// next restart.
	Bands func() common.BandConfig

	// Recall, when set, is consulted once per human turn to pull archived
	// material the user did not ask for by name. Nil disables recall
	// completely, which is what every chapter before this one wants, and is
	// also the honest state of a brand-new agent that has nothing archived
	// yet.
	//
	// It is deliberately an interface held by the hub rather than a concrete
	// type: the engine knows that something can answer a query with parts,
	// and knows nothing whatsoever about BM25, chunking, or the existence of
	// a judge. Swapping the whole retrieval strategy is a change to one
	// constructor call at the top of the program.
	Recall common.Recaller

	// Cache, when set, observes every request and the usage it produced, and
	// reports where the cacheable prefix stopped matching the previous turn.
	// Nil disables the analysis, which is what every chapter before this one
	// wants.
	//
	// It observes rather than participates: it cannot alter the request, and a
	// failure inside it must never fail the turn. An agent that refuses to
	// answer because its instrumentation is unhappy is worse than an
	// uninstrumented one.
	Cache common.CacheLens

	// warned is advisory only: it keeps the ninety percent notice from
	// repeating every turn. The authoritative fact, which tools are
	// withdrawn, lives in the log as a ToolsChanged event and survives a
	// restart. This does not need to.
	warned     bool
	requestCtx context.Context
}

func NewEngine(cfg common.Config, path string, jobs common.JobManager, tools common.ToolRegistry, host common.Host) *Engine {
	// Match the single-surface renderer's default in persisted provenance.
	if cfg.Surface == 0 && (cfg.Vendor == common.VendorAnthropic || cfg.Vendor == common.VendorGemini) {
		cfg.Surface = common.SurfaceForModel(cfg.Model, cfg.Vendor)
	}
	return &Engine{
		Log:   NewLog(),
		Ctx:   common.NewContext(),
		Cfg:   cfg,
		HTTP:  &http.Client{Timeout: 120 * time.Second},
		Path:  path,
		Jobs:  jobs,
		Tools: tools,
		Host:  host,
	}
}

// Record appends to the log and advances the context. One path in.
//
// Chapter 15 adds two things, both here because this is the only door. The
// event reaches the journal before anything else happens to it, so a crash a
// microsecond later still has it on disk (rule 9). And the reducer is total
// (rule 8): an event that cannot be applied is kept in the log, skipped by
// the context, and reported in the agent's log. The same skip happens on
// every replay, because replay runs the same Apply, so live and restored
// contexts cannot drift apart over it.
func (e *Engine) Record(ev common.Event) error {
	stored := Append(e.Log, ev)
	if e.Journal != nil {
		if err := e.Journal.Append(stored); err != nil {
			return err
		}
	}
	if err := Apply(e.Ctx, stored); err != nil {
		e.logf("reducer: skipped event %d (%s): %v", stored.Seq, stored.Type, err)
	}
	return nil
}

func (e *Engine) logf(format string, args ...any) {
	if e.Host != nil {
		e.Host.Logf(format, args...)
	}
}

// Say records a human prompt.
func (e *Engine) Say(text string) error {
	return e.Record(common.Event{Type: common.MessageReceived, Message: &common.MessageData{
		Actor: common.ActorHuman, Parts: common.PartList{common.TextPart{Text: text}},
	}})
}

// Attach records volatile data — the time, the screen, live status.
//
// It is an ordinary common.MessageReceived with Actor: System. The reducer is what
// decides it is ephemeral rather than dialogue, which is the chapter's rule
// about classification made concrete: the capture site does not know, and
// cannot know, what an arriving message means.
func (e *Engine) Attach(text string) error {
	return e.Record(common.Event{Type: common.MessageReceived, Message: &common.MessageData{
		Actor: common.ActorSystem, Parts: common.PartList{common.TextPart{Text: text}},
	}})
}

// requestCfg returns a copy of Cfg whose APIKey field holds a bearer
// credential that is valid right now.
//
// Copying matters. The resolved token is per-request state, and writing it
// back onto e.Cfg would turn a cache of the last refresh into the
// configuration itself: a later reader could not tell the operator's
// long-lived key from an access token that expires in forty minutes. The
// engine's Cfg keeps holding whatever was configured; only the copy handed to
// the renderer carries the resolved value.
//
// A nil provider returns Cfg unchanged. See the Creds field for why that is
// modelled absence and not a fallback.
// noteCredential records which credential is in force, the first time one is
// used and on every change after that.
//
// It never prints the credential itself: DescribeCredential names the kind and
// has no path that can emit a token.
//
// Logging only on change is deliberate. A line every turn is a line nobody
// reads, and the fault worth catching here is a credential that changes when
// nothing asked it to -- a model switch moving the route while the token stays
// behind, which bills the wrong account without failing.
func (e *Engine) noteCredential() {
	desc := common.DescribeCredential(e.Creds)
	if desc == e.lastCredDesc {
		return
	}
	e.lastCredDesc = desc
	e.logf("credential: vendor=%v using %s", e.Cfg.Vendor, desc)
}

func (e *Engine) requestCfg() (common.Config, error) {
	e.noteCredential()
	cfg := e.Cfg
	if e.Creds == nil {
		return cfg, nil
	}
	tok, err := e.Creds.GetBearerToken(context.Background())
	if err != nil {
		return cfg, fmt.Errorf("resolve credential: %w", err)
	}
	cfg.APIKey = tok

	// The credential also selects which deployment the request lands on, and
	// those deployments do not accept the same parameters. Resolving the
	// route here, in the one place that already resolves the token, keeps the
	// two facts from drifting apart: a renderer can never be handed a plan
	// token with a metered route's permissions.
	cfg.Route = common.RouteFor(e.Creds.Kind())
	return cfg, nil
}

// Turn renders the current context, sends it, and folds the response back in.
//
// watch is the caller's half of the stream: set OnDelta to watch content
// arrive and OnPartFinal to be told when a part is complete. Both are
// optional, and a zero StreamCallbacks is the ordinary non-observing call.
// The engine overwrites OnEvent and OnFrame with its own recording, because
// the event log and the API log belong to it — there is deliberately no way
// for a caller to intercept what gets logged.
func (e *Engine) Turn(watch common.StreamCallbacks) (string, error) {
	return e.TurnContext(context.Background(), watch)
}

func (e *Engine) TurnContext(ctx context.Context, watch common.StreamCallbacks) (string, error) {
	e.requestCtx = ctx
	defer func() { e.requestCtx = nil }()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	// Loud refusal: reject unknown models before doing anything else.
	if _, known := common.LookupModel(e.Cfg.Model); !known {
		err := fmt.Errorf("unknown model %q: not in the supported model table; refusing to proceed", e.Cfg.Model)
		_ = e.Record(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{Message: err.Error()}})
		return "", err
	}

	renderer, parser, err := SeamFor(e.Cfg.Vendor, e.Cfg.Surface)
	if err != nil {
		return "", err
	}

	// Chapter 16: turn finished work into memory before deciding what to
	// cut. Compaction is the coarse move and curation is the fine one, so
	// compaction goes first: there is no point spending the redaction
	// ladder on a span that is about to become a memory, and a span that
	// survives compaction is exactly the span worth curating.
	e.maybeCompact()

	// ...and then check whether that was enough. Compaction first, because
	// it is what usually brings the context back under the line; forcing is
	// for when there is nothing checkpointed left to compact.
	e.maybeForce()

	// Chapter 15: decide this request's cuts and record them as events
	// BEFORE rendering, so the request carries exactly what the log says.
	if err := e.curate(); err != nil {
		return "", err
	}

	// A call with no result makes the conversation unsendable, so close
	// any that are still waiting before the render reads the context.
	// Recorded as events, for the same reason the cuts above are.
	if err := e.closeLostCalls(); err != nil {
		return "", err
	}

	// A result that arrived LATE is not lost, so the repair above cannot see
	// it: the call is answered, just not where a vendor demands. Report it
	// rather than repair it, because moving an entry would make the context
	// disagree with the log it is projected from.
	e.reportMisplacedCalls()

	// RENDER BEFORE RECORDING common.RequestSent.
	//
	// Reverse these two lines and the bug is subtle and expensive: the reducer
	// clears pending ephemera on common.RequestSent, so recording first means the
	// renderer never sees them and the volatile data is silently never
	// delivered. Nothing errors. The model just quietly does not know what
	// time it is. Chapter 4 hits the identical ordering trap with hints.

	// The credential is resolved immediately before rendering rather than at
	// startup. For a static API key the two are the same thing. For an OAuth
	// access token they are the difference between a request and a 401: the
	// token may have expired since the previous turn, and the provider
	// refreshes on demand when asked for one.
	cfg, err := e.requestCfg()
	if err != nil {
		return "", err
	}
	req, err := renderer.Render(e.Ctx, cfg)
	if err != nil {
		return "", err
	}

	// Capture the exact bytes about to go on the wire.
	//
	// This hangs off the call site rather than off Render for two reasons.
	// First, all three vendors route through here, so one hook covers them
	// all. Second and more important, Render has other callers that build
	// entirely different conversations: the relevance judge and the memory
	// compressor. If those were captured, one of them landing between two
	// turns would become the "previous request" and the diff would compare a
	// judge prompt against a conversation. Hooking Turn excludes them by
	// construction instead of by filtering.
	//
	// BodyOf reads from req.GetBody, which hands back a fresh reader, so the
	// live body is not consumed.
	if e.Cache != nil {
		if body, bErr := common.BodyOf(req); bErr == nil {
			e.Cache.ObserveRequest(e.Cfg.Model, body)
		}
	}
	if err := e.Record(common.Event{Type: common.RequestSent, Request: &common.RequestData{To: common.Provenance{
		Vendor: e.Cfg.Vendor, Model: e.Cfg.Model, Surface: e.Cfg.Surface,
	}}}); err != nil {
		return "", err
	}

	resp, err := e.send(req)
	if err != nil {
		// A transport failure is infrastructure: it ends the turn.
		_ = e.Record(common.Event{Type: common.ErrorOccurred, Error: &common.ErrorData{Message: err.Error()}})
		return "", err
	}
	// The PARSER reads the body; the ENGINE owns closing it. Handing a live
	// body to the seam and keeping the close here is what lets one Parse
	// method serve a trickle and a whole document without caring which it got.
	defer resp.Body.Close()

	var vendorErr, recErr error

	// The caller supplies the watching half of these callbacks (deltas and
	// finals, which are observations). The engine supplies the recording
	// half, because the event log and the API log are its responsibility and
	// a stateless vendor parser has no Host to write to.
	watch.OnEvent = func(ev common.Event) {
		if recErr != nil {
			return
		}
		if err := e.Record(ev); err != nil {
			recErr = err
			return
		}
		if ev.Type == common.ErrorOccurred && ev.Error != nil {
			// The event is the record; this is the report. Without it a 404
			// on the model name prints {"assistant":""} and exits 0 — measured
			// live against Gemini, and invisible unless you read the log.
			vendorErr = fmt.Errorf("%s: %s", e.Cfg.Vendor, ev.Error.Message)
		}
		// Usage arrives on this event and nowhere else, and it is the
		// per-request figure rather than the running session total the reducer
		// keeps. The lens needs the per-request one: the question is what this
		// request was charged, not what the session has cost so far.
		if ev.Type == common.ResponseEnded && ev.Response != nil {
			if e.Cache != nil {
				e.Cache.ObserveUsage(ev.Response.Usage)
			}
			// Report spend up the parent chain. The reducer's total is durable
			// and spans every run the conversation has had; this one spans this
			// process, which is what a human means by "this session".
			if e.Host != nil {
				e.Host.RecordUsage(ev.Response.Usage)
			}
		}
	}
	watch.OnFrame = func(eventType string, data []byte) {
		// Per-frame, so the API log stays byte-complete under streaming. The
		// body is consumed by the parser now, so logging it whole is no
		// longer possible and the response half of the log would otherwise
		// simply go dark.
		if eventType != "" {
			e.Host.APILogf("<<< [%s] %s", eventType, string(data))
			return
		}
		e.Host.APILogf("<<< %s", string(data))
	}

	if err := parser.Parse(resp, watch); err != nil {
		return "", err
	}
	if recErr != nil {
		return "", recErr
	}
	return e.lastAgentText(), vendorErr
}

func (e *Engine) send(req *http.Request) (*http.Response, error) {
	// Log the outbound JSON request.
	if req.Body != nil {
		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
		e.Host.APILogf(">>> %s %s\n%s", req.Method, req.URL, string(reqBody))
		req.Body = io.NopCloser(strings.NewReader(string(reqBody)))
	}

	if e.requestCtx != nil {
		req = req.WithContext(e.requestCtx)
	}
	resp, err := e.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	// Only the status line here. The body is NOT read: reading it would
	// consume the stream the parser is about to walk, and buffering it whole
	// would give up streaming entirely while still looking like it worked.
	e.Host.APILogf("<<< %d %s", resp.StatusCode, resp.Header.Get("Content-Type"))

	return resp, nil
}

// PendingCalls returns the tool calls that have been asked for and not yet
// answered, in the order they were asked.
//
// It is computed from the dialogue rather than stored, for the same reason
// everything else in Chapter 2 is computed from the dialogue: the log is the
// truth, and a second copy of the truth is a second thing to get wrong.
func (e *Engine) PendingCalls() []common.ToolCallPart {
	answered := map[string]bool{}
	var calls []common.ToolCallPart
	for _, entry := range e.Ctx.Dialogue {
		for _, p := range entry.Parts {
			switch v := p.(type) {
			case common.ToolCallPart:
				calls = append(calls, v)
			case common.ToolResultPart:
				answered[v.CallID] = true
			}
		}
	}
	var out []common.ToolCallPart
	for _, c := range calls {
		if !answered[c.CallID] {
			out = append(out, c)
		}
	}
	return out
}

func (e *Engine) Save() error { return SaveLogFile(e.Log, e.Path) }

// CallEphemeral runs all tools with the given ephemeral mode ("round" or "turn"),
// combines their output, and records it as a system message so the reducer puts
// it into context.Ephemera.
//
// Ephemeral tools are NOT included in the tool declarations sent to the LLM —
// the model never sees them as callable. It sees their output as context data.
func (e *Engine) CallEphemeral(mode string) error {
	tools := e.Tools.EphemeralTools(mode)
	if len(tools) == 0 {
		return nil
	}

	var parts []string
	for _, t := range tools {
		c := &common.Call{Host: e.Host, Jobs: e.Jobs}
		out, err := t.Run(c, nil)
		if err != nil {
			// Ephemeral tool errors are logged but not fatal —
			// a snapshot failure should not abort the turn.
			e.Host.Logf("ephemeral tool %s error: %v", t.Name, err)
			continue
		}
		if out != "" {
			parts = append(parts, fmt.Sprintf("## %s (auto-updated)\n\n%s", t.Name, out))
		}
	}

	if len(parts) == 0 {
		return nil
	}

	combined := strings.Join(parts, "\n\n")
	return e.Attach(combined)
}

func (e *Engine) lastAgentText() string {
	for i := len(e.Ctx.Dialogue) - 1; i >= 0; i-- {
		if e.Ctx.Dialogue[i].Actor != common.ActorAgent {
			continue
		}
		var b strings.Builder
		for _, p := range e.Ctx.Dialogue[i].Parts {
			if t, ok := p.(common.TextPart); ok {
				b.WriteString(t.Text)
			}
		}
		return b.String()
	}
	return ""
}

// RenderOnly plays a log to a context and renders it, making no network call.
// This is what makes replay, redaction, ephemera and the seam into byte
// comparisons — and if your architecture cannot offer it cheaply, your context
// is not actually separate from your transport.
func RenderOnly(path string, cfg common.Config) ([]byte, error) {
	log, err := LoadLogFile(path)
	if err != nil {
		return nil, err
	}
	ctx, err := Replay(log)
	if err != nil {
		return nil, err
	}
	renderer, _, err := SeamFor(cfg.Vendor, cfg.Surface)
	if err != nil {
		return nil, err
	}
	req, err := renderer.Render(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}
	return common.BodyOf(req)
}

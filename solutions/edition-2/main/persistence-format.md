# Persistence format, version 1

Normative student codec specification for Chapter 10; implementation status is
in evidence/ch10/implementation-status.md. It is not a runtime acceptance claim.

## Storage and exactness

Store leaves are regular, non-symlink files: events.log, optional checkpoint.json,
owner.lock, and origin.json only for explicit imports. No prefix is rewritten.
Log header is exactly `{"log_version":1}` followed by LF. Session events are
contiguous, beginning with session_initialized at 1 or session_anchor at S+1.
Existing version-1 event payload schemas remain those of Chapters 2–9, extended
by Chapter 10's session, limits and turn.request_index fields. Event annotations
remain allowed; duplicate members and additional known payloads are invalid.

The checkpoint root has exactly these required members:

| Member | Type / constraint |
|---|---|
| version | integer 1 |
| session_id | 32 lowercase hex digits |
| identity | Identity below |
| as_of | positive uint64 |
| high_watermarks | Watermarks below |
| state_version | integer 1 |
| state | State below, or null for rebuild using available full history/origin |
| state_sha256 | lowercase 64-digit hex SHA-256 of canonical state, or null exactly when state is null |

Identity is `{mode,system,skills,handlers}`. mode is plain or skills. Plain has
nonempty exact system string and skills:null; skills mode has system:null and
skills:{primary:string,catalog_sha256:hash,bindings_sha256:hash}. handlers is a
name-sorted array of `{name:string,description:string,schema:JSON object}` with
unique names; it is the installed ceiling, not current grants. Catalog hash
covers all parsed definitions sorted by name, each with name, description, type,
tools, depends, loadable-skills and body; sets sort lexically and empty sets are
[]. Bindings hash covers the complete copied scalar map. Paths are excluded.

Watermarks is `{event:uint64,request:uint64,activation:uint64,job:uint64}`.
event equals as_of; activation/job are at least represented durable maxima. request
is at least the maximum recorded request_index and may include burned admissions.
Every subsequent recorded index exceeds that captured cursor; recorded IDs cannot
be reused. Zero is allowed only for unused cursors, never allocated identities.

Canonical JSON uses UTF-8 byte-sorted object keys, no whitespace, preserved
array order and literal noncontrol UTF-8 (no HTML/slash escaping or normalization).
Use short JSON control escapes and lowercase \\u00xx for remaining controls.
Normalize numbers using exact signed coefficient and decimal exponent: remove
leading zeroes, canonicalize any zero to 0, remove trailing coefficient zeroes
and adjust exponent; omit exponent iff zero, otherwise use e and shortest signed
exponent without plus or leading zeroes. Never use binary64 or expand the exponent.
Exact integers accept only integer JSON tokens in structural integer fields.
Canonical tokens are hash material; stored structural integers remain decimal.

`Raw` below means null for absent raw data, otherwise a JSON string containing
its **exact original JSON text**, not a JSON value encoded by this codec. Decode
and validate that text with duplicate/depth/UTF-8/type checks and inherited
payload constraints. JSON syntax cannot hide inside a string wrapper. Arguments
must decode to objects; opaque/data/raw_usage retain inherited JSON types.
Compare replay Raw strings and text byte-for-byte. Compare schema/identity and
call argument correspondence via canonical parsed values, without changing Raw.

## Closed nested grammar

Unless explicitly marked nullable, every member listed below is required and
non-null; no other member is allowed. All arrays, including empty ones, are
present. All maps are objects with dynamic keys only where explicitly stated.
String means a valid UTF-8 string. Integers are uint64 unless a signed type is
specified. Structural null is never shorthand for an omitted array/map.
Names/IDs/provenance and strings retain inherited validity rules. The fixed
shape of a snapshot Part deliberately differs from sparse event Part encoding.

State = `{session,context,usage,skills,limits,window}`.

- session = `{id,identity,as_of,high_watermarks}`; exactly equals outer metadata.
- usage = array of `{from:Provenance,usage:Usage}`, sorted by vendor/model/surface,
  unique provenance, agreeing with context.responses summed once.
- skills = null in plain mode, otherwise SkillSnapshot below.
- limits = null or nonempty LimitValues below; agrees with accepted limit facts.
- window = `{events:[WindowEvent],renderable_count:uint64,event_count:uint64}`;
  <=100 events, strictly increasing by seq, exactly the most recent retained
  renderable accepted facts. count includes omitted renderable events.

Provenance = `{vendor:string,model:string,surface:string}`; supported vendor and
surface with exact nonempty model. Usage = `{input,cache_write,cache_read,output}`,
all nonnegative signed int64 with checked sums. Ref = `{kind:integer,locator:string}`
with kind 1/2/3 and nonempty locator; no automatic reference access.

Part = `{type,text,from,opaque,call_id,name,args,parts,is_error,data,mime,ref,stub}`.
text is nullable string; from nullable Provenance; opaque/args/data Raw; parts
array of Part; ref nullable Ref; is_error Boolean; remaining members strings.
Unused strings are empty, unused raw/pointers null, unused parts []. Variants:

| type | Active fields / constraints |
|---|---|
| text | text present (empty allowed); from and opaque both present or both absent |
| tool_call | call_id,name nonempty; from present; args object Raw; opaque optional |
| tool_result | call_id nonempty; parts contain only text/blob/redacted; is_error Boolean |
| opaque | from present; data present Raw |
| blob | mime nonempty; ref present |
| redacted | stub nonempty; ref optional |

Entry = `{skill_name,activation,anchor,seq,actor,purpose,parts}`. seq positive,
anchor 0 or accepted call-batch response sequence. actor/purpose combinations:
human/dialogue or hint; agent/dialogue; tool/dialogue; system/instruction,
ephemeral or skill. Skill entry has skill_name and activation, parts=[] on disk;
material is resolved from that activation with the exact Chapter 9 envelope.
Other entries have skill_name="" and activation=0. Arrays retain reducer projection
order; anchors can legitimately make projection order differ from sequence order.

Context = `{SkillMode,SkillPrimary,SkillBatch,DeferredSkills,PendingSkills,Hints,
TurnID,TurnIDs,ExplicitTurns,FinalResponse,Jobs,Entries,Instructions,Ephemera,
Pending,Active,Continuation,Calls,LastSeq,Session,RequestCursor,RequestSeqs,Turns,Responses,
Redactions,Guidance,LimitFacts}`. These field spellings follow shared Context.

- SkillMode, ExplicitTurns, FinalResponse, Active, Continuation: Booleans.
- SkillPrimary: empty on disk (resolved from primary activation on installation).
- SkillBatch, LastSeq, RequestCursor: uint64; LastSeq equals as_of.
- DeferredSkills, PendingSkills, Hints, Entries, Instructions, Ephemera: [Entry].
- Pending: nullable Entry; TurnID string (empty if none).
- TurnIDs: object mapping nonempty request IDs to true; equals keys of Turns.
- Calls: object mapping exact call IDs to CallState.
- Jobs: object keyed by canonical positive decimal handles, values JobSnapshot.
- Session: nullable SessionFact; for snapshots must be initializer shape matching
  metadata. An anchor changes LastSeq, not the creation identity.
- RequestSeqs: sorted unique positive accepted request_sent coordinates, <=LastSeq;
  every nonzero Guidance.consumed_at references one of these. No request body copy.
- Turns: object mapping request IDs to TurnFact.
- Responses: array of ResponseFact in sequence order.
- Redactions: array of RedactionFact in event order.
- Guidance: array of GuidanceFact in receipt order.
- LimitFacts: array of LimitFact in event order.

CallState = `{JobHandle,Part,Dispatched,Returned,CalledAt,ReturnedAt}`: uint64 (0 means no Job),
Part of tool_call type, Booleans. CalledAt/ReturnedAt are event coordinates (0 iff absent); dispatch precedes result and both follow the original call response. Its key equals Part.call_id; call part exactly
matches a response entry; returned requires matching result entry. Historical
call/result identities survive redaction. No accepted call can disappear.

JobSnapshot = `{handle,status,output,bytes,exit_code,is_error,reason,cwd}`.
handle positive; status running/done/killed; output is kind-3 Ref cr/io/HANDLE;
bytes nonnegative int64; exit_code nullable signed int; is_error Boolean;
reason empty except kill_job/shutdown for killed; cwd string (empty or absolute).
No runtime ownership, process, input/output cursor or worker is encoded.

TurnFact = `{index,start,end,outcome,policy}`: positive index/start/end and
terminal outcome success/interrupted/canceled/error/round_limit/stopped; policy
nullable `{revision,max_model_requests,effective_max_model_requests}` (signed
integer request limits 0..256, effective zero maps to 16). Snapshot live boundary
requires every turn terminal; historical inspection may diagnose unfinished logs.

ResponseFact = `{seq,from,requested,model_reported,usage,raw_usage,stop_reason}`:
seq positive, from Provenance, requested nullable Provenance, model_reported
nullable Boolean, usage Usage, raw_usage Raw, stop_reason string. References
one accepted agent entry. No second copy of response parts.

RedactionFact = `{seq,from,to,level,reason}`: positive ordered span strictly before
seq, level redact_result, nonempty reason. Context result children must equal the
corresponding effective redaction projection, without erasing original raw logs.
GuidanceFact = `{seq,kind,consumed_at,anchor}`: kind hint/ephemeral; consumed_at 0
or later request sequence, anchor 0 or completed batch coordinate. Unconsumed
facts correspond exactly to pending hints/ephemera, with matching anchors.

LimitValues = object containing a nonempty subset of ai_callback_delay (finite
nonnegative JSON number within duration range), ai_callback_pattern (string,
valid Go regex including explicit empty), max_output_bytes (positive signed int).
No other member; omission and explicit empty pattern differ.
LimitFact = `{seq,kind,call_id,name,overrides}`: kind set/consumed; name empty on
set and exact attempted name on consumed; overrides LimitValues. Facts match
accepted calls/arguments, consumption before dispatch, set after dispatch before
successful setter result; consumed overrides equal prior pending setting.

SkillSnapshot = `{state,material,transitions,last_id,ceiling,contributors}`.
state = `{revision,primary,roots,active,available,tools,retired}` with Chapter 9
sorted arrays: roots/tools strings; active `{name,type,activation}`; available
`{name,description}`; retired `{name,activation}` sorted by activation.
material = [{record:Activation,event_seq:uint64,retired:Boolean}] sorted by ID.
Activation = `{activation,name,type,body,sha256,tools,dependencies,offers}` exactly
as Chapter 9 (arrays required). Bodies occur **once** in material. dependencies
sorted positive activation IDs; offers sorted `{name,description}`. last_id
is max allocated ID. ceiling sorted unique names equals identity handler names.
contributors = [{tool:string,activations:[uint64],mandatory:Boolean}] sorted by
tool; derived active grant contributors plus mandatory management pair.
transitions = [{seq,action,name,state,activated:[uint64]}] in durable sequence
order. initialize first, then load/unload; activated references material, never
copies bodies. Validate every historical visibility/closure/action/retirement,
revision, dependency/material order, no reused IDs and exact hashes. For live
mount also compare frozen current candidate expansion and grants at each step.

WindowEvent = `{event:SnapshotEvent,activations:[uint64]}`. SnapshotEvent has all
Event struct fields required: skills,turn,hint,job,seq,type,time,message,request,
response,tool,redact,error,session,limits. Exactly matching payload non-null;
other payloads null. Nested fields are the inherited event grammar, encoded with
all struct fields present, nullable pointers/raw as null, omitted numeric/string/
Boolean fields as 0/""/false, lists as []. Replay raw fields use Raw strings.
For skill payloads, activated=[] and activations contains material references;
reconstitute activated from material. Skill payload has action,name,ceiling,state,
activated; ceiling=[] on changes. This is a bounded witness only. Payload shapes:

- turn: policy nullable TurnPolicy, request_id string, outcome string,
  request_index uint64 (0 except starts).
- hint: request_id,text strings.
- message: Entry; request: delivery string,hints [uint64],configuration nullable
  `{system:string,max_tokens:int,tools:[ToolDefinition],resolved_model:string}`,
  to Provenance,ephemera [uint64]. ToolDefinition is name,description,input_schema
  (object Raw; schema field in Identity remains semantic JSON).
- response: stop_reason,from,requested,model_reported,parts,usage,raw_usage;
  requested/model_reported/usage nullable, parts [Part]; remaining as above.
- tool: job nullable JobSnapshot,call_id,name,args Raw,parts [Part],is_error Boolean.
- redact: from,to,level,reason; error: code,message strings; job as above.
- session: SessionFact; limits: `{call_id,name,overrides:LimitValues}`.

SessionFact in Context/window encodes all five fields: session_id,identity,
origin_as_of,origin_sha256,high_watermarks. identity nullable Identity,
high_watermarks nullable Watermarks, other unused fields 0/"". Initializer has
identity present; anchor has identity null and the other three fields present.
Actual log session payloads remain **exactly** Chapter 10's sparse initializer
or anchor shapes. Session/limit bookkeeping never enters renderable window.

## Semantic acceptance and bounds

Check all references, exact bytes, ordering, pairing, provenance, collection
counts and cross-owner projections before installation. Settled snapshot has no
TurnID/Pending/Active/unanswered call/unresolved batch/deferred material. Pending
completed-batch hints and independent historical running jobs are allowed. Never
restore processes or runtime queues. Pending limits restore to Jobs; usage to
Engine; skills to Skills; Context to Agent. No second mutable saved Context.

Full-origin loading independently reduces available prefix and compares this
projection, then applies tail once. Null state rebuild uses available origin
and history. Imported origin is immutable accepted checkpoint bytes, whose
whole-file SHA-256 includes LF. Anchor validates identity, hash, as_of, watermarks
and S+1. Later checkpoints cannot replace origin. Events/Dump expose actual
anchor/tail only; pre-origin reconstruction returns history_unavailable.

Limits: 1 GiB log; 64 MiB each session physical event including LF; 512 MiB
checkpoint/origin; 256 MiB canonical state; 1,000,000 each events/entries/parts/
activations/seen calls/requests; depth 128; 1,024 installed handlers and 16 MiB
canonical handler definitions. Inherited smaller input/source/policy limits
remain. Standalone/header bounds stay inherited. Skill candidate size refusal
remains controlled skill_too_large; subsequent storage admission is terminal
session_limit. Reads over bounds are session_corrupt. Parse incrementally before
unbounded allocation. Exact limits pass when all other constraints hold.

Unicode clarification cf73a64: all strict session JSON string values and keys,
including parsed inner raw-JSON strings, must encode Unicode scalar values.
Reject lone escaped high/low surrogates before decoding; accept adjacent valid
pairs, genuine U+FFFD and escaped-backslash literal text. Legacy standalone
JSON behavior is unchanged by this added session rule.

Prepared-event clarification c5ad6c0: new session candidates validate before one
bounded JSONL preparation. Those prepared bytes define accepted Raw fragments;
append exactly them before applying/publishing. Preparation preserves number
lexemes and order, decoded text/manual/signature values and opaque meaning, while
normalizing line formatting/escape spelling as needed. Existing record reads
preserve their actual accepted fragments and physical sizes. Strict snapshot and
prefix comparisons remain byte-exact for Raw. This is distinct from canonical
hash/equality normalization. See student-review Q5 for above-maximum watermark
wording awaiting confirmation; ordinary exact-max paths are unchanged.

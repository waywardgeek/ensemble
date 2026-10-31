# Edition 2 Chapter 21 / original 23: grader enhancement

Final coder submission, October 11; independent acceptance pending. First green
stage is `b1f91580`. Final actual full results are accepted 100/100 and legacy
75/100; no composed score is needed. Mandatory
`.agents/skills/ensemble-coding/SKILL.md`, full workflow and carryover reloaded
after compaction. No runtime, frozen export or original grader edits.

The separate grader retains all seven original rows and weights: file paths 15,
commands/descendants 20, network 15, credentials 15, child narrowing 10, safe-mode
absence 15, host-path refusal 10. Total 100; no invented parity. It builds the
actual core CLI, supplies a declared older Sonnet-5 model and typed primary,
observes the completed answer, sends EOF and checks successful process exit.
Only the external model peer is fake; file operations, jobs, Seatbelt and the
public Child/Clamp consumer are real. macOS is the original supported platform.

Run `go run ./cmd/grade-edition2 -ch 21 solutions/edition-2/ch21`.
For legacy, use `-sandbox-library .` before `solutions/ch23`; this selects its
existing public import path for the grader's temporary external consumer. The
accepted default is the public `ensemble` subpackage. No product flag is added.

Initial accepted grading passed 100/100. The first legacy 40/100 contained
invalid fixture assumptions and is preserved, not treated as four security
failures. The original safe-mode list names four exec/job tools; requiring
`tool_limits` removal added an unstated condition. The accepted CLI puts its
API trace under the sandbox root, whereas legacy uses cwd. Launching from the
actual sandbox root makes the log reachable on both paths. A PATH-selected
Python failed, while explicit `/usr/bin/python3` succeeds on the corrected
legacy run. The synthetic credential in a fake model response seeds the trace;
its returned assistant history is not itself graded as a filtering obligation.
Only actual environment/log tool results establish credential confinement.

The next actual full runs were accepted 100/100 (1.850 s, exit 0), legacy 75/100
(1.658 s, exit 1). Legacy's default log read still truncated before the seeded
response, so that result did not prove a redaction failure. An explicit inherited
65,536-byte file/output budget exposed the response and produced an actual full
75/100 in 1.090 s. The readable log contains the synthetic credential. Source
confirms legacy Logger writes directly; its separate SanitizeForLog helper does
not protect that path. The other failure is real: legacy Within resolves the
parent of an existing symlink instead of the leaf, accepting a child root outside
its parent's tree. Legitimate existing and not-yet-created subroots remain
positive controls; narrowing and the warning are checked independently, using
physical containment rather than exact returned path spelling.

The current accepted focused full test passes under race in 2.599 s; grader and
command vet pass. Reports, commands and source hashes are in
`solutions/edition-2/evidence/ch23/enhancement21/`. Initial raw reports and source
copies remain distinct from the corrected stage. Original grading stays 0/100:
missing model launch and wrong public import masked its behavior, as previously
independently accepted. Actual historical 12 + 3 Astra/high request/completion
pairs and original 27 mutation attempts retain their original source provenance.

New numeric checks pair real useful file/edit/search/list and command/Python/Git
work with correlated denied relative/absolute/symlink reads, outside writes and
nested-shell reads. Network denial uses the same live local HTTP target as a
successful host control, plus an unchanged hit count. Safe mode pairs file
success with declaration absence and actual guessed-name unknown-tool refusals.
The public consumer checks inheritance, cleared single-file paths, effective
restriction, independent diagnostics and symlink escape. These are bounded
observations, not a general security certification. Rooted-open races, state and
media paths, web disable, immediate toolchain-cache path scope and unsupported
platform behavior retain their accepted source/integration evidence separately.
EOF success does not prove all possible cancellation or descendant lifetimes.


Final validation and exact examples
-----------------------------------

`accepted-final.json` and `legacy-final.json` are actual complete reports from
the final predicates: 0.578 s and 0.569 s after binary builds. The enclosing
external-capture test passes in 1.817 s. Each ordinary and safe CLI completes
its final response and exits normally after EOF. A capture-only Go overlay
retains wire bytes without changing grading predicates. Its first version
accidentally added an unused import inside the embedded consumer as well as the
driver, producing accepted 90 and legacy 75 with setup errors. Those reports
and the bad overlay are preserved as `before-overlay-import-fix-*`; they are
not the final baseline. The corrected overlay only adds request capture.
`credential-observation.json` independently isolates each successful log result:
the seeded marker is present on both implementations, the complete synthetic
credential is absent from both environment results, and only legacy's actual
log tool result contains it. No credential is drawn from a live profile.

The final refusal predicate also requires the caller's supplied relative,
symlink or cwd path to remain visible; absence of the host root alone had not
established that original obligation. A generic refusal defect now fails only
`no-host-path-leak`, and restored grading returns 100. The final source binding
records the exact command/check/driver/test hashes and unchanged frozen tree
identities. Final additions after execution are explanatory comments only.
Focused race passed at the prior stage; final vet passes. There was no repeat
of the original known-wrong launch or paid exercise.

The exact manuscript execution binding is `20c231d9`; all 12 fences remain
byte-identical in final prose `687a63e5`, as recorded in `final-source.json`.
Four Go excerpts match actual owning source modulo whitespace; the published
core build compiles that source. The complete typed primary, scratch setup,
Astra/high subscription launch, two ordinary prompts and separate safe launch
all run through the real frozen CLI with a signed external issuer and faithful
scripted Responses peer. Nineteen main requests and three successful empty
judge replies are retained. Assertions inspect correlated real file, edit,
search/list, Python assertion, Git commit, outside read/write refusals, absent
synthetic environment token, safe file effects and guessed-command refusal.
Both clients exit 0 after observed final responses. No GUI or port 8088 is used.
The grader fence reuses the measured complete grade instead of repeating it.

The initial example guard stopped before execution when the author changed
prose. A subsequent completed example attempt failed validation because the
peer expected main reasoning fields on a tool-free judge and the validator
assumed a different refusal phrase. Both failed receipts remain. The corrected
peer distinguishes auxiliary requests, includes actual completed-response usage
and checks the observed refusal. This is scripted mechanics validation, not a
new autonomous coding, model-quality, real account-eligibility or live-network
result. Historical live evidence remains the separate 12 + 3 request sessions.

Sensitivity and retained limits
-------------------------------

Thirteen distinct source defects are detected in isolated copies. Ten first
cases exercise absolute file bypass, host-root disclosure, kernel reads, kernel
writes, network permission, lost child inheritance, missing safe restriction,
missing diagnostic, safe-mode removal and environment-token leakage. All fail
their intended weighted rows. The network defect raises the actual same-target
hit count; no marker-only denial is substituted. Empty successful file reads
fail file and safe positive controls and also the log positive, so those three
failures are coupled. Exit 17 after successful work fails six CLI rows through
the shared exit guard, while the independent child consumer still passes; it
is one lifetime observation, not six separate security proofs. The thirteenth
case is the generic refusal described above. Both restored runs return 100.

The first script stopped before applying empty-read because its replacement
matched both read and edit implementations. The saved correction bounds it to
the first read function; no compilation/setup failure is counted as sensitivity.
The first 12 cases bind the stage predicates; the final generic-path case and
restored control bind the corrected predicate. The separate real legacy log
and independently verified symlink-root failures supply additional actual
negative controls. No claim is made that all 13 cases ran against final comment
bytes or that removing one of several redundant redaction layers must fail.

Unchanged stronger claims reuse the accepted original-numbered source tests,
27 historical attempts and corrected regressions, with their original source
versions. In particular: `TestSandboxFiles`, `TestSandboxCommands`,
`TestSandboxChildSpec`, `TestSandboxSafeMode`, `TestSandboxCredentialLogs`,
`TestSandboxWebGate`, media/state checks and independent cache/cwd regressions.
These cover rooted operations and path races, full environment/Git policy,
web disable and immediate cache-file scope beyond the numeric fixture. The
historical OpenFile survivor and misplaced media positive control stay recorded;
25 intended failures among 27 attempts are not 25 independent security proofs.
No expanded platform, arbitrary filesystem-race, mount/hardlink or general
sandbox certification is claimed.

The new check/driver contain 440 physical lines, 25 comment-only lines and 399
nonblank non-comment lines (including the embedded consumer). The 32-line test
has 3 comment-only lines. The comment ratio is below the teaching target;
comments explain completion/EOF, positive controls, log budget and redaction
scope, four-tool contract and public physical-boundary comparisons without
padding the straightforward fixture lists. The only new exported API,
`Chapter21`, is documented. Original grader files and both frozen solutions
remain unchanged; shared runtime/crossover work is excluded from all commits.

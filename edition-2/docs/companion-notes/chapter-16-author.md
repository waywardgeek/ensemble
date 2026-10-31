# Chapter 16 author note

## Scope and editorial decisions

Read the current companion voice guide, the complete current Edition 3 Chapter 16 including exercises and final restart note, accepted companion Chapter 15, and its author note. Recovered clipped passages. Used only that approved set and the two owned output files; no legacy guides, implementation inspection, Git, children, builds, graders, runtime tests, or lint.

The organizing problem is preserving client control of context while changing provider, credentials, and funding. The explanation resolves through separate owners for credentials, rendering, transport, and billing. The manuscript translates exercise contracts into prose rather than assignments. It retains the credential-replay incident, refresh race, subscription transport failures, accessibility motivation, and paid tool-loop observations because they explain decisions. It omits command recipes, code listings, tables, and the local grading scoreboard.

Whitespace-delimited word counts: source **6,607**, companion **3,465**, about **48% shorter**. The H1 is exactly `# Chapter 16: Leaving Anthropic for OpenAI`.

## Mechanisms and evidence

- Preserved private-reasoning replay authorization versus encryption, the dated preserved-thinking/context-editing conflict, and uncertainty about the precise OpenAI mitigation. Research and API claims are inherited from the current source, not independently rechecked online.
- Retained credential ownership outside renderers; PKCE, fresh state/nonce, random installation identity, callback registration, issued-ID persistence, ID-token validation, stable subject identity, granted scopes, and separation of login from inference permission.
- Retained cross-process refresh serialization with reload, atomic replacement before use, the unavoidable server-rotation/local-crash gap, six terminal errors versus transient failures, and bounded same-source recovery after 401.
- Retained per-model surface selection and per-route request capabilities; namespaces and call-result correlation; terminal completion before committing streamed items; empty-terminal-output recovery; and route-specific cache behavior. The plan route's options rejection was observed, while omitting paired markers is conservative.
- Retained summary event separation, accessible incremental delivery, final-only suffix recovery, default `auto` versus explicit `detailed`, and actual-model selection of pre-tool prose policy, including worker overrides. Public summaries do not satisfy ordinary assistant explanation requirements.
- Preserved proposed fallback versus accepted two-mode/no-fallback behavior, error-specific funding semantics, announcements, logout persistence, account catalog distinctions, profile isolation, and credential secrecy.
- Kept the scheduling proof and real tool-result requirement. Historical API-key success remains distinct from the subscription answer's launcher-induced kill and absence of graceful-save evidence. Cached-prefix predictions remain distinct from zero provider grants. Local fake evidence, historical paid evidence, and the October 22 restart are not interchangeable. Live OAuth was deferred; no fresh native GUI/browser/listening observation is claimed.

## Checks and carryover

Completed one full manuscript reread. The textual H1 check caught and corrected an incorrect draft title. Corrected the refresh boundary wording and made installation-ID privacy, credential isolation, and model-selected pre-tool explanation explicit. Checked only changed passages afterward. Simple checks confirmed exact H1, one H1, and no fenced code or tables. No implementation or runtime results were produced by this authoring pass.

Both owned files are stable and ready for independent review; remain available for material corrections. Chapter 17 was not inspected. Carry forward client-owned neutral history, offline rendering, request-local credentials, route-specific capability selection, terminal-success gating, separate public-summary events, and explicit funding authorization. Do not promote the synthetic model or historical paid runs into current live-provider proof, or infer that accepted automatic fallback exists.

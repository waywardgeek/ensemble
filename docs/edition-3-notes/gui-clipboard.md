# Copyable text snippets

This independent feature branch is based on `origin/main` at `123d6017`.
It ports only the accepted snippet-box feature from main `cedf90d7`;
timestamps, agent renaming, theme defaults and context optimizations are not
dependencies and are not included.

Completed Markdown fenced blocks in text replies become read-only textareas
with independent Copy buttons. Copying includes only the enclosed snippet,
not its fences, surrounding prose or the containing artifact. Thinking, tool
and user cards have no copy controls. Longer backtick fences can contain
shorter fences literally; payload whitespace, Unicode and HTML-like text are
not interpreted as markup.

Unclosed blocks remain noninteractive while streaming. For long replies the
preview offers no partial snippet controls; Full artifact renders complete
boxes. Unchanged boxes retain identity, focus and selection while later prose
streams. Authoritative finals and replay use the same renderer. There are no
core, provider, server or system-prompt changes.

The browser uses its native Clipboard API, then a synchronous copy fallback
for HTTP LAN sessions. When both are blocked, it selects the snippet for a
native keyboard copy. Production never reads the clipboard. Firefox tests
exercise real secure-context, insecure-HTTP and blocked-copy behavior under
explicit testing preferences, restoring the original clipboard afterward.
Those preferences are not evidence for default-permission behavior; async API
rejection followed by successful legacy copying remains a coverage limitation.

The matching main implementation passed native Firefox tests and a real local
model conversation, including a tool roundtrip, snippet-only copying, agent
selection, reload and process restart. Five targeted mutation checks protected
snippet scope, fence length, node reuse, card-kind isolation and preview safety.
Branch-specific validation is recorded separately below; main is left intact.

Branch validation: all three native Firefox clipboard scenarios pass with no
skips. GUI `go test ./... -short -timeout=120s` passes with the installed Chromium
path, and GUI `go vet ./...` passes. These runs use the independent upstream
base, not the combined main implementation. A first test-launch attempt lacked
the temporary Playwright dependency link; after correcting that test setup,
the real Firefox suite passed. No application dependency or source change was
needed, and the dependency link is not part of the branch.

Independent review (Galileo) accepts all eight owned paths with no blockers.
An independent Firefox probe also confirms that the upstream Thinking expander
cannot change Full artifact expansion state or expose reasoning copy controls.
The matching 30-line codebook additions preserve the original chapter and
assembled text; two inherited prose diagnostics remain nonblocking.

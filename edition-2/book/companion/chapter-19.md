# Chapter 19: The Sandbox

In Bill's October 2025 account, the CodeRhapsody secrets manager did exactly what he designed it to do. Credentials lived in an encrypted store; an approval dialog controlled their release. When Claude Sonnet 4 needed the sudo password, it ran `cat` and asked permission. Bill was following the reasoning and clicked Approve before noticing the consequence: the password had become a tool result. The agent never needed to ask again. Three hours later, he found it committed to Git. Scrubbing the history took longer than writing the secrets manager.

The mistake was in the data path. Approval released the credential into a conversation from which it could be repeated, logged, or committed. A better dialog might have caught this particular decision, but it would still have depended on Bill noticing it. He also reports mistaking Codex's unrestricted mode for autonomy inside a sandbox while starting the second edition. Repeated approvals had made full access attractive; the name had not supplied the missing boundary.

Chapter 18 gave the agent access to untrusted web content. Combined with private data and external action, that forms Simon Willison's [“lethal trifecta,” described in June 2025](https://simonwillison.net/2025/Jun/16/the-lethal-trifecta/). A poisoned page can ask the model to read a secret and send it elsewhere using legitimate tools. No exploit in the tools is necessary. This chapter restricts what those tools can do even when the model requests the wrong thing.

## The runtime is trusted; its instructions are not

Ensemble's design does not put the entire binary in a process jail. The binary is trusted host code, selected and compiled by the operator. It holds credentials and attaches them to provider requests. The untrusted material is model output, including tool calls influenced by retrieved content. The boundary confines the operations that output can request while leaving the host able to communicate with the model.

That is a specific threat model: a poisoned conversation driving legitimate tools. Compromised host code and operating-system escapes are outside it. The host, its chosen root, and its trusted configuration must remain outside the model's control.

Keeping secrets out of context closes one disclosure path, but not all of them. A model can ask a command to copy bytes it never sees. Conversely, denying network access to commands does not prevent the trusted host from sending model requests or a granted web tool from sending a query. Files placed inside the workspace and capabilities granted to the agent still determine what can leave through those routes.

Two kinds of operation need enforcement. File tools run inside the trusted Go process, so they use confined file operations. `run_command` launches arbitrary programs, which do not call Ensemble's path checker. Those processes and their descendants need an operating-system sandbox. Neither mechanism can substitute for the other.

The policy has one owner. The existing public `Config` carries `SandboxRoot`, `SafeMode`, and `EnableWebSearch`; Sandbox retains its required Agent parent and obtains the configuration through it. Tools reach that Sandbox through the Agent. A copied root in a helper or callback would create another policy value that could become stale. An empty root selects explicit full access; a nonempty root selects confinement.

## A checked path is not a confined operation

For a file-tool request, a relative name is interpreted under the sandbox root. Canonicalization removes traversal components and resolves symlinks, including existing ancestors when the requested file has not yet been created. The canonical candidate must be inside the canonical root as a directory relationship, not merely share a string prefix. `workspace-other` is not a child of `workspace`.

Canonicalizing only the candidate caused an early failure described in the source. On macOS, `/tmp` resolves to `/private/tmp`, and `/var` to `/private/var`. A root created under an alias no longer matched its own resolved contents, so legitimate accesses were refused. Both sides of the comparison need the same interpretation. The root is canonicalized at use rather than trusted as a permanently resolved string.

Even a correct comparison cannot make a later ordinary open safe. A process can replace a symlink between the check and the open. Worse, an interface that only returns a checked name invites callers to check one string and accidentally pass the original string to `os.ReadFile`. The common interface therefore exposes operations such as `ReadFile`, `OpenFile`, and `ReadDir`, not just validation.

The source implementation uses Go 1.25's `os.Root` file, directory, and rename operations, passing relative names to an opened directory root. The traversal resistance belongs to the operation itself, rather than to a prior promise about a pathname. Go's [traversal-resistant API explanation](https://go.dev/blog/osroot), cited by the source, explains that distinction. This uses the standard library's protection; it is not a new guarantee against every hard-link, mount, or hostile-native-code scenario.

The same reasoning covers less obvious reads and writes. A command can replace `save.json` with a symlink before the trusted host saves state. Media reads and later reads of retained job output can cross the boundary just as ordinary file tools can. They need the same confined operations. Application-selected skill files are a separate trusted input and may live outside the workspace; their location does not grant file tools permission to read there, nor does confinement remove secrets mistakenly placed in that trusted corpus.

Refusals have a smaller information boundary of their own. They report the caller's supplied string, not a resolved host path or expanded working-directory diagnostic. An error should explain the refusal without teaching the model the filesystem layout it was denied access to.

## The shell needs a kernel boundary

Redirecting `HOME` and removing environment credentials changes a command's defaults. It does not stop an absolute read of `/etc/passwd`. On macOS, the chapter uses Seatbelt through `sandbox-exec` to constrain the child process. Descendants inherit the sandbox, so adding another shell or launching Python does not bypass it. The trusted Ensemble process remains outside, and command output returns through the existing job supervisor.

The profile denies command network access. Applying a single network-enabled profile to the whole process tree would lose the distinction between the host's provider requests and a command's attempted connection. Confined commands on unsupported platforms fail closed with an unsupported-platform error; this chapter does not implement Linux or Windows confinement. `sandbox-exec` is marked deprecated in the macOS man page, so the source's successful local use is not evidence of indefinite availability or portability.

The implemented profile starts with default allowance, then denies file reads and writes globally and adds specific exceptions. An earlier deny-default experiment failed before `main`; importing `bsd.sb` did not rescue it. That showed missing startup allowances in that experiment, not the impossibility of a deny-default design.

Useful commands need more than their workspace. The profile allows reads of the root directory entry, system paths such as `/usr`, `/bin`, `/System`, `/Library`, and `/dev`, and the TLS trust store under `/private/etc/ssl`. It permits writes to selected device files, including `/dev/null`, which Git opens for both reading and writing. Required ancestor and system-selection metadata is allowed without opening sibling data. The workspace is readable and writable. This is not `chroot`: `/` still names the host root, and the system exceptions are real exceptions to workspace-only access.

Failures in those allowances can look like unrelated defects. A missing TLS configuration read looked like a network problem. In the restart record, Python and Git could not inspect `/var/select/developer_dir`; narrow metadata allowances restored both on that host. That result did not establish that every added literal was necessary, and shell-selection warnings remained even after the commands worked. The correction did not justify opening the whole selection subtree.

The Xcode shim cache exposed the opposite danger: allowing too much to make tools work. Git, Python, and Clang can pass through shims that use randomly suffixed `xcrun_db*` files in the per-user temporary directory. Allowing that whole directory also allows temporary workspaces belonging to other processes. A narrower-looking rule still failed during the second-edition review: a temporary-directory subpath rule combined with a basename pattern allowed a planted `xcrun_db-private-marker` inside an unrelated nested directory. A real command read its contents.

The correction anchored the allowance to immediate cache files under the physical temporary directory, preserving Git and Python without admitting the nested marker. Location mattered as much as the name. Physical paths mattered too: Seatbelt matches `/private/var`, not the `/var` alias returned by the temporary-directory query. An earlier investigation changed both canonicalization and a trailing slash; isolating the changes showed that canonicalization fixed the failure and the slash was irrelevant.

## Credentials stay with the transport

The trusted binary adds authentication at the HTTP boundary. Provider-specific headers carry the credential; the conversation carries the request and response content, not those headers. OAuth refresh likewise belongs to the credential provider inside the binary. The model neither receives the refreshed token nor participates in obtaining it.

A server can reflect a credential back in a successful response or an error. Logs are another route: a key written to `api.log` inside the workspace is now readable data. Known secrets must therefore be scrubbed before events, traces, logs, or model-visible results retain them. The source's logger replaces the entire known value with `[redacted]`, while transport handling includes current and renewed Bearer-token echoes. Transport failures expose a generic safe error or the context's cancellation error rather than credential-bearing details.

Scrubbing after writing is too late; the model may already have read the file. Matching known strings is still only a backstop. It cannot recognize every encoding, split disclosure, or credential unknown to the logger, and it does not justify placing a credential store in the workspace.

Commands receive an allowlisted environment rather than a blacklist of suspicious variable names. A blacklist misses an unusually named credential and can retain executable startup hooks. The confined environment keeps selected locale and timezone settings, an explicit system PATH, and workspace-rooted HOME, TMPDIR, and Git configuration. SSH agent access and arbitrary inherited variables are absent. HOME redirection makes ordinary lookups local; the kernel must still refuse absolute outside paths.

Git illustrates why redirection is often better than another exception. Its real user configuration can contain tokens in URL rewrites, credential helpers, or includes that lead elsewhere. The source instead points `GIT_CONFIG_GLOBAL` at a minimal generated workspace configuration and `GIT_CONFIG_SYSTEM` at `/dev/null`. Local commits get an identity without credential helpers, rewrites, or includes. The recorded local `init`, `add`, `commit`, and `log` operations worked without exposing the user's configuration.

The original credential discussion proposed delivering a needed secret directly to a trusted process's stdin, with only a confirmation returned to the model. That would avoid the October 2025 failure's direct path into context, but the recipient could still echo it, so output would need protection too. This chapter adds neither a `send_secret` tool nor a secret store. It establishes where credentials belong, not a new means of handing them to arbitrary programs.

## Narrowing capabilities is an operation

Confinement, safe mode, and web access are independent choices. The application exposes them as `--sandbox`, `--safe-mode`, and `--no-web-search`, with corresponding configuration and environment inputs. Removing shell tools does not confine file tools. Selecting a root does not disable web search. CLI and GUI application defaults enable search, while an explicitly supplied `Config` is authoritative: its zero `EnableWebSearch` value disables it.

Safe mode removes `run_command`, `send_input`, `wait_for_job`, and `kill_job` from declarations and dispatch; the source's accepted implementation also removes `tool_limits`. Ordinary file tools remain usable. Hiding a declaration is insufficient if a model can guess the name and call it anyway.

The first edition's initial safe-mode test found exactly that kind of reassuring non-boundary. `RemoveTool` normalized `run_command` to `runcommand` and deleted a registry key that did not exist. The entry under its real name survived. Tools with underscores stayed callable, and normal successful-tool tests gave no indication that removal was broken. The relevant property was absence and refusal after removal, not whether the removal function returned without error.

Web search has a different gate because Chapter 18 supplies it as a skill, not built-in tools. When disabled, `web-search` is not loadable, its MCP server does not connect, and its tools are not bridged in. This closes the named skill's route. It is not a general network policy for arbitrary custom tools or other MCP servers.

Later child agents need to inherit these choices without silently regaining authority. `Config.Child(dataDir)` starts from the parent configuration, changes the data directory, and clears parent-specific log and save paths. Copying restrictions avoids the ambiguity of constructing a fresh configuration whose false Boolean might mean either “explicitly disabled” or “nobody considered this field.” Untouched fields already have the parent's values.

`Clamp(parent)` then returns a configuration no wider than the parent's and reports any attempted widening. A child cannot turn off a parent's safe mode, turn on search the parent disabled, or select an empty or escaping root when the parent is confined. A canonically contained narrower root is allowed. The Boolean directions differ: safe mode is a restriction and narrows with OR; web search is a capability and narrows with AND.

Enforcement and diagnosis are separate. Even if the caller ignores the reported error, the returned configuration is narrowed. The caller must still use that returned value; the method cannot protect construction that bypasses it. This chapter prepares the policy operation, not a fleet of spawned agents, and it does not claim that a restricted child's text cannot influence its parent.

## What the observations establish

The source's first profile experiment appeared to pass every escape test because it denied everything, including reads inside the workspace. Positive controls exposed the useless result. Evidence for confinement needs both a refused operation and useful work that still finishes.

The outside fixture must exist and be readable by the host; failure to read a nonexistent file proves nothing. A network target must be reachable outside the sandbox and remain alive during the confined attempt. A compound shell command can end successfully after an earlier denial, so the last exit code alone does not establish each operation's outcome. Refused writes also need an absent file effect. These distinctions make the evidence about the boundary rather than about a model's description of it.

The source records earlier live sessions that completed file work and a local Git commit while refusing outside paths, symlinks, descendant reads, and command network access. Those sessions preceded the nested-cache and working-directory refusal corrections. Direct regressions covered the corrections; the account does not claim another paid session for them.

The restart record, dated October 22, 2026, describes Sonnet 4.6 sessions through the API-key CLI. Paired results and artifacts showed file work, Python, and Git alongside traversal, symlink, and outside-write refusals. A Python-to-shell-to-`cat` chain failed to read outside. A confined network attempt added no request to a reachable local peer, while the host continued provider communication. A known synthetic token was filtered from returned output and absent from retained logs. A narrower safe-mode session retained file work, refused a parent read, and omitted command and job tools from declarations. The completed sessions exited normally at EOF. These are recorded observations, not a new runtime verification; they did not exercise the deferred OAuth/browser flow or establish universal isolation.

Full access remains an explicit option for work that confinement would prevent. Bill reports using it during a late-2025 smart-device compromise at home. Unexpected speech from a Wiim speaker led him to change the wifi password and investigate. A supervised agent read configurations, scanned the network, checked logs, and helped respond across a dozen devices. He reports a much faster response, without a retained timing measurement. Private data, potentially hostile logs, and external reconfiguration were all essential to the task.

That account explains the choice to accept broader exposure, not a proof that observation makes it safe. The sudo-password incident came from a person watching closely too. Where a confined workspace fits the job, the boundary takes a decision away from the poisoned conversation that neither an approval dialog nor a prompt can reliably enforce.

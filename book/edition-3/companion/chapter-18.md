# Chapter 18: The Sandbox

In October 2025, Bill built a secrets manager around an approval dialog. Credentials stayed in an encrypted store until the Agent asked and he approved. The first time the Agent needed his sudo password, it ran `cat`. He followed the reasoning and clicked Approve before the implication landed. Three hours later, he found the password committed to git.

That is his reported historical incident, not a fresh attack against the C# implementation. Its useful detail is that approval worked. The human supplied the permission while missing the consequence. More persuasive wording on the same button would not have established a file boundary.

Chapter 17 completed the dangerous combination: private data, untrusted content, and external action. A poisoned page can ask an Agent to read a credential and send it elsewhere using legitimate tools. This chapter constrains those tools without pretending that a prompt can make hostile instructions mechanically unenforceable.

## The host is trusted; the conversation is not

The binary needs its provider credential and an HTTP connection to the model. Confining the entire application as though those were accidental privileges misses the intended boundary. The threat here is a poisoned conversation driving tools, not compromised host code.

Each Agent strongly owns one Sandbox. Common declares the relationship, persistent root and handle state, and callable operations; functional partial sections implement the same object. Tools reach it through Agent, Jobs through their owners, and Engine through its parent. There is no copied root in each tool and no interface wrapper joining declaration to implementation.

An empty root retains full access. A nonempty root must exist and confines file tools, file-backed media, Jobs output, and History persistence. Safe mode is a separate choice: it removes shell and job-control tools. Web authority separately controls whether the web-search skill can load at all. None of these choices is an instruction asking the model to behave as though a capability were absent.

Launch choices are immutable. Hosts translate their explicit settings; the library does not inspect ambient environment. Neither host changes the process-wide working directory. A model switch preserves the chosen policy.

The two enforcement mechanisms are deliberately different. File tools run inside trusted C# and use Sandbox's confined operations. A shell never calls that code. Commands and their descendants need an operating-system boundary too.

## Resolving a name does not grant access forever

A path resolver follows existing links, appends missing components without creating them, and checks containment. An ordinary inward symlink can work; an outward link refuses. Relative-path containment avoids accepting a sibling such as `project-old` merely because its name begins with `project`. Refusals report the requested spelling rather than disclosing the resolved outside target.

Resolution alone leaves a check/use race. Between validation and open, another process can replace a component with a symlink. The earlier answer to “where does this name point?” does not authorize whatever it points to now.

The current native file layer retains an open handle to the physical root. Actual operations repeat policy resolution and open each component relative to retained directory handles with no-follow semantics. The final open refuses a replacement link too. A race may make a legitimate operation fail; it must not trigger a fallback to ordinary host path access.

Editing retains one read/write handle across reading, exact-anchor validation, and replacement. New files use exclusive creation instead of checking existence and then racing to create. Directory enumeration returns names, not reusable access grants; every later descent or read goes through Sandbox again. Recursive search tracks canonical directories to avoid inward-link cycles.

Native details produced real C# failures. A runtime diagnostic FIFO inside the root made a blocking open hold the Sandbox gate and stall search. Acquisition now uses nonblocking opens and rejects non-seekable handles. Search skips refused or vanished entries while retaining ordinary matches. The Apple ARM64 calling convention also required special handling for creation permissions, and append needed the managed stream position initialized at the end. A declaration that compiled had not established either behavior.

The same owner covers less obvious paths. Engine resolves file-backed media and cache captures. Jobs confines both its output writer and later readback; protecting only creation would leave observation as another escape. History uses confined access for snapshots, replacement files, backups, journals, archives, and logs while retaining its write-before-publication rules. An append refusal cannot publish a new event or reuse its sequence.

Relative host-state defaults live beneath the selected root. An explicit outside path refuses rather than being silently moved. Failed GUI settings persistence changes neither the old file nor the live value. Model-facing Job reports use rooted-relative output names; trusted host metadata can retain the real path without broadcasting it to the conversation.

## The shell gets a kernel boundary

On macOS, Sandbox prepares a launch through the fixed `/usr/bin/sandbox-exec` path and an in-memory Seatbelt profile. It does not start or own the process. Job keeps responsibility for PTY, stdin, output, cancellation, process groups, and joining.

The profile denies filesystem reads, writes, and networking, then allows the root and the system material necessary for useful local commands. Ancestors receive metadata-only traversal, not sibling-directory reads. Toolchain support uses a narrow exception for immediate Darwin cache files, not writable access to the shared temporary directory. Physical paths matter: `/var` commonly resolves to `/private/var`, and a policy written for the wrong spelling can deny perfectly legitimate work.

The confined environment is rebuilt from a short allowlist. It supplies a fixed system path, ordinary locale and terminal settings, rooted home and temporary directories, and generated Git configuration. Startup hooks, injected libraries, arbitrary tokens, and inherited provider credentials do not ride along. Git gets a minimal local identity without reading the user's configuration, whose includes, helpers, aliases, and signing commands can carry far more authority than a name and email.

Process control has its own injection risk. The trusted PTY bootstrap reports its process group through one bounded, strictly parsed startup line before command input and output are admitted. A writable `.pgid` file would let the command nominate an unrelated process for the host to signal. A malformed handshake refuses and stops the child.

The historical first confinement profile denied everything. Every escape failed—and so did reading a file inside the workspace. That apparent success was useless. The current source pairs refusals with useful local file, Python, Git staging, stdin, and cancellation behavior. A failed network request only establishes denial when the same owned target remains reachable outside the profile.

`sandbox-exec` is deprecated but is the mechanism exercised here. Confined commands require macOS; unsupported execution does not retry unconfined. The current native evidence is macOS ARM64, not a portable security claim.

## Close the diagnostic side door

The host can correctly withhold its bearer from tools and still leak it through a readable API log. A provider may echo a credential in diagnostic data; redacting when `read_file` opens the log is too late because another reader can obtain the same bytes.

Known credentials are masked before diagnostic writers receive them. Ensemble's shared scrubber obtains configured keys through live Agent relationships, handles overlapping values longest-first and common JSON spellings, and covers labels as well as payloads. A separate GUI log writer initially escaped both confinement and masking; observing the actual outside file exposed the missed path.

Plan bearers remain local to the Engine exchange rather than becoming Configuration fields or a new global token cache. The one permitted 401 retry can involve both rejected and rotated bearers; both pass through the trace scrubber for that exchange. Remote HTTP error bodies remain omitted at their source.

This does not rewrite successful provider text, tool content, or canonical History. Unknown secrets and arbitrary encodings remain outside known-value masking. Trusted skills, assets, explicitly selected credential storage, and extra MCP services are host inputs. A stdio MCP server does not automatically inherit the command profile. The shell's denied network is not a ban on every external action an Agent can request through its other granted tools.

## Children can narrow, not widen

Policy derivation prepares for the fleet without constructing one. A child can inherit the root or choose a descendant, but cannot escape through an existing symlink. Validation can recognize a future directory without creating it; actual Agent construction still needs the root to exist. Safe mode combines restrictively, and web authority survives only if both parent and requested child permit it.

The important return value is the effective Configuration. A widening request produces a refusal *and still returns the narrowed value*. If a caller ignores the alarm, the wall remains. Returning the dangerous request beside an error would make every caller part of the enforcement mechanism.

The first edition demonstrated the difference between displayed policy and real authority. Its tool-removal function deleted a normalized name that was not the registry's actual key. `run_command` remained callable while safe mode could look enabled. Absence must reach dispatch, including a guessed call, rather than stop at the advertised tool list.

Current C# results exercise local hosts, rooted file effects, diagnostic masking, and the macOS command boundary. Earlier model-led sandbox sessions are historical evidence with their own scope. None excludes privileged mount changes, host compromise, or every possible escape. What this design provides is more precise: a defined set of file and command operations cannot acquire broader authority merely because the model supplies a convincing reason.

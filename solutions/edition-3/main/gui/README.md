# Embedding the GUI

The GUI is the separate `ensemble/gui` module. The core `ensemble` module does
not import it. Construction does not start an HTTP server or an MCP listener:

```go
view, err := gui.New(root, selectedAgent, logWriter)
if err != nil { return err }
defer view.Close() // closes view/relay resources, not the application-owned Agent
handler := view.Handler("") // embedded assets; an explicit directory overrides them
// Mount handler or serve it using the application's HTTP owner.
```

`selectedAgent` must be owned by `root`; nil/typed-nil inputs are rejected.
Settings read and update that selected Agent only. Viewing another Agent, replay,
or later child registration does not widen that Agent's runtime policy.
`gui.Assets()` exposes the rooted embedded `fs.FS` independently of the shell.
The GUI executable defaults to these embedded assets, regardless of cwd.

`view.ServeMCP(listener)` uses an application-created `net.Listener`. Each accepted
session binds to one connected browser and has an independent routing key, so
JSON-RPC IDs can overlap between sessions. Requests without a browser receive
correlated errors. `view.Close()` closes listeners and sessions. This does not
replace the core's separate stdio MCP transport.

Applications can explicitly present human questions using
`root.SetHumanQuestionSink(view.Hub.PresentHumanQuestion)`. The library owns
question correlation and `RespondToHumanQuestion`; the view only presents/replies.
Do not install multiple competing sinks for the same application.

## Reusing rendering without the stock page

These files are public assets and are also imported by the default page:

- `part-renderer.js`: `renderPart({frame, node}, {onSpeak} = {})` renders a public
  Wire part into a caller-owned node. It does not select Agents or choose layout.
- `tool-result.js`: `renderToolResult(part, open = false)` returns the result
  preview/full-details element; the complete text is retained.
- `artifact-scroll.js`: `createArtifactScroll(container, {followButton,
  threshold = 100} = {})` owns follow behavior for that container. Its result has
  `nearBottom()`, `follow(wasNear)`, and `destroy()`.

For example, a custom page can import `renderPart` and `createArtifactScroll`,
measure `wasNear` before rendering into its own node, and call `follow(wasNear)`
afterwards. The renderer's optional `onSpeak(agentID, partID)` is supplied by the
embedding application. No application-global selection, socket, or DOM IDs are
required by these modules.

## Fleet construction

The application installs an `ensemble.AgentFactory`. It composes inherited
Config, provider/auth/skills, and distinct child state paths, then calls
`root.NewAgent(config)`. The library admits immutable ID/parent/schema, owns
parentage, questions, Jobs, wait/join outcomes and recursive shutdown. Until
successful admission the factory owns its intended construction: on failure it
must clean that exact Agent, or return `Agent, err` so the library can roll back
that attempt. It must not return foreign/preexisting Agents, and remains
responsible for unrelated resources it creates. This is not a transaction over
arbitrary factory side effects.

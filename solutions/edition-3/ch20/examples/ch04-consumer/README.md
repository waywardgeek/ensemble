# Public-library consumer

This separate module imports `ensemble`, registers `shout` on its own Agent,
then asks the shared Agent loop to uppercase `hello framework`. The handler
writes through the normal job path and logs `shout called` through its actual
Agent and Ensemble parents. No internal imports or second execution loop.

Build from this directory with `go build .`. Configuration uses the public
`ConfigFromEnv` helper (`LLM_VENDOR`, `LLM_MODEL`, `LLM_API_KEY`, `LLM_BASE_URL`
and the established provider-specific fallbacks). `CH02_LOG`, when supplied,
receives the typed history after shutdown; each job's output locator identifies
its retained file. The default Agent data directory is independent of the core
CLI's `cr` directory. No provider credentials belong in source or history.

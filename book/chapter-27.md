# Chapter 27: Running Local Models

> **In development.** Local Models is the last closeout development slice.
> This placeholder supplies neither a verified setup recipe nor completed exercises.

This chapter will establish an explicit local inference endpoint, test buffered
and streaming responses, and exercise real tool conversations without cloud
fallback. Deterministic protocol tests and actual model usefulness are separate
acceptance gates. The first target is Ollama inside WSL; hardware, installed
models and server capabilities must be verified before making recommendations.

Junien's [PR 3](https://github.com/waywardgeek/ensemble/pull/3) supplies the
sleeping-inference-machine use case, and
[PR 6](https://github.com/waywardgeek/ensemble/pull/6) highlights explicit streaming
policy. Wake-on-LAN remains optional and separately authorized.

The [closeout plan](../docs/edition-3-notes/first-book-closeout-plan.md) separates
synthetic grading, real-target trials and human acceptance. No model download,
listener change or inference campaign is authorized by this placeholder.

Next: [Crossing Over to Ensemble](chapter-28.md).

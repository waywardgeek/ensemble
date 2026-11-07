# Local-model usage review for Julien

Ensemble now supports explicitly configured local inference in its ordinary GUI.
Bill verified a separate MiMo Agent could answer and call a tool that returned
results. That integration goal is accepted. This guide invites practical feedback
on Julien's own larger GPU setup; it does not ask him to reproduce Bill's machine
or prove that a small MiMo model is a good coder.

Thanks to Julien for requesting local-model support and for the nine PRs reviewed
during the earlier integration work. This review is about everyday use, not a new
student reconstruction or a requirement to run every internal grader.

## 1. Choose the server you already operate

Use a server with a compatible OpenAI Chat Completions endpoint and structured
function calling. Ollama is one tested server, not a runtime requirement. Other
servers still need their own compatibility check; an OpenAI-compatible label alone
does not prove tool or streaming behavior.

Record the server version, exact model/package name, quantization if applicable,
and configured serving context. Advertised maximum context is not the loaded
allocation. Choose a working limit that fits the actual allocation and workload.
No model download, GPU reconfiguration, public listener or firewall change is
required merely to review the feature.

The model runs on the inference server. Tools still execute on the Ensemble host,
under that host's authority and Sandbox policy. A prompt naming a project directory
is not filesystem confinement. Begin with a disposable, nonprivate task directory.

## 2. Supply an explicit model document

Create a private configuration file based on this version1 example. Replace the
wire model name, address and serving limit with your actual values. The example's
4,096 tokens are a placeholder, not a recommendation for your larger machine.
Keep local IDs stable because saved Agent selections refer to them.

```json
{
  "version": 1,
  "vendors": [{
    "id": "local",
    "label": "Local inference",
    "endpoints": [{
      "id": "home",
      "address": "http://127.0.0.1:11434",
      "surface": "chat_completions",
      "authentication": "none",
      "models": [{
        "id": "coding-model",
        "name": "REPLACE_WITH_EXACT_MODEL_NAME",
        "label": "My local model",
        "features": { "stream": true },
        "serving_context_tokens": 4096
      }]
    }]
  }]
}
```

Here `stream: true` declares the selected server/model supports streaming. Verify
that contract; do not copy speculative media or reasoning capabilities into the
file. The address is the server base URL, not a complete `/chat/completions` URL.
For a separate machine, use its approved reachable address and retain appropriate
network restrictions. Do not expose an unauthenticated inference API publicly.

`none` is explicit unauthenticated funding, not permission to borrow cloud keys.
If your server requires authentication, use the configured binding contracts in
[Chapter27](../book/chapter-27.md); do not put secrets into this model document.
The vendor label need not be OpenAI merely because the wire protocol is compatible.

Keep your normal launch options and add `--model-config /absolute/path/models.json`
when you deliberately restart Ensemble. Confirm the host uses the intended skills
root. The supplied `agent/skills/local-coding` primary skill is a small starting
point. Do not replace the main collaborator's selection merely to test this route.
Existing saved Agents retain their own selections and funding.

## 3. Check the ordinary workflow

- In Create Agent, choose the configured model and `local-coding`. The picker
  should show one vendor/model entry, not one copy per funding source.
- For a clean first check, disable that Agent's six memory bands and automatic
  recall, and save both forms. These forms still have Save buttons; funding
  defaults auto-save. Memory and recall can be evaluated separately later.
- Confirm its working-context limit reflects the configured serving allocation.
  The displayed usage is the last measured input, not a prediction of the next
  request. Missing usage is unknown, not zero.
- Send a short greeting. Then ask for one harmless tool on a small fixture, such
  as listing a directory containing only two test files. Look in Actions for the
  actual call and returned result. Tool-shaped prose is not tool execution.
- Stop there if your purpose is integration verification. If you want to assess
  coding usefulness on your hardware, choose one modest real task and its tests.
  Judge correct edits and actual test output separately from runtime support.

The small skill grants six coding tools plus the runtime's load_skill,
unload_skill and micro_handoff controls. It offers no additional loadable skills.
It is not the special six-schema harness used in the historical MiMo trial.
Commands can return a running Job; do not repeat uncertain effects. Inspect Actions
and report the handle rather than treating a running observation as failure.

Changing a default funding choice in Settings does not retarget existing Agents.
Select the model explicitly to apply a new default. Missing or ambiguous funding
must refuse without a cloud fallback. A stopped saved peer has recovery controls;
this is not a general repair editor for an invalid initial-Agent or host file.

## 4. Useful feedback

Please report only what you actually tried:

- Ensemble revision, host OS, inference server/version and exact model.
- GPU arrangement, quantization and configured serving context, if known.
- Whether the model and primary skill were easy to find and select.
- Whether replies visibly streamed and Actions showed a call and its result.
- Whether status, errors, focus, text size and tool output were usable with your
  preferred display and assistive workflow.
- Optional coding results: task size, tests actually run and practical latency.
- Any refusal, lost selection, unexpected cloud routing or confusing save behavior.
  Stop on an unexpected route rather than retrying with another payer.

A short sanitized description or minimal reproducer is enough. Do not send keys,
tokens, credential files, private source, full histories or raw API traces. State
which observations were not tested; no full transcript or benchmark campaign is
required. Controlled cancellation/restart tests are optional and should use owned
disposable work, never interrupt an important production run just for this review.

Bill will arrange the invitation. Preparing this guide does not publish a release,
contact Julien, or authorize work on his machine. The public Local Models checker
still has bounded, ungraded coverage; this human usage review adds a different kind
of evidence and does not manufacture a full certification score.

# Prepare MiMo for an ordinary GUI trial

This is an explicit example for Bill's previously tested Windows Ollama host ET.
It is not automatic LAN discovery or a portable default. No startup probe,
model download, credential fallback or inference is required to load it.

## What the host loads

`mimo-et.models.json` registers vendor `ollama`, endpoint `et`, and model
`mimo-v2_6-9b`. Its provider-facing name remains `mimo-v2.6:9b`. The endpoint is
`http://et.lan:11434`, using Chat Completions and the host's explicit `none`
authentication binding. The model streams text and has a configured serving
allocation of 4,096 tokens. This is the last observed allocation, not a fresh
inspection of ET or a promise about its advertised maximum. Media and reasoning
options are not declared. No cloud credentials belong to this offering.

The repository skill `agent/skills/local-coding/SKILL.md` grants read_file,
list_directory, search_files, write_file, edit_file and run_command. Ordinary
skill activation also exposes load_skill, unload_skill and micro_handoff. No
loadable skills are offered. This is six coding tools plus three standard controls,
not the exact six-schema experimental harness. A running command is not a failed
command; do not rerun it. Without wait_for_job in this small skill, report its
handle and let the human observe completion through Actions.

## Installation is a separate human decision

Keep the existing launch command and its funding, global-settings, skills,
workspace, data and public-origin options. After stopping the application and
waiting for its shell prompt, add this option to the approved launch command:

```text
--model-config /Users/bill/projects/ensemble/agent/Examples/Gui/mimo-et.models.json
```

The example path is specific to Bill's checkout. Other hosts must choose their
own approved path and endpoint. The skills root must include `local-coding`;
a launcher using the repository's `agent/skills` root will find it on restart.
Adding this document does not select MiMo for Cadence or change a funding default.
Do not edit the installed launcher, copy credentials or restart automatically.

## Before the first prompt

1. Confirm the GUI lists Ollama / MiMo 9B and the `local-coding` primary skill.
2. Create a fresh separate Agent with those choices. Keep Cadence on Astra.
3. In that Agent's Settings, disable all six memory bands and save. Separately
   disable recall and save. A small skill does not disable either policy.
4. Confirm the working-context limit is 4,096 input tokens. Start with a short
   task and no attachments. Last-request usage does not guarantee the next
   request will fit; request bytes are not token estimates.
5. Agree on a disposable task directory and inspect the inherited host Sandbox
   policy. A directory named in a prompt and a small tool list are not confinement.
   GUI peers inherit host policy; this setup adds no per-peer security boundary.
6. Only after approval, send a short greeting and request one harmless tool on a
   small controlled fixture. Confirm the actual call and result in Actions. Do not
   retry uncertain effects, install models, increase
   serving context, switch funding or fall back to cloud inference automatically.

The earlier bounded live MiMo coding experiment already passed. Bill has now also
accepted the ordinary GUI integration: a greeting and an observed tool call/result
worked. No further MiMo coding trial is required. Use
[the evaluation guide](../../../docs/local-model-evaluation.md) for another host;
model quality is a separate optional question.

## Offline preparation evidence

An outside-assembly consumer loaded the actual model document, activated the
actual skill, selected the route and rendered without inference. It confirmed
explicit no-auth funding, streaming, a 4,096-token working limit and exactly nine
tool schemas. With one short synthetic prompt and memory/recall disabled, that
request was 7,818 UTF-8 bytes; no provider token count was measured.

A separate isolated real GUI host exposed MiMo and the small skill, created a
peer through the public HTTP route, and retained the mounted Agent's old model.
It shut down cleanly; versioned event logs contained no RequestSent events.
Evidence is under the directory recorded by
`/private/tmp/cadence-mimo-setup-current`. No ET request, live settings change,
installation or restart occurred during preparation. This is source/loader and
host-route evidence, not live MiMo readiness or browser usability acceptance.

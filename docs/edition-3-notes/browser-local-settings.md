# Browser-profile settings audit and migration

Bill authorized autonomous implementation after the settings audit. His explicit
clarification is localStorage semantics, not a machine-wide device identity:
profiles and origins are independent; same-origin tabs in one profile share.
External publication remains paused. No install, restart, provider request,
credential read, native microphone, or host capture is part of this work.

## Ownership inventory

- Browser profile: theme; narration enabled, rate, voice and speech-pausing;
  now also font size and hold-to-talk shortcut.
- Tab: selected conversation, drafts, scroll, expanded cards/tree, selected
  mobile pane and current pane widths. No new persistence requirement for these.
- Selected Agent: model, tool-batch/context policies, memory band enabled/size
  settings, recall enablement/judge/model/threshold/quotas, goal file/delay/enablement.
- Shared user profile: name and description. Future authenticated multi-user
  identity is a separate design, not a browser-local display setting.
- Host: accounts, credentials, funding, transcription service authorization.

The audit found a separate budget-persistence mismatch: GUI edits apply to the
selected Agent but save in the shared GUI settings file; startup loads those
budgets into the root Agent. This is not fixed by this presentation slice and
must not be disguised as a reason to make policy browser-local.

## Implementation

`device-preferences.js` owns one `ensemble.device.v1` localStorage record.
Integer font sizes clamp to 12–32 pixels. The first server value seeds an unset
font once; local choices win subsequent frames. New shortcuts default Off and
accept Off/F8/Right Alt, without inheriting legacy server shortcut values. The
existing voice controller cancels recording on binding changes; its guards and
host funding are unchanged. The microphone button remains independent.

Same-origin storage events synchronize tabs, ignore other storage areas, and
reset receiving tabs to font18/shortcutOff on removal. Unavailable storage keeps
page-session choices. Both controls work offline and never send server settings
patches. Legacy server fields remain for old clients/files. Current speech rate
was already local and remains consumed directly at utterance creation.

The Settings dialog labels browser, selected-Agent, shared-user, and host scopes.
Chapters7/8/26/27 teach these choices at their original construction points,
including Off as the initial shortcut and profile/origin separation. The old
shared-font save example is replaced by an Agent tool-budget example.

## Verification

- GUI build and fixture build: zero warnings/errors.
- New real-page test: two profiles, same-origin tabs, actual document font and
  live voice-controller binding/ARIA, reload, conflicting server values, offline
  changes, storage removal, blocked-storage fallback, invalid/bounded values,
  and zero backend writes for local controls.
- Negative control serves the old app.js only in the disposable test browser:
  it fails the Off assertion with server AltRight. The current page passes.
- Existing automatic-pause/local-speech suite passes; no pause-policy change.
- Five public GUI groups and 18 voice-shortcut guard cases pass.
- Settings dialog preserves local appearance across reload while shared profile,
  budgets, memory and recall retain acknowledged saves; six category layouts,
  both themes, narrow/enlarged text, focus and drafts pass.
- Chapters7/8/26/27 pass hard prose lint, with style advisories retained.

Final exact-asset profile and dialog tests passed, followed by joined fixture
shutdown and a successful complete publication build (13 publication regressions,
520 interior/521 reading pages, 31 contents references, 106 reading links, and
1402 EPUB local links/resources). All five generated files were byte-compared
into book/. Final evidence: `/private/tmp/cadence-device-acceptance.yI4X9G`; book/release.json there contains hashes.
The earlier book-release-2026-11-03 receipt remains historical; this receipt
supersedes its artifact hashes. Tests use isolated hosts and browser
profiles with joined cleanup, not Bill's live environment. No live phone/laptop
acceptance is claimed.

Lesson: browser preferences, tab state, Agent policy and host authority are four
different scopes. A local preference migration must change both the editor and
its consumer, prevent server replay from overriding it, and revise the original
teaching contract rather than only document a repair.

## Updated artifact receipt

```json
{
  "local_epub_links": 1402,
  "interior_pages": 520,
  "reading_pages": 521,
  "files": {
    "the-art-of-building-ai-coding-agents.md": {
      "bytes": 1168263,
      "sha256": "c2f7b0391dfc29494f80e6aea318e8ed8f8d42a8ebe64b4c014464de154c221a"
    },
    "The Art of Building AI Coding Agents.epub": {
      "bytes": 17101643,
      "sha256": "93356bfa6cdb92046b06f5e208c89373d32a1a5ff0a8344352c90cb355578135"
    },
    "The Art of Building AI Coding Agents - Print Interior.pdf": {
      "bytes": 2443218,
      "sha256": "6085deeba81bbea6eba59dc539c1f0beec8e3d811854e44af1a533f2f4a424ae"
    },
    "The Art of Building AI Coding Agents.pdf": {
      "bytes": 19645053,
      "sha256": "9180c04c5b294b7415ec195676fd6e6922a18a4a88e84faea99040bb9c9299e4"
    },
    "interior-print.html": {
      "bytes": 1528415,
      "sha256": "70f1e6345f19434212934cd5ae196e83b9ce6e400e276cc82c685f20f57241f0"
    }
  }
}
```

# Advanced Artifacts figures

Authorized by Bill on 2026-11-07. Actual implemented browser views with synthetic
inputs, captured in disposable fixtures. No live conversation, host screen, provider
request or installed application modification. Crops preserve the actual component
styles; no pixels retouched. Dark appearance is the application fixture default.

Capture source: `agent/GuiTests/advanced-artifact-figures.cjs`. Start the existing
GuiTests `--artifact-renderer-fixture DIRECTORY` with an isolated environment, wait
for its `url` file, and set `GUI_FIXTURE_URL` to that URL. Run with the installed Node:

```sh
node agent/GuiTests/advanced-artifact-figures.cjs book/figures/advanced-artifacts DISTRIBUTION/skills/shared
```

The capture requires the existing Playwright dependency and installed Chrome; it
downloads neither. Always create the fixture `stop` file and await host exit in a
finally block. The capture closes its browser in its own finally block. The run
used the previously built isolated GuiTests assembly; no runtime source changed.
All capture runs joined their host. Final captures passed content checks and were
inspected as pixels. This is not screen-reader or human usability acceptance.

The command response contains exactly 65536 synthetic bytes; its illustrative test
count is not an Ensemble test receipt. The approval candidate is staged, not tried
or kept. Dice runs from the copied distribution fixture, not a new game tool.

## File identity

- `approval.png`: 1688 × 1052; SHA-256 `f27767a3ed4a0d8c72ad275fb0c8d280cddf0eec11eba6c8ad1b5076a9cf6bf4`.
- `command.png`: 1338 × 1134; SHA-256 `95e65ef4e1a1c23176e48af837558fcd5d29a951d92eebe41da56a7c308eea8a`.
- `directory.png`: 1318 × 294; SHA-256 `c48ec57219f0a26ac83429db510b81042c9822ca3fd75f24967711321e121242`.
- `inline-dice.png`: 1552 × 510; SHA-256 `2d4997bf7c19060b51f5e35851d10fe263a13eab4b07e648dfcccd8ca2dfe401`.
- `replacement.png`: 1318 × 442; SHA-256 `a3488c288f608d7526fe606859ea414b9a8a0b50108febf1ca5e02f99ac7b7e4`.
- `search.png`: 1318 × 368; SHA-256 `8f769bb10b3638c858b7467ad7def91628f62ad223652970999f472f94ef387e`.

Fixture assembly SHA-256: `4a1ce27ce3b0bffe0dd52b43074bf53db2a7b20e231e60e61d1731a4c3488daa`.

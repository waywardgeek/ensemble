---
name: host-screen
type: loadable
description: Explicitly authorized one-shot screenshots of the server computer, not the browser or phone.
tools: list_displays screenshot
---
# Host screen capture

These tools access the computer running Ensemble. Loading this skill grants host-screen authority; ordinary file access, code-tools and browser access do not imply it. Safe mode refuses both tools. OS screen-recording permission is also required. Do not elevate privileges or bypass permission.

Use `list_displays` for current zero-based display indices and logical bounds. `screenshot` with no selection captures primary display 0; `display` chooses another. Alternatively provide all of `x`, `y`, `width`, `height` and no `display`. Negative secondary-display origins are valid. The current backend is macOS-only and refuses rectangles spanning displays or lying outside them; other platforms fail explicitly.

A screenshot returns actual bounded, resized PNG pixels with original and prepared dimensions through the normal tool-image path. A filename or text description is not a screenshot. Model tool-image capability is checked before capture. Temporary files and the native helper are owned and joined on cancellation and failure. An observation delay may return a Job; use its handle to collect the result. Stopping observation does not cancel capture; explicit Job cancellation does.

Capture only the display or region the human authorized. No background monitoring or speculative private-window capture. Announce native tests and agree synthetic content/region first. This needs no Chrome picker and does not capture the requesting phone. macOS permission must be handled by the human; preflight refuses without opening a permission prompt or changing settings.

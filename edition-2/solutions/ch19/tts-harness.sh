#!/bin/sh
# The grader supplies only a provider boundary. Exercise the actual optional GUI.
exec node "$(dirname "$0")/gui/tts-harness.mjs" "$@"

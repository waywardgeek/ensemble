#!/bin/sh
# Public executable speech-check entry point. No build or installation side effects.
set -eu
HERE=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
if [ "${1:-}" = --describe ]; then
    printf '%s\n' '{"read_file_tool":"read_file","read_file_arg":"path","supports_type_during_turn":true,"supports_streaming_off":true}'
    exit 0
fi
: "${TTS_NODE:?Set TTS_NODE to the existing Node executable}"
exec "$TTS_NODE" "$HERE/tts-harness.mjs" "$@"

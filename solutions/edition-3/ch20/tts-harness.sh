#!/bin/sh
# Chapter 13 §13.9: real GUI + owned private browser; no dependency installation.
# tts-harness.sh PROMPT [--type-during-turn TEXT]
# LLM_BASE_URL, WORKSPACE, TTS_LOG are required for a normal run.
set -eu
root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cd "$root/examples/ch13-speech"
exec go run . --source-root "$root" "$@"

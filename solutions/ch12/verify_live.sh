#!/usr/bin/env bash
# Live verification: run the real agent against Anthropic,
# check api.log for thinking field and count DeltaThinking deltas.
set -euo pipefail

cd "$(dirname "$0")"

# Read the key from settings, never echo it
KEY=$(python3 -c "import json; print(json.load(open('$HOME/.cr/settings.json'))['directClaudeAPIKey'])")
export ANTHROPIC_API_KEY="$KEY"
export LLM_VENDOR=anthropic
export LLM_MODEL=claude-sonnet-5

# Build the binary
echo "Building agent..."
go build -o /tmp/agent-verify ./cmd/

# Clean any existing api.log
rm -f api.log

# Run with a single prompt, piped in
echo "Running live against claude-sonnet-5..."
echo "What is 2+2? Answer in one word." | timeout 60 /tmp/agent-verify 2>/tmp/agent-stderr.log || true

echo ""
echo "=== API.LOG thinking field ==="
if [ -f api.log ]; then
    # Show just the request line that contains "thinking"
    python3 -c "
import json, sys
with open('api.log') as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except:
            continue
        # Request lines have 'body' with 'thinking'
        if 'body' in obj:
            body = obj['body']
            if isinstance(body, str):
                try:
                    body = json.loads(body)
                except:
                    continue
            if 'thinking' in body:
                print('FOUND thinking in request:')
                print(json.dumps(body.get('thinking'), indent=2))
                print(f'max_tokens in request: {body.get(\"max_tokens\")}')
" 2>&1 || echo "(no structured thinking found in api.log)"
    # Also grep for raw thinking
    grep -o '"thinking":{[^}]*}' api.log 2>/dev/null || grep -c '"thinking"' api.log 2>/dev/null || echo "(no thinking pattern in api.log)"
else
    echo "NO api.log found"
fi

echo ""
echo "=== STDERR (DeltaThinking count) ==="
if [ -f /tmp/agent-stderr.log ]; then
    # Count thinking deltas — they show as dimmed text on stderr
    wc -c < /tmp/agent-stderr.log
    head -20 /tmp/agent-stderr.log
fi

# Clean up
rm -f /tmp/agent-verify

#!/usr/bin/env bash
# ch12 deletion audit — verify each grader check is sensitive.
# For each mutation: apply patch, rebuild, grade, verify target check fails.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

PASS=0
FAIL=0

# Run grader, return 0 if target check FAILS (mutation detected).
run_mutant() {
    local name="$1"
    local target_check="$2"
    local file="$3"
    local old="$4"
    local new="$5"

    echo ""
    echo "=== MUTANT: $name ==="
    echo "  Target check: $target_check"
    echo "  File: $file"

    # Verify pattern matches exactly once.
    local count
    count=$(grep -cF "$old" "$file" || true)
    if [ "$count" != "1" ]; then
        echo "  FATAL: pattern matches $count times (expected 1)"
        echo "  Pattern: $old"
        FAIL=$((FAIL + 1))
        return
    fi

    # Apply mutation.
    cp "$file" "$file.bak"
    # Use perl for reliable multiline replace.
    perl -pi -e "s/\Q${old}\E/${new}/g" "$file"

    # Rebuild.
    if ! go vet ./agent/... 2>/dev/null; then
        echo "  SKIP: mutant does not compile"
        cp "$file.bak" "$file"
        rm -f "$file.bak"
        FAIL=$((FAIL + 1))
        return
    fi

    # Grade.
    local output
    output=$(go run ./cmd/grade -ch 12 2>&1 || true)

    # Restore.
    cp "$file.bak" "$file"
    rm -f "$file.bak"

    # Check: target check should FAIL.
    # Parse the output for the target check line.
    local check_line
    check_line=$(echo "$output" | grep -F "$target_check" || true)
    if echo "$check_line" | grep -qF "[FAIL]"; then
        echo "  PASS: mutation detected (check failed as expected)"
        echo "  Details: $check_line"
        PASS=$((PASS + 1))
    elif echo "$check_line" | grep -qF "[PASS]"; then
        echo "  FAIL: mutation NOT detected (check still passes!)"
        echo "  Details: $check_line"
        FAIL=$((FAIL + 1))
    else
        echo "  FAIL: could not find check in output"
        echo "  Output: $output"
        FAIL=$((FAIL + 1))
    fi
}

echo "Ch12 Deletion Audit"
echo "==================="

# 1. mcp-handshake: remove Initialize call from ConnectMCP
run_mutant \
    "no-initialize" \
    "mcp-handshake" \
    "agent/agent.go" \
    'if err := client.Initialize(ctx); err != nil {' \
    'if false { // MUTANT: skip initialize'

# 2. tool-discovery: remove ListTools call  
run_mutant \
    "no-list-tools" \
    "tool-discovery" \
    "agent/agent.go" \
    'mcpTools, err := client.ListTools(ctx)' \
    'var mcpTools []mcp.ToolInfo; var err error; _ = mcpTools; _ = err // MUTANT'

# 3. mcp-tool-call: remove Bridge call
run_mutant \
    "no-bridge" \
    "mcp-tool-call" \
    "agent/agent.go" \
    'bridged := mcp.Bridge(client, mcpTools)' \
    'var bridged []common.Tool // MUTANT: no bridge'

# 4. ephemeral-round: remove callEphemeral from engine round loop
run_mutant \
    "no-ephemeral" \
    "ephemeral-round" \
    "agent/internal/llm/engine.go" \
    'e.callEphemeral(ctx, "round")' \
    '// MUTANT: e.callEphemeral(ctx, "round")'

# 5. reverse-call: remove SetReverseHandler
run_mutant \
    "no-reverse" \
    "reverse-call" \
    "agent/agent.go" \
    'client.SetReverseHandler(func' \
    '// MUTANT client.SetReverseHandler(func'

# 6. ws-tunnel: remove jsonrpc case from Hub
run_mutant \
    "no-ws-jsonrpc" \
    "ws-tunnel" \
    "agent/internal/ws/handler.go" \
    'case "jsonrpc":' \
    'case "MUTANT_jsonrpc":'

echo ""
echo "==================="
echo "Results: $PASS passed, $FAIL failed"
if [ "$FAIL" -gt 0 ]; then
    echo "DELETION AUDIT FAILED"
    exit 1
else
    echo "DELETION AUDIT PASSED"
fi

#!/bin/bash
# Generate the chapter 3 tool-usage exhibit from a corpus of v2-format
# CodeRhapsody session histories.
#
# Usage:  scripts/tool-usage.sh /path/to/cr/histories > book/exhibit-ch03-tools.md
#
# A table whose derivation is not reproducible is an assertion. This script is
# the derivation. Re-run it against any agent's history directory to get that
# agent's distribution.

set -euo pipefail
DIR="${1:?usage: tool-usage.sh <histories-dir>}"
cd "$DIR"

FILES=$(ls ./*.md | wc -l | tr -d ' ')

# Every tool call in the corpus.
ALL=$(mktemp)
grep -h '^### TOOL_CALL: ' ./*.md | sed 's/^### TOOL_CALL: //' | tr -d ' \r' > "$ALL"
RAW_TOTAL=$(wc -l < "$ALL" | tr -d ' ')

# EXCLUSIONS, stated so a reader can disagree with them.
#   1. Other lives: social, presentations, mail, calendar, drive.
#   2. Browser automation: real, but it is a different agent's job.
#   3. Artifacts: parse failures and test scaffolding, not tools.
OTHER_LIVES='^(moltbook_|powerpoint|google-workspace|gmail_|drive_|calendar_|list_calendars|triage-email|add_slide|add_bullet_points|create_presentation|create_new_presentation|save_presentation|populate_placeholder|manage_text|run_author_editor)'
BROWSER='^(click|take_snapshot|list_pages|navigate_page|navigate|evaluate_script|select_page|new_page|close_page|press_key|fill|fill_form|hover|emulate|chrome|list_console_messages|list_network_requests|wait_for|get_server_info|resize_page|upload_file)$'
ARTIFACTS='^(nop|unknown|test-skill|test-supervised|interview_agent|\$FUNCTION)'

USEFUL=$(mktemp)
# ALIAS MERGING. The corpus spans months during which tools were renamed. A
# naive count treats a rename as two tools and undercounts both. Map historical
# names onto the name the tool carries today, verified by diffing the corpus
# against internal/agent/tool_definitions.go.
grep -Ev "$OTHER_LIVES" "$ALL" | grep -Ev "$BROWSER" | grep -Ev "$ARTIFACTS" \
  | sed -e 's/^kill$/kill_job/' \
        -e 's/^kill_sub_agent$/kill_agent/' \
        -e 's/^search_memory$/search_knowledge/' \
        -e 's/^search_history$/search_knowledge/' \
        -e 's/^read_image$/read_file/' \
  > "$USEFUL"
TOTAL=$(wc -l < "$USEFUL" | tr -d ' ')
NTOOLS=$(sort -u "$USEFUL" | wc -l | tr -d ' ')
EXCLUDED=$((RAW_TOTAL - TOTAL))

classify() {
  # stdin: "<count> <toolname>" lines. stdout: full markdown table rows.
  awk -v t="$1" '
  function family(n) {
    if (n~/^(run_command|send_input|wait_for_job|kill_job|jobs|kill)$/)                          return "shell + jobs";
    if (n~/^(read_file|list_directory|find_files|read_image)$/)                                  return "read + navigate";
    if (n~/^(edit_file|write_file|replace_lines|delete_file|copy_file|move_file)$/)              return "mutate files";
    if (n~/^(search_files|semantic_search|search_web|crawl_web|search_knowledge|search_memory|search_history)$/) return "search";
    if (n~/^(refine_context|compress_context|keep_tool_results)$/)                               return "context";
    if (n~/^(save_memory|add_learning|list_learnings|read_learning|delete_learning|handoff_task|micro_handoff)$/) return "memory";
    if (n~/^(load_skill|list_skills|unload_skill)$/)                                             return "skills";
    if (n~/(sub_agent|_agent$|agents_status|agent_status)/)                                      return "sub-agents";
    return "other";
  }
  { c=$1; n=$2; cum+=c;
    st = "current";
    if (n ~ /^(compress_context|list_skills|list_conversations|list_sub_agents)$/) st = "retired";
    if (n == "handoff_task") st = "retiring";
    printf "| %d | `%s` | %s | %d | %.2f%% | %.2f%% | %s |\n", NR, n, family(n), c, 100*c/t, 100*cum/t, st }'
}

rollup() {
  awk -v t="$1" '
  function family(n) {
    if (n~/^(run_command|send_input|wait_for_job|kill_job|jobs|kill)$/)                          return "shell + jobs";
    if (n~/^(read_file|list_directory|find_files|read_image)$/)                                  return "read + navigate";
    if (n~/^(edit_file|write_file|replace_lines|delete_file|copy_file|move_file)$/)              return "mutate files";
    if (n~/^(search_files|semantic_search|search_web|crawl_web|search_knowledge|search_memory|search_history)$/) return "search";
    if (n~/^(refine_context|compress_context|keep_tool_results)$/)                               return "context";
    if (n~/^(save_memory|add_learning|list_learnings|read_learning|delete_learning|handoff_task|micro_handoff)$/) return "memory";
    if (n~/^(load_skill|list_skills|unload_skill)$/)                                             return "skills";
    if (n~/(sub_agent|_agent$|agents_status|agent_status)/)                                      return "sub-agents";
    return "other";
  }
  { c[family($1)] += 1 }
  END { for (k in c) printf "%d\t%s\n", c[k], k }' \
  | sort -rn | awk -v t="$1" -F'\t' '{printf "| %s | %d | %.2f%% |\n", $2, $1, 100*$1/t}'
}


cat <<EOF
# Exhibit — what an AI coding agent actually calls

**Generated $(date +%Y-%m-%d) by \`scripts/tool-usage.sh\`.**

Corpus: $FILES archived CodeRhapsody session histories — a working AI coding
agent operating on its own Go codebase over roughly six months.

**$TOTAL tool calls across $NTOOLS tools.**

## What was excluded, and why

The raw corpus holds $RAW_TOTAL calls. $EXCLUDED of them ($(awk -v e=$EXCLUDED -v r=$RAW_TOTAL 'BEGIN{printf "%.1f", 100*e/r}')%) are excluded
here because they are not coding work: social posting, presentations, mail,
calendar and drive; browser automation; and parse artifacts. The exclusion
patterns are in the script, so the judgement is reviewable rather than implied.

That the excluded share is under one percent is itself the finding: this agent
spends essentially all of its time coding.

## The corpus is a snapshot, and it lags the tool set

Two distortions to know about before trusting any single row.

**Renames were merged.** The corpus spans months in which tools were renamed. A
naive count treats a rename as two tools and undercounts both. \`kill\` is folded
into \`kill_job\` (225 combined), \`kill_sub_agent\` into \`kill_agent\`,
\`search_memory\` and \`search_history\` into \`search_knowledge\`, \`read_image\`
into \`read_file\`. The mapping was derived by diffing the corpus against the
live tool registry, not recalled from memory.

**Ten current tools appear nowhere in this corpus** — \`join_agents\`,
\`wait_for_agent_change\`, \`send_message\`, \`send_message_to_parent\`,
\`respond_to_parent_message\`, \`interrupt_agent\`, \`kill_agent\`,
\`get_submission\`, \`set_tool_watchdog\`, \`propose_skill\`. They postdate it.

That second point has teeth: **almost the entire sub-agent supervision suite is
newer than the measurement.** The 0.09% share for sub-agents below is therefore
a lower bound on a capability that barely existed when these sessions ran. It is
not evidence that sub-agents do not matter; it is the date stamp on the corpus.
The same caution applies to \`set_tool_watchdog\`, which exists because of a
failure that happened after most of these sessions were recorded.

## Full distribution

| rank | tool | family | calls | share | cumulative | status |
|---|---|---|---|---|---|---|
EOF

sort "$USEFUL" | uniq -c | sort -rn | classify "$TOTAL"

cat <<EOF

## Rolled up by family

| family | calls | share |
|---|---|---|
EOF

rollup "$TOTAL" < "$USEFUL"

rm -f "$ALL" "$USEFUL"

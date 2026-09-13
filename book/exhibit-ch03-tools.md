# Exhibit — what an AI coding agent actually calls

**Generated 2026-09-13 by `scripts/tool-usage.sh`.**

Corpus: 511 archived CodeRhapsody session histories — a working AI coding
agent operating on its own Go codebase over roughly six months.

**70401 tool calls across 51 tools.**

## What was excluded, and why

The raw corpus holds 71032 calls. 631 of them (0.9%) are excluded
here because they are not coding work: social posting, presentations, mail,
calendar and drive; browser automation; and parse artifacts. The exclusion
patterns are in the script, so the judgement is reviewable rather than implied.

That the excluded share is under one percent is itself the finding: this agent
spends essentially all of its time coding.

## The corpus is a snapshot, and it lags the tool set

Two distortions to know about before trusting any single row.

**Renames were merged.** The corpus spans months in which tools were renamed. A
naive count treats a rename as two tools and undercounts both. `kill` is folded
into `kill_job` (225 combined), `kill_sub_agent` into `kill_agent`,
`search_memory` and `search_history` into `search_knowledge`, `read_image`
into `read_file`. The mapping was derived by diffing the corpus against the
live tool registry, not recalled from memory.

**Ten current tools appear nowhere in this corpus** — `join_agents`,
`wait_for_agent_change`, `send_message`, `send_message_to_parent`,
`respond_to_parent_message`, `interrupt_agent`, `kill_agent`,
`get_submission`, `set_tool_watchdog`, `propose_skill`. They postdate it.

That second point has teeth: **almost the entire sub-agent supervision suite is
newer than the measurement.** The 0.09% share for sub-agents below is therefore
a lower bound on a capability that barely existed when these sessions ran. It is
not evidence that sub-agents do not matter; it is the date stamp on the corpus.
The same caution applies to `set_tool_watchdog`, which exists because of a
failure that happened after most of these sessions were recorded.

## Full distribution

| rank | tool | family | calls | share | cumulative | status |
|---|---|---|---|---|---|---|
| 1 | `run_command` | shell + jobs | 25586 | 36.34% | 36.34% | current |
| 2 | `read_file` | read + navigate | 18098 | 25.71% | 62.05% | current |
| 3 | `edit_file` | mutate files | 11671 | 16.58% | 78.63% | current |
| 4 | `search_files` | search | 7247 | 10.29% | 88.92% | current |
| 5 | `write_file` | mutate files | 1616 | 2.30% | 91.22% | current |
| 6 | `refine_context` | context | 1167 | 1.66% | 92.88% | current |
| 7 | `send_input` | shell + jobs | 800 | 1.14% | 94.01% | current |
| 8 | `save_memory` | memory | 530 | 0.75% | 94.76% | current |
| 9 | `list_directory` | read + navigate | 530 | 0.75% | 95.52% | current |
| 10 | `find_files` | read + navigate | 526 | 0.75% | 96.26% | current |
| 11 | `semantic_search` | search | 497 | 0.71% | 96.97% | current |
| 12 | `replace_lines` | mutate files | 385 | 0.55% | 97.52% | current |
| 13 | `wait_for_job` | shell + jobs | 322 | 0.46% | 97.97% | current |
| 14 | `kill_job` | shell + jobs | 225 | 0.32% | 98.29% | current |
| 15 | `search_web` | search | 206 | 0.29% | 98.59% | current |
| 16 | `compress_context` | context | 154 | 0.22% | 98.81% | retired |
| 17 | `crawl_web` | search | 152 | 0.22% | 99.02% | current |
| 18 | `load_skill` | skills | 143 | 0.20% | 99.22% | current |
| 19 | `delete_file` | mutate files | 94 | 0.13% | 99.36% | current |
| 20 | `jobs` | shell + jobs | 80 | 0.11% | 99.47% | current |
| 21 | `list_skills` | skills | 66 | 0.09% | 99.57% | retired |
| 22 | `screenshot` | other | 51 | 0.07% | 99.64% | current |
| 23 | `search_knowledge` | search | 26 | 0.04% | 99.67% | current |
| 24 | `add_learning` | memory | 26 | 0.04% | 99.71% | current |
| 25 | `spawn_sub_agent` | sub-agents | 23 | 0.03% | 99.74% | current |
| 26 | `handoff_task` | memory | 23 | 0.03% | 99.78% | retiring |
| 27 | `list_learnings` | memory | 19 | 0.03% | 99.80% | current |
| 28 | `copy_file` | mutate files | 19 | 0.03% | 99.83% | current |
| 29 | `read_learning` | memory | 16 | 0.02% | 99.85% | current |
| 30 | `take_screenshot` | other | 14 | 0.02% | 99.87% | current |
| 31 | `get_sub_agent_status` | sub-agents | 13 | 0.02% | 99.89% | current |
| 32 | `kill_agent` | sub-agents | 12 | 0.02% | 99.91% | current |
| 33 | `delete_learning` | memory | 10 | 0.01% | 99.92% | current |
| 34 | `move_file` | mutate files | 9 | 0.01% | 99.94% | current |
| 35 | `unload_skill` | skills | 7 | 0.01% | 99.95% | current |
| 36 | `list_secrets` | other | 7 | 0.01% | 99.96% | current |
| 37 | `list_conversations` | other | 7 | 0.01% | 99.97% | retired |
| 38 | `agent_status` | sub-agents | 6 | 0.01% | 99.97% | current |
| 39 | `list_sub_agents` | sub-agents | 3 | 0.00% | 99.98% | retired |
| 40 | `list_displays` | other | 3 | 0.00% | 99.98% | current |
| 41 | `send_secret` | other | 2 | 0.00% | 99.99% | current |
| 42 | `watchdog_remove` | other | 1 | 0.00% | 99.99% | current |
| 43 | `watchdog_list` | other | 1 | 0.00% | 99.99% | current |
| 44 | `watchdog_add` | other | 1 | 0.00% | 99.99% | current |
| 45 | `wait_for_agent` | sub-agents | 1 | 0.00% | 99.99% | current |
| 46 | `shutdown_agent` | sub-agents | 1 | 0.00% | 99.99% | current |
| 47 | `report_bug` | other | 1 | 0.00% | 99.99% | current |
| 48 | `micro_handoff` | memory | 1 | 0.00% | 100.00% | current |
| 49 | `keep_tool_results` | context | 1 | 0.00% | 100.00% | current |
| 50 | `ask_agent` | sub-agents | 1 | 0.00% | 100.00% | current |
| 51 | `agents_status` | sub-agents | 1 | 0.00% | 100.00% | current |

## Rolled up by family

| family | calls | share |
|---|---|---|
| shell + jobs | 27013 | 38.37% |
| read + navigate | 19154 | 27.21% |
| mutate files | 13794 | 19.59% |
| search | 8128 | 11.55% |
| context | 1322 | 1.88% |
| memory | 625 | 0.89% |
| skills | 216 | 0.31% |
| other | 88 | 0.12% |
| sub-agents | 61 | 0.09% |

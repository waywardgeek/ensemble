# Tool usage distribution — measured corpus

**Generated 2026-09-13.** Corpus: `~/projects/coderhapsody/cr/histories`,
511 archived CodeRhapsody session histories, after Hewitt/Lyric sessions were
scrubbed. **71032 tool calls**, 108 distinct tools.

Regenerate (run inside any directory of v2-format history files):

```
grep -h '^### TOOL_CALL: ' *.md | sed 's/^### TOOL_CALL: //' | sort | uniq -c | sort -rn
```

## Full distribution

| rank | tool | calls | share | cumulative |
|---|---|---|---|---|
| 1 | `run_command` | 25586 | 36.020% | 36.02% |
| 2 | `read_file` | 18094 | 25.473% | 61.49% |
| 3 | `edit_file` | 11671 | 16.431% | 77.92% |
| 4 | `search_files` | 7247 | 10.202% | 88.13% |
| 5 | `write_file` | 1616 | 2.275% | 90.40% |
| 6 | `refine_context` | 1167 | 1.643% | 92.04% |
| 7 | `send_input` | 800 | 1.126% | 93.17% |
| 8 | `save_memory` | 530 | 0.746% | 93.92% |
| 9 | `list_directory` | 530 | 0.746% | 94.66% |
| 10 | `find_files` | 526 | 0.741% | 95.40% |
| 11 | `semantic_search` | 497 | 0.700% | 96.10% |
| 12 | `replace_lines` | 385 | 0.542% | 96.65% |
| 13 | `wait_for_job` | 322 | 0.453% | 97.10% |
| 14 | `search_web` | 206 | 0.290% | 97.39% |
| 15 | `compress_context` | 154 | 0.217% | 97.61% |
| 16 | `crawl_web` | 152 | 0.214% | 97.82% |
| 17 | `load_skill` | 143 | 0.201% | 98.02% |
| 18 | `kill` | 128 | 0.180% | 98.20% |
| 19 | `kill_job` | 97 | 0.137% | 98.34% |
| 20 | `delete_file` | 94 | 0.132% | 98.47% |
| 21 | `jobs` | 80 | 0.113% | 98.58% |
| 22 | `click` | 70 | 0.099% | 98.68% |
| 23 | `take_snapshot` | 69 | 0.097% | 98.78% |
| 24 | `list_skills` | 66 | 0.093% | 98.87% |
| 25 | `screenshot` | 51 | 0.072% | 98.94% |
| 26 | `chrome` | 44 | 0.062% | 99.00% |
| 27 | `moltbook_verify` | 43 | 0.061% | 99.07% |
| 28 | `list_pages` | 37 | 0.052% | 99.12% |
| 29 | `navigate_page` | 35 | 0.049% | 99.17% |
| 30 | `moltbook_read_post` | 33 | 0.046% | 99.21% |
| 31 | `moltbook_comment` | 26 | 0.037% | 99.25% |
| 32 | `add_learning` | 26 | 0.037% | 99.29% |
| 33 | `moltbook_status` | 25 | 0.035% | 99.32% |
| 34 | `moltbook_comments` | 24 | 0.034% | 99.36% |
| 35 | `spawn_sub_agent` | 23 | 0.032% | 99.39% |
| 36 | `handoff_task` | 23 | 0.032% | 99.42% |
| 37 | `evaluate_script` | 22 | 0.031% | 99.45% |
| 38 | `moltbook_mark_read` | 21 | 0.030% | 99.48% |
| 39 | `nop` | 20 | 0.028% | 99.51% |
| 40 | `select_page` | 19 | 0.027% | 99.54% |
| 41 | `list_learnings` | 19 | 0.027% | 99.56% |
| 42 | `copy_file` | 19 | 0.027% | 99.59% |
| 43 | `search_knowledge` | 18 | 0.025% | 99.61% |
| 44 | `moltbook_post` | 17 | 0.024% | 99.64% |
| 45 | `moltbook_feed` | 17 | 0.024% | 99.66% |
| 46 | `read_learning` | 16 | 0.023% | 99.68% |
| 47 | `interview_agent` | 15 | 0.021% | 99.71% |
| 48 | `take_screenshot` | 14 | 0.020% | 99.73% |
| 49 | `get_sub_agent_status` | 13 | 0.018% | 99.74% |
| 50 | `kill_sub_agent` | 12 | 0.017% | 99.76% |
| 51 | `delete_learning` | 10 | 0.014% | 99.77% |
| 52 | `triage-email` | 9 | 0.013% | 99.79% |
| 53 | `move_file` | 9 | 0.013% | 99.80% |
| 54 | `unload_skill` | 7 | 0.010% | 99.81% |
| 55 | `list_secrets` | 7 | 0.010% | 99.82% |
| 56 | `list_conversations` | 7 | 0.010% | 99.83% |
| 57 | `test-supervised` | 6 | 0.008% | 99.84% |
| 58 | `press_key` | 6 | 0.008% | 99.85% |
| 59 | `powerpoint` | 6 | 0.008% | 99.85% |
| 60 | `new_page` | 6 | 0.008% | 99.86% |
| 61 | `agent_status` | 6 | 0.008% | 99.87% |
| 62 | `add_slide` | 6 | 0.008% | 99.88% |
| 63 | `fill` | 5 | 0.007% | 99.89% |
| 64 | `wait_for` | 4 | 0.006% | 99.89% |
| 65 | `search_memory` | 4 | 0.006% | 99.90% |
| 66 | `search_history` | 4 | 0.006% | 99.90% |
| 67 | `read_image` | 4 | 0.006% | 99.91% |
| 68 | `google-workspace` | 4 | 0.006% | 99.92% |
| 69 | `create_presentation` | 4 | 0.006% | 99.92% |
| 70 | `add_bullet_points` | 4 | 0.006% | 99.93% |
| 71 | `unknown` | 3 | 0.004% | 99.93% |
| 72 | `manage_text` | 3 | 0.004% | 99.94% |
| 73 | `list_sub_agents` | 3 | 0.004% | 99.94% |
| 74 | `list_displays` | 3 | 0.004% | 99.94% |
| 75 | `send_secret` | 2 | 0.003% | 99.95% |
| 76 | `save_presentation` | 2 | 0.003% | 99.95% |
| 77 | `moltbook_upvote` | 2 | 0.003% | 99.95% |
| 78 | `list_network_requests` | 2 | 0.003% | 99.95% |
| 79 | `list_calendars` | 2 | 0.003% | 99.96% |
| 80 | `get_server_info` | 2 | 0.003% | 99.96% |
| 81 | `watchdog_remove` | 1 | 0.001% | 99.96% |
| 82 | `watchdog_list` | 1 | 0.001% | 99.96% |
| 83 | `watchdog_add` | 1 | 0.001% | 99.96% |
| 84 | `wait_for_agent` | 1 | 0.001% | 99.97% |
| 85 | `upload_file` | 1 | 0.001% | 99.97% |
| 86 | `test-skill` | 1 | 0.001% | 99.97% |
| 87 | `shutdown_agent` | 1 | 0.001% | 99.97% |
| 88 | `resize_page` | 1 | 0.001% | 99.97% |
| 89 | `report_bug` | 1 | 0.001% | 99.97% |
| 90 | `populate_placeholder` | 1 | 0.001% | 99.97% |
| 91 | `navigate` | 1 | 0.001% | 99.98% |
| 92 | `moltbook_follow` | 1 | 0.001% | 99.98% |
| 93 | `moltbook_delete_comment` | 1 | 0.001% | 99.98% |
| 94 | `micro_handoff` | 1 | 0.001% | 99.98% |
| 95 | `list_console_messages` | 1 | 0.001% | 99.98% |
| 96 | `keep_tool_results` | 1 | 0.001% | 99.98% |
| 97 | `hover` | 1 | 0.001% | 99.98% |
| 98 | `gmail_list_messages` | 1 | 0.001% | 99.99% |
| 99 | `fill_form` | 1 | 0.001% | 99.99% |
| 100 | `emulate` | 1 | 0.001% | 99.99% |
| 101 | `drive_search_files` | 1 | 0.001% | 99.99% |
| 102 | `drive_list_files` | 1 | 0.001% | 99.99% |
| 103 | `create_new_presentation` | 1 | 0.001% | 99.99% |
| 104 | `close_page` | 1 | 0.001% | 99.99% |
| 105 | `calendar_list_events` | 1 | 0.001% | 100.00% |
| 106 | `ask_agent` | 1 | 0.001% | 100.00% |
| 107 | `agents_status` | 1 | 0.001% | 100.00% |
| 108 | `$FUNCTION_NAME_NOT_PROVIDED` | 1 | 0.001% | 100.00% |

## Rolled up by family

More useful than 108 rows: what *kind* of work the agent does.

| family | calls | share |
|---|---|---|
| shell + job supervision | 27013 | 38.03% |
| read + navigate | 19154 | 26.97% |
| mutate files | 13794 | 19.42% |
| search | 8128 | 11.44% |
| context management | 1322 | 1.86% |
| memory + handoff | 625 | 0.88% |
| browser | 395 | 0.56% |
| other | 309 | 0.44% |
| skills | 216 | 0.30% |
| sub-agents | 76 | 0.11% |

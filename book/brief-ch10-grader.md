# Ch10 Grader — P9 Deletion Audit

## Checks (7 checks, 100 points)

| ID | Points | What it protects |
|----|--------|-----------------|
| initial-tools | 15 | Only primary skill's tools + load/unload visible at startup |
| load-skill | 15 | load_skill tool call adds new tools to declarations |
| progressive-disclosure | 15 | Loading a skill reveals new loadable skills |
| depends-autoload | 15 | Dependency skills auto-load, their tools appear |
| var-substitution | 10 | $VAR placeholders replaced in skill body |
| blocked-skill | 10 | Non-loadable skills rejected with error |
| ch9-parity | 20 | All ch9 checks still pass |

## Mutation Matrix

| Mutation | What was deleted | Failing checks | Score |
|----------|-----------------|----------------|-------|
| 1. no-skill-filtering | `IsToolEnabled` returns true unconditionally | initial-tools, load-skill | 65/100 |
| 2. no-dependency-loading | Dependency loop in `loadSkillRecursive` | depends-autoload | 85/100 |
| 3. no-var-substitution | `VarRegistry.Render` returns body unchanged | var-substitution | 90/100 |
| 4. no-blocked-check | `IsLoadable` returns true unconditionally | blocked-skill | 90/100 |

4/4 mutations caught. All protected behaviors are independently verified.

## Cascade Notes

- Mutation 1 (no filtering) kills both `initial-tools` (too many tools visible) AND `load-skill` (count doesn't increase because all tools already visible). Same root cause, two symptoms.
- Mutations 2-4 each kill exactly one check — clean 1:1 mapping.

## Gotcha

Go build cache serves stale binaries across worktrees with the same module path. Always `go clean -cache` before mutation testing, or use `-a` flag on `go build`.

# Chapter 12 validation

GUI observation and control chapter preparation, October 8, 2026. New Chapter 12
maps to first-edition Chapter 13. Bill requires an actual MCP tunnel over
WebSocket through the optional GUI module; Chapter 11 teaches its public,
transport-independent construction boundary. Bill's editorial approval remains
separate from technical validation.

| Gate | Owner | Status and evidence | Next action |
|---|---|---|---|
| Research and outline | Author `/root/coder_ch04` | Original outline/evidence frozen at `54806a6`; historical fake-server/default-browser discrepancy explicitly retained | Publish coordinator decisions and complete the student-facing contract |
| Contract/design review | Coordinator, author | Bootstrap, scoped views, Agent leases, direct request capture and strict policy identity extension accepted as working directions; frame-size extension explicitly chosen | Review exact wire, ordering, bounds and lifecycle contract before student release |
| Independent checks | Grader engineer | Not started | Prepare real-browser tunnel controls from the published contract; a fake stdio peer cannot establish the required capability |
| Student implementation and local checks | Fresh student, unassigned | Not started; accepted Chapter 11 source is a prerequisite | Freeze permitted inputs and review the student's ownership plan |
| Initial live use | Student, coordinator | Not started | Review complete bounded CLI/browser/public/all-provider plan before live calls |
| Historical comparison and revisions | Independent code reviewer | Not started | Preserve initial implementation, runs and teaching experience first |
| Manuscript and feedback | Author, student, proofreader | Outline only; no successful demonstration claimed | Reconcile actual evidence and resolve teaching/reviewer findings |
| Export and checkpoint | Coordinator | Not started | Complete all gates before export and immutable chapter tag |

## Working design decisions

These are coordinator choices under Bill's authorized architecture, not additional
decisions attributed to Bill. The full chapter must teach their precise behavior.

- Bootstrap an explicitly selected unbound view before discovering and freezing
  aliases, then mount the Agent's Page into that view. The proposed browser wait
  is bounded and cancelable; ordinary startup without debugging stays usable.
- Keep typed integration policy in Agent creation configuration. Skills retain
  their existing frontmatter and grant rules. Agent leases and canceled automatic
  collection cannot dispose another Agent's calls or the GUI-owned socket.
- Keep collected observations transient until Actor accepts the exact ordered
  samples in `request_sent`, atomically with existing request consumption. An
  abandoned pre-request collection leaves no pending sample for another attempt.
  Historical reconstruction selects recorded samples; pure prefix rendering
  performs no collection. External observations do not become system authority.
- Version the nonempty integration policy explicitly as session format 3;
  preserve the existing format-1 and format-2 contracts. Logical view scope is
  creation identity; physical browser instances and pending UI actions are not
  executable state restored from history.
- Bind controls and targeted artifact reads to their view/mount and appropriate
  data version. Preserve human draft ownership and require actual acknowledgments
  for applied settings. Report bounded omissions truthfully.
- Extend the physical frame bound explicitly for an enabled MCP tunnel while
  retaining the 65,536-byte limit for inherited commands. A dedicated envelope
  carries canonical base64 for one complete message, at most 8 MiB decoded and
  12 MiB for the envelope. The full contract must account for queued bytes and
  shared-socket scheduling; operation counts alone are insufficient. No chunk
  reassembly protocol is required.

The author assessed direct request capture against earlier pure-render and exact
reconstruction rules and found no requirement for a second durable sample queue.
This is a contract assessment, not implementation evidence. The original outline
preserves the earlier proposal and questions at their original revision.

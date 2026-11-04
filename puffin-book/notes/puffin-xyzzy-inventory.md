# New Puffin: xyzzy read-only deployment inventory

Current design decisions and later clarifications are maintained in
[scheduling-and-puffin.md](scheduling-and-puffin.md). Scheduling is separate from
Puffin; composing its assembly with Gateway remains a deployment proposal.

Bill redirected the work from debugging the old Puffin to building a replacement
and migrating her identity/memories. He specified three components, all on xyzzy:
Gateway; one scheduler/calendar/task service; and Puffin the Ensemble Agent.
He authorized read-only SSH inventory. No service, schedule, credential, package,
power setting, or file on xyzzy was changed. No private logs/configuration contents
or family calendar contents were read. Commands inspected metadata and sizes.

## Observed host

- Raspberry Pi host xyzzy, Debian12/bookworm, Linux6.12.25+rpt-rpi-2712, aarch64.
- Memory7.9GiB total,6.8GiB available; swap199MiB unused at inspection.
- Single29.7GiB mmc device; root ext4 reports29GiB total,18GiB used,9.8GiB free.
- Python3.11.2, system Go1.19.8, Bubblewrap0.8.0 installed. dotnet not on PATH;
  this does not prove no private runtime exists elsewhere.
- gateway.service and puffin.service active, run as bill, Restart=always,
  RestartSec5; cloudflared active and left untouched.
- Gateway working directory/executable /home/bill/gateway[/gateway].
- Puffin working directory/executable /home/bill/puffin[/puffin].
- Gateway directory21MiB; Puffin441MiB, including pf/agents about396MiB;
  separate pf_clone about4MiB. Calendar/data directory about2.3MiB. These are
  size classifications, not an audit of memory contents or authority.
- /var/log about1.1GiB, /var/cache1.2GiB, user cache2.4GiB. No cleanup authorized
  or performed. Do not treat caches as disposable without a separate review.

There is no measured need to upgrade flash now. Publish exact required linux-arm64
artifacts elsewhere rather than carrying SDK/package/build caches onto xyzzy.
Measure final footprint with current+rollback copies and durable-state growth;
per-file64MiB diagnostic caps are not a global disk quota. Flash health/endurance,
backup reliability and long-term growth were not established by this inventory.

## Proven implementation prerequisite

agent/Sandbox/Command.cs currently refuses rooted command execution on every
non-macOS platform. Linux rooted file support does not establish command
confinement. Installed Bubblewrap and namespace capacity are prerequisites, not
proof of an effective sandbox. Implement and test fail-closed Linux command
confinement (filesystem, environment, network and descendant lifetime) before
Puffin gains run_command. Do not silently use full-access mode as fallback.

## Proposed ownership and next plan

Gateway authenticates/routes; the scheduler/calendar/task service owns durable
records, recurrence/exceptions and occurrence/delivery state; Puffin owns her
identity, memories and agentic work. Proposed schedule context is ephemeral:
Puffin receives a fresh replaceable view of the durable service, rather than
remembering appointments as authoritative prose. Bill's phrase ephemeral data
must not be interpreted as permission to discard the authoritative schedule.
Exactly one timing authority, including a deliberate decision about existing
goal timers. No simultaneous old/new live scheduling during migration.

Next: inventory Puffin's identity/memory formats under the granted read authority,
resolve live versus clone provenance, and present a first-build/migration plan.
Keep originals intact. Imported memories do not import credentials, permissions,
active schedules, pending calls, or old process state. Build approval remains
separate from deployment/cutover; no new implementation was started here.

Lesson: sufficient RAM/disk and an installed sandbox utility do not prove the
application's confinement path exists. Check the exact platform implementation
before promising a more capable yet confined Agent.

// This is a projection of the core snapshot, never a browser-owned wake timer.
export function renderGoalStatus(element, status, unavailable = "Goal: loading") {
  let label = unavailable;
  const details = [];
  if (status) {
    const names = { working: "working", waiting: "waiting", paused: "paused", off: "reminders off", error: "error" };
    label = status.file ? `Goal: ${names[status.phase] || "configured"}` : "No goal";
    if (status.file) {
      if (!status.enabled && status.phase === "working") label += " · reminders off";
      details.push(`File: ${status.file}`, `Idle delay: ${status.idle_seconds} seconds`, `Reminders: ${status.enabled ? "enabled" : "off"}`);
      if (status.enabled && status.phase === "waiting" && status.next_reminder) {
        const next = new Date(status.next_reminder);
        if (Number.isFinite(next.getTime()) && next.getTime() > 0) details.push(`Next reminder: ${next.toLocaleString()}`);
      }
      if (status.error) details.push(`Error: ${status.error}`);
    }
  }
  element.textContent = label;
  element.title = [label, ...details].join(". ");
  element.setAttribute("aria-label", element.title);
}

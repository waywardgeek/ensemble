// Owns only a page's continuing views, never Jobs or their model-observation cursor.
export function createJobSessions(pane, request = fetch) {
    const sessions = new Map();
    let generation = 0;
    function reset() {
        generation++;
        for (const session of sessions.values()) {
            clearTimeout(session.timer);
            session.controller?.abort();
            session.node.remove();
        }
        sessions.clear();
    }
    function selected(node) {
        const selection = window.getSelection();
        if (!selection || selection.isCollapsed) return false;
        for (let i = 0; i < selection.rangeCount; i++) if (selection.getRangeAt(i).intersectsNode(node)) return true;
        return false;
    }
    function paint(session, value) {
        session.latest = value;
        const renderer = window.EnsembleArtifactRenderers?.command;
        if (session.renderedStart === value.start && session.renderedData === value.data && session.renderer === renderer) {
            session.status = value.status;
            session.label.textContent = `Command session · Job ${session.handle} · ${value.status}`;
            return;
        }
        if (selected(session.output)) {
            session.pending = value;
            return;
        }
        session.pending = null;
        const bytes = Uint8Array.from(atob(value.data), c => c.charCodeAt(0));
        if (bytes.length !== value.end - value.start) throw new Error("Invalid window length");
        const oldTop = session.output.scrollTop;
        const atBottom = !session.output.contains(document.activeElement) && session.output.scrollHeight - oldTop - session.output.clientHeight < 24;
        try {
            if (typeof renderer !== "function") throw new Error("Command renderer unavailable");
            renderer(session.output, bytes.slice(), value.rolled_out);
        } catch {
            // Optional presentation cannot suppress the recorded output. This
            // literal fallback does not emulate ANSI or undo custom side effects.
            session.output.replaceChildren(document.createTextNode(new TextDecoder().decode(bytes)));
        }
        session.renderer = renderer;
        session.renderedStart = value.start;
        session.renderedData = value.data;
        session.output.scrollTop = atBottom ? session.output.scrollHeight : oldTop;
        session.notice.textContent = value.rolled_out
            ? `Earlier output rolled out; retained on disk. Showing bytes ${value.start}–${value.end}. Styling resets at this boundary; partial control fragments may appear.`
            : `Showing ${value.end} output bytes. Terminal log, not an interactive terminal.`;
        session.status = value.status;
        session.label.textContent = `Command session · Job ${session.handle} · ${value.status}`;
    }
    async function refresh(session) {
        if (session.controller || !session.node.isConnected) return;
        const epoch = generation;
        const controller = new AbortController();
        session.controller = controller;
        const deadline = setTimeout(() => controller.abort(), 10000);
        try {
            if (document.hidden || pane.getClientRects().length === 0) return;
            const url = new URL("./jobs/window", document.baseURI);
            url.searchParams.set("agent", session.agent);
            url.searchParams.set("handle", session.handle);
            const response = await request(url, { headers: { "X-Ensemble-Output": "1" }, credentials: "same-origin", redirect: "error", cache: "no-store", signal: controller.signal });
            if (!response.ok) throw new Error("Output window unavailable");
            const value = await response.json();
            if (epoch !== generation || !session.node.isConnected) return;
            if (value.handle !== session.handle || value.origin_call_id !== session.origin || value.origin_tool !== "run_command" ||
                value.encoding !== "base64" || !Number.isSafeInteger(value.start) || !Number.isSafeInteger(value.end) ||
                value.start < 0 || value.end < value.start || value.end - value.start > 65536 ||
                typeof value.data !== "string" || value.data.length > 87384 || value.rolled_out !== (value.start > 0) ||
                typeof value.status !== "string") throw new Error("Invalid output window");
            paint(session, value);
        } catch (error) {
            if (epoch === generation && session.node.isConnected && !controller.signal.aborted)
                session.notice.textContent = "Live output unavailable; recorded observations remain below. Refresh to retry.";
        } finally {
            clearTimeout(deadline);
            if (session.controller === controller) session.controller = null;
            if (epoch === generation && session.node.isConnected && session.status === "running")
                session.timer = setTimeout(() => refresh(session), 1000);
        }
    }
    // A final response can arrive while the human selects text. Retain that
    // bounded snapshot and apply it after selection clears, even if polling ends.
    function selectionChanged() {
        for (const session of sessions.values()) if (session.pending && !selected(session.output))
            paint(session, session.pending);
    }
    document.addEventListener("selectionchange", selectionChanged);
    document.addEventListener("ensemble-renderers-changed", () => {
        for (const session of sessions.values()) if (session.latest) paint(session, session.latest);
    });
    function observe(card, foreground = false) {
        const job = card.job;
        if (job?.origin_tool !== "run_command" || typeof job.origin_call_id !== "string" || !job.origin_call_id ||
            typeof card.agent !== "string" || !Number.isSafeInteger(job.handle) || job.handle <= 0) return;
        const key = JSON.stringify([card.agent, job.handle, job.origin_call_id]);
        let session = sessions.get(key);
        if (!session) {
            const node = document.createElement("section");
            node.className = "job-session part";
            node.dataset.jobHandle = job.handle;
            const label = document.createElement("strong");
            const notice = document.createElement("p");
            const output = document.createElement("pre");
            output.className = "terminal-output";
            output.tabIndex = 0;
            output.setAttribute("aria-label", "Continuing command output");
            const retry = document.createElement("button");
            retry.type = "button";
            retry.className = "btn btn-sm btn-outline-primary";
            retry.textContent = "Refresh output";
            session = { node, label, notice, output, agent: card.agent, handle: job.handle, origin: job.origin_call_id, status: job.status };
            retry.onclick = () => { clearTimeout(session.timer); refresh(session); };
            label.textContent = `Command session · Job ${job.handle}`;
            notice.textContent = "Recorded observations remain separate; loading the bounded output window…";
            node.append(label, notice, output, retry);
            pane.append(node);
            sessions.set(key, session);
        }
        session.status = job.status;
        if (foreground && pane.lastElementChild !== session.node) {
            // moveBefore preserves focus/selection where supported; never clone.
            const scroll = pane.scrollTop;
            if (pane.moveBefore) pane.moveBefore(session.node, null);
            else pane.append(session.node);
            pane.scrollTop = scroll;
        }
        clearTimeout(session.timer);
        refresh(session);
    }
    return { observe, reset };
}

import test from "node:test";
import assert from "node:assert/strict";
import { access, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { createServer } from "node:http";
import { spawn, execFileSync } from "node:child_process";
import { once } from "node:events";
import { firefox } from "playwright";

const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));
async function until(check, label) {
  for (let attempt = 0; attempt < 200; attempt++) {
    if (await check()) return;
    await delay(25);
  }
  assert.fail(`Timed out: ${label}`);
}

test("turn timing component distinguishes known zero, absent categories and legacy facts", { timeout: 20000 }, async () => {
  const server = createServer(async (request, response) => {
    if (request.url === "/") {
      response.setHeader("content-type", "text/html");
      response.end('<link rel="stylesheet" href="/style.css"><section id="scroll"></section>');
      return;
    }
    const name = request.url.slice(1);
    if (!["artifact-scroll.js", "renderers.js", "style.css"].includes(name)) {
      response.writeHead(404).end();
      return;
    }
    response.setHeader("content-type", name.endsWith(".css") ? "text/css" : "text/javascript");
    response.end(await readFile(new URL(`./web/gui/${name}`, import.meta.url)));
  });
  let browser;
  try {
    await new Promise(resolve => server.listen(0, "127.0.0.1", resolve));
    browser = await firefox.launch({ headless: true });
    const page = await browser.newPage();
    await page.goto(`http://127.0.0.1:${server.address().port}`);
    const result = await page.evaluate(async () => {
      const { ArtifactScroll } = await import("/artifact-scroll.js");
      const element = document.querySelector("#scroll");
      const scroll = new ArtifactScroll(element);
      scroll.receive({ type: "user_message", part_id: "legacy-user", at: "2026-10-25T12:00:00Z", part: { type: "text", text: "legacy" } });
      scroll.receive({ type: "part_final", part_id: "legacy-reply", at: "2026-10-25T12:01:00Z", part: { type: "text", text: "legacy reply", elapsed: 12 } });
      scroll.receive({ type: "turn_ended", text: "legacy reply" });
      const legacyRows = element.querySelectorAll(".turn-timing").length;
      const cardSpans = [...element.querySelectorAll(".artifact__timing")].map(node => node.title);
      const summary = { type: "turn_timing", turn_seq: 9, started: "2026-10-25T12:00:00Z", total_ms: 0, model_ms: 0, tool_ms: 0, other_ms: 0, calls: 2, requests: 1, interrupted: true, error: "<script>unsafe</script>" };
      scroll.receive(summary);
      scroll.receive({ ...summary, replay: true });
      const rows = element.querySelectorAll(".turn-timing");
      const palettes = ["light", "dark"].map(theme => {
        document.body.dataset.theme = theme;
        return [...rows[0].children].every(child => getComputedStyle(child).color === getComputedStyle(document.body).color);
      });
      const knownZero = { count: rows.length, text: rows[0].textContent, title: rows[0].querySelector(".turn-timing__other").title,
        role: rows[0].getAttribute("role"), label: rows[0].getAttribute("aria-label"), last: element.lastElementChild === rows[0],
        cards: element.querySelectorAll(".artifact").length, scripts: element.querySelectorAll("script").length,
        cardSpans: [...element.querySelectorAll(".artifact__timing")].map(node => node.title) };
      scroll.reset();
      const resetEmpty = element.children.length === 0;
      scroll.receive({ ...summary, calls: 0, requests: 0, interrupted: false, error: "", replay: true });
      const absent = { count: element.querySelectorAll(".turn-timing").length, text: element.textContent,
        model: !!element.querySelector(".turn-timing__model"), tools: !!element.querySelector(".turn-timing__tools") };
      scroll.receive({ ...summary, calls: 1, requests: 2, interrupted: false, error: "provider failed", total_ms: 1250, model_ms: 1000, tool_ms: 200, other_ms: 50 });
      const updated = { count: element.querySelectorAll(".turn-timing").length, text: element.textContent, title: element.querySelector(".turn-timing__other")?.title };
      scroll.receive({ type: "turn_timing", turn_seq: 10 });
      const valid = { ...summary, turn_seq: 10, total_ms: 10, model_ms: 4, tool_ms: 3, other_ms: 3 };
      const malformed = [
        { ...valid, model_ms: 1000, other_ms: 0 },
        { ...valid, other_ms: 0 },
        { ...valid, started: "0001-01-01T00:00:00Z" },
        { ...valid, started: "0001-01-01T00:00:00.000Z" },
        { ...valid, started: "invalid" },
        { ...valid, requests: 0 },
        { ...valid, calls: 0 },
        { ...valid, total_ms: Number.MAX_SAFE_INTEGER, model_ms: Number.MAX_SAFE_INTEGER, tool_ms: 1, other_ms: 0 },
      ];
      for (const field of ["total_ms", "model_ms", "tool_ms", "other_ms", "calls", "requests", "turn_seq"]) {
        for (const value of [-1, 0.5, "1", null, NaN, Infinity, Number.MAX_SAFE_INTEGER + 1])
          malformed.push({ ...valid, [field]: value });
      }
      for (const field of ["started", "total_ms", "model_ms", "tool_ms", "other_ms", "calls", "requests", "interrupted"]) {
        const missing = { ...valid };
        delete missing[field];
        malformed.push(missing);
      }
      malformed.push({ ...valid, interrupted: "false" });
      for (const phases of [null, [], "invalid", {}, { context_copy: -1, unattributed: 4 }, { context_copy: 2 }, { context_copy: 4 }, { "": 3 }, { context_copy: 3.1 }, { context_copy: "3" }])
        malformed.push({ ...valid, other_phases_ms: phases });
      const rejected = malformed.map((message, index) => {
        scroll.receive(message);
        if (message.turn_seq === 10) scroll.receive({ ...message, turn_seq: 9, replay: true });
        return { index, count: element.querySelectorAll(".turn-timing").length, text: element.textContent };
      });
      scroll.receive({ ...summary, total_ms: 5, model_ms: 0, tool_ms: 0, other_ms: 5, calls: 0, requests: 0, interrupted: false, error: "", other_phases_ms: { context_copy: 3, "<script>unsafe</script>": 1, unattributed: 1 } });
      const detail = element.querySelector(".turn-timing__phases");
      if (!detail) throw new Error("Missing Other breakdown");
      detail.querySelector("summary").click();
      const phases = { open: detail.open, text: detail.textContent, scripts: element.querySelectorAll("script").length };
      return { legacyRows, cardSpans, knownZero, palettes, resetEmpty, absent, updated, rejected, phases, finalRows: element.querySelectorAll(".turn-timing").length };
    });
    assert.equal(result.legacyRows, 0);
    assert.deepEqual(result.knownZero.cardSpans, result.cardSpans);
    assert.equal(result.knownZero.count, 1);
    assert.match(result.knownZero.text, /Total 0msModel 0msTools \(2\) 0msOther 0msInterrupted/);
    assert.match(result.knownZero.title, /Interrupted; Error: <script>unsafe<\/script>/);
    assert.equal(result.knownZero.role, "group");
    assert.equal(result.knownZero.label, "Turn timing");
    assert.equal(result.knownZero.last, true);
    assert.equal(result.knownZero.cards, 2);
    assert.equal(result.knownZero.scripts, 0);
    assert.deepEqual(result.palettes, [true, true], "all timing text, including interrupted/error, uses the theme foreground");
    assert.equal(result.resetEmpty, true);
    assert.deepEqual(result.absent, { count: 1, text: "Total 0msOther 0ms", model: false, tools: false });
    assert.equal(result.updated.count, 1);
    assert.equal(result.updated.text, "Total 1.3sModel 1.0sTools (1) 200msOther 50msError");
    assert.match(result.updated.title, /Error: provider failed/);
    for (const rejected of result.rejected) {
      assert.equal(rejected.count, 1, `malformed frame ${rejected.index} created a row`);
      assert.equal(rejected.text, result.updated.text, `malformed frame ${rejected.index} changed a valid row`);
    }
    assert.equal(result.finalRows, 1);
    assert.equal(result.phases.open, true);
    assert.match(result.phases.text, /context copy: 3ms/);
    assert.match(result.phases.text, /<script>unsafe<\/script>: 1ms/);
    assert.equal(result.phases.scripts, 0);
  } finally {
    await browser?.close();
    server.closeAllConnections();
    await new Promise(resolve => server.close(resolve));
  }
});

test("turn timing survives real tool roundtrips, background work, agent selection and restart", { timeout: 60000 }, async () => {
  const directory = await mkdtemp(join(tmpdir(), "ensemble-turn-timing-"));
  const binary = join(directory, "gui");
  const requests = [], frames = [], errors = [];
  const steps = new Map();
  const provider = createServer(async (request, response) => {
    let body = "";
    for await (const chunk of request) body += chunk;
    const value = JSON.parse(body);
    requests.push(value);
    const prompt = value.messages.filter(message => message.role === "user" && typeof message.content === "string").at(-1)?.content;
    const step = steps.get(prompt) || 0;
    steps.set(prompt, step + 1);
    await delay(prompt === "background" && step > 0 ? 650 : 160);
    let message = { role: "assistant", content: `${prompt} finished` };
    if (step === 0 && ["roundtrip", "background"].includes(prompt)) {
      const commands = prompt === "roundtrip"
        ? ["sleep 0.2; printf ROUNDTRIP_RESULT"]
        : ["while [ ! -f release-background ]; do sleep 0.05; done; printf BACKGROUND_ONE", "while [ ! -f release-background ]; do sleep 0.05; done; printf BACKGROUND_TWO"];
      message = { role: "assistant", content: "Running tools", tool_calls: commands.map((command, index) => ({
        id: `${prompt}-${index}`, type: "function", function: { name: "run_command", arguments: JSON.stringify({ command, ai_callback_delay: prompt === "background" ? 0.05 : 3 }) },
      })) };
    }
    response.setHeader("content-type", "application/json");
    response.end(JSON.stringify({ choices: [{ message, finish_reason: message.tool_calls ? "tool_calls" : "stop" }], usage: { prompt_tokens: 8, completion_tokens: 8 } }));
  });
  let browser, proc;
  const stop = async () => {
    if (proc && proc.exitCode === null && proc.signalCode === null) {
      const exited = once(proc, "exit");
      proc.kill("SIGTERM");
      const force = setTimeout(() => proc.kill("SIGKILL"), 2000);
      try { await exited; } finally { clearTimeout(force); }
    }
  };
  try {
    await new Promise(resolve => provider.listen(0, "127.0.0.1", resolve));
    execFileSync("go", ["build", "-o", binary, "./cmd"], { cwd: new URL(".", import.meta.url), timeout: 30000 });
    const probe = createServer();
    await new Promise(resolve => probe.listen(0, "127.0.0.1", resolve));
    const port = probe.address().port;
    await new Promise(resolve => probe.close(resolve));
    const url = `http://127.0.0.1:${port}/?speech=transcript`;
    const start = async () => {
      proc = spawn(binary, ["--port", String(port), "--no-stream"], {
        cwd: directory,
        env: { PATH: process.env.PATH, HOME: directory, LLM_VENDOR: "openai", LLM_MODEL: "local-fixture", LLM_API_KEY: "fixture", LLM_BASE_URL: `http://127.0.0.1:${provider.address().port}` },
        stdio: ["pipe", "pipe", "pipe"],
      });
      let log = "";
      proc.stdout.on("data", data => { log += data; });
      proc.stderr.on("data", data => { log += data; });
      await until(async () => {
        if (proc.exitCode !== null) throw Error(log);
        try { return (await fetch(url)).ok; } catch { return false; }
      }, "GUI startup");
    };
    await start();
    browser = await firefox.launch({ headless: true });
    const page = await browser.newPage();
    page.setDefaultTimeout(5000);
    page.on("pageerror", error => errors.push(error.message));
    page.on("websocket", socket => socket.on("framereceived", frame => frames.push(JSON.parse(frame.payload))));
    await page.goto(url);
    await page.locator('[data-agent-key="main"]').waitFor();
    const rows = () => page.locator("#artifacts .turn-timing").evaluateAll(nodes => nodes.map(node => ({ seq: node.dataset.turnSeq, text: node.textContent, titles: [...node.children].map(child => child.title) })));
    const prompt = async (text, count) => {
      await page.fill("#input", text);
      await page.click("#input-form button");
      await page.waitForFunction(count => document.querySelectorAll("#artifacts .turn-timing").length === count, count);
    };
    const before = Date.now();
    await prompt("roundtrip", 1);
    const first = frames.find(frame => frame.type === "turn_timing" && !frame.replay);
    assert.equal(first.calls, 1);
    assert.equal(first.requests, 2);
    assert.ok(first.model_ms >= 300, JSON.stringify(first));
    assert.ok(first.tool_ms >= 190, JSON.stringify(first));
    assert.ok(first.total_ms >= first.model_ms + first.tool_ms, JSON.stringify(first));
    assert.equal(first.total_ms, first.model_ms + first.tool_ms + first.other_ms);
    assert.equal(Object.values(first.other_phases_ms).reduce((sum, value) => sum + value, 0), first.other_ms);
    assert.ok(Object.hasOwn(first.other_phases_ms, "context_copy"));
    await page.locator(".turn-timing__phases summary").click();
    await page.locator(".turn-timing__phases[open] pre").waitFor();
    assert.match(await page.locator(".turn-timing__phases pre").textContent(), /context copy: \d+ms/);
    assert.ok(Date.parse(first.started) >= before - 10);
    assert.ok(first.total_ms <= Date.now() - before);
    assert.ok(requests[1].messages.some(message => message.role === "tool" && message.content.includes("ROUNDTRIP_RESULT")));
    assert.equal(await page.locator("#actions .turn-timing").count(), 0);
    await page.getByRole("group", { name: "Turn timing", exact: true }).waitFor();

    await page.click("#settings-tab");
    const themes = await page.locator("#theme option").evaluateAll(options => options.map(option => option.value));
    for (const theme of ["light", "dark", ...(themes.includes("system") ? ["system"] : [])]) {
      await page.selectOption("#theme", theme);
      for (const scheme of theme === "system" ? ["light", "dark"] : [theme]) {
        await page.emulateMedia({ colorScheme: scheme });
        await page.waitForFunction(expected => document.body.dataset.theme === expected, scheme);
        const colors = await page.locator(".turn-timing").evaluate(row => ({
          row: getComputedStyle(row).color, foreground: getComputedStyle(document.body).color,
        }));
        assert.equal(colors.row, colors.foreground, `${theme}/${scheme}: timing must use the readable theme foreground`);
      }
    }
    await page.click("#agents-tab");

    await prompt("background", 2);
    const background = frames.filter(frame => frame.type === "turn_timing" && !frame.replay).at(-1);
    assert.equal(background.calls, 2);
    assert.equal(background.requests, 2);
    assert.ok(background.model_ms >= 800, JSON.stringify(background));
    assert.ok(background.tool_ms < 400, JSON.stringify(background));
    assert.equal(background.total_ms, background.model_ms + background.tool_ms + background.other_ms);
    assert.equal(frames.filter(frame => frame.type === "tool_dispatched" && frame.call_id.startsWith("background-")).length, 2);
    assert.ok(requests.at(-1).messages.filter(message => message.role === "tool").some(message => message.content.includes("running")));
    const original = await rows();
    await delay(300);
    await writeFile(join(directory, "release-background"), "");
    await until(async () => (await readFile(join(directory, "event.log"), "utf8")).includes("BACKGROUND_TWO"), "background completion");
    assert.deepEqual(await rows(), original);

    await page.fill("#agent-name", "peer");
    await page.fill("#agent-model", "local-fixture");
    await page.click("#create-agent button");
    await page.locator('[data-agent-key="peer"]').click();
    await page.waitForFunction(() => document.querySelector("#conversation-title").textContent === "peer");
    assert.deepEqual(await rows(), []);
    await prompt("peer", 1);
    const peerRows = await rows();
    assert.equal(await page.locator(".turn-timing__tools").count(), 0);
    await page.locator('[data-agent-key="main"]').click();
    await page.waitForFunction(() => document.querySelectorAll("#artifacts .turn-timing").length === 2);
    assert.deepEqual(await rows(), original);
    await page.locator('[data-agent-key="peer"]').click();
    await page.waitForFunction(() => document.querySelectorAll("#artifacts .turn-timing").length === 1);
    assert.deepEqual(await rows(), peerRows);
    await page.locator('[data-agent-key="main"]').click();
    await page.waitForFunction(() => document.querySelectorAll("#artifacts .turn-timing").length === 2);
    await page.reload();
    await page.waitForFunction(() => document.querySelectorAll("#artifacts .turn-timing").length === 2);
    assert.deepEqual(await rows(), original);
    await stop();
    await access(join(directory, "save.json"));
    await start();
    await page.reload();
    await page.waitForFunction(() => document.querySelectorAll("#artifacts .turn-timing").length === 2);
    assert.deepEqual(await rows(), original);
    assert.ok(frames.some(frame => frame.type === "turn_timing" && frame.replay && frame.turn_seq === first.turn_seq));
    await prompt("after restart", 3);
    assert.equal(new Set((await rows()).map(row => row.seq)).size, 3);
    assert.deepEqual((await rows()).slice(0, 2), original);
    assert.deepEqual(errors, []);
  } finally {
    await writeFile(join(directory, "release-background"), "");
    await browser?.close();
    await stop();
    provider.closeAllConnections();
    await new Promise(resolve => provider.close(resolve));
    await rm(directory, { recursive: true, force: true });
  }
});

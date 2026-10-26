import test from "node:test";
import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { createServer } from "node:http";
import { spawn, execFileSync } from "node:child_process";
import { once } from "node:events";
import { firefox } from "playwright";

const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));

for (const scenario of [
  { name: "Anthropic", vendor: "anthropic", model: "claude-sonnet-4-6", streaming: true, env: {}, args: [] },
  { name: "OpenAI EN_STREAMING=all", vendor: "openai", model: "local-fixture", streaming: true, env: { EN_STREAMING: "all" }, args: [] },
  { name: "OpenAI --stream all", vendor: "openai", model: "local-fixture", streaming: true, env: {}, args: ["--stream", "all"] },
  { name: "OpenAI --no-stream", vendor: "openai", model: "local-fixture", streaming: false, env: { EN_STREAMING: "all" }, args: ["--no-stream"] },
]) {
test(`artifact timestamps survive replay and measure real parts (${scenario.name})`, { timeout: 60000 }, async () => {
  const directory = await mkdtemp(join(tmpdir(), "ensemble-timestamps-"));
  const binary = join(directory, "gui");
  let release;
  const firstObserved = new Promise(resolve => { release = resolve; });
  let streamingRequested = false;
  const provider = createServer(async (request, response) => {
    let body = "";
    for await (const chunk of request) body += chunk;
    streamingRequested = JSON.parse(body).stream === true;
    if (scenario.vendor === "openai") {
      const usage = { prompt_tokens: 8, completion_tokens: 8 };
      if (!streamingRequested) {
        response.writeHead(200, { "content-type": "application/json" });
        response.end(JSON.stringify({ choices: [{ message: { role: "assistant", content: "Hello world." }, finish_reason: "stop" }], usage }));
        return;
      }
      response.writeHead(200, { "content-type": "text/event-stream" });
      const chunk = delta => response.write(`data: ${JSON.stringify({ choices: [{ index: 0, delta }] })}\n\n`);
      chunk({ role: "assistant", content: "Hello " });
      await firstObserved;
      await delay(130);
      chunk({ content: "world." });
      await delay(200);
      response.write(`data: ${JSON.stringify({ choices: [{ index: 0, delta: {}, finish_reason: "stop" }] })}\n\n`);
      response.end(`data: ${JSON.stringify({ choices: [], usage })}\n\ndata: [DONE]\n\n`);
      return;
    }
    response.writeHead(200, { "content-type": "text/event-stream" });
    const frame = (type, fields = {}) => response.write(`event: ${type}\ndata: ${JSON.stringify({ type, ...fields })}\n\n`);
    frame("message_start", { message: { usage: { input_tokens: 8, output_tokens: 1 } } });
    frame("content_block_start", { index: 0, content_block: { type: "thinking" } });
    frame("content_block_delta", { index: 0, delta: { type: "thinking_delta", thinking: "Checking " } });
    await firstObserved;
    await delay(90);
    frame("content_block_delta", { index: 0, delta: { type: "thinking_delta", thinking: "carefully." } });
    frame("content_block_stop", { index: 0 });
    await delay(200);
    frame("content_block_start", { index: 1, content_block: { type: "text" } });
    frame("content_block_delta", { index: 1, delta: { type: "text_delta", text: "Hello " } });
    await delay(130);
    frame("content_block_delta", { index: 1, delta: { type: "text_delta", text: "world." } });
    frame("content_block_stop", { index: 1 });
    await delay(170);
    frame("message_delta", { usage: { output_tokens: 8 } });
    frame("message_stop");
    response.end();
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
      proc = spawn(binary, ["--port", String(port), ...scenario.args], {
        cwd: directory,
        env: { PATH: process.env.PATH, HOME: directory, LLM_VENDOR: scenario.vendor, LLM_MODEL: scenario.model, LLM_API_KEY: "fixture", LLM_BASE_URL: `http://127.0.0.1:${provider.address().port}`, ...scenario.env },
        stdio: ["pipe", "pipe", "pipe"],
      });
      let log = "";
      proc.stdout.on("data", data => { log += data; });
      proc.stderr.on("data", data => { log += data; });
      for (let attempt = 0; attempt < 200; attempt++) {
        if (proc.exitCode !== null) throw Error(log);
        try { if ((await fetch(url)).ok) return; } catch {}
        await delay(25);
      }
      throw Error(`GUI did not start: ${log}`);
    };
    await start();
    browser = await firefox.launch({ headless: true });
    const page = await browser.newPage();
    page.setDefaultTimeout(5000);
    const frames = [], errors = [];
    page.on("pageerror", error => errors.push(error.message));
    page.on("websocket", socket => socket.on("framereceived", frame => frames.push(JSON.parse(frame.payload))));
    await page.goto(url);
    await page.locator('[data-agent-key="main"]').waitFor();
    await page.fill("#input", "Check timing");
    await page.click("#input-form button");
    const firstKind = scenario.vendor === "anthropic" ? "thinking" : "text";
    await page.locator(`.artifact--${firstKind} .content`).waitFor();
    assert.equal(streamingRequested, scenario.streaming, `${scenario.name}: provider request streaming`);
    if (scenario.streaming) assert.equal(await page.locator(`.artifact--${firstKind} .artifact__timing`).isHidden(), true);
    release();
    const cardCount = scenario.vendor === "anthropic" ? 3 : 2;
    await page.waitForFunction(count => [...document.querySelectorAll("#artifacts .artifact__timing")].filter(node => !node.hidden).length === count, cardCount);
    const finals = frames.filter(frame => frame.type === "part_final");
    assert.equal(finals.length, cardCount - 1);
    assert.equal(await page.locator(".artifact--text .content").textContent(), "Hello world.");
    for (const frame of frames.filter(frame => ["part_final", "user_message"].includes(frame.type))) {
      const title = await page.locator(`[data-id="${frame.part_id}"] .artifact__timing`).getAttribute("title");
      assert.ok(title.startsWith(new Date(frame.at).toISOString()));
      if (frame.type === "part_final" && scenario.streaming) {
        assert.ok(frame.part.elapsed >= 70, JSON.stringify(frame));
        assert.ok(Date.parse(frame.at) - frame.part.started > frame.part.elapsed + 100);
        assert.match(title, /to produce$/);
      } else {
        assert.ok(!frame.part.started && !frame.part.elapsed, JSON.stringify(frame));
        assert.doesNotMatch(title, /to produce/);
      }
    }
    const chips = () => page.locator("#artifacts .artifact__timing").evaluateAll(nodes => nodes.map(node => ({ text: node.textContent, title: node.title })));
    const original = await chips();
    await page.reload();
    await page.waitForFunction(count => document.querySelectorAll("#artifacts .artifact__timing").length === count, cardCount);
    assert.deepEqual(await chips(), original);
    await stop();
    await start();
    await page.reload();
    await page.waitForFunction(count => document.querySelectorAll("#artifacts .artifact__timing").length === count, cardCount);
    assert.deepEqual(await chips(), original);

    // Exercise the reusable component's legacy-input boundary in the same real DOM.
    const legacy = await page.evaluate(async () => {
      const { ArtifactScroll } = await import("/artifact-scroll.js");
      const element = document.createElement("section");
      document.body.append(element);
      const scroll = new ArtifactScroll(element);
      const values = [undefined, "not-a-date", "0001-01-01T00:00:00Z", "2026-01-02T03:04:05Z", "2026-01-02T03:04:05Z", "2026-01-02T03:04:05Z", true];
      const spans = [undefined, -10, "120", 119900, undefined, 0.1];
      values.forEach((at, index) => scroll.receive({ type: "part_final", part_id: index, at, part: { type: "text", text: "legacy", elapsed: spans[index] } }));
      return [...element.querySelectorAll(".artifact__timing")].map(node => ({ hidden: node.hidden, text: node.textContent, title: node.title }));
    });
    for (const card of legacy.slice(0, 3)) assert.deepEqual(card, { hidden: true, text: "", title: "" });
    assert.match(legacy[3].text, /2m 00s$/);
    for (const card of legacy.slice(4, 6)) assert.equal(card.title, "2026-01-02T03:04:05.000Z");
    assert.deepEqual(legacy[6], { hidden: true, text: "", title: "" });
    assert.deepEqual(errors, []);
  } finally {
    release();
    await browser?.close();
    await stop();
    provider.closeAllConnections();
    await new Promise(resolve => provider.close(resolve));
    await rm(directory, { recursive: true, force: true });
  }
});
}

import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import http from "node:http";
import net from "node:net";
import { spawn, execFileSync } from "node:child_process";
import { once } from "node:events";
import { chromium } from "playwright";

async function until(check, label, timeout = 15000) {
  const end = Date.now() + timeout;
  while (Date.now() < end) {
    if (await check()) return;
    await new Promise(resolve => setTimeout(resolve, 20));
  }
  assert.fail(`Timed out: ${label}`);
}

test("switching releases outgoing speech/typing gates without unpausing the incoming agent", { timeout: 120000 }, async () => {
  const dir = await fs.mkdtemp(path.join(os.tmpdir(), "ensemble-selection-pause-"));
  const bin = path.join(dir, "gui");
  const skillDir = path.join(dir, "skills", "test-primary");
  await fs.mkdir(skillDir, { recursive: true });
  await fs.writeFile(path.join(skillDir, "SKILL.md"), "---\nname: test-primary\ndescription: Selection regression fixture.\ntype: primary\ntools: write_file\n---\nExplain the intended write, perform it, then answer normally.\n");
  execFileSync("go", ["build", "-o", bin, "./cmd"], { cwd: new URL(".", import.meta.url) });
  let releaseReply, nextFile, toolReply = true;
  const provider = http.createServer(async (req, res) => {
    for await (const _ of req) { /* consume the real request */ }
    let message;
    if (toolReply) {
      toolReply = false;
      await new Promise(resolve => { releaseReply = resolve; });
      message = { role: "assistant", content: "I will write the requested marker file.", tool_calls: [{ id: "write-marker", type: "function", function: { name: "write_file", arguments: JSON.stringify({ path: nextFile, content: "released" }) } }] };
    } else {
      toolReply = true;
      message = { role: "assistant", content: "The marker is written." };
    }
    res.setHeader("Content-Type", "application/json");
    res.end(JSON.stringify({ id: "selection", model: "gpt-4o-mini", choices: [{ message, finish_reason: message.tool_calls ? "tool_calls" : "stop" }], usage: { prompt_tokens: 1, completion_tokens: 1 } }));
  });
  provider.listen(0, "127.0.0.1");
  await once(provider, "listening");
  const probe = net.createServer(); probe.listen(0, "127.0.0.1"); await once(probe, "listening");
  const port = probe.address().port; await new Promise(resolve => probe.close(resolve));
  const assets = process.env.GUI_ASSETS || new URL("./web/gui", import.meta.url).pathname;
  const proc = spawn(bin, ["--port", String(port), "--no-stream", "--gui-dir", assets], { cwd: dir, env: { PATH: process.env.PATH, HOME: process.env.HOME, LLM_VENDOR: "openai", LLM_MODEL: "gpt-4o-mini", LLM_BASE_URL: `http://127.0.0.1:${provider.address().port}`, LLM_API_KEY: "fixture", EN_SKILLS_DIR: path.dirname(skillDir), EN_PRIMARY_SKILL: "test-primary" }, stdio: ["pipe", "pipe", "pipe"] });
  let log = "";
  proc.stdout.on("data", b => { log += b; }); proc.stderr.on("data", b => { log += b; });
  let browser;
  try {
    const url = `http://127.0.0.1:${port}/`;
    await until(async () => {
      if (proc.exitCode !== null) throw Error(log);
      try { return (await fetch(url)).ok; } catch { return false; }
    }, "isolated GUI startup");
    browser = await chromium.launch({ headless: true, executablePath: process.env.CHROME_BIN || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" });
    const page = await browser.newPage();
    await page.addInitScript(() => {
      window.sent = [];
      const send = WebSocket.prototype.send;
      WebSocket.prototype.send = function (data) { window.sent.push(JSON.parse(data)); return send.call(this, data); };
      window.speechSynthesis.speak = utterance => { utterance.onstart?.(); };
      window.speechSynthesis.cancel = () => {};
    });
    const frames = [], held = [];
    let holding = false, delaySelection = false, deliver;
    await page.routeWebSocket("**/ws", ws => {
      deliver = data => ws.send(data);
      const server = ws.connectToServer();
      server.onMessage(data => {
        const message = JSON.parse(String(data)); frames.push(message);
        if (delaySelection && message.type === "agent_selected" && message.key === "peer") holding = true;
        if (holding) held.push(data); else ws.send(data);
      });
    });
    await page.goto(url + "?speech=transcript");
    await page.waitForFunction(() => document.querySelector("#conversation-title").textContent === "main");
    await page.fill("#agent-name", "peer");
    await page.click("#create-agent button");
    await page.waitForSelector('[data-agent-key="peer"]');
    // A second real connection establishes B's pause independently of A's view.
    const peer = await browser.newPage();
    await peer.goto(url + "?speech=transcript");
    await peer.click('[data-agent-key="peer"]');
    await peer.waitForFunction(() => document.querySelector("#conversation-title").textContent === "peer");
    await peer.fill("#input", "Keep B paused");
    await peer.waitForFunction(() => document.querySelector("#pause-state").textContent === "paused");

    for (const mode of ["speech", "typing"]) {
      nextFile = path.join(dir, `${mode}-released.txt`);
      releaseReply = undefined;
      await page.fill("#input", `Write the ${mode} marker.`);
      await page.click("#input-form button");
      await until(() => releaseReply, "provider request waiting at external boundary");
      if (mode === "speech") {
        await page.evaluate(() => {
          window.ensembleSpeech.audio = true;
          window.ensembleSpeech.enqueue("Outgoing speech is still active.");
          window.oldUtterance = window.ensembleSpeech.active;
        });
      } else {
        await page.fill("#input", "An unfinished human draft");
      }
      await page.waitForFunction(() => document.querySelector("#pause-state").textContent === "paused");
      const start = frames.length;
      releaseReply();
      await until(() => frames.slice(start).some(f => f.type === "part_final" && f.part?.type === "tool_call"), "model's completed tool call while dispatch is paused");
      await assert.rejects(fs.access(nextFile), { code: "ENOENT" });
      await page.evaluate(() => { window.sent = []; });
      delaySelection = true;
      await page.click('[data-agent-key="peer"]');
      await until(() => holding, "selection acknowledgment held at browser transport boundary");
      assert.deepEqual(await page.evaluate(() => window.sent.filter(m => ["pause", "unpause", "select_agent"].includes(m.type)).map(m => m.type)), ["unpause", "select_agent"]);
      assert.equal(await page.locator('[data-agent-key="peer"]').isDisabled(), true);
      // Also reject an already queued click rather than emitting duplicate
      // acknowledgments that could reopen pacing between replay boundaries.
      await page.locator('[data-agent-key="peer"]').dispatchEvent("click");
      assert.deepEqual(await page.evaluate(() => window.sent.filter(m => ["pause", "unpause", "select_agent"].includes(m.type)).map(m => m.type)), ["unpause", "select_agent"]);
      await until(async () => { try { return await fs.readFile(nextFile, "utf8") === "released"; } catch { return false; } }, "outgoing real tool dispatch released without a new prompt");
      // Exercise queued outgoing text and delayed native cancellation while the
      // selection acknowledgment is in flight. Neither may pace the new agent.
      await page.evaluate(() => {
        window.ensembleSpeech.audio = true;
        window.ensembleSpeech.enqueue("An outgoing phrase arriving during selection.");
        window.oldUtterance?.onerror?.({ type: "error", error: "canceled" });
      });
      assert.equal(await page.evaluate(() => window.ensembleSpeech.pending().length), 0);
      assert.deepEqual(await page.evaluate(() => window.sent.filter(m => ["pause", "unpause", "select_agent"].includes(m.type)).map(m => m.type)), ["unpause", "select_agent"]);
      holding = delaySelection = false;
      for (const data of held.splice(0)) deliver(data);
      await page.waitForFunction(() => document.querySelector("#conversation-title").textContent === "peer");
      await page.waitForFunction(() => document.querySelector("#pause-state").textContent === "paused");
      assert.equal(await peer.locator("#pause-state").textContent(), "paused");
      await page.evaluate(() => {
        window.oldUtterance?.onend?.({ type: "end" });
        window.ensembleSpeech.enqueue("New selection speech.");
        window.oldUtterance?.onerror?.({ type: "error", error: "canceled" });
      });
      assert.equal(await page.evaluate(() => window.ensembleSpeech.pending()[0]?.text), "New selection speech.");
      assert.equal(await page.evaluate(() => window.sent.filter(m => m.type === "unpause").length), 1, "late callbacks and acknowledgment must not unpause B");
      // Return to A. Re-establish B's independent pause for the next case.
      await page.click('[data-agent-key="main"]');
      await page.waitForFunction(() => document.querySelector("#conversation-title").textContent === "main");
      await page.evaluate(() => { window.ensembleSpeech.audio = false; });
      await peer.fill("#input", "");
      await peer.fill("#input", "Keep B paused again");
      await peer.waitForFunction(() => document.querySelector("#pause-state").textContent === "paused");
      await until(() => toolReply, "outgoing turn completed");
    }
  } finally {
    releaseReply?.();
    await browser?.close();
    // Dispose only this isolated fixture; the negative control can strand a
    // paused tool even during graceful server shutdown.
    if (proc.exitCode === null) {
      const exited = once(proc, "exit");
      proc.kill("SIGKILL");
      await exited;
    }
    await new Promise(resolve => provider.close(resolve));
    await fs.rm(dir, { recursive: true, force: true });
  }
});

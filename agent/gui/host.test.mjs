import test from "node:test";
import assert from "node:assert/strict";
import { mkdtemp, readFile, writeFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { createServer, request as httpRequest } from "node:http";
import net from "node:net";
import { spawn, execFileSync } from "node:child_process";
import { once } from "node:events";
import { firefox } from "playwright";

const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));
const reachable = async url => {
  try {
    const response = await fetch(url, { signal: AbortSignal.timeout(500) });
    await response.arrayBuffer();
    return response.ok;
  } catch { return false; }
};
const stop = async child => {
  if (!child || child.exitCode !== null || child.signalCode !== null) return;
  const exited = once(child, "exit");
  child.kill("SIGTERM");
  const force = setTimeout(() => child.kill("SIGKILL"), 2000);
  try { await exited; } finally { clearTimeout(force); }
};

test("GUI host selection preserves origin checks and browser chat", { timeout: 90000 }, async parent => {
  const directory = await mkdtemp(join(tmpdir(), "ensemble-host-"));
  const binary = join(directory, "gui");
  let requests = 0, browser;
  const provider = createServer(async (request, response) => {
    for await (const chunk of request) {}
    requests++;
    response.setHeader("Content-Type", "application/json");
    response.end(JSON.stringify({ choices: [{ message: { role: "assistant", content: "Host binding works." }, finish_reason: "stop" }], usage: { prompt_tokens: 1, completion_tokens: 1 } }));
  });
  try {
    execFileSync("go", ["build", "-o", binary, "./cmd"], { cwd: new URL(".", import.meta.url), timeout: 30000 });
    provider.listen(0, "127.0.0.1");
    await once(provider, "listening");
    browser = await firefox.launch({ headless: true });
    const example = JSON.parse(await readFile(new URL("./settings.example.json", import.meta.url), "utf8"));
    for (const mode of [
      { name: "default loopback", host: null, address: "127.0.0.1", warning: false },
      { name: "wildcard", host: "0.0.0.0", address: "127.0.0.2", warning: true },
      { name: "config override", host: "0.0.0.0", address: "127.0.0.2", warning: true, config: true },
      { name: "hostname", host: "localhost", address: "localhost", warning: true },
      { name: "IPv6 loopback", host: "::1", address: "[::1]", warning: false },
      { name: "empty rejected", host: "", address: "127.0.0.1", warning: false },
    ]) {
      await parent.test(mode.name, { timeout: 15000 }, async current => {
        const workspace = await mkdtemp(join(directory, "launch-"));
        const probe = net.createServer();
        try {
          probe.listen(0, mode.host === "::1" ? "::1" : "127.0.0.1");
          await once(probe, "listening");
        } catch (error) {
          if (mode.host === "::1" && error.code === "EADDRNOTAVAIL") { current.skip("IPv6 loopback unavailable"); return; }
          throw error;
        }
        const port = probe.address().port;
        await new Promise(resolve => probe.close(resolve));
        const providerURL = `http://127.0.0.1:${provider.address().port}`;
        const args = ["--port", String(port)];
        if (mode.config) {
          const config = structuredClone(example);
          Object.assign(config, { workspace: ".", skills_dir: "", primary_skill: "", gui_port: port, enable_web_search: false, recall: { disabled: true } });
          Object.assign(config.provider, { vendor: "openai", model: "host-fixture", billing_mode: "api", credential_file: "key.txt", base_url: providerURL, reasoning_effort: "", reasoning_summary: "" });
          config.gui.tts_enabled = false;
          const path = join(workspace, "settings.json");
          await writeFile(join(workspace, "key.txt"), "fixture", { mode: 0o600 });
          await writeFile(path, JSON.stringify(config));
          args.splice(0, args.length, "--config", path);
        }
        if (mode.host !== null) args.push("--host", mode.host);
        const child = spawn(binary, args, {
          cwd: workspace,
          env: { PATH: process.env.PATH, HOME: workspace, LLM_VENDOR: "openai", LLM_MODEL: "host-fixture", LLM_API_KEY: "fixture", LLM_BASE_URL: providerURL, EN_RECALL_DISABLED: "1" },
          stdio: ["pipe", "ignore", "pipe"],
        });
        let output = "", page;
        child.stderr.on("data", data => { output = (output + data).slice(-65536); });
        const watchdog = setTimeout(() => child.kill("SIGKILL"), 12000);
        try {
          if (mode.host === "") {
            const [code] = await once(child, "exit");
            assert.notEqual(code, 0);
            assert.ok(output.includes("--host must not be empty"), "empty host must have an explicit validation error");
            return;
          }
          const origin = `http://${mode.address}:${port}`;
          const deadline = Date.now() + 5000;
          let ready = false;
          while (Date.now() < deadline) {
            if (child.exitCode !== null || child.signalCode !== null) assert.fail(output.includes("flag provided but not defined: -host") ? "GUI rejected --host" : "GUI exited before startup");
            if (await reachable(origin)) { ready = true; break; }
            await delay(25);
          }
          assert.ok(ready, "configured listener did not become reachable");
          assert.equal(output.includes("without authentication"), mode.warning, "listener authentication warning");
          if (mode.host === null) assert.equal(await reachable(`http://127.0.0.2:${port}`), false, "default listener escaped its loopback address");
          const status = await new Promise((resolve, reject) => {
            const request = httpRequest(`${origin}/ws`, { headers: { Connection: "Upgrade", Upgrade: "websocket", "Sec-WebSocket-Version": "13", "Sec-WebSocket-Key": Buffer.alloc(16).toString("base64"), Origin: "https://cross-origin.invalid" } });
            request.once("response", response => { response.resume(); resolve(response.statusCode); });
            request.once("upgrade", (_response, socket) => { socket.destroy(); resolve(101); });
            request.once("error", reject);
            request.setTimeout(2000, () => request.destroy(new Error("upgrade probe timed out")));
            request.end();
          });
          assert.equal(status, 403, "cross-origin websocket upgrade must be forbidden");
          page = await browser.newPage();
          page.setDefaultTimeout(4000);
          const errors = [];
          page.on("pageerror", error => errors.push(error.message));
          await page.goto(`${origin}/?speech=transcript`);
          await page.locator('[data-agent-key="main"]').waitFor();
          const before = requests;
          await page.fill("#input", "Check host binding.");
          await page.click("#input-form button");
          await page.waitForFunction(() => document.querySelector(".artifact--text .content")?.textContent === "Host binding works.");
          assert.equal(requests, before + 1, "chat must reach the real HTTP boundary");
          assert.deepEqual(errors, []);
        } finally {
          clearTimeout(watchdog);
          await stop(child);
          await page?.close();
        }
      });
    }
  } finally {
    await browser?.close();
    provider.closeAllConnections();
    await new Promise(resolve => provider.close(resolve));
    await rm(directory, { recursive: true, force: true });
  }
});

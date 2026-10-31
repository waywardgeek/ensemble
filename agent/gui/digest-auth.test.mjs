import test from "node:test";
import assert from "node:assert/strict";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { createServer, request as httpRequest } from "node:http";
import { createHash, randomBytes } from "node:crypto";
import { spawn, spawnSync } from "node:child_process";
import { once } from "node:events";
import { firefox } from "playwright";

const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));
const sha256 = text => createHash("sha256").update(text).digest("hex");

async function readCredentials(path) {
  try { return JSON.parse(await readFile(path, "utf8")); }
  catch { throw new Error("Cannot read generated credential JSON"); }
}

async function stop(child) {
  if (!child || child.exitCode !== null || child.signalCode !== null) return;
  const exited = once(child, "exit");
  child.kill("SIGTERM");
  const watchdog = setTimeout(() => child.kill("SIGKILL"), 2000);
  try { await exited; } finally { clearTimeout(watchdog); }
}

async function unusedPort() {
  const probe = createServer();
  probe.listen(0, "127.0.0.1");
  await once(probe, "listening");
  const port = probe.address().port;
  await new Promise(resolve => probe.close(resolve));
  return port;
}

function request(url, headers = {}) {
  return new Promise((resolve, reject) => {
    const outgoing = httpRequest(url, { headers, agent: false });
    outgoing.once("response", response => {
      response.resume();
      response.once("end", () => resolve({ status: response.statusCode, headers: response.headers }));
      response.once("error", () => reject(new Error("HTTP response failed")));
    });
    outgoing.once("upgrade", (response, socket) => {
      socket.destroy();
      resolve({ status: response.statusCode, headers: response.headers });
    });
    outgoing.once("error", () => reject(new Error("HTTP probe failed")));
    outgoing.setTimeout(2000, () => outgoing.destroy(new Error("HTTP probe timed out")));
    outgoing.end();
  });
}

function digest(challenge, password, uri) {
  assert.equal(challenge?.startsWith("Digest "), true, "server must challenge with Digest");
  assert.equal(/algorithm=SHA-256(?:,|$)/i.test(challenge), true, "SHA-256 required");
  assert.equal(challenge.includes('qop="auth"'), true, "qop auth required");
  const realm = challenge.match(/realm="([^"]+)"/)?.[1];
  const nonce = challenge.match(/nonce="([^"]+)"/)?.[1];
  assert.equal(realm, "Ensemble");
  assert.equal(typeof nonce, "string");
  const cnonce = randomBytes(16).toString("hex");
  const response = sha256(`${sha256(`ensemble:${realm}:${password}`)}:${nonce}:00000001:${cnonce}:auth:${sha256(`GET:${uri}`)}`);
  return `Digest username="ensemble", realm="${realm}", nonce="${nonce}", uri="${uri}", algorithm=SHA-256, qop=auth, nc=00000001, cnonce="${cnonce}", response="${response}"`;
}

test("native Firefox Digest protects GUI assets, WebSocket chat and restored sessions", { timeout: 120000 }, async parent => {
  const directory = await mkdtemp(join(tmpdir(), "ensemble-digest-"));
  const binary = join(directory, "gui");
  const authFile = join(directory, "gui-auth.json");
  const environment = { PATH: process.env.PATH, HOME: directory, XDG_CONFIG_HOME: directory, XDG_DATA_HOME: directory,
    LLM_VENDOR: "openai", LLM_MODEL: "digest-fixture", LLM_API_KEY: "fixture", EN_RECALL_DISABLED: "1" };
  let browser, child, password, credential;
  let modelCalls = 0;
  const provider = createServer(async (incoming, response) => {
    for await (const chunk of incoming) {}
    modelCalls++;
    if (incoming.method !== "POST" || incoming.url !== "/v1/chat/completions") {
      response.writeHead(404).end();
      return;
    }
    response.setHeader("Content-Type", "application/json");
    response.end(JSON.stringify({ id: `digest-${modelCalls}`, model: "digest-fixture",
      choices: [{ message: { role: "assistant", content: `Authenticated reply ${modelCalls}.` }, finish_reason: "stop" }],
      usage: { prompt_tokens: 1, completion_tokens: 1 } }));
  });
  let output = "";
  const launch = args => {
    output = "";
    child = spawn(binary, args, { cwd: directory, env: environment, stdio: ["pipe", "pipe", "pipe"] });
    for (const stream of [child.stdout, child.stderr]) stream.on("data", data => { output = (output + data).slice(-65536); });
    return child;
  };
  const start = async () => {
    const port = await unusedPort();
    launch(["--port", String(port), "--no-stream", "--gui-auth-file", authFile]);
    const origin = `http://127.0.0.1:${port}`;
    const deadline = Date.now() + 10000;
    while (Date.now() < deadline) {
      assert.equal(child.exitCode, null, "GUI exited before startup");
      assert.equal(child.signalCode, null, "GUI terminated before startup");
      try {
        if ((await request(origin)).status === 401) return origin;
      } catch {}
      await delay(25);
    }
    assert.fail("protected GUI did not become ready");
  };
  const safeOutput = () => {
    assert.equal(output.includes(password), false, "GUI output exposed password");
    assert.equal(output.includes(credential.ha1), false, "GUI output exposed HA1");
  };
  const watchdog = setTimeout(() => { child?.kill("SIGKILL"); void browser?.close(); }, 110000);
  try {
    const build = spawnSync("go", ["build", "-o", binary, "./cmd"], {
      cwd: new URL(".", import.meta.url), timeout: 30000, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"],
    });
    assert.equal(build.status, 0, "GUI binary must build before integration testing");
    const initialized = spawnSync(binary, ["auth-init", "--file", authFile], {
      cwd: directory, env: environment, timeout: 5000, encoding: "utf8", stdio: ["ignore", "pipe", "pipe"],
    });
    assert.equal(initialized.status, 0, "auth-init must succeed");
    password = initialized.stdout.trim();
    assert.equal(password.length >= 22 && !/\s/.test(password), true, "CLI must print one generated secret");
    credential = await readCredentials(authFile);
    assert.equal(credential.version === 1 && credential.username === "ensemble" && credential.realm === "Ensemble", true);
    assert.equal(credential.ha1 === sha256(`ensemble:Ensemble:${password}`), true, "stored HA1 must match generated password");
    assert.equal(JSON.stringify(credential).includes(password), false, "credential file must not store plaintext password");
    assert.equal(initialized.stderr.includes(password) || initialized.stderr.includes(credential.ha1), false);

    provider.listen(0, "127.0.0.1");
    await once(provider, "listening");
    environment.LLM_BASE_URL = `http://127.0.0.1:${provider.address().port}`;
    browser = await firefox.launch({ headless: true });
    parent.diagnostic(`Native Firefox ${browser.version()}`);

    await parent.test("missing explicit credentials fail closed", { timeout: 10000 }, async () => {
      launch(["--gui-auth-file", join(directory, "absent.json"), "--port", String(await unusedPort())]);
      const failedStartup = setTimeout(() => child.kill("SIGKILL"), 5000);
      try {
        const [code, signal] = await once(child, "exit");
        assert.equal(signal, null, "missing credentials must exit without needing a watchdog");
        assert.notEqual(code, 0);
        assert.equal(output.includes("GUI authentication required"), true, "missing explicit credentials must prevent startup");
        assert.equal(modelCalls, 0);
        safeOutput();
      } finally {
        clearTimeout(failedStartup);
        await stop(child);
      }
    });

    let origin = await start();
    const websocketHeaders = { Connection: "Upgrade", Upgrade: "websocket", "Sec-WebSocket-Version": "13",
      "Sec-WebSocket-Key": randomBytes(16).toString("base64"), Origin: origin };
    await parent.test("unauthenticated assets and WebSocket cannot reach the model", async () => {
      for (const path of ["/", "/app.js", "/style.css", "/ws"]) {
        const result = await request(origin + path, path === "/ws" ? websocketHeaders : {});
        assert.equal(result.status, 401, `unauthenticated ${path}`);
        digest(result.headers["www-authenticate"], password, path);
      }
      assert.equal(modelCalls, 0);
    });

    await parent.test("native wrong-password login fails", { timeout: 15000 }, async () => {
      const context = await browser.newContext({ httpCredentials: { username: "ensemble", password: "incorrect-fixture-password" } });
      try {
        const page = await context.newPage();
        const responses = [];
        page.on("response", response => responses.push(response.status()));
        await page.goto(origin, { timeout: 5000 }).catch(() => {});
        assert.equal(responses.includes(401), true, "wrong credentials must receive an HTTP challenge");
        assert.equal(await page.locator("#input-form").count(), 0, "wrong credentials must not load the GUI");
        assert.equal(modelCalls, 0);
      } finally { await context.close(); }
    });

    let context = await browser.newContext({ httpCredentials: { username: "ensemble", password } });
    let page = await context.newPage();
    page.setDefaultTimeout(8000);
    const errors = [];
    const replies = number => page.locator(".artifact--text .content").filter({ hasText: `Authenticated reply ${number}.` });
    const chat = async number => {
      const before = modelCalls;
      await page.fill("#input", `Digest conversation ${number}`);
      await page.click("#input-form button");
      await replies(number).waitFor();
      assert.equal(modelCalls, before + 1, "one chat must reach the real HTTP model boundary");
    };
    await parent.test("native login loads assets and opens an authenticated WebSocket for chat", async () => {
      page.on("pageerror", () => errors.push("browser script error"));
      const script = page.waitForResponse(response => new URL(response.url()).pathname === "/app.js" && response.status() === 200);
      const style = page.waitForResponse(response => new URL(response.url()).pathname === "/style.css" && response.status() === 200);
      const socket = page.waitForEvent("websocket");
      assert.equal((await page.goto(`${origin}/?speech=transcript`)).status(), 200);
      await Promise.all([script, style, socket]);
      await page.locator('[data-agent-key="main"][aria-pressed="true"]').waitFor();
      await chat(1);
      assert.deepEqual(errors, []);
    });

    await parent.test("independent Digest accepts correct requests, rejects replay, and preserves Origin checks", async () => {
      const uri = "/app.js?digest=independent";
      const challenge = (await request(origin + uri)).headers["www-authenticate"];
      const authorization = digest(challenge, password, uri);
      assert.equal((await request(origin + uri, { Authorization: authorization })).status, 200);
      assert.equal((await request(origin + uri, { Authorization: authorization })).status, 401, "reused authorization must fail");
      const wsChallenge = (await request(origin + "/ws", websocketHeaders)).headers["www-authenticate"];
      assert.equal((await request(origin + "/ws", { ...websocketHeaders, Origin: "https://malicious.invalid",
        Authorization: digest(wsChallenge, password, "/ws") })).status, 403, "valid Digest must not bypass Origin validation");
      assert.equal((await request(origin + "/ws", { ...websocketHeaders,
        Authorization: digest(wsChallenge, password, "/ws") })).status, 101, "same-origin authenticated upgrade must work");
      assert.equal(modelCalls, 1);
    });

    await parent.test("reload and agent selection replay authenticated conversation", async () => {
      await page.reload();
      await replies(1).waitFor();
      assert.equal(await replies(1).count(), 1);
      await page.fill("#agent-name", "digest-peer");
      await page.fill("#agent-model", "digest-fixture");
      await page.click("#create-agent button");
      await page.locator('[data-agent-key="digest-peer"]').click();
      await page.locator('[data-agent-key="digest-peer"][aria-pressed="true"]').waitFor();
      assert.equal(await replies(1).count(), 0, "main history must not leak into another agent");
      await page.locator('[data-agent-key="main"]').click();
      await replies(1).waitFor();
      assert.equal(await replies(1).count(), 1);
      assert.equal(modelCalls, 1, "replay and selection must not invoke the model");
    });

    await parent.test("fresh native login after process restart preserves credentials and history", async () => {
      await context.close();
      await stop(child);
      assert.equal(child.exitCode, 0, "GUI should save on clean shutdown");
      safeOutput();
      origin = await start();
      const saved = await readCredentials(authFile);
      assert.equal(saved.ha1 === credential.ha1 && saved.username === credential.username, true, "restart must preserve credentials");
      assert.equal((await request(origin + "/app.js")).status, 401);
      context = await browser.newContext({ httpCredentials: { username: "ensemble", password } });
      page = await context.newPage();
      page.setDefaultTimeout(8000);
      await page.goto(`${origin}/?speech=transcript`);
      await replies(1).waitFor();
      assert.equal(await replies(1).count(), 1);
      assert.equal(modelCalls, 1, "restoration must not invoke the model");
      await chat(2);
      await context.close();
      await stop(child);
      safeOutput();
    });
  } finally {
    clearTimeout(watchdog);
    await browser?.close();
    await stop(child);
    provider.closeAllConnections();
    await new Promise(resolve => provider.close(resolve));
    await rm(directory, { recursive: true, force: true });
  }
});

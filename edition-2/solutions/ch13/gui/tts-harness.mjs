import { chromium } from "playwright";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import net from "node:net";
import { fileURLToPath } from "node:url";
import { spawn, spawnSync } from "node:child_process";

if (process.argv[2] === "--describe") {
  console.log(
    JSON.stringify({
      read_file_tool: "read_file",
      read_file_arg: "path",
      supports_type_during_turn: true,
      supports_streaming_off: true,
      log_format: "jsonl",
    }),
  );
  process.exit(0);
}
const directory = path.dirname(fileURLToPath(import.meta.url));
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), "ensemble-tts-"));
let browser, child;
try {
  const build = spawnSync(
    "go",
    ["build", "-o", path.join(scratch, "gui"), "./cmd"],
    { cwd: directory, encoding: "utf8" },
  );
  if (build.status !== 0) throw new Error(build.stderr);
  const listener = net.createServer();
  await new Promise((resolve) => listener.listen(0, "127.0.0.1", resolve));
  const port = listener.address().port;
  await new Promise((resolve) => listener.close(resolve));
  child = spawn(
    path.join(scratch, "gui"),
    [
      "--port",
      String(port),
      "--gui-dir",
      path.join(directory, "web/gui"),
      "--workspace",
      process.env.WORKSPACE || scratch,
      ...(process.env.STREAMING === "off" ? ["--no-stream"] : []),
    ],
    {
      cwd: scratch,
      stdio: ["pipe", "ignore", "pipe"],
      env: {
        PATH: process.env.PATH,
        HOME: process.env.HOME,
        LLM_VENDOR: "anthropic",
        LLM_MODEL: "claude-haiku-5-5",
        LLM_API_KEY: "fake",
        LLM_BASE_URL: process.env.LLM_BASE_URL,
      },
    },
  );
  let diagnostic = "";
  child.stderr.on("data", (data) => {
    diagnostic += data;
  });
  const url = `http://127.0.0.1:${port}`;
  for (let i = 0; i < 200; i++) {
    try {
      if ((await fetch(url)).ok) break;
    } catch {}
    if (child.exitCode !== null) throw new Error(diagnostic);
    await new Promise((resolve) => setTimeout(resolve, 25));
  }
  const chrome = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
  browser = await chromium.launch({
    executablePath:
      process.env.CHROME_BIN || (fs.existsSync(chrome) ? chrome : undefined),
  });
  const page = await browser.newPage();
  await page.addInitScript(() => {
    const start = Date.now();
    window.speechEvents = [];
    const record = (entry) =>
      window.speechEvents.push({
        ms: Date.now() - start,
        seq: window.speechEvents.length + 1,
        ...entry,
      });
    window.addEventListener("ensemble-speech", (event) =>
      record({ kind: "utterance", ...event.detail }),
    );
    // Observe real outgoing gate commands without changing the WebSocket peer.
    const original = WebSocket.prototype.send;
    WebSocket.prototype.send = function (data) {
      const message = JSON.parse(data);
      if (["pause", "unpause"].includes(message.type))
        record({
          kind: message.type === "pause" ? "pause" : "resume",
          cause: document.querySelector("#input").value.length
            ? "typing"
            : "speech",
        });
      return original.call(this, data);
    };
    // The only fake is the external synthesizer: asynchronous start/end,
    // cancellation and paused state, with the real app's queue/renderer intact.
    let timer;
    Object.defineProperty(window, "speechSynthesis", {
      value: {
        paused: false,
        speak(utterance) {
          utterance.onstart?.({ type: "start" });
          timer = setTimeout(() => utterance.onend?.({ type: "end" }), 150);
        },
        cancel() {
          clearTimeout(timer);
        },
      },
    });
  });
  await page.goto(url);
  await page.waitForFunction(
    () => document.querySelector("#state").textContent !== "connecting",
  );
  await page.locator("#input").fill(process.argv[2]);
  await page.locator("#input").press("Enter");
  if (process.argv[3] === "--type-during-turn") {
    await page.waitForFunction(() => window.ensembleSpeech.active);
    await page.locator("#input").fill(process.argv[4]);
    await page.waitForFunction(
      () =>
        !window.ensembleSpeech.active && !window.ensembleSpeech.queue.length,
    );
    await page.locator("#input").fill("");
  }
  await page.waitForFunction(() => window.ensembleTurn > 0);
  await page.waitForFunction(
    () => !window.ensembleSpeech.active && !window.ensembleSpeech.queue.length,
  );
  fs.writeFileSync(
    process.env.TTS_LOG,
    (await page.evaluate(() => window.speechEvents))
      .map((entry) => JSON.stringify(entry))
      .join("\n") + "\n",
  );
} finally {
  await browser?.close();
  if (child && child.exitCode === null) {
    const exited = new Promise((resolve) => child.once("exit", resolve));
    child.kill("SIGTERM");
    await exited;
  }
  fs.rmSync(scratch, { recursive: true, force: true });
}

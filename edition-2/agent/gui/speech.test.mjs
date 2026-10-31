// Real browser, GUI, Agent, jobs and parser; only external provider/speech peers are fakes.
import { test } from "node:test";
import assert from "node:assert/strict";
import { chromium } from "playwright";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import http from "node:http";
import { spawn, spawnSync } from "node:child_process";
import { createInterface } from "node:readline";
const directory = path.dirname(new URL(import.meta.url).pathname);
const chrome =
  process.env.CHROME_BIN ||
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
test(
  "speech pipeline, real delivery and blind capabilities",
  { timeout: 60000 },
  async (t) => {
    const scratch = fs.mkdtempSync(
      path.join(os.tmpdir(), "ensemble-speech-test-"),
    );
    const built = spawnSync("go", ["build", "-o", scratch + "/gui", "./cmd"], {
      cwd: directory,
      encoding: "utf8",
    });
    assert.equal(built.status, 0, built.stderr);
    let mode = "whole",
      calls = 0,
      driverTools;
    const provider = http.createServer(async (req, res) => {
      let bytes = "";
      for await (const chunk of req) bytes += chunk;
      const body = JSON.parse(bytes);
      calls++;
      if (mode === "whole")
        await new Promise((resolve) => setTimeout(resolve, 100));
      if (mode === "driver") driverTools = body.tools.map((tool) => tool.name);
      if (mode === "error") {
        res.writeHead(500);
        res.end("SPEECH_ERROR");
        return;
      }
      const parts =
        mode === "tool" && calls === 1
          ? [
              { type: "text", text: "Before tool:" },
              {
                type: "tool_use",
                id: "r",
                name: "read_file",
                input: { path: "result.txt" },
              },
            ]
          : [{ type: "text", text: "Whole queueChunk and max_tool_rounds." }];
      if (!body.stream) {
        res.setHeader("Content-Type", "application/json");
        res.end(
          JSON.stringify({
            content: parts,
            usage: { input_tokens: 1, output_tokens: 1 },
          }),
        );
        return;
      }
      res.setHeader("Content-Type", "text/event-stream");
      const frame = (value) =>
        res.write("data: " + JSON.stringify(value) + "\n\n");
      frame({
        type: "message_start",
        message: { usage: { input_tokens: 1, output_tokens: 1 } },
      });
      frame({
        type: "content_block_start",
        index: 0,
        content_block: { type: "text", text: "" },
      });
      for (const text of ["A bro", "ken word\nwraps.", "\n\nTail:"])
        frame({
          type: "content_block_delta",
          index: 0,
          delta: { type: "text_delta", text },
        });
      frame({ type: "content_block_stop", index: 0 });
      frame({ type: "message_stop" });
      res.end();
    });
    await new Promise((r) => provider.listen(0, "127.0.0.1", r));
    fs.writeFileSync(scratch + "/result.txt", "UNSPOKEN_TOOL_RESULT");
    fs.mkdirSync(scratch + "/skills/target", { recursive: true });
    fs.writeFileSync(
      scratch + "/skills/target/SKILL.md",
      "---\nname: target\ntype: primary\ntools: read_file\n---\nAnswer the user.\n",
    );
    let child, browser, peer;
    const launch = async (stream = false) => {
      const child = spawn(
        scratch + "/gui",
        [
          "--port",
          "19414",
          "--gui-dir",
          directory + "/web/gui",
          ...(stream ? [] : ["--no-stream"]),
        ],
        {
          cwd: scratch,
          env: {
            PATH: process.env.PATH,
            HOME: process.env.HOME,
            EN_SKILLS_DIR: scratch + "/skills",
            EN_PRIMARY_SKILL: "target",
            LLM_VENDOR: "anthropic",
            LLM_MODEL: "claude-sonnet-5",
            LLM_API_KEY: "fake",
            LLM_BASE_URL: "http://127.0.0.1:" + provider.address().port,
          },
          stdio: ["pipe", "ignore", "pipe"],
        },
      );
      let stderr = "";
      child.stderr.on("data", (d) => (stderr += d));
      for (let i = 0; i < 100; i++) {
        try {
          if ((await fetch("http://127.0.0.1:19414")).ok) return child;
        } catch {}
        if (child.exitCode !== null) throw Error(stderr);
        await new Promise((r) => setTimeout(r, 20));
      }
      throw Error("GUI did not start");
    };
    const stop = async (p) => {
      if (p && p.exitCode === null) {
        const done = new Promise((r) => p.once("exit", r));
        if (p === peer) p.stdin.end();
        else p.kill("SIGTERM");
        await done;
      }
    };
    try {
      child = await launch();
      browser = await chromium.launch({ executablePath: chrome });
      const page = await browser.newPage();
      await page.addInitScript(() => {
        window.nativeUtterances = [];
        window.gateCommands = [];
        const send = WebSocket.prototype.send;
        WebSocket.prototype.send = function (data) {
          const message = JSON.parse(data);
          if (["pause", "unpause"].includes(message.type))
            window.gateCommands.push(message.type);
          return send.call(this, data);
        };
        Object.defineProperty(window, "speechSynthesis", {
          value: {
            paused: false,
            speak(u) {
              window.nativeUtterances.push(u);
              u.onstart?.({ type: "start" });
            },
            cancel() {},
          },
        });
      });
      await page.goto("http://127.0.0.1:19414?speech=transcript");
      await page.waitForFunction(
        () =>
          window.ensembleSpeech &&
          document.querySelector("#state").textContent !== "connecting",
      );
      const heard = () => page.evaluate(() => window.ensembleSpeech.transcript);
      const clear = () =>
        page.evaluate(() => {
          window.ensembleSpeech.cancel();
          window.ensembleSpeech.transcript = [];
        });
      await page.evaluate(() => {
        const s = window.ensembleSpeech;
        s.chunk(1, "A bro");
        s.chunk(1, "ken wor");
        s.chunk(1, "d wraps\nwithout a pause.");
        s.chunk(1, "\n\nTrailing:");
        s.flush();
      });
      assert.deepEqual(
        (await heard()).map((x) => x.text),
        ["A broken word wraps without a pause.", "Trailing:"],
        "fragments/wrap/boundaries/flush",
      );
      await clear();
      await page.evaluate(() => {
        const s = window.ensembleSpeech;
        s.chunk(5, "First paragraph\n\nSecond paragraph");
        s.flush();
      });
      assert.deepEqual(
        (await heard()).map((x) => x.text),
        ["First paragraph", "Second paragraph"],
        "blank line alone ends a phrase",
      );
      for (const split of [false, true]) {
        await clear();
        await page.evaluate((split) => {
          const s = window.ensembleSpeech;
          const text =
            "### Heading\n- **queueChunk** uses `max_tool_rounds` and HTTPServer.\n\n[Read more](https://example.test).\n```go\nDO_NOT_SPEAK();\n```\nAfter:";
          for (const chunk of split ? [...text] : [text])
            s.chunk(2, chunk, false, "thinking");
          s.flush();
        }, split);
        assert.deepEqual(
          (await heard()).map((x) => x.text),
          [
            "Heading queue Chunk uses max tool rounds and HTTP Server.",
            "Read more.",
            "code block After:",
          ],
          "markup and fences " + split,
        );
        assert.ok(
          (await heard()).every((x) => x.source === "thinking"),
          "source attribution",
        );
      }
      await clear();
      await page.evaluate(() =>
        window.ensembleSpeech.chunk(3, "Hello ordinary World.", true),
      );
      assert.equal(
        (await heard())[0].text,
        "Hello ordinary World.",
        "ordinary capitalization",
      );
      assert.equal(
        await page.evaluate(() => nativeUtterances.length),
        0,
        "transcript bypass has no native audio",
      );
      await t.test("5x slider, Eloquence calibration and reload", async () => {
        // Exercise the real slider, settings broadcast and speech queue. Only
        // the platform audio boundary is recorded; no timing claim is inferred.
        await page.locator("#settings-tab").click();
        assert.equal(await page.locator("#tts-speed").getAttribute("max"), "5");
        for (const [speed, rate] of [[0.5, 0.5], [1.7, 1.7], [2, 2], [3, 2.4], [3.5, 2.6], [4.2, 2.74], [5, 2.9]]) {
          await page.locator("#tts-speed").fill(String(speed));
          await page.waitForFunction(s => Number(document.querySelector("#speed-value").value) === s, speed);
          const actual = await page.evaluate(() => {
            window.ensembleSpeech.cancel();
            window.ensembleSpeech.audio = true;
            window.ensembleSpeech.enqueue("Calibrated speed.", "text");
            const rate = window.nativeUtterances.at(-1).rate;
            window.ensembleSpeech.cancel();
            window.ensembleSpeech.audio = false;
            return rate;
          });
          assert.ok(Math.abs(actual - rate) < 1e-6, `${speed}x uses native rate ${rate}, got ${actual}`);
        }
        await page.reload();
        await page.waitForFunction(() => Number(document.querySelector("#speed-value").value) === 5);
        await page.locator("#settings-tab").click();
        assert.equal(await page.locator("#tts-speed").inputValue(), "5", "high speed survives reload");
        await page.locator("#tts-speed").fill("1");
        await page.waitForFunction(() => Number(document.querySelector("#speed-value").value) === 1);
      });
      // Drop transient deltas at the external network boundary, retaining actual
      // finals. Plain provider replies normally get synthetic deltas from Engine.
      await page.routeWebSocket("**/ws", (ws) => {
        const server = ws.connectToServer();
        server.onMessage((message) => {
          if (JSON.parse(message).type !== "part_delta") ws.send(message);
        });
      });
      await page.reload();
      await page.waitForFunction(
        () =>
          window.ensembleSpeech &&
          document.querySelector("#state").textContent !== "connecting",
      );
      await clear();
      const submit = async (text) => {
        const turn = await page.evaluate(() => window.ensembleTurn);
        await page.locator("#input").fill(text);
        await page.locator("#input").press("Enter");
        await page.waitForFunction((t) => window.ensembleTurn > t, turn);
      };
      await submit("whole delivery");
      assert.deepEqual(
        await heard(),
        [{ text: "Whole queue Chunk and max tool rounds.", source: "text" }],
        "whole final is spoken",
      );
      await clear();
      await page.evaluate(() =>
        window.ensembleSpeech.chunk(81, "Tool boundary tail:"),
      );
      mode = "tool";
      ((calls = 0), driverTools);
      await submit("read file");
      assert.deepEqual(
        (await heard()).map((x) => x.text),
        [
          "Before tool:",
          "Tool boundary tail:",
          "read file",
          "Whole queue Chunk and max tool rounds.",
        ],
        "tool boundary order and silent result",
      );
      assert.ok(
        (await page.locator("#actions").innerText()).includes(
          "UNSPOKEN_TOOL_RESULT",
        ),
        "real tool result displayed",
      );
      await clear();
      await page.evaluate(() =>
        window.ensembleSpeech.chunk(80, "Unfinished error tail:"),
      );
      mode = "error";
      await submit("error");
      assert.equal(
        (await heard())[0].text,
        "Unfinished error tail:",
        "error flushes unfinished text",
      );
      assert.ok(
        (await heard()).some(
          (x) => x.source === "error" && x.text.includes("500"),
        ),
        "displayed provider error spoken",
      );
      // Controlled external completion tests each ordering of the two real causes.
      await clear();
      await page.evaluate(() => {
        window.gateCommands = [];
        window.ensembleSpeech.audio = true;
        window.ensembleSpeech.enqueue("Gate phrase.", "text");
      });
      await page.waitForFunction(
        () => document.querySelector("#pause-state").textContent === "paused",
      );
      await page.locator("#input").fill(" ");
      assert.deepEqual(
        await page.evaluate(() => window.gateCommands),
        ["pause"],
        "unchanged cause emits no second pause",
      );
      await page.evaluate(() =>
        window.ensembleSpeech.active.onend({ type: "end" }),
      );
      assert.equal(
        await page.locator("#pause-state").innerText(),
        "paused",
        "speech end cannot release typing",
      );
      await page.locator("#input").fill("");
      await page.waitForFunction(
        () => document.querySelector("#pause-state").textContent === "",
      );
      await page.locator("#input").fill(" ");
      await page.evaluate(() => window.ensembleSpeech.enqueue("Second gate."));
      await page.locator("#input").fill("");
      assert.equal(
        await page.locator("#pause-state").innerText(),
        "paused",
        "typing end cannot release speech",
      );
      await page.evaluate(() =>
        window.ensembleSpeech.active.onend({ type: "end" }),
      );
      await page.waitForFunction(
        () => document.querySelector("#pause-state").textContent === "",
      );
      await page.locator("#input").fill("typed");
      await page.evaluate(() => window.ensembleSpeech.enqueue("Cancel me."));
      await page.keyboard.press("Escape");
      assert.equal(
        await page.evaluate(() => window.ensembleSpeech.active),
        null,
        "Escape cancels with nonempty input",
      );
      assert.equal(
        await page.locator("#pause-state").innerText(),
        "paused",
        "Escape preserves typing pause",
      );
      await page.locator("#input").fill("");
      // The external persona's exact registry and dispatch agree; guessed DOM tools fail.
      mode = "whole";
      await page.close();
      peer = spawn(
        process.execPath,
        [
          directory + "/virtual-user.mjs",
          "--agent-url",
          "ws://127.0.0.1:19414/ws",
          "--persona",
          "blind",
          "--browser",
          chrome,
        ],
        {
          env: { PATH: process.env.PATH, HOME: process.env.HOME },
          stdio: ["pipe", "pipe", "pipe"],
        },
      );
      const lines = createInterface({ input: peer.stdout })[
        Symbol.asyncIterator
      ]();
      let id = 0;
      const rpc = async (method, params) => {
        peer.stdin.write(
          JSON.stringify({ jsonrpc: "2.0", id: ++id, method, params }) + "\n",
        );
        const line = await lines.next();
        assert.equal(line.done, false);
        return JSON.parse(line.value).result;
      };
      assert.deepEqual(
        (await rpc("tools/list")).tools.map((x) => x.name),
        ["hear", "type", "submit", "wait", "sleep"],
        "exact blind registry",
      );
      for (const name of ["gui_snapshot", "gui_click"])
        assert.equal(
          (await rpc("tools/call", { name })).isError,
          true,
          "blind rejects " + name,
        );
      await rpc("tools/call", {
        name: "type",
        arguments: { text: "Describe identifiers" },
      });
      await rpc("tools/call", { name: "submit" });
      const listened = await rpc("tools/call", { name: "wait" });
      assert.deepEqual(
        JSON.parse(listened.content[0].text),
        [{ text: "Whole queue Chunk and max tool rounds.", source: "text" }],
        "blind completion hears actual pipeline",
      );
      assert.equal(
        (
          await rpc("tools/call", {
            name: "sleep",
            arguments: { milliseconds: 1 },
          })
        ).isError,
        undefined,
      );
      assert.deepEqual(
        JSON.parse((await rpc("tools/call", { name: "hear" })).content[0].text),
        [],
        "hear cursor avoids duplicates",
      );
      peer.stdin.end();
      await new Promise((r) => peer.once("exit", r));
      peer = null;
      // Check the shipped client, too: a narrow MCP peer cannot prove the
      // consuming Agent omitted unrelated builtins and skill lifecycle tools.
      const clientBuild = spawnSync(
        "go",
        ["build", "-o", scratch + "/listener", "./cmd/virtual-user"],
        { cwd: directory, encoding: "utf8" },
      );
      assert.equal(clientBuild.status, 0, clientBuild.stderr);
      mode = "driver";
      const client = spawn(
        scratch + "/listener",
        [
          "--agent-url",
          "http://127.0.0.1:19414",
          "--peer",
          directory + "/virtual-user.mjs",
          "--browser",
          chrome,
          "--task",
          "Report your available capabilities.",
        ],
        {
          cwd: scratch,
          env: {
            PATH: process.env.PATH,
            HOME: process.env.HOME,
            LLM_VENDOR: "anthropic",
            LLM_MODEL: "claude-sonnet-5",
            LLM_API_KEY: "fake",
            LLM_BASE_URL: "http://127.0.0.1:" + provider.address().port,
            EN_DISABLE_STREAMING: "1",
          },
          stdio: ["ignore", "pipe", "pipe"],
        },
      );
      let clientError = "";
      client.stderr.on("data", (data) => (clientError += data));
      client.stdout.resume();
      assert.equal(
        await new Promise((resolve) => client.once("exit", resolve)),
        0,
        "actual virtual client: " + clientError,
      );
      assert.deepEqual(
        driverTools.slice().sort(),
        ["hear", "sleep", "submit", "type", "wait"],
        "actual model receives exactly five persona tools",
      );
      mode = "whole";
      await stop(child);
      child = null;
      fs.rmSync(scratch + "/save.json", { force: true });
      child = await launch(true);
      const streamPage = await browser.newPage();
      await streamPage.goto("http://127.0.0.1:19414?speech=transcript");
      await streamPage.waitForFunction(
        () =>
          window.ensembleSpeech &&
          document.querySelector("#state").textContent !== "connecting",
      );
      await streamPage.locator("#input").fill("stream");
      await streamPage.locator("#input").press("Enter");
      await streamPage.waitForFunction(() => window.ensembleTurn > 0);
      assert.deepEqual(
        await streamPage.evaluate(() => window.ensembleSpeech.transcript),
        [
          { text: "A broken word wraps.", source: "text" },
          { text: "Tail:", source: "text" },
        ],
        "streamed final not duplicated",
      );
    } finally {
      await stop(peer);
      await browser?.close();
      await stop(child);
      await new Promise((r) => provider.close(r));
      fs.rmSync(scratch, { recursive: true, force: true });
    }
  },
);

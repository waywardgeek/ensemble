// Exercise owned peer lifetime through actual Chrome, GUI and the public client.
// The only fake is the external provider returning a long sleep then a final.
import { test } from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import http from "node:http";
import { spawn, spawnSync } from "node:child_process";
import { createInterface } from "node:readline";

const directory = path.dirname(fileURLToPath(import.meta.url));
const chrome =
  process.env.CHROME_BIN ||
  "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome";
// Bounds are assertions about shutdown, not limits on the product's sleep tool.
async function within(promise, milliseconds, label = "shutdown") {
  let timer;
  try {
    return await Promise.race([
      promise,
      new Promise((_, reject) => {
        timer = setTimeout(
          () => reject(new Error(`${label} exceeded bound`)),
          milliseconds,
        );
      }),
    ]);
  } finally {
    clearTimeout(timer);
  }
}

test(
  "peer EOF cancels pending sleep and the actual client can finish",
  { timeout: 20000 },
  async (t) => {
    const scratch = fs.mkdtempSync(
      path.join(os.tmpdir(), "ensemble-peer-lifetime-"),
    );
    let gui, peer, client;
    let sleeping = false,
      requests = [],
      finalSent;
    const provider = http.createServer(async (req, res) => {
      let bytes = "";
      for await (const chunk of req) bytes += chunk;
      requests.push(JSON.parse(bytes));
      const content =
        sleeping && requests.length === 1
          ? [
              {
                type: "tool_use",
                id: "long-sleep",
                name: "sleep",
                input: { milliseconds: 60000 },
              },
            ]
          : [{ type: "text", text: "Listener turn finished." }];
      res.setHeader("Content-Type", "application/json");
      res.end(
        JSON.stringify({
          content,
          usage: { input_tokens: 1, output_tokens: 1 },
        }),
      );
      if (content[0].type === "text") finalSent?.();
    });
    await new Promise((resolve) => provider.listen(0, "127.0.0.1", resolve));
    const env = {
      PATH: process.env.PATH,
      HOME: process.env.HOME,
      LLM_VENDOR: "anthropic",
      LLM_MODEL: "claude-sonnet-5",
      LLM_API_KEY: "fake",
      LLM_BASE_URL: "http://127.0.0.1:" + provider.address().port,
      EN_DISABLE_STREAMING: "1",
    };
    const owned = (command, args, options = {}) => {
      const process = spawn(command, args, {
        cwd: scratch,
        env,
        stdio: ["pipe", "pipe", "pipe"],
        ...options,
      });
      process.errors = "";
      process.stderr.on("data", (data) => (process.errors += data));
      process.exited = new Promise((resolve) =>
        process.once("exit", (code, signal) => resolve({ code, signal })),
      );
      return process;
    };
    const stop = async (process) => {
      if (!process || process.exitCode !== null || process.signalCode !== null)
        return;
      process.kill("SIGTERM");
      await within(process.exited, 2000);
    };
    try {
      for (const [name, entry] of [
        ["gui", "./cmd"],
        ["listener", "./cmd/virtual-user"],
      ]) {
        const result = spawnSync(
          "go",
          ["build", "-o", scratch + "/" + name, entry],
          { cwd: directory, encoding: "utf8" },
        );
        assert.equal(result.status, 0, result.stderr);
      }
      gui = owned(scratch + "/gui", [
        "--port",
        "19416",
        "--gui-dir",
        directory + "/web/gui",
      ]);
      gui.stdout.resume();
      for (let i = 0; i < 100; i++) {
        try {
          if ((await fetch("http://127.0.0.1:19416")).ok) break;
        } catch {}
        if (gui.exitCode !== null) throw new Error(gui.errors);
        await new Promise((resolve) => setTimeout(resolve, 20));
      }
      await t.test(
        "peer keeps normal sleep but exits on EOF during sleep",
        async () => {
          for (const sleep of [0, 4000]) {
            peer = owned(
              process.execPath,
              [
                directory + "/virtual-user.mjs",
                "--agent-url",
                "http://127.0.0.1:19416",
                "--browser",
                chrome,
              ],
              { env: { PATH: process.env.PATH, HOME: process.env.HOME } },
            );
            const lines = createInterface({ input: peer.stdout })[
              Symbol.asyncIterator
            ]();
            let id = 0;
            const request = (method, params) =>
              peer.stdin.write(
                JSON.stringify({ jsonrpc: "2.0", id: ++id, method, params }) +
                  "\n",
              );
            request("initialize");
            assert.ok(
              JSON.parse((await lines.next()).value).result.protocolVersion,
              "real peer initialized",
            );
            if (!sleep) {
              const start = performance.now();
              request("tools/call", {
                name: "sleep",
                arguments: { milliseconds: 150 },
              });
              assert.equal(
                JSON.parse((await lines.next()).value).result.content[0].text,
                "slept",
              );
              assert.ok(
                performance.now() - start >= 140,
                "normal sleep retains its requested delay",
              );
              request("tools/call", { name: "wait" });
              assert.equal(
                JSON.parse((await lines.next()).value).result.content[0].text,
                "[]",
                "ordinary idle wait still works",
              );
            } else {
              request("tools/call", {
                name: "sleep",
                arguments: { milliseconds: sleep },
              });
              // The initialized peer has entered its asynchronous timer before EOF.
              await new Promise((resolve) => setTimeout(resolve, 100));
            }
            peer.stdin.end();
            assert.deepEqual(
              await within(
                peer.exited,
                1000,
                sleep ? "pending sleep EOF" : "idle EOF",
              ),
              { code: 0, signal: null },
              sleep ? "pending sleep EOF exits" : "idle EOF exits",
            );
            peer = null;
          }
        },
      );
      // First prove ordinary application shutdown. Then leave a real 60-second
      // MCP job active: the existing job callback lets the model finish its turn.
      await t.test(
        "real client exits after its answer with a sleep job still active",
        async () => {
          for (sleeping of [false, true]) {
            requests = [];
            const answered = new Promise((resolve) => (finalSent = resolve));
            client = owned(
              scratch + "/listener",
              [
                "--agent-url",
                "http://127.0.0.1:19416",
                "--peer",
                directory + "/virtual-user.mjs",
                "--browser",
                chrome,
                "--task",
                "Exercise the speech peer.",
              ],
              { detached: true },
            );
            let answer = "";
            client.stdout.on("data", (data) => (answer += data));
            await within(answered, 8000);
            if (sleeping) {
              assert.equal(requests.length, 2);
              const results = requests[1].messages
                .flatMap((message) => message.content || [])
                .filter((part) => part.type === "tool_result");
              assert.ok(
                results.some((part) =>
                  JSON.stringify(part.content).includes("running"),
                ),
                "actual job callback returned running before the final answer",
              );
            }
            assert.deepEqual(
              await within(
                client.exited,
                1500,
                "client exit with pending sleep=" + sleeping,
              ),
              { code: 0, signal: null },
              "actual client exits after final answer with pending sleep=" +
                sleeping,
            );
            assert.ok(answer.includes("Listener turn finished."));
            client = null;
          }
        },
      );
    } finally {
      if (client && client.exitCode === null) {
        // A failed lifetime assertion must clean up the client AND its peer.
        try {
          process.kill(-client.pid, "SIGTERM");
        } catch {}
        await within(client.exited, 2000);
      }
      await stop(peer);
      await stop(gui);
      await new Promise((resolve) => provider.close(resolve));
      fs.rmSync(scratch, { recursive: true, force: true });
    }
  },
);

import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import http from "node:http";
import net from "node:net";
import { spawn, execFileSync } from "node:child_process";
import { once } from "node:events";
import { setTimeout as delay } from "node:timers/promises";
import { firefox } from "playwright";



function pdf(label) {
  const content = `BT /F1 12 Tf 20 100 Td (${label}) Tj ET\n`;
  const objects = [
    "<< /Type /Catalog /Pages 2 0 R >>",
    "<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
    "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
    "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
    `<< /Length ${Buffer.byteLength(content)} >>\nstream\n${content}endstream`,
  ];
  let bytes = Buffer.concat([Buffer.from("%PDF-1.4\n%"), Buffer.from([0xe2, 0xe3, 0xcf, 0xd3, 0, 0xff]), Buffer.from("\n")]);
  const offsets = [0];
  for (const [index, object] of objects.entries()) {
    offsets.push(bytes.length);
    bytes = Buffer.concat([bytes, Buffer.from(`${index + 1} 0 obj\n${object}\nendobj\n`)]);
  }
  const xref = bytes.length;
  const entries = offsets.slice(1).map(offset => `${String(offset).padStart(10, "0")} 00000 n \n`).join("");
  return Buffer.concat([bytes, Buffer.from(`xref\n0 6\n0000000000 65535 f \n${entries}trailer\n<< /Size 6 /Root 1 0 R >>\nstartxref\n${xref}\n%%EOF\n`)]);
}

for (const streaming of [false, true]) {
  test(`offer_download uses real tools and Firefox downloads (stream=${streaming})`, { timeout: 120000 }, async () => {
    const directory = await fs.mkdtemp(path.join(os.tmpdir(), "ensemble-downloads-"));
    const workspace = path.join(directory, "workspace");
    const binary = path.join(directory, "gui");
    const filename = "Report #1.pdf";
    const mainPDF = pdf("Main agent PDF");
    const peerPDF = pdf("Peer agent PDF");
    const model = "gpt-4o-mini";
    const requests = [], providerErrors = [], frames = [], sent = [], pageErrors = [];
    let scenario, browser, child, processError;
    let processLog = "";
    const provider = http.createServer(async (request, response) => {
      try {
        let body = "";
        for await (const chunk of request) body += chunk;
        const value = JSON.parse(body);
        requests.push(value);
        assert.equal(request.method, "POST");
        assert.equal(request.url, "/v1/chat/completions");
        assert.equal(value.model, model);
        assert.equal(Boolean(value.stream), streaming);
        assert.ok(scenario, "only a submitted user request may launch the provider");
        const user = value.messages.filter(message => message.role === "user").at(-1);
        const userText = typeof user?.content === "string" ? user.content : user?.content?.map(part => part.text || "").join("");
        assert.equal(userText, scenario.prompt, "the actual user request must reach the model");
        const offers = (value.tools || []).filter(tool => tool.function?.name === "offer_download");
        assert.equal(offers.length, 1, "the model must see the built-in offer_download tool exactly once");
        assert.equal(offers[0].type, "function");
        const schema = offers[0].function.parameters;
        assert.equal(schema.type, "object");
        assert.deepEqual(Object.keys(schema.properties), ["path"]);
        assert.equal(schema.properties.path.type, "string");
        assert.deepEqual(schema.required, ["path"]);
        let message = { role: "assistant", content: scenario.answer };
        let finish = "stop";
        if (scenario.args && scenario.phase === 0) {
          assert.match(scenario.prompt, /please offer .* for download/i);
          message = { role: "assistant", content: "Preparing your requested download.", tool_calls: [{
            id: scenario.callID, type: "function",
            function: { name: "offer_download", arguments: JSON.stringify(scenario.args) },
          }] };
          finish = "tool_calls";
        } else if (scenario.args) {
          assert.equal(scenario.phase, 1, "one tool call followed by one normal answer");
          const results = value.messages.filter(message => message.role === "tool" && message.tool_call_id === scenario.callID);
          assert.equal(results.length, 1, "the real tool result must return to the model");
          assert.ok(results[0].content.length > 0);
        } else {
          assert.equal(scenario.phase, 0, "ordinary replies must not launch tools or another request");
        }
        scenario.phase++;
        const usage = { prompt_tokens: 100, completion_tokens: 20 };
        if (streaming) {
          response.setHeader("Content-Type", "text/event-stream");
          if (message.tool_calls && message.content) {
            response.write(`data: ${JSON.stringify({ choices: [{ index: 0, delta: { content: message.content } }] })}\n\n`);
          }
          const delta = message.tool_calls ? { tool_calls: message.tool_calls.map(call => ({ index: 0, ...call })) } : { content: message.content };
          response.write(`data: ${JSON.stringify({ choices: [{ index: 0, delta }] })}\n\n`);
          response.write(`data: ${JSON.stringify({ choices: [{ index: 0, delta: {}, finish_reason: finish }] })}\n\n`);
          response.end(`data: ${JSON.stringify({ choices: [], usage })}\n\ndata: [DONE]\n\n`);
        } else {
          response.setHeader("Content-Type", "application/json");
          response.end(JSON.stringify({ choices: [{ message, finish_reason: finish }], usage }));
        }
      } catch (error) {
        providerErrors.push(error.message);
        response.writeHead(500).end("Download fixture rejected the provider request");
      }
    });
    try {
      await fs.mkdir(path.join(workspace, "main"), { recursive: true });
      await fs.mkdir(path.join(workspace, "peer"));
      await fs.writeFile(path.join(workspace, "main", filename), mainPDF);
      await fs.writeFile(path.join(workspace, "peer", filename), peerPDF);
      execFileSync("go", ["build", "-o", binary, "./cmd"], { cwd: new URL(".", import.meta.url), timeout: 60000 });
      provider.listen(0, "127.0.0.1");
      await once(provider, "listening");
      const probe = net.createServer();
      probe.listen(0, "127.0.0.1");
      await once(probe, "listening");
      const port = probe.address().port;
      await new Promise(resolve => probe.close(resolve));
      const url = `http://127.0.0.1:${port}/?speech=transcript`;
      const start = async () => {
        processError = undefined;
        processLog = "";
        child = spawn(binary, ["--port", String(port), "--workspace", workspace, ...(streaming ? [] : ["--no-stream"])], {
          cwd: directory,
          env: {
            PATH: process.env.PATH, HOME: directory, EN_HOME: path.join(directory, "auth"),
            LLM_VENDOR: "openai", LLM_MODEL: model, LLM_API_KEY: "fixture",
            LLM_BASE_URL: `http://127.0.0.1:${provider.address().port}`, OPENAI_BILLING_MODE: "api",
            EN_RECALL_DISABLED: "1", EN_WEB_SEARCH: "0",
          },
          stdio: ["pipe", "pipe", "pipe"],
        });
        child.on("error", error => { processError = error; });
        child.stdout.on("data", bytes => { processLog = (processLog + bytes).slice(-8000); });
        child.stderr.on("data", bytes => { processLog = (processLog + bytes).slice(-8000); });
        let ready = false;
        for (let attempt = 0; attempt < 200 && !ready; attempt++) {
          if (processError) throw processError;
          assert.equal(child.exitCode, null, processLog);
          try { ready = (await fetch(url, { signal: AbortSignal.timeout(500) })).ok; } catch {}
          if (!ready) await delay(25);
        }
        assert.ok(ready, `isolated GUI did not start: ${processLog}`);
      };
      const stop = async () => {
        const exited = once(child, "exit");
        child.kill("SIGTERM");
        const timer = setTimeout(() => child.kill("SIGKILL"), 3000);
        try { await exited; } finally { clearTimeout(timer); }
        assert.equal(child.exitCode, 0, `isolated GUI did not shut down cleanly: ${processLog}`);
      };
      await start();
      browser = await firefox.launch({ headless: true });
      assert.ok(browser, "Firefox must actually run");
      const page = await browser.newPage({ acceptDownloads: true });
      page.setDefaultTimeout(5000);
      page.on("pageerror", error => pageErrors.push(error.message));
      page.on("websocket", socket => {
        socket.on("framereceived", frame => frames.push(JSON.parse(frame.payload)));
        socket.on("framesent", frame => sent.push(JSON.parse(frame.payload)));
      });
      await page.goto(url);
      await page.waitForFunction(() => document.querySelector("#conversation-title").textContent === "main");
      const buttons = () => page.getByRole("button", { name: /^Download /, includeHidden: true });
      const chatButtons = () => page.locator("#artifacts").getByRole("button", { name: `Download ${filename}`, exact: true });
      const dispatches = () => frames.filter(frame => frame.type === "tool_dispatched" && !frame.replay);
      const select = async key => {
        await page.locator(`[data-agent-key="${key}"]`).click();
        await page.waitForFunction(key => document.querySelector("#conversation-title").textContent === key, key);
      };
      const prompt = async (entry, buttonCount) => {
        scenario = { ...entry, phase: 0 };
        const before = requests.length;
        const dispatched = dispatches().length;
        const turns = await page.evaluate(() => window.ensembleTurn);
        await page.fill("#input", scenario.prompt);
        await page.click("#input-form button");
        await page.waitForFunction(turns => window.ensembleTurn > turns, turns);
        assert.deepEqual(providerErrors, [], "fake provider request contract");
        assert.equal(requests.length - before, scenario.args ? 2 : 1);
        assert.equal(dispatches().length - dispatched, scenario.args ? 1 : 0);
        await page.locator("#artifacts .artifact--text").filter({ hasText: scenario.answer }).waitFor();
        if (scenario.args) {
          const call = dispatches().at(-1);
          assert.equal(call.name, "offer_download");
          assert.equal(call.call_id, scenario.callID);
          assert.deepEqual(call.input, scenario.args);
          const result = frames.filter(frame => frame.type === "tool_finished" && !frame.replay && frame.call_id === scenario.callID);
          assert.equal(result.length, 1);
          assert.equal(result[0].is_error, Boolean(scenario.error), "failed tools must remain failures");
        }
        assert.equal(await buttons().count(), buttonCount, "only successful offers create download buttons");
        scenario = undefined;
      };
      const download = async expected => {
        assert.equal(await buttons().count(), 1, "replay must not duplicate offers or leak another agent's offer");
        assert.equal(await chatButtons().count(), 1, "the accessible download button belongs in chat");
        const before = { requests: requests.length, dispatches: dispatches().length, sent: sent.length, turns: await page.evaluate(() => window.ensembleTurn) };
        const [attachment] = await Promise.all([page.waitForEvent("download"), chatButtons().click()]);
        assert.equal(attachment.suggestedFilename(), filename);
        assert.equal(await attachment.failure(), null);
        assert.deepEqual(await fs.readFile(await attachment.path()), expected, "Firefox must receive the exact binary PDF bytes");
        await delay(150);
        assert.equal(requests.length, before.requests, "clicking must not launch the model");
        assert.equal(dispatches().length, before.dispatches, "clicking must not execute a tool");
        assert.equal(await page.evaluate(() => window.ensembleTurn), before.turns, "clicking must not start a turn");
        assert.deepEqual(sent.slice(before.sent).filter(frame => !["pause", "unpause"].includes(frame.type)), [], "download clicks must not send actor commands");
        assert.deepEqual(providerErrors, []);
      };

      assert.equal(requests.length, 0);
      assert.equal(await buttons().count(), 0);
      await prompt({
        prompt: "Describe the report, without offering a download.",
        answer: `Ordinary prose mentioning main/${filename} and offer_download({"path":"main/${filename}"}) is not an offer.`,
      }, 0);
      for (const [index, args] of [{ path: "missing.pdf" }, { path: "main" }, { path: "" }].entries()) {
        await prompt({
          prompt: `Please offer ${JSON.stringify(args.path)} for download.`, args,
          callID: `failed-${index}`, error: true, answer: `Download failure ${index} was reported.`,
        }, 0);
      }
      await page.reload();
      await page.locator("#artifacts .artifact--text").filter({ hasText: "Download failure 2 was reported." }).waitFor();
      assert.equal(await buttons().count(), 0, "failed offers must stay absent on replay");
      await prompt({
        prompt: `Please offer main/${filename} for download.`, args: { path: `main/${filename}` },
        callID: "main-pdf", answer: "Main report is ready.",
      }, 1);
      await download(mainPDF);
      await download(mainPDF);
      await prompt({ prompt: "Thanks. Reply normally without another download.", answer: "You are welcome." }, 1);
      await page.fill("#agent-name", "peer");
      await page.click("#create-agent button");
      await page.locator('[data-agent-key="peer"]').waitFor();
      await select("peer");
      assert.equal(await buttons().count(), 0, "main's offer must not appear in an empty peer conversation");
      await prompt({
        prompt: `Please offer peer/${filename} for download.`, args: { path: `peer/${filename}` },
        callID: "peer-pdf", answer: "Peer report is ready.",
      }, 1);
      await download(peerPDF);
      const completedRequests = requests.length;
      const completedDispatches = dispatches().length;
      for (const [key, expected, answer] of [["main", mainPDF, "Main report is ready."], ["peer", peerPDF, "Peer report is ready."]]) {
        await select(key);
        await page.locator("#artifacts .artifact--text").filter({ hasText: answer }).waitFor();
        await download(expected);
        await page.reload();
        await page.waitForFunction(() => document.querySelector("#conversation-title").textContent === "main");
        if (key !== "main") await select(key);
        await page.locator("#artifacts .artifact--text").filter({ hasText: answer }).waitFor();
        await download(expected);
      }
      assert.equal(requests.length, completedRequests, "selection and reload must not invoke the model");
      assert.equal(dispatches().length, completedDispatches, "selection and reload must not rerun tools");

      await select("main");
      await page.locator("#artifacts .artifact--text").filter({ hasText: "Main report is ready." }).waitFor();
      const oldCapability = await chatButtons().getAttribute("href");
      assert.match(oldCapability, /^\/download\/[0-9a-f]{64}$/);
      const original = await page.evaluate(async href => {
        const response = await fetch(href, { cache: "no-store" });
        return { status: response.status, bytes: Array.from(new Uint8Array(await response.arrayBuffer())) };
      }, oldCapability);
      assert.equal(original.status, 200, "the original capability must work before restart");
      assert.deepEqual(Buffer.from(original.bytes), mainPDF);
      await stop();
      const calls = async () => JSON.parse(await fs.readFile(path.join(directory, "save.json"), "utf8"))
        .log.filter(event => event.type === "tool_called");
      const callsBeforeRestart = await calls();
      assert.ok(callsBeforeRestart.some(event => event.tool.call_id === "main-pdf"), "the saved history contains the original offer");
      await start();
      await page.reload();
      await page.locator("#artifacts .artifact--text").filter({ hasText: "Main report is ready." }).waitFor();
      await chatButtons().waitFor();
      const newCapability = await chatButtons().getAttribute("href");
      assert.match(newCapability, /^\/download\/[0-9a-f]{64}$/);
      assert.notEqual(newCapability, oldCapability, "restart must issue a fresh capability for the replayed offer");
      const oldStatus = await page.evaluate(async href => (await fetch(href, { cache: "no-store" })).status, oldCapability);
      assert.ok(oldStatus >= 400, `the old capability must fail after restart; got ${oldStatus}`);
      await download(mainPDF);
      assert.equal(requests.length, completedRequests, "restart and replay must not invoke the model");
      assert.equal(dispatches().length, completedDispatches, "restart and replay must not dispatch a tool");
      await stop();
      assert.deepEqual(await calls(), callsBeforeRestart, "restart must not execute tools before a browser subscribes");
      assert.deepEqual(providerErrors, []);
      assert.deepEqual(pageErrors, []);
    } finally {
      await browser?.close();
      if (child?.pid && child.exitCode === null && child.signalCode === null) {
        const exited = once(child, "exit");
        child.kill("SIGTERM");
        const timer = setTimeout(() => child.kill("SIGKILL"), 3000);
        try { await exited; } finally { clearTimeout(timer); }
      }
      provider.closeAllConnections();
      if (provider.listening) await new Promise(resolve => provider.close(resolve));
      await fs.rm(directory, { recursive: true, force: true });
    }
  });
}

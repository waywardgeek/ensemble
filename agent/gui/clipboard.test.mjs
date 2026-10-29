import test from "node:test";
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { networkInterfaces } from "node:os";
import { createServer } from "node:http";
import { firefox } from "playwright";

const lanAddress = Object.values(networkInterfaces()).flat()
  .find(address => address.family === "IPv4" && !address.internal)?.address;

async function serve() {
  const server = createServer(async (request, response) => {
    const name = request.url.slice(1);
    if (!name) {
      response.setHeader("content-type", "text/html");
      response.end('<link rel="stylesheet" href="/style.css"><input id="prompt" value="unsent prompt"><section id="scroll"></section>');
    } else if (["artifact-scroll.js", "renderers.js", "style.css"].includes(name)) {
      response.setHeader("content-type", name.endsWith(".css") ? "text/css" : "text/javascript");
      response.end(await readFile(new URL(`./web/gui/${name}`, import.meta.url)));
    } else response.writeHead(404).end();
  });
  await new Promise(resolve => server.listen(0, "0.0.0.0", resolve));
  return { server, port: server.address().port };
}

for (const scenario of ["secure", "http", "blocked"]) {
  test(`clipboard snippets use real Firefox APIs: ${scenario}`, { timeout: 30000 }, async () => {
    assert.ok(lanAddress, "A non-loopback IPv4 address is required to exercise insecure HTTP");
    const { server, port } = await serve();
    let browser, reader, original;
    try {
      browser = await firefox.launch({ headless: true, firefoxUserPrefs: {
        "dom.events.testing.asyncClipboard": true,
        ...(scenario !== "http" ? { "dom.allow_cut_copy": false } : {}),
      } });
      const context = await browser.newContext();
      reader = await context.newPage();
      await reader.goto(`http://127.0.0.1:${port}`);
      original = await reader.evaluate(() => navigator.clipboard.readText());
      const page = await context.newPage();
      const errors = [];
      page.on("pageerror", error => errors.push(error.message));
      await page.goto(`http://${scenario === "secure" ? "127.0.0.1" : lanAddress}:${port}`);
      const capabilities = await page.evaluate(() => ({ secure: isSecureContext, clipboard: !!navigator.clipboard }));
      assert.deepEqual(capabilities, { secure: scenario === "secure", clipboard: scenario === "secure" });
      await page.evaluate(async () => {
        const { ArtifactScroll } = await import("/artifact-scroll.js");
        window.scrollView = new ArtifactScroll(document.querySelector("#scroll"));
      });
      const receive = message => page.evaluate(message => window.scrollView.receive(message), message);
      const card = id => page.locator(`article[data-id="${id}"]`);
      const reply = (id, text, replay = false) => receive({ type: "part_final", part_id: id, replay, part: { type: "text", text } });
      const noControls = async id => {
        const artifact = card(id);
        assert.equal(await artifact.count(), 1);
        assert.equal(await artifact.locator(".clipboard-block").count(), 0);
        assert.equal(await artifact.getByRole("button", { name: /copy/i, includeHidden: true }).count(), 0);
        assert.equal(await artifact.getByRole("textbox", { name: "Copyable text", includeHidden: true }).count(), 0);
      };
      const snippet = async (id, index, text) => {
        const block = card(id).locator(".clipboard-block").nth(index);
        const field = block.getByRole("textbox", { name: "Copyable text", exact: true });
        assert.equal(await field.inputValue(), text, "only payload lines belong in the copyable field");
        assert.equal(await field.evaluate(field => field.tagName === "TEXTAREA" && field.readOnly), true);
        assert.equal(await block.getByRole("button", { name: "Copy text", exact: true }).count(), 1);
        assert.equal(await block.getByRole("status").count(), 1);
        return block;
      };
      const copied = async (id, index, text) => {
        await page.bringToFront();
        const block = await snippet(id, index, text);
        const button = block.getByRole("button", { name: "Copy text", exact: true });
        await button.click();
        if (scenario === "blocked") {
          await block.getByRole("status").filter({ hasText: /Copy blocked/ }).waitFor({ timeout: 3000 });
          const manual = block.getByRole("textbox", { name: "Copyable text", exact: true });
          const selection = await manual.evaluate(field => ({ start: field.selectionStart, end: field.selectionEnd, focused: document.activeElement === field }));
          assert.deepEqual(selection, { start: 0, end: text.length, focused: true });
          await manual.press("ControlOrMeta+c");
        } else {
          await block.getByRole("status").filter({ hasText: /^Copied$/ }).waitFor({ timeout: 3000 });
          assert.equal(await button.evaluate(button => document.activeElement === button), true, "copy must retain keyboard focus");
        }
        assert.equal(await page.locator("textarea[aria-hidden=true]").count(), 0, "temporary copy field must be removed");
        await reader.bringToFront();
        const actual = await reader.evaluate(() => navigator.clipboard.readText());
        assert.equal(actual === text, true, "clipboard must exactly match the snippet, excluding fences and outside prose");
        assert.equal(await page.locator("#prompt").inputValue(), "unsent prompt");
      };

      await reply("ordinary", "Prose alone has no copy button.");
      await noControls("ordinary");
      await reply("ordinary", "Outside prose\n```text\nplain code\n```\n```js\nconst value = 42;\n```\n```\nunlabeled snippet\n```\nOutside conclusion");
      assert.equal(await card("ordinary").locator(".clipboard-block").count(), 3);
      await copied("ordinary", 0, "plain code");
      await copied("ordinary", 1, "const value = 42;");
      await copied("ordinary", 2, "unlabeled snippet");
      const nested = "```js\nconst nested = true;\n```";
      await reply("nested", `Outside prose\n\`\`\`\`text\n${nested}\n\`\`\`\`\nOutside conclusion`);
      assert.equal(await card("nested").locator(".clipboard-block").count(), 1, "nested js fences remain payload within one outer snippet box");
      await copied("nested", 0, nested);

      const restricted = "```clipboard\nnot a reply snippet\n```";
      await receive({ type: "part_delta", part_id: "thinking", kind: "thinking", chunk: restricted });
      await noControls("thinking");
      await receive({ type: "part_final", part_id: "thinking", replay: true, part: { type: "opaque", from: { surface: "chat_completions" }, data: { reasoning_content: restricted } } });
      await noControls("thinking");
      await receive({ type: "part_final", part_id: "summary", part: { type: "opaque", from: { surface: "responses" }, data: { type: "reasoning", summary: [{ text: restricted }] } } });
      await noControls("summary");
      await receive({ type: "user_message", part_id: "user", part: { type: "text", text: restricted } });
      await noControls("user");
      await receive({ type: "tool_dispatched", call_id: "job", name: "run_command", input: { command: restricted } });
      await noControls("call-job");
      await receive({ type: "tool_finished", call_id: "job", result: restricted, is_error: true });
      await noControls("call-job");

      await receive({ type: "part_delta", part_id: "answer", kind: "text", chunk: "Before\n```clipboard\nprovisional payload" });
      await noControls("answer");
      await receive({ type: "part_delta", part_id: "answer", kind: "text", chunk: "\n```" });
      assert.equal(await card("answer").locator(".clipboard-block").count(), 1);
      await copied("answer", 0, "provisional payload");
      await page.bringToFront();
      await card("answer").getByRole("textbox", { name: "Copyable text" }).evaluate(field => {
        window.selectedSnippet = field;
        field.focus();
        field.setSelectionRange(2, 9, "backward");
      });
      await receive({ type: "part_delta", part_id: "answer", kind: "text", chunk: "\nFollowing prose continues streaming." });
      assert.deepEqual(await card("answer").getByRole("textbox", { name: "Copyable text" }).evaluate(field => ({
        same: field === window.selectedSnippet, focused: document.activeElement === field,
        start: field.selectionStart, end: field.selectionEnd, direction: field.selectionDirection,
      })), { same: true, focused: true, start: 2, end: 9, direction: "backward" });
      await reply("answer", "Authoritative introduction\n```clipboard\nfinal payload differs\n```\nAuthoritative conclusion");
      assert.equal(await card("answer").locator(".clipboard-block").count(), 1);
      await copied("answer", 0, "final payload differs");

      const literal = '<script>window.clipboardHTMLExecuted = true</script>\n</textarea><img src="/missing" onerror="window.clipboardHTMLExecuted = true">\n& <b>literal</b> **unchanged** café 🦊';
      const first = `\n  keep leading spaces\n\n${literal}\n\n`;
      const second = "second independent snippet\n```js\nconst nested = true;\n```\nend nested code";
      await reply("multiple", `Outside introduction\n\n\`\`\`clipboard\n${first}\n\`\`\`\nOutside middle\n\`\`\`\`clipboard\n${second}\n\`\`\`\`\`\nOutside conclusion`);
      assert.equal(await card("multiple").locator(".clipboard-block").count(), 2);
      const firstBlock = await snippet("multiple", 0, first);
      const secondBlock = await snippet("multiple", 1, second);
      assert.equal(await secondBlock.getByRole("status").textContent(), "");
      await copied("multiple", 0, first);
      const firstStatus = await firstBlock.getByRole("status").textContent();
      assert.equal(await secondBlock.getByRole("status").textContent(), "", "copying one snippet must not change its neighbor's status");
      await copied("multiple", 1, second);
      assert.equal(await firstBlock.getByRole("status").textContent(), firstStatus);
      assert.equal(await firstBlock.getByRole("textbox", { name: "Copyable text" }).inputValue(), first);
      assert.equal(await card("multiple").locator("script, img, b").count(), 0);
      assert.equal(await page.evaluate(() => window.clipboardHTMLExecuted), undefined);
      assert.match(await card("multiple").locator(".content").textContent(), /Outside middle/);

      await receive({ type: "part_delta", part_id: "wide", kind: "text", chunk: "`````clipboard\nwide payload\n````\nstill payload" });
      await noControls("wide");
      await receive({ type: "part_delta", part_id: "wide", kind: "text", chunk: "\n``````\nOutside wide fence" });
      assert.equal(await card("wide").locator(".clipboard-block").count(), 1);
      await copied("wide", 0, "wide payload\n````\nstill payload");

      const longPayload = "Unicode: café 🦊\n" + "full output\n".repeat(550) + "END_OF_FULL_SNIPPET";
      const long = `## Full Markdown\n\`\`\`clipboard\nshort complete snippet\n\`\`\`\n**Outside prose**\n\`\`\`clipboard\n${longPayload}\n\`\`\`\nOutside ending`;
      await reply("long", long);
      const details = card("long").locator("details").filter({ has: page.locator("summary", { hasText: /^Full artifact$/ }) });
      assert.equal(await details.count(), 1);
      assert.equal(await details.evaluate(element => element.open), false);
      assert.equal(await card("long").locator(".clipboard-block:visible").count(), 0, "truncated previews must not expose any copy blocks");
      assert.equal(await card("long").locator(".clipboard-block").evaluateAll(blocks => blocks.every(block => block.closest("details"))), true);
      await details.locator("summary").click();
      assert.equal(await details.locator(".clipboard-block").count(), 2);
      assert.equal(await details.locator("h2").textContent(), "Full Markdown");
      await copied("long", 0, "short complete snippet");
      await copied("long", 1, longPayload);
      await page.bringToFront();
      await details.getByRole("textbox", { name: "Copyable text" }).nth(1).evaluate(field => {
        window.selectedFullSnippet = field;
        field.focus();
        field.setSelectionRange(3, 12);
      });
      await receive({ type: "part_delta", part_id: "long", kind: "text", chunk: "\nMore streaming prose after the closed snippets." });
      assert.equal(await details.evaluate(element => element.open), true, "streaming must preserve the open full-artifact view");
      assert.deepEqual(await details.getByRole("textbox", { name: "Copyable text" }).nth(1).evaluate(field => ({
        same: field === window.selectedFullSnippet, focused: document.activeElement === field,
        start: field.selectionStart, end: field.selectionEnd,
      })), { same: true, focused: true, start: 3, end: 12 });

      assert.equal(await page.locator(".artifact__copy").count(), 0, "whole-artifact Copy controls must never exist");
      assert.equal(await page.locator("article > header").getByRole("button", { name: /copy/i, includeHidden: true }).count(), 0);
      await page.evaluate(() => window.scrollView.reset());
      assert.equal(await page.locator(".artifact, .clipboard-block").count(), 0);
      const restored = "Replayed introduction\n```clipboard\nrestored snippet\n```\nReplayed conclusion";
      await reply("answer", restored, true);
      await reply("answer", restored, true);
      assert.equal(await page.locator(".artifact").count(), 1);
      assert.equal(await page.locator(".clipboard-block").count(), 1);
      await copied("answer", 0, "restored snippet");
      if (scenario === "secure") {
        const switching = "distinct click-time snippet while switching views";
        await reply("answer", `Never copy this introduction\n\`\`\`clipboard\n${switching}\n\`\`\`\nNever copy this conclusion`);
        await page.bringToFront();
        await page.evaluate(() => {
          const block = document.querySelector(".clipboard-block");
          const button = block.querySelector('button[aria-label="Copy text"]');
          window.pendingCopyButton = button;
          window.pendingCopyStatus = block.querySelector('[role="status"]');
          button.addEventListener("click", () => {
            window.copyWasPending = button.getAttribute("aria-busy") === "true";
            window.scrollView.receive({ type: "part_final", part_id: "answer", part: { type: "text", text: "```clipboard\nreplacement after the click\n```" } });
            window.scrollView.reset();
            const prompt = document.querySelector("#prompt");
            prompt.focus();
            prompt.setSelectionRange(1, 5);
          }, { once: true });
        });
        await page.locator(".clipboard-block").getByRole("button", { name: "Copy text", exact: true }).click();
        await page.waitForFunction(() => !window.pendingCopyButton.hasAttribute("aria-busy"));
        assert.equal(await page.evaluate(() => window.copyWasPending), true);
        assert.equal(await page.evaluate(() => window.pendingCopyStatus.textContent), "Copied");
        assert.equal(await page.locator(".artifact").count(), 0);
        const prompt = await page.locator("#prompt").evaluate(field => ({ focused: document.activeElement === field, start: field.selectionStart, end: field.selectionEnd }));
        assert.deepEqual(prompt, { focused: true, start: 1, end: 5 });
        await reader.bringToFront();
        assert.equal(await reader.evaluate(() => navigator.clipboard.readText()) === switching, true, "a pending copy must retain its click-time text after reset");
      }
      assert.deepEqual(errors, []);
    } finally {
      try {
        if (reader && original !== undefined) {
          await reader.bringToFront();
          await reader.evaluate(text => navigator.clipboard.writeText(text), original);
        }
      } finally {
        await browser?.close();
        server.closeAllConnections();
        await new Promise(resolve => server.close(resolve));
      }
    }
  });
}

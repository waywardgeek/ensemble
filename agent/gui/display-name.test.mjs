import test from "node:test";
import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import net from "node:net";
import { spawn, execFileSync } from "node:child_process";
import { once } from "node:events";
import { firefox } from "playwright";

test("GUI rename preserves selection and updates both tabs and reload", { timeout: 60000 }, async () => {
  const directory = await mkdtemp(join(tmpdir(), "ensemble-display-name-"));
  const binary = join(directory, "gui");
  let browser, proc;
  try {
    execFileSync("go", ["build", "-o", binary, "./cmd"], { cwd: new URL(".", import.meta.url) });
    const probe = net.createServer();
    probe.listen(0, "127.0.0.1");
    await once(probe, "listening");
    const port = probe.address().port;
    await new Promise(resolve => probe.close(resolve));
    proc = spawn(binary, ["--port", String(port)], {
      cwd: directory,
      env: { PATH: process.env.PATH, HOME: directory, LLM_VENDOR: "openai", LLM_MODEL: "gpt-4o-mini", LLM_API_KEY: "fixture", LLM_BASE_URL: "http://127.0.0.1:1" },
      stdio: ["pipe", "pipe", "pipe"],
    });
    let log = "";
    proc.stdout.on("data", data => { log += data; });
    proc.stderr.on("data", data => { log += data; });
    const url = `http://127.0.0.1:${port}/?speech=transcript`;
    let ready = false;
    for (let attempt = 0; attempt < 200; attempt++) {
      if (proc.exitCode !== null) throw Error(log);
      try { ready = (await fetch(url)).ok; } catch {}
      if (ready) break;
      await new Promise(resolve => setTimeout(resolve, 25));
    }
    assert.ok(ready, log);
    browser = await firefox.launch({ headless: true });
    assert.ok(browser);
    const first = await browser.newPage(), second = await browser.newPage(), errors = [];
    for (const page of [first, second]) {
      page.on("pageerror", error => errors.push(error.message));
      await page.goto(url);
      await page.locator('[data-agent-key="main"]').waitFor();
    }
    const title = (page, name) => page.waitForFunction(name => document.querySelector("#conversation-title").textContent === name, name);
    const rename = async (page, oldName, newName) => {
      page.once("dialog", dialog => newName === null ? dialog.dismiss() : dialog.accept(newName));
      await page.getByRole("button", { name: `Rename ${oldName}`, exact: true }).click();
    };
    const mainName = 'Astra <img src=x onerror="window.injected=true">';
    await rename(first, "main", mainName);
    for (const page of [first, second]) {
      await title(page, mainName);
      assert.ok((await page.locator('[data-agent-key="main"]').textContent()).includes(mainName));
      assert.equal(await page.locator("#agent-tree img").count(), 0);
      assert.equal(await page.evaluate(() => window.injected), undefined);
    }
    await first.fill("#agent-name", "peer");
    await first.click("#create-agent button");
    await second.locator('[data-agent-key="peer"]').click();
    await title(second, "peer");
    await rename(first, "peer", "Research / peer");
    await title(second, "Research / peer");
    await title(first, mainName);
    assert.equal(await first.locator('[data-agent-key="main"]').getAttribute("aria-pressed"), "true");
    await rename(first, mainName, "   ");
    await first.waitForFunction(() => document.querySelector("#settings-status").textContent.includes("must not be blank"));
    await title(first, mainName);
    await rename(first, mainName, null);
    for (const page of [first, second]) {
      await page.reload();
      await title(page, mainName);
      await page.getByRole("button", { name: "Rename Research / peer", exact: true }).waitFor();
    }
    await second.locator('[data-agent-key="peer"]').click();
    await title(second, "Research / peer");
    await first.getByRole("button", { name: "Delete Research / peer", exact: true }).click();
    await title(second, mainName);
    assert.equal(await first.locator('[data-agent-key="peer"]').count(), 0);
    assert.deepEqual(errors, []);
  } finally {
    await browser?.close();
    if (proc && proc.exitCode === null) {
      const exited = once(proc, "exit");
      proc.kill("SIGTERM");
      await exited;
    }
    await rm(directory, { recursive: true, force: true });
  }
});

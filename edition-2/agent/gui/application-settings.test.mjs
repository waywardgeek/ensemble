import test from "node:test";
import assert from "node:assert/strict";
import { createServer } from "node:http";
import { readFile } from "node:fs/promises";
import { chromium } from "playwright";

// Real shipped shell/DOM with a websocket fixture; no Go server, provider or credentials.
test("application settings are explicit-save, restart-only drafts; legacy preferences still autosave", { timeout: 30000 }, async () => {
  const server = createServer(async (request, response) => {
    const name = new URL(request.url, "http://fixture").pathname.slice(1) || "index.html";
    if (!/^[a-z-]+\.(html|js|css)$/.test(name)) { response.writeHead(404).end(); return; }
    try {
      const content = await readFile(new URL(`./web/gui/${name}`, import.meta.url));
      response.setHeader("content-type", name.endsWith("js") ? "text/javascript" : name.endsWith("css") ? "text/css" : "text/html");
      response.end(content);
    } catch { response.writeHead(404).end(); }
  });
  await new Promise(resolve => server.listen(0, "127.0.0.1", resolve));
  let browser;
  try {
    browser = await chromium.launch({ headless: true,
      executablePath: process.env.CHROME_BIN || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" });
    const page = await browser.newPage();
    const errors = [], commands = [];
    page.on("pageerror", error => errors.push(error.message));
    const provider = { vendor: "openai", model: "saved-model", billing_mode: "api", base_url: "", credential_file: "/home/test/auth.json", reasoning_effort: "", reasoning_summary: "" };
    const settings = { model: "active-model", theme: "dark", font_size: 16, tts_enabled: false, tts_speed: 1, context_target: 0, log_retention: 0,
      application: { path: '/config/<img src=x onerror="alert(1)">.json', provider,
        active_provider: { ...provider, vendor: "anthropic", model: "active-model" }, restart_required: true } };
    let peer;
    const publish = (value = settings, type = "settings_changed", request_id) => peer.send(JSON.stringify({ type, settings: value, request_id }));
    await page.routeWebSocket("**/ws", socket => {
      peer = socket;
      socket.onMessage(raw => {
        const command = JSON.parse(raw);
        commands.push(command);
        if (command.type === "subscribe") {
          socket.send(JSON.stringify({ type: "agent_selected", key: "main" }));
          publish(settings, "current_settings");
        }
      });
    });
    const updates = () => commands.filter(command => command.type === "update_settings");
    const waitForUpdates = async count => {
      for (let n = 0; n < 100 && updates().length < count; n++) await new Promise(resolve => setTimeout(resolve, 10));
      assert.equal(updates().length, count);
    };
    await page.goto(`http://127.0.0.1:${server.address().port}/?speech=transcript`);
    await page.locator("#settings-tab").click();
    await page.waitForFunction(() => !document.querySelector("#application-settings").hidden);
    assert.match(await page.locator("#application-settings legend").textContent(), /Application settings — changes apply after restart/);
    assert.equal(await page.locator("#model").isVisible(), false);
    assert.equal(await page.locator("#model").isDisabled(), true);
    assert.equal(await page.locator("#provider-model").inputValue(), "saved-model");
    assert.equal(await page.locator("#provider-reasoning-effort").inputValue(), "");
    assert.equal(await page.locator("#provider-reasoning-summary").inputValue(), "");
    assert.equal(await page.locator("#application-active").textContent(), "Active now: anthropic · active-model · api");
    assert.equal(await page.locator("#application-path").textContent(), settings.application.path);
    assert.equal(await page.locator("#application-path img").count(), 0);
    assert.equal(await page.locator("#application-restart").isVisible(), true);
    assert.equal(await page.locator("#application-advanced").getAttribute("open"), null);
    assert.equal(await page.locator("#save-application-settings").isDisabled(), true);

    await page.locator("#provider-vendor").selectOption("gemini");
    await page.locator("#provider-model").fill('next-model " <script>');
    await page.locator("#provider-billing-mode").selectOption("chatgpt");
    await page.locator("#provider-reasoning-effort").selectOption("xhigh");
    await page.locator("#provider-reasoning-summary").selectOption("detailed");
    await page.locator("#application-advanced summary").click();
    await page.locator("#provider-base-url").fill("https://example.test/v1");
    await page.locator("#provider-credential-file").fill('credentials/"auth".json');
    assert.equal(updates().length, 0, "provider edits do not autosave");
    // A regular preference saves immediately and its broadcast cannot clobber the draft.
    await page.locator("#theme").selectOption("light");
    await waitForUpdates(1);
    assert.deepEqual(updates()[0].settings, { theme: "light" });
    publish({ ...settings, theme: "light" }, "settings_update");
    await page.waitForFunction(() => document.body.dataset.theme === "light");
    assert.equal(await page.locator("#provider-model").inputValue(), 'next-model " <script>');
    assert.equal(await page.locator("#provider-vendor").inputValue(), "gemini");

    await page.locator("#save-application-settings").click();
    await waitForUpdates(2);
    const submitted = { vendor: "gemini", model: 'next-model " <script>', billing_mode: "chatgpt", base_url: "https://example.test/v1", credential_file: 'credentials/"auth".json', reasoning_effort: "xhigh", reasoning_summary: "detailed" };
    const firstRequest = updates()[1].request_id;
    assert.match(firstRequest, /^[0-9a-f-]{36}$/);
    assert.deepEqual(updates()[1], { type: "update_settings", request_id: firstRequest, settings: { provider: submitted } });
    // Neither a matching snapshot without our ID nor another tab's error is an acknowledgement.
    publish({ ...settings, theme: "dark", application: { ...settings.application, provider: submitted } });
    peer.send(JSON.stringify({ type: "settings_error", request_id: "another-tab", error: "Unrelated error" }));
    await page.waitForFunction(() => document.querySelector("#settings-status").textContent === "Unrelated error");
    assert.equal(await page.locator("#save-application-settings").isDisabled(), false);
    assert.match(await page.locator("#application-status").textContent(), /^Saving/);
    peer.send(JSON.stringify({ type: "settings_error", request_id: firstRequest, error: "Invalid provider combination <b>not markup</b>" }));
    await page.waitForFunction(() => document.querySelector("#application-status").textContent.includes("not saved"));
    assert.equal(await page.locator("#provider-model").inputValue(), submitted.model);
    assert.equal(await page.locator("#save-application-settings").isDisabled(), false);
    assert.equal(await page.locator("#application-status b").count(), 0);
    publish({ ...settings, theme: "light" });
    await page.waitForFunction(() => document.body.dataset.theme === "light");
    assert.equal(await page.locator("#provider-model").inputValue(), submitted.model);

    await page.locator("#save-application-settings").click();
    await waitForUpdates(3);
    // Editing while the save is in flight must not be lost to its acknowledgement.
    await page.locator("#provider-model").fill("newer-draft");
    const normalized = { ...submitted, credential_file: '/config/credentials/"auth".json' };
    const saved = { ...settings, theme: "dark", application: { ...settings.application, provider: normalized } };
    assert.notEqual(updates()[2].request_id, firstRequest);
    publish(saved, "settings_changed", updates()[2].request_id);
    await page.waitForFunction(() => document.body.dataset.theme === "dark");
    assert.equal(await page.locator("#provider-model").inputValue(), "newer-draft");
    assert.equal(await page.locator("#save-application-settings").isDisabled(), false);
    await page.locator("#save-application-settings").click();
    await waitForUpdates(4);
    assert.equal(updates()[3].settings.provider.credential_file, submitted.credential_file);
    assert.notEqual(updates()[3].request_id, updates()[2].request_id);
    saved.application.provider = { ...normalized, model: "newer-draft" };
    publish(saved, "settings_changed", updates()[3].request_id);
    await page.waitForFunction(() => document.querySelector("#save-application-settings").disabled);
    assert.match(await page.locator("#application-status").textContent(), /saved.*restart/);
    assert.equal(await page.locator("#provider-credential-file").inputValue(), normalized.credential_file,
      "successful acknowledgement clears the draft and displays normalized server values");
    assert.equal(await page.locator("#application-active").textContent(), "Active now: anthropic · active-model · api");

    // Selection is a fresh draft boundary; no application capability means legacy behavior.
    peer.send(JSON.stringify({ type: "agent_selected", key: "child" }));
    const { application, ...legacy } = settings;
    publish(legacy, "current_settings");
    await page.waitForFunction(() => !document.querySelector("#model").disabled);
    assert.equal(await page.locator("#application-settings").isVisible(), false);
    assert.equal(await page.locator("#model").isVisible(), true);
    await page.locator("#model").fill("legacy-new-model");
    await page.locator("#model").press("Tab");
    await waitForUpdates(5);
    assert.deepEqual(updates()[4].settings, { model: "legacy-new-model" });
    peer.send(JSON.stringify({ type: "agent_selected", key: "main" }));
    publish(settings, "current_settings");
    await page.waitForFunction(() => !document.querySelector("#application-settings").hidden);
    assert.equal(await page.locator("#provider-model").inputValue(), "saved-model");
    assert.equal(await page.locator("#save-application-settings").isDisabled(), true);
    publish({ ...settings, application: { ...settings.application, restart_required: false } });
    await page.waitForFunction(() => document.querySelector("#application-restart").hidden);
    assert.equal(await page.locator("#application-restart").isVisible(), false);
    assert.deepEqual(errors, []);
  } finally {
    await browser?.close();
    await new Promise(resolve => server.close(resolve));
  }
});

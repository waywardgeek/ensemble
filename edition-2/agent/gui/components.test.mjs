import assert from "node:assert/strict";
import { chromium } from "playwright";

// The Go external consumer serves these bytes directly from gui.Assets().
// No shell, websocket, speech object or source filesystem is supplied to the page.
const browser = await chromium.launch({ headless: true,
  executablePath: process.env.CHROME_BIN || "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" });
try {
  const page = await browser.newPage();
  await page.goto(process.env.ASSET_URL + "/index.html");
  const errors = [];
  page.on("pageerror", error => errors.push(error.message));
  const actual = await page.evaluate(async () => {
    document.body.innerHTML = '<section id="standalone"></section>';
    const { ArtifactScroll } = await import("/artifact-scroll.js");
    const scroll = new ArtifactScroll(document.querySelector("section"));
    scroll.receive({ type: "part_delta", part_id: 1, kind: "text", chunk: "partial" });
    scroll.receive({ type: "part_final", part_id: 1, part: { type: "text", text: "**Independent** <script>unsafe</script>" } });
    scroll.receive({ type: "tool_dispatched", call_id: "tool", name: "read_file", input: { path: "file" } });
    scroll.receive({ type: "tool_finished", call_id: "tool", result: "COMPONENT_RESULT", is_error: false });
    const result = { cards: document.querySelectorAll(".artifact").length,
      text: document.body.textContent, bold: document.querySelector("strong")?.textContent,
      scripts: document.querySelectorAll("script").length,
      speechHidden: document.querySelector("button").hidden };
    scroll.reset();
    result.empty = !document.querySelector(".artifact");
    return result;
  });
  assert.equal(actual.cards, 2);
  assert.equal(actual.bold, "Independent");
  assert.ok(actual.text.includes("COMPONENT_RESULT"));
  assert.ok(!actual.text.includes("partial"));
  assert.equal(actual.scripts, 0);
  assert.equal(actual.speechHidden, true);
  assert.equal(actual.empty, true);
  // Exercise the shipped browser MCP module, not a copy of its tool registry.
  const tools = await page.evaluate(async () => {
    const { browserMCP } = await import("/mcp.js");
    const replies = [];
    let queueReads = 0;
    const pending = [{ text: "ON_DEMAND_SPEECH", state: "queued" }];
    const receive = browserMCP(message => replies.push(message.payload), {
      pending() { queueReads++; return pending; },
    });
    receive({ jsonrpc: "2.0", id: 1, method: "tools/list" });
    const definitions = replies.pop().result.tools;
    const readsBeforeCall = queueReads;
    const marker = document.createElement("button");
    marker.id = "snapshot-marker";
    marker.textContent = "ON_DEMAND_GUI";
    document.body.append(marker);
    receive({ jsonrpc: "2.0", id: 2, method: "tools/call",
      params: { name: "gui_snapshot", arguments: {} } });
    const snapshot = replies.pop();
    marker.textContent = "UPDATED_ON_DEMAND_GUI";
    receive({ jsonrpc: "2.0", id: 3, method: "tools/call",
      params: { name: "gui_snapshot", arguments: {} } });
    const refreshed = replies.pop();
    receive({ jsonrpc: "2.0", id: 4, method: "tools/call",
      params: { name: "tts_queue", arguments: {} } });
    return { definitions, readsBeforeCall, queueReads, snapshot, refreshed,
      queue: replies.pop(), pending };
  });
  assert.deepEqual(tools.definitions.map(tool => tool.name).sort(),
    ["gui_click", "gui_input", "gui_snapshot", "tts_queue"]);
  for (const definition of tools.definitions)
    assert.equal(definition.ephemeral, undefined, `${definition.name} must be on-demand`);
  assert.equal(tools.readsBeforeCall, 0, "discovery must not read the speech queue");
  assert.equal(tools.queueReads, 1, "one explicit call reads the speech queue once");
  assert.equal(tools.snapshot.id, 2);
  assert.equal(tools.snapshot.result.isError, false);
  assert.match(tools.snapshot.result.content[0].text, /#snapshot-marker "ON_DEMAND_GUI"/);
  assert.equal(tools.refreshed.id, 3);
  assert.match(tools.refreshed.result.content[0].text, /"UPDATED_ON_DEMAND_GUI"/);
  assert.equal(tools.queue.id, 4);
  assert.equal(tools.queue.result.isError, false);
  assert.deepEqual(JSON.parse(tools.queue.result.content[0].text), tools.pending);
  console.log("browser observation tools are discoverable and callable on demand");
  assert.deepEqual(errors, []);
  console.log("embedded component rendered without application shell");
} finally { await browser.close(); }

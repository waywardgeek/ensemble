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
  assert.deepEqual(errors, []);
  console.log("embedded component rendered without application shell");
} finally { await browser.close(); }

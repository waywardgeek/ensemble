// A stdio MCP peer, driven by the existing Ensemble CLI. The model receives
// only these five capabilities; Playwright and its DOM access stay private.
import { chromium } from "playwright";
import { parseArgs } from "node:util";
import { createInterface } from "node:readline";

const { values } = parseArgs({
  options: {
    "agent-url": { type: "string" },
    persona: { type: "string", default: "blind" },
    browser: { type: "string" },
    headed: { type: "boolean", default: false },
  },
});
if (!values["agent-url"] || values.persona !== "blind")
  throw new Error("Use --agent-url http://localhost:8088 --persona blind");

const tools = [
  {
    name: "hear",
    description:
      "Hear new speech transcript entries with their source; no screen contents",
    inputSchema: { type: "object", properties: {} },
  },
  {
    name: "type",
    description: "Replace the chat input with text",
    inputSchema: {
      type: "object",
      properties: { text: { type: "string" } },
      required: ["text"],
    },
  },
  {
    name: "submit",
    description: "Submit the typed chat input",
    inputSchema: { type: "object", properties: {} },
  },
  {
    name: "wait",
    description: "Wait until the submitted turn ends, then hear its new speech",
    inputSchema: { type: "object", properties: {} },
  },
  {
    name: "sleep",
    description: "Sleep for milliseconds before listening again",
    inputSchema: {
      type: "object",
      properties: { milliseconds: { type: "number" } },
      required: ["milliseconds"],
    },
  },
];

// This external application owns its browser lifetime. It neither owns nor
// reconstructs the target Agent; actual input and speech use the shipped page.
class VirtualUser {
  cursor = 0;
  turn = 0;
  async open() {
    this.browser = await chromium.launch({
      headless: !values.headed,
      executablePath: values.browser,
    });
    this.page = await this.browser.newPage();
    const url = new URL(values["agent-url"]);
    url.protocol =
      url.protocol === "wss:"
        ? "https:"
        : url.protocol === "ws:"
          ? "http:"
          : url.protocol;
    if (url.pathname === "/ws") url.pathname = "/";
    url.searchParams.set("speech", "transcript");
    await this.page.goto(url.href);
    await this.page.waitForFunction(
      () =>
        window.ensembleSpeech &&
        document.querySelector("#state").textContent !== "connecting",
    );
  }
  async hear() {
    const entries = await this.page.evaluate(
      (cursor) => window.ensembleSpeech.transcript.slice(cursor),
      this.cursor,
    );
    this.cursor += entries.length;
    return JSON.stringify(entries);
  }
  async call(name, args) {
    // Check capability at dispatch too. Guessing gui_snapshot or gui_click
    // cannot reach an undeclared operation or an arbitrary selector/evaluation.
    if (!tools.some((tool) => tool.name === name))
      throw new Error(`Unknown tool ${name}`);
    switch (name) {
      case "hear":
        return this.hear();
      case "type":
        if (typeof args.text !== "string") throw new Error("text is required");
        await this.page.locator("#input").fill(args.text);
        return "typed";
      case "submit":
        if (!(await this.page.locator("#input").inputValue()).trim())
          throw new Error("input is empty");
        this.turn = await this.page.evaluate(() => window.ensembleTurn + 1);
        await this.page.locator("#input").press("Enter");
        return "submitted";
      case "wait":
        await this.page.waitForFunction(
          (turn) => window.ensembleTurn >= turn,
          this.turn,
          { timeout: 0 },
        );
        return this.hear();
      case "sleep":
        if (!Number.isFinite(args.milliseconds) || args.milliseconds < 0)
          throw new Error("milliseconds must be nonnegative");
        await new Promise((resolve) => setTimeout(resolve, args.milliseconds));
        return "slept";
    }
  }
}

const user = new VirtualUser();
try {
  await user.open();
  // Session shutdown owns the whole peer, including timers unrelated to the
  // browser. Closing only the page leaves a pending sleep holding Node alive.
  const close = async () => {
    await user.browser.close();
    process.exit(0);
  };
  process.once("SIGTERM", close);
  const lines = createInterface({ input: process.stdin });
  // EOF and SIGTERM end the same lifetime, even during wait or sleep.
  lines.once("close", close);
  for await (const line of lines) {
    const request = JSON.parse(line);
    if (request.id === undefined) continue;
    const reply = { jsonrpc: "2.0", id: request.id };
    if (request.method === "initialize")
      reply.result = {
        protocolVersion: "2024-11-05",
        capabilities: { tools: {} },
        serverInfo: { name: "blind virtual user", version: "1" },
      };
    else if (request.method === "tools/list") reply.result = { tools };
    else if (request.method === "tools/call") {
      try {
        const text = await user.call(
          request.params.name,
          request.params.arguments || {},
        );
        reply.result = { content: [{ type: "text", text }] };
      } catch (error) {
        reply.result = {
          content: [{ type: "text", text: error.message }],
          isError: true,
        };
      }
    } else reply.error = { code: -32601, message: "Method not found" };
    console.log(JSON.stringify(reply));
  }
} finally {
  await user.browser?.close();
}

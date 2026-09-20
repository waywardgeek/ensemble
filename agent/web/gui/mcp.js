// mcp.js — Browser MCP server over WebSocket.
// Speaks JSON-RPC 2.0 through the agent's WebSocket hub.
// Provides gui_snapshot (ephemeral/round), gui_click, gui_input,
// and tts_queue (ephemeral/round) tools to the agent.

(function () {
  "use strict";

  // ---- JSON-RPC helpers ----

  function jsonrpcResponse(id, result) {
    return { jsonrpc: "2.0", id: id, result: result };
  }

  function jsonrpcError(id, code, message) {
    return { jsonrpc: "2.0", id: id, error: { code: code, message: message } };
  }

  // ---- Tool definitions ----

  var TOOLS = [
    {
      name: "gui_snapshot",
      description:
        "Returns a markdown summary of the current GUI state: visible panes, interactive elements with CSS selectors, artifact summaries.",
      inputSchema: {
        type: "object",
        properties: {},
        additionalProperties: false,
      },
      ephemeral: "round",
    },
    {
      name: "gui_click",
      description: "Clicks an element identified by a CSS selector.",
      inputSchema: {
        type: "object",
        properties: {
          selector: {
            type: "string",
            description: "CSS selector of the element to click",
          },
        },
        required: ["selector"],
        additionalProperties: false,
      },
    },
    {
      name: "gui_input",
      description:
        "Sets the text value of an input element identified by a CSS selector.",
      inputSchema: {
        type: "object",
        properties: {
          selector: {
            type: "string",
            description: "CSS selector of the input element",
          },
          text: {
            type: "string",
            description: "Text to set as the input value",
          },
        },
        required: ["selector", "text"],
        additionalProperties: false,
      },
    },
    {
      name: "tts_queue",
      description:
        "Returns pending TTS utterances as a JSON array with text, state (playing/pending), and timing.",
      inputSchema: {
        type: "object",
        properties: {},
        additionalProperties: false,
      },
      ephemeral: "round",
    },
    {
      name: "gui_submit",
      description:
        "Dispatch an Enter keydown event on an element. Use after gui_input to submit a prompt.",
      inputSchema: {
        type: "object",
        properties: {
          selector: {
            type: "string",
            description: "CSS selector of the element to submit",
          },
        },
        required: ["selector"],
        additionalProperties: false,
      },
    },
    {
      name: "wait_for_idle",
      description:
        "Blocks until the coding agent becomes idle (finishes its current turn). Returns immediately if already idle. Use this after sending a prompt to wait for the agent to finish working.",
      inputSchema: {
        type: "object",
        properties: {
          timeout_seconds: {
            type: "number",
            description: "Maximum seconds to wait before returning (default: 120)",
          },
        },
        additionalProperties: false,
      },
    },
    {
      name: "sleep",
      description:
        "Pauses for the given number of seconds. Use as a simple delay between actions.",
      inputSchema: {
        type: "object",
        properties: {
          seconds: {
            type: "number",
            description: "Number of seconds to sleep (default: 5)",
          },
        },
        additionalProperties: false,
      },
    },
  ];

  // ---- Tool handlers ----

  // gui_snapshot: walk visible DOM, produce markdown under ~4KB.
  function guiSnapshot() {
    var lines = [];
    lines.push("## GUI Snapshot");
    lines.push("");

    // Visible panes.
    var panes = document.querySelectorAll("[data-pane]");
    if (panes.length > 0) {
      lines.push("### Panes");
      panes.forEach(function (p) {
        var name = p.getAttribute("data-pane") || p.id || p.className;
        var vis = p.offsetParent !== null ? "visible" : "hidden";
        var rect = p.getBoundingClientRect();
        lines.push(
          "- **" +
            name +
            "** (" +
            vis +
            ", " +
            Math.round(rect.width) +
            "x" +
            Math.round(rect.height) +
            ")"
        );
      });
      lines.push("");
    }

    // Interactive elements: buttons, inputs, selects, textareas.
    var interactives = document.querySelectorAll(
      "button, input, select, textarea, [role='button'], [tabindex]"
    );
    if (interactives.length > 0) {
      lines.push("### Interactive Elements");
      var count = 0;
      interactives.forEach(function (el) {
        if (count >= 50) return; // cap for context size
        if (el.offsetParent === null) return; // skip hidden

        var tag = el.tagName.toLowerCase();
        var sel = cssSelector(el);
        var label =
          el.textContent ||
          el.getAttribute("aria-label") ||
          el.getAttribute("placeholder") ||
          "";
        label = label.trim().substring(0, 60);
        if (tag === "input" || tag === "textarea") {
          var val = el.value || "";
          lines.push(
            "- `" +
              sel +
              "` " +
              tag +
              '[type="' +
              (el.type || "text") +
              '"] value="' +
              val.substring(0, 40) +
              '" ' +
              (label ? "(" + label + ")" : "")
          );
        } else {
          lines.push(
            "- `" + sel + "` " + tag + (label ? ' "' + label + '"' : "")
          );
        }
        count++;
      });
      lines.push("");
    }

    // Artifacts (content areas).
    var artifacts = document.querySelectorAll(
      "[data-artifact], .artifact, .code-block, pre"
    );
    if (artifacts.length > 0) {
      lines.push("### Artifacts");
      var ac = 0;
      artifacts.forEach(function (el) {
        if (ac >= 10) return;
        var type =
          el.getAttribute("data-artifact") || el.className || el.tagName;
        var text = (el.textContent || "").trim();
        var preview = text.substring(0, 120);
        if (text.length > 120) preview += "...";
        lines.push("- **" + type + "**: " + preview);
        ac++;
      });
      lines.push("");
    }

    var md = lines.join("\n");
    // Truncate to ~4KB.
    if (md.length > 4000) {
      md = md.substring(0, 4000) + "\n\n(truncated)";
    }
    return md;
  }

  // Build a reasonable CSS selector for an element.
  function cssSelector(el) {
    if (el.id) return "#" + el.id;

    var tag = el.tagName.toLowerCase();
    // Try data attributes.
    if (el.getAttribute("data-action")) {
      return tag + '[data-action="' + el.getAttribute("data-action") + '"]';
    }
    if (el.getAttribute("name")) {
      return tag + '[name="' + el.getAttribute("name") + '"]';
    }
    // Fallback: tag + nth-of-type.
    var parent = el.parentElement;
    if (!parent) return tag;
    var siblings = parent.querySelectorAll(":scope > " + tag);
    if (siblings.length === 1) return tag;
    for (var i = 0; i < siblings.length; i++) {
      if (siblings[i] === el) {
        return tag + ":nth-of-type(" + (i + 1) + ")";
      }
    }
    return tag;
  }

  // gui_click: click an element by selector.
  function guiClick(args) {
    var el = document.querySelector(args.selector);
    if (!el) return "error: no element matches selector: " + args.selector;
    el.click();
    return "clicked: " + args.selector;
  }

  // gui_input: set text on an input element.
  function guiInput(args) {
    var el = document.querySelector(args.selector);
    if (!el) return "error: no element matches selector: " + args.selector;
    el.value = args.text;
    // Mark as programmatic so the pause handler ignores it.
    window._mcpProgrammaticInput = true;
    el.dispatchEvent(new Event("input", { bubbles: true }));
    el.dispatchEvent(new Event("change", { bubbles: true }));
    window._mcpProgrammaticInput = false;
    return "set value on " + args.selector + ": " + args.text;
  }

  // gui_submit: dispatch Enter keydown on an element (e.g. to submit a prompt).
  function guiSubmit(args) {
    var el = document.querySelector(args.selector);
    if (!el) return "error: no element matches selector: " + args.selector;
    el.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Enter", code: "Enter", bubbles: true })
    );
    return "submitted: " + args.selector;
  }

  // tts_queue: return pending TTS utterances.
  function ttsQueue() {
    // Check for a global TTS queue (set by tts.js).
    if (window._ttsQueue && Array.isArray(window._ttsQueue)) {
      return JSON.stringify(window._ttsQueue);
    }

    // Fallback: check speechSynthesis API.
    if (window.speechSynthesis && window.speechSynthesis.speaking) {
      return JSON.stringify([
        { text: "(speaking)", state: "playing", startedAt: Date.now() },
      ]);
    }

    return "[]";
  }

  // wait_for_idle: block until the agent state becomes idle.
  function waitForIdle(args) {
    var timeout = (args.timeout_seconds || 120) * 1000;
    return new Promise(function (resolve) {
      if (window.agentState === "idle") {
        resolve("agent is idle");
        return;
      }
      var timer = null;
      var check = setInterval(function () {
        if (window.agentState === "idle") {
          clearInterval(check);
          if (timer) clearTimeout(timer);
          resolve("agent is idle");
        }
      }, 500);
      timer = setTimeout(function () {
        clearInterval(check);
        resolve("timeout: agent still " + (window.agentState || "unknown"));
      }, timeout);
    });
  }

  // sleep: block for a given number of seconds.
  function sleepTool(args) {
    var ms = (args.seconds || 5) * 1000;
    return new Promise(function (resolve) {
      setTimeout(function () {
        resolve("slept " + args.seconds + " seconds");
      }, ms);
    });
  }

  var TOOL_HANDLERS = {
    gui_snapshot: function () {
      return guiSnapshot();
    },
    gui_click: function (args) {
      return guiClick(args);
    },
    gui_input: function (args) {
      return guiInput(args);
    },
    gui_submit: function (args) {
      return guiSubmit(args);
    },
    tts_queue: function () {
      return ttsQueue();
    },
    wait_for_idle: function (args) {
      return waitForIdle(args);
    },
    sleep: function (args) {
      return sleepTool(args);
    },
  };

  // ---- MCP Server Protocol ----

  // Handle incoming JSON-RPC message from the agent (via WebSocket hub).
  function handleMessage(msg, sendResponse) {
    if (!msg || msg.jsonrpc !== "2.0") return null;

    switch (msg.method) {
      case "initialize":
        return jsonrpcResponse(msg.id, {
          protocolVersion: "2024-11-05",
          capabilities: {
            tools: { listChanged: false },
          },
          serverInfo: { name: "gui-mcp", version: "1.0.0" },
        });

      case "tools/list":
        return jsonrpcResponse(
          msg.id,
          {
            tools: TOOLS.map(function (t) {
              var info = {
                name: t.name,
                description: t.description,
                inputSchema: t.inputSchema,
              };
              if (t.ephemeral) {
                info.ephemeral = t.ephemeral;
              }
              return info;
            }),
          }
        );

      case "tools/call":
        return handleToolCall(msg, sendResponse);

      case "notifications/initialized":
        // Acknowledgement from client, no response needed.
        return null;

      default:
        if (msg.id !== undefined && msg.id !== null) {
          return jsonrpcError(msg.id, -32601, "Method not found: " + msg.method);
        }
        return null; // Notification we don't handle.
    }
  }

  function handleToolCall(msg, sendResponse) {
    var params = msg.params || {};
    var name = params.name;
    var args = params.arguments || {};

    var handler = TOOL_HANDLERS[name];
    if (!handler) {
      return jsonrpcResponse(msg.id, {
        content: [{ type: "text", text: "unknown tool: " + name }],
        isError: true,
      });
    }

    try {
      var result = handler(args);
      // Async handler: returns a Promise. Response sent via callback.
      if (result && typeof result.then === "function") {
        result.then(function (text) {
          sendResponse(jsonrpcResponse(msg.id, {
            content: [{ type: "text", text: text }],
          }));
        }).catch(function (e) {
          sendResponse(jsonrpcResponse(msg.id, {
            content: [{ type: "text", text: "error: " + e.message }],
            isError: true,
          }));
        });
        return null; // No synchronous response.
      }
      return jsonrpcResponse(msg.id, {
        content: [{ type: "text", text: result }],
      });
    } catch (e) {
      return jsonrpcResponse(msg.id, {
        content: [{ type: "text", text: "error: " + e.message }],
        isError: true,
      });
    }
  }

  // ---- WebSocket integration ----

  // Hook into the existing WebSocket connection.
  // The hub sends jsonrpc messages with type: "jsonrpc".
  // We intercept them and respond.

  function attachToWebSocket(ws) {
    var origOnMessage = ws.onmessage;

    ws.onmessage = function (event) {
      try {
        var envelope = JSON.parse(event.data);
        if (envelope.type === "jsonrpc" && envelope.payload) {
          var rpcMsg =
            typeof envelope.payload === "string"
              ? JSON.parse(envelope.payload)
              : envelope.payload;
          // Callback for async tool handlers to send responses later.
          var sendAsyncResponse = function (response) {
            var reply = { type: "jsonrpc", payload: response };
            if (envelope.source) {
              reply.source = envelope.source;
            }
            ws.send(JSON.stringify(reply));
          };
          var response = handleMessage(rpcMsg, sendAsyncResponse);
          if (response) {
            sendAsyncResponse(response);
          }
          return; // Consumed by MCP server.
        }
      } catch (e) {
        // Not a JSON message or not for us — fall through.
      }

      // Pass to original handler.
      if (origOnMessage) {
        origOnMessage.call(ws, event);
      }
    };
  }

  // Auto-attach when WebSocket is created.
  // Override WebSocket constructor to hook in.
  var OrigWebSocket = window.WebSocket;
  window.WebSocket = function (url, protocols) {
    var ws = new OrigWebSocket(url, protocols);
    // Attach MCP server after connection opens.
    ws.addEventListener("open", function () {
      attachToWebSocket(ws);
    });
    return ws;
  };
  window.WebSocket.prototype = OrigWebSocket.prototype;
  window.WebSocket.CONNECTING = OrigWebSocket.CONNECTING;
  window.WebSocket.OPEN = OrigWebSocket.OPEN;
  window.WebSocket.CLOSING = OrigWebSocket.CLOSING;
  window.WebSocket.CLOSED = OrigWebSocket.CLOSED;

  // Export for direct use.
  window.MCPServer = {
    handleMessage: handleMessage,
    tools: TOOLS,
    handlers: TOOL_HANDLERS,
  };
})();

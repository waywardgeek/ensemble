// gui.js — Three-pane workbench: routing, drag bars, settings, theming.

(function() {
  'use strict';

  // ── Scroll views ──
  const chatScroll = new ArtifactScroll(document.getElementById('chat-scroll'));
  const actionsScroll = new ArtifactScroll(document.getElementById('actions-scroll'));

  // ── DOM references ──
  const input = document.getElementById('prompt-input');
  const stateEl = document.getElementById('agent-state');
  const sidebar = document.getElementById('sidebar');
  const rightPane = document.getElementById('right-pane');
  const settingsOverlay = document.getElementById('settings-overlay');
  const agentDot = document.querySelector('.agent-dot');
  const agentLabel = document.querySelector('.agent-label');

  let ws;
  let agentState = 'idle';
  window.agentState = 'idle';

  // ── WebSocket ──
  function connect() {
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    ws = new WebSocket(`${proto}//${location.host}/ws`);

    ws.onopen = () => {
      ws.send(JSON.stringify({type: 'subscribe'}));
    };

    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data);
      routeMessage(msg);
    };

    ws.onclose = () => {
      setTimeout(connect, 2000);
    };
  }

  // ── Message routing ──
  // Thinking and chat → center pane. Tools → right pane.
  function routeMessage(msg) {
    switch (msg.type) {
      case 'event_range':
        if (msg.last > 0) {
          ws.send(JSON.stringify({type: 'fetch', from: msg.first, to: msg.last}));
        }
        break;

      case 'part_delta':
        if (msg.kind === 'tool_call') {
          actionsScroll.handleMessage(msg);
        } else {
          chatScroll.handleMessage(msg);
        }
        break;

      case 'part_final':
        if (msg.tool) {
          actionsScroll.handleMessage(msg);
        } else {
          chatScroll.handleMessage(msg);
        }
        break;

      case 'part_partial':
        // Reconnection partial — route based on kind if available.
        chatScroll.handleMessage(msg);
        break;

      case 'tool_dispatched':
      case 'tool_finished':
        actionsScroll.handleMessage(msg);
        break;

      case 'message':
        chatScroll.handleMessage(msg);
        break;

      case 'state_changed':
        agentState = msg.to;
        window.agentState = agentState;
        updateStateUI();
        break;

      case 'turn_ended':
        agentState = 'idle';
        window.agentState = 'idle';
        updateStateUI();
        chatScroll.handleMessage(msg);
        break;

      case 'current_settings':
        if (msg.models) populateModelDropdown(msg.models);
        if (msg.settings) applySettings(msg.settings);
        break;

      case 'settings_changed':
        if (msg.settings) applySettings(msg.settings);
        break;

      case 'usage':
        renderUsage(msg);
        break;

      case 'error':
        chatScroll.handleMessage(msg);
        break;
    }
  }

  function updateStateUI() {
    stateEl.textContent = agentState;
    input.placeholder = agentState === 'idle'
      ? 'Type a message...'
      : 'Type a hint...';

    // Update agent tree dot.
    agentDot.className = 'agent-dot';
    if (agentState === 'idle') {
      agentDot.classList.add('idle');
    } else if (agentState === 'thinking') {
      agentDot.classList.add('thinking');
    } else {
      agentDot.classList.add('tools');
    }
  }

  // ── Pause gate ──
  // Two independent causes block tool dispatch: the agent is speaking, and the user
  // is typing. The gate is the OR of the two, and only a CHANGE in that OR goes on
  // the wire. Clearing one cause therefore cannot resume the agent while the other
  // still holds, whichever one happened to fire.
  let userTyping = false;
  let gatePaused = false;

  function updateGate() {
    const blocked = userTyping || TTS.speaking;
    if (blocked === gatePaused) return;
    gatePaused = blocked;
    if (ws && ws.readyState === WebSocket.OPEN) {
      // Both halves of the gate produce the same pause, so the cause travels
      // with it. Without that, a speech log shows that the agent stopped but
      // never why, and the two causes need very different fixes.
      ws.send(JSON.stringify({
        type: blocked ? 'pause' : 'unpause',
        typing: userTyping,
        speaking: TTS.speaking,
      }));
    }
  }

  // The speaking half of the gate. This subscription is the whole reason TTS state
  // is exported: without it the agent runs tools while it is still talking.
  TTS.onStateChange = updateGate;

  // The client half of the speech log. The browser cannot write files, so every
  // entry that enters the speech channel is forwarded to the server, which appends
  // it to the log. The gate's pause and resume lines already travel this wire; this
  // adds the utterances themselves, so the log shows what was said and when it was
  // held back.
  TTS.onRecord = (entry) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'tts',
        tts: {kind: 'utterance', seq: entry.seq, text: entry.text, raw: entry.raw || ''},
      }));
    }
  };


  // ── Input handling ──

  // Grow the composer to fit its content, then scroll. The cap comes from the
  // stylesheet rather than a constant here, so it stays in rem and tracks the
  // font-size slider. Setting height to auto first lets scrollHeight shrink
  // again when text is deleted; without that the box would only ever grow.
  function autoGrow() {
    input.style.height = 'auto';
    const maxH = parseFloat(getComputedStyle(input).maxHeight) || Infinity;
    const full = input.scrollHeight;
    input.style.height = Math.min(full, maxH) + 'px';
    input.style.overflowY = full > maxH ? 'auto' : 'hidden';
  }

  input.addEventListener('input', autoGrow);

  input.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      TTS.cancel();          // Escape always silences speech, whatever is typed.
      input.value = '';
      autoGrow();            // .value = '' fires no input event, so shrink by hand
      userTyping = false;    // assigning .value fires no input event
      updateGate();
      return;
    }

    if (e.key !== 'Enter' || e.shiftKey) return;
    e.preventDefault();
    const text = input.value.trim();
    if (!text) return;
    input.value = '';
    autoGrow();

    chatScroll.handleMessage({type: 'message', actor: 'user', text});

    if (ws && ws.readyState === WebSocket.OPEN) {
      if (agentState === 'idle') {
        ws.send(JSON.stringify({type: 'prompt', text}));
      } else {
        ws.send(JSON.stringify({type: 'hint', text}));
      }
    }
    userTyping = false;      // assigning .value fires no input event
    updateGate();
  });

  // Any character counts, a space included: a half-typed hint is still in progress.
  // (Programmatic input from MCP tools is excluded, or the driver gates itself.)
  input.addEventListener('input', () => {
    if (window._mcpProgrammaticInput) return;
    userTyping = input.value.length > 0;
    updateGate();
  });

  // ── Sidebar toggle ──
  // Derive the initial attribute from the actual class rather than trusting the
  // markup default: anything that collapses the sidebar at startup would leave a
  // hardcoded aria-expanded lying, and a lying label is worse than none.
  document
    .getElementById('hamburger')
    .setAttribute('aria-expanded', String(!sidebar.classList.contains('collapsed')));
  document.getElementById('hamburger').addEventListener('click', (e) => {
    const collapsed = sidebar.classList.toggle('collapsed');
    // A CSS class is invisible to a screen reader and to a DOM snapshot alike.
    // Mirror the state onto the button so the control reports what it just did.
    e.currentTarget.setAttribute('aria-expanded', String(!collapsed));
  });

  // Sidebar tabs.
  document.querySelectorAll('#sidebar-tabs .tab').forEach(tab => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('#sidebar-tabs .tab').forEach(t => {
        t.classList.remove('active');
        // The 'active' class styles the tab but says nothing to a screen reader
        // or a DOM snapshot. Mirror it onto aria-selected so which panel is
        // showing is a readable fact rather than something to be inferred.
        t.setAttribute('aria-selected', 'false');
      });
      document.querySelectorAll('.tab-panel').forEach(p => p.classList.remove('active'));
      tab.classList.add('active');
      tab.setAttribute('aria-selected', 'true');
      const panel = document.querySelector(`.tab-panel[data-panel="${tab.dataset.tab}"]`);
      if (panel) panel.classList.add('active');
    });
  });

  // ── Drag bars ──
  function initDragBar(barId, leftEl, rightEl, minLeft, minRight) {
    const bar = document.getElementById(barId);
    let startX, startLeftW, startRightW;

    bar.addEventListener('mousedown', (e) => {
      e.preventDefault();
      bar.classList.add('dragging');
      startX = e.clientX;
      startLeftW = leftEl.offsetWidth;
      if (rightEl) startRightW = rightEl.offsetWidth;

      function onMove(e) {
        const dx = e.clientX - startX;
        const newLeft = Math.max(minLeft, startLeftW + dx);
        leftEl.style.width = newLeft + 'px';
        if (rightEl) {
          const newRight = Math.max(minRight, startRightW - dx);
          rightEl.style.width = newRight + 'px';
        }
      }

      function onUp() {
        bar.classList.remove('dragging');
        document.removeEventListener('mousemove', onMove);
        document.removeEventListener('mouseup', onUp);
      }

      document.addEventListener('mousemove', onMove);
      document.addEventListener('mouseup', onUp);
    });
  }

  initDragBar('drag-left', sidebar, null, 0, 0);
  initDragBar('drag-right', document.getElementById('center'), rightPane, 300, 0);

  // ── Settings ──
  document.getElementById('settings-btn').addEventListener('click', () => {
    settingsOverlay.classList.remove('hidden');
  });

  document.getElementById('settings-close').addEventListener('click', () => {
    settingsOverlay.classList.add('hidden');
  });

  settingsOverlay.addEventListener('click', (e) => {
    if (e.target === settingsOverlay) {
      settingsOverlay.classList.add('hidden');
    }
  });

  // Settings tabs.
  document.querySelectorAll('.settings-tab').forEach(tab => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('.settings-tab').forEach(t => t.classList.remove('active'));
      document.querySelectorAll('.settings-tab-panel').forEach(p => p.classList.remove('active'));
      tab.classList.add('active');
      const panel = document.querySelector(`.settings-tab-panel[data-panel="${tab.dataset.tab}"]`);
      if (panel) panel.classList.add('active');
    });
  });

  // Range slider live display — update the paired <span> as the slider moves.
  function updateRangeDisplay(inputEl) {
    const id = inputEl.id.replace('set-', 'rv-');
    const span = document.getElementById(id);
    if (span) span.textContent = inputEl.value;
  }

  document.querySelectorAll('.settings-tab-panel input[type="range"]').forEach(el => {
    el.addEventListener('input', () => {
      updateRangeDisplay(el);
      // Apply font size immediately while dragging.
      if (el.id === 'set-font-size') applyFontSize(parseInt(el.value, 10));
    });
  });

  // Populate model dropdown from server's model catalog.
  let modelCatalog = [];  // Saved for context-target slider range updates.

  function populateModelDropdown(models) {
    modelCatalog = models;
    const sel = document.getElementById('set-model');
    sel.innerHTML = '';
    let currentVendor = '';
    let group = null;
    for (const m of models) {
      if (m.vendor !== currentVendor) {
        currentVendor = m.vendor;
        group = document.createElement('optgroup');
        group.label = currentVendor;
        sel.appendChild(group);
      }
      const opt = document.createElement('option');
      opt.value = m.id;
      opt.textContent = m.display_name;
      group.appendChild(opt);
    }
  }

  // Update the context-target slider range when the model changes.
  function updateContextSliderForModel(modelId) {
    const m = modelCatalog.find(e => e.id === modelId);
    const slider = document.getElementById('set-context-target');
    if (m && m.context_window > 0) {
      slider.max = m.context_window;
      slider.step = Math.max(1000, Math.round(m.context_window / 100));
    }
  }

  document.getElementById('set-model').addEventListener('change', (e) => {
    updateContextSliderForModel(e.target.value);
  });

  // Send a settings patch to the server.
  //
  // Every control routes through here rather than calling ws.send itself, so
  // there is one place that knows the message shape. Its absence was a real
  // bug: the handlers below called sendPatch before it existed, so each one
  // threw a ReferenceError and no setting was ever persisted. The dialog still
  // looked correct, because the controls that change something visible — theme
  // and font size — also apply locally on the way past. The visible half
  // worked and the durable half never ran.
  function sendPatch(patch) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({type: 'update_settings', settings: patch}));
  }

  // Settings change handlers — each sends a flat patch.
  //
  // Flat, not nested by tab, because this is the published WebSocket contract:
  // a client sends {"theme":"dark"} and gets the whole settings object back
  // with flat keys. The tabs are presentation only. The on-disk file groups
  // these into sections, but that translation belongs to the server, which
  // owns the file — not to the client, which owns neither.
  //
  // Map: element ID → {key, type}
  const settingFields = {
    'set-model':           {key: 'model', type: 'string'},
    'set-max-tool-rounds': {key: 'max_tool_rounds', type: 'int'},
    'set-thinking-level':  {key: 'thinking_budget', type: 'thinking'},
    'set-context-target':  {key: 'context_target', type: 'int'},
    'set-log-retention':   {key: 'log_retention', type: 'int'},
    'set-tts-enabled':     {key: 'tts_enabled', type: 'bool'},
    'set-tts-speed':       {key: 'tts_speed', type: 'float'},
    'set-font-size':       {key: 'font_size', type: 'int'},
    'set-theme':           {key: 'theme', type: 'string'},
  };

  // Map thinking dropdown to budget tokens.
  const thinkingLevels = {low: 4096, medium: 16384, high: 32000, max: 200000};

  function thinkingLevelFromBudget(budget) {
    if (budget <= 4096) return 'low';
    if (budget <= 16384) return 'medium';
    if (budget <= 32000) return 'high';
    return 'max';
  }

  Object.entries(settingFields).forEach(([id, spec]) => {
    const el = document.getElementById(id);
    if (!el) return;
    el.addEventListener('change', () => {
      let val;
      if (spec.type === 'bool') val = el.checked;
      else if (spec.type === 'int') val = parseInt(el.value, 10) || 0;
      else if (spec.type === 'float') val = parseFloat(el.value) || 0;
      else if (spec.type === 'thinking') val = thinkingLevels[el.value] || 32000;
      else val = el.value;

      const patch = {};
      patch[spec.key] = val;
      sendPatch(patch);
    });
  });

  // Band config change handlers.
  document.querySelectorAll('.band-row[data-band]').forEach(row => {
    const band = row.dataset.band;
    const toggle = row.querySelector('.band-enabled');
    const budget = row.querySelector('.band-budget-input');

    function sendBandPatch() {
      const mem = {};
      mem[band] = {
        disabled: !toggle.checked,
        budget: parseInt(budget.value, 10) || 0,
      };
      sendPatch({memory: mem});
    }

    toggle.addEventListener('change', sendBandPatch);
    budget.addEventListener('change', sendBandPatch);
  });

  // ── Session usage meter ──

  // Compact token counts. A working session runs to millions of tokens, and
  // the raw digits are both hard to read and wide enough to reflow the bar
  // every time they gain a place.
  function fmtTokens(n) {
    if (n >= 1e6) return (n / 1e6).toFixed(2) + 'M';
    if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K';
    return String(n);
  }

  // Small amounts need more decimals, or the first several turns of a session
  // all read a flat "$0.00" and the meter looks broken rather than cheap.
  function fmtCost(d) {
    if (d >= 1) return '$' + d.toFixed(2);
    if (d > 0) return '$' + d.toFixed(4);
    return '$0.00';
  }

  function renderUsage(u) {
    const input = u.input || 0;
    const cacheWrite = u.cache_write || 0;
    const cacheRead = u.cache_read || 0;

    // The three input categories are disjoint, so this is the true total the
    // session was charged for on the way in. Cache read is a subset of it,
    // which is what makes the pair readable: "187K in, 184K of it cached".
    const totalIn = input + cacheWrite + cacheRead;

    const cost = document.getElementById('u-cost');
    if (u.priced) {
      cost.textContent = fmtCost(u.cost_usd || 0);
      cost.classList.remove('unpriced');
      cost.title = 'Session cost so far';
    } else {
      // An unpriced model and a model that has spent nothing both cost zero
      // dollars. A dash keeps the first from looking like the second.
      cost.textContent = '$—';
      cost.classList.add('unpriced');
      cost.title = 'This model has no price sheet, so cost is unknown';
    }

    document.getElementById('u-in').textContent = 'in ' + fmtTokens(totalIn);
    document.getElementById('u-cr').textContent = 'cache ' + fmtTokens(cacheRead);
    document.getElementById('u-hit').textContent = ((u.hit_rate || 0) * 100).toFixed(1) + '%';

    // Exact figures on hover, since the compact forms round.
    document.getElementById('usage-meter').title =
      'input ' + input.toLocaleString() +
      ' · cache write ' + cacheWrite.toLocaleString() +
      ' · cache read ' + cacheRead.toLocaleString() +
      ' · output ' + (u.output || 0).toLocaleString();
  }

  function applySettings(s) {
    // Flat keys: this is the wire contract. The settings dialog groups these
    // into tabs and the file groups them into sections, but neither grouping
    // appears here — the server sends one flat object with every field
    // present, so "off" is distinguishable from "not mentioned".

    // AI
    if (s.model) {
      document.getElementById('set-model').value = s.model;
      updateContextSliderForModel(s.model);
    }
    if (s.max_tool_rounds !== undefined) document.getElementById('set-max-tool-rounds').value = s.max_tool_rounds;
    if (s.thinking_budget !== undefined) {
      document.getElementById('set-thinking-level').value = thinkingLevelFromBudget(s.thinking_budget);
    }
    if (s.context_target !== undefined) {
      const slider = document.getElementById('set-context-target');
      const val = s.context_target || Math.round(parseInt(slider.max, 10) / 2);
      slider.value = val;
      updateRangeDisplay(slider);
    }
    if (s.log_retention !== undefined) document.getElementById('set-log-retention').value = s.log_retention;

    // Appearance
    if (s.theme) document.getElementById('set-theme').value = s.theme;
    if (s.font_size !== undefined) {
      const size = s.font_size || 16;  // 0 means unset; default to 16px
      document.getElementById('set-font-size').value = size;
      updateRangeDisplay(document.getElementById('set-font-size'));
    }
    applyTheme(s.theme || 'dark');
    applyFontSize(s.font_size);

    // Accessibility
    document.getElementById('set-tts-enabled').checked = !!s.tts_enabled;
    if (s.tts_speed !== undefined) {
      document.getElementById('set-tts-speed').value = s.tts_speed;
      updateRangeDisplay(document.getElementById('set-tts-speed'));
    }
    if (typeof TTS !== 'undefined') {
      TTS.enabled = !!s.tts_enabled;
      if (s.tts_speed !== undefined) TTS.rate = s.tts_speed;
    }

    // Memory band config. This one really is nested on the wire, because the
    // field itself is a struct of six bands rather than six sibling keys.
    if (s.memory) {
      const bands = {soul: s.memory.soul, memory: s.memory.memory,
        '64x': s.memory['64x'], '8x': s.memory['8x'],
        session: s.memory.session, conversation: s.memory.conversation};
      Object.entries(bands).forEach(([name, band]) => {
        if (!band) return;
        const row = document.querySelector(`.band-row[data-band="${name}"]`);
        if (!row) return;
        row.querySelector('.band-enabled').checked = !band.disabled;
        row.querySelector('.band-budget-input').value = band.budget || 0;
      });
    }
  }

  function applyFontSize(size) {
    const px = size || 16;
    document.documentElement.style.fontSize = px + 'px';
  }

  function applyTheme(theme) {
    const html = document.documentElement;
    html.classList.remove('light');
    if (theme === 'light') {
      html.classList.add('light');
    } else if (theme === 'system') {
      if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
        html.classList.add('light');
      }
    }
  }

  // Listen for system theme changes.
  if (window.matchMedia) {
    window.matchMedia('(prefers-color-scheme: light)').addEventListener('change', () => {
      const theme = document.getElementById('set-theme').value;
      if (theme === 'system') applyTheme('system');
    });
  }

  // ── Start ──
  connect();
})();

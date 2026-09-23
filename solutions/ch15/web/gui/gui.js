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
        if (msg.settings) applySettings(msg.settings);
        break;

      case 'settings_changed':
        if (msg.settings) applySettings(msg.settings);
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
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      TTS.cancel();          // Escape always silences speech, whatever is typed.
      input.value = '';
      userTyping = false;    // assigning .value fires no input event
      updateGate();
      return;
    }

    if (e.key !== 'Enter') return;
    const text = input.value.trim();
    if (!text) return;
    input.value = '';

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

  // Settings change handlers — send update_settings on any change.
  const settingFields = {
    'set-model': {key: 'model', type: 'string'},
    'set-temperature': {key: 'temperature', type: 'float'},
    'set-max-tokens': {key: 'max_tokens', type: 'int'},
    'set-thinking-budget': {key: 'thinking_budget', type: 'int'},
    'set-max-tool-rounds': {key: 'max_tool_rounds', type: 'int'},
    'set-context-target': {key: 'context_target', type: 'int'},
    'set-log-retention': {key: 'log_retention', type: 'int'},
    'set-tts-enabled': {key: 'tts_enabled', type: 'bool'},
    'set-tts-speed': {key: 'tts_speed', type: 'float'},
    'set-font-size': {key: 'font_size', type: 'int'},
    'set-theme': {key: 'theme', type: 'string'},
  };

  Object.entries(settingFields).forEach(([id, spec]) => {
    const el = document.getElementById(id);
    if (!el) return;
    el.addEventListener('change', () => {
      let val;
      if (spec.type === 'bool') val = el.checked;
      else if (spec.type === 'int') val = parseInt(el.value, 10) || 0;
      else if (spec.type === 'float') val = parseFloat(el.value) || 0;
      else val = el.value;

      const patch = {};
      patch[spec.key] = val;
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({type: 'update_settings', settings: patch}));
      }
    });
  });

  function applySettings(s) {
    // The server is authoritative and, since Settings dropped omitempty, always
    // sends every field. So test PRESENCE, not truthiness: a truthy test silently
    // discards a legitimate zero, which is how a clamped value could be corrected
    // on the server and still display the old number here. Strings keep the truthy
    // test, because an empty one means "unset" and must not blank a live control.
    if (s.model) document.getElementById('set-model').value = s.model;
    if (s.temperature !== undefined) document.getElementById('set-temperature').value = s.temperature;
    if (s.max_tokens !== undefined) document.getElementById('set-max-tokens').value = s.max_tokens;
    if (s.thinking_budget !== undefined) document.getElementById('set-thinking-budget').value = s.thinking_budget;
    if (s.max_tool_rounds !== undefined) document.getElementById('set-max-tool-rounds').value = s.max_tool_rounds;
    if (s.context_target !== undefined) document.getElementById('set-context-target').value = s.context_target || 400000;
    if (s.log_retention !== undefined) document.getElementById('set-log-retention').value = s.log_retention;
    document.getElementById('set-tts-enabled').checked = !!s.tts_enabled;
    if (s.tts_speed !== undefined) document.getElementById('set-tts-speed').value = s.tts_speed;
    if (s.font_size !== undefined) document.getElementById('set-font-size').value = s.font_size;
    if (s.theme) document.getElementById('set-theme').value = s.theme;

    // Apply theme.
    applyTheme(s.theme || 'dark');

    // Apply font size.
    if (s.font_size) {
      document.documentElement.style.fontSize = s.font_size + 'px';
    }

    // Apply TTS. Speed is a float: Bill runs 2.5 or 3.8, so this must never be
    // rounded or coerced to an integer anywhere along the path.
    if (typeof TTS !== 'undefined') {
      TTS.enabled = !!s.tts_enabled;
      if (s.tts_speed !== undefined) TTS.rate = s.tts_speed;
    }
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

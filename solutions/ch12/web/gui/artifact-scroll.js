// ArtifactScroll — reusable scroll view for agent observations.

class ArtifactScroll {
  constructor(container, options = {}) {
    this.container = container;
    this.artifacts = new Map(); // part_id -> element
    this.toolCards = new Map(); // call_id -> element
    this.accumulated = new Map(); // part_id -> accumulated text
    this.options = options;
  }

  handleMessage(msg) {
    switch (msg.type) {
      case 'message':
        this._handleUserMessage(msg);
        break;
      case 'part_delta':
        this._handleDelta(msg);
        break;
      case 'part_final':
        this._handleFinal(msg);
        break;
      case 'part_partial':
        this._handlePartial(msg);
        break;
      case 'tool_dispatched':
        this._handleToolDispatched(msg);
        break;
      case 'tool_finished':
        this._handleToolFinished(msg);
        break;
      case 'error':
        this._handleError(msg);
        break;
      case 'state_changed':
        // Handled by the main page
        break;
      case 'turn_ended':
        this._scrollToBottom();
        break;
    }
  }

  _handleUserMessage(msg) {
    const el = document.createElement('div');
    el.className = 'artifact user-message';
    el.textContent = msg.text || '';
    this.container.appendChild(el);
    this._scrollToBottom();
  }

  _handleDelta(msg) {
    const id = msg.part_id;
    let el = this.artifacts.get(id);
    if (!el) {
      el = document.createElement('div');
      el.className = 'artifact';
      if (msg.kind === 'thinking') el.classList.add('thinking');
      this.container.appendChild(el);
      this.artifacts.set(id, el);
      this.accumulated.set(id, '');
    }

    const text = (this.accumulated.get(id) || '') + msg.chunk;
    this.accumulated.set(id, text);

    if (msg.kind === 'tool_call') {
      el.innerHTML = Renderers.json(text);
    } else {
      el.innerHTML = Renderers.markdown(text);
    }
    this._scrollToBottom();

    // TTS for text and thinking chunks
    if (msg.kind === 'text' || msg.kind === 'thinking') {
      TTS.queueChunk(msg.chunk);
    }
  }

  _handleFinal(msg) {
    const id = msg.part_id;
    let el = this.artifacts.get(id);
    if (!el) {
      el = document.createElement('div');
      el.className = 'artifact';
      this.container.appendChild(el);
      this.artifacts.set(id, el);
    }

    if (msg.text !== undefined) {
      el.innerHTML = Renderers.markdown(msg.text);
    } else if (msg.tool) {
      el.innerHTML = Renderers.json(`${msg.tool}: ${msg.args || ''}`);
    }
  }

  _handleToolDispatched(msg) {
    const el = document.createElement('div');
    el.className = 'artifact tool-card';

    let inputStr = '';
    if (msg.input) {
      try {
        inputStr = JSON.stringify(msg.input, null, 2);
      } catch(e) {
        inputStr = String(msg.input);
      }
    }

    el.innerHTML = `<span class="tool-name">${msg.name}</span> <code>${msg.call_id}</code>
<pre>${Renderers.truncate(inputStr, 500)}</pre>`;
    this.container.appendChild(el);
    this.toolCards.set(msg.call_id, el);
    this._scrollToBottom();

    // TTS summary
    TTS.speakToolDispatch(msg.name, msg.input);
  }

  _handleToolFinished(msg) {
    const el = this.toolCards.get(msg.call_id);
    if (!el) return;

    const resultDiv = document.createElement('div');
    resultDiv.className = msg.is_error ? 'tool-result tool-error' : 'tool-result';
    resultDiv.textContent = Renderers.truncate(msg.result || '', 1000);
    el.appendChild(resultDiv);
    this._scrollToBottom();
  }

  _scrollToBottom() {
    this.container.scrollTop = this.container.scrollHeight;
  }

  _handlePartial(msg) {
    const id = msg.part_id;
    let el = this.artifacts.get(id);
    if (!el) {
      el = document.createElement('div');
      el.className = 'artifact';
      this.container.appendChild(el);
      this.artifacts.set(id, el);
    }
    el.innerHTML = Renderers.markdown(msg.content || '');
    this.accumulated.set(id, msg.content || '');
    this._scrollToBottom();
  }

  _handleError(msg) {
    const el = document.createElement('div');
    el.className = 'artifact tool-card tool-error';
    el.textContent = msg.message || 'Unknown error';
    this.container.appendChild(el);
    this._scrollToBottom();
  }
}

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

    // End of a part is a phrase boundary. Streaming routinely stops on a colon or a
    // bare clause, so text still buffered here would otherwise sit waiting for a
    // period that never arrives, and never be heard at all.
    TTS.flush();
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

    // Build these nodes rather than interpolating into innerHTML. Tool names and
    // arguments carry attacker-influenceable text (a filename, a fetched URL, the
    // contents of a file the agent just read), so string-building HTML here is an
    // injection hole: a tool argument containing markup would execute in the GUI.
    // textContent cannot be escaped out of.
    const nameSpan = document.createElement('span');
    nameSpan.className = 'tool-name';
    nameSpan.textContent = msg.name;

    const idCode = document.createElement('code');
    idCode.textContent = msg.call_id;

    const pre = document.createElement('pre');
    pre.textContent = Renderers.truncate(inputStr, 500);
    // The cap keeps the card readable, but the hidden remainder must stay reachable:
    // hovering reads the full value, so nothing is permanently invisible.
    if (inputStr.length > 500) pre.title = inputStr;

    el.replaceChildren(nameSpan, document.createTextNode(' '), idCode, pre);
    this.container.appendChild(el);
    this.toolCards.set(msg.call_id, el);
    this._scrollToBottom();

    // TTS summary
    // Flush first: narration buffered so far belongs ahead of the tool
    // announcement, or the listener hears the two interleaved out of order.
    TTS.flush();
    TTS.speakToolDispatch(msg.name, msg.input);
  }

  _handleToolFinished(msg) {
    const el = this.toolCards.get(msg.call_id);
    if (!el) return;

    const resultDiv = document.createElement('div');
    resultDiv.className = msg.is_error ? 'tool-result tool-error' : 'tool-result';
    const fullResult = msg.result || '';
    resultDiv.textContent = Renderers.truncate(fullResult, 1000);
    // Same rule as the input above: cap the display, keep the remainder reachable.
    if (fullResult.length > 1000) resultDiv.title = fullResult;
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
    // Errors are spoken as well as shown. Displaying an error to a reader who works
    // by speech and never voicing it means the one message class that most needs to
    // interrupt is the only class that is silent.
    TTS.speakError(msg.message || 'Unknown error');
  }
}

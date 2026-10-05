// ArtifactScroll — reusable scroll view for agent observations.

class ArtifactScroll {
  constructor(container, options = {}) {
    this.container = container;
    this.artifacts = new Map(); // part_id -> element
    this.toolCards = new Map(); // call_id -> element
    this.accumulated = new Map(); // part_id -> accumulated text
    this.options = options;
    // View cap, set from settings. Zero means no cap.
    this.maxEvents = Number.isFinite(options.maxEvents) && options.maxEvents > 0 ? options.maxEvents : 0;
  }

  // clear removes every rendered artifact and forgets the correlation state.
  //
  // The three maps have to go with the DOM, not just the DOM. They key
  // streaming deltas to the elements those deltas belong to, so a map entry
  // that outlived its element would route the next turn's text into a node
  // that is no longer on the page: text vanishing with no error anywhere.
  clear() {
    this.container.replaceChildren();
    this.artifacts.clear();
    this.toolCards.clear();
    this.accumulated.clear();
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
        // A turn that ends in failure has to say so. The reason arrives as a
        // field on turn_ended rather than as a separate error message, so
        // ignoring it here leaves a listener hearing the agent fall silent
        // with no idea whether it finished, stalled, or died.
        if (msg.error) {
          this._handleError({ message: msg.error });
        }
        this._scrollToBottom();
        break;
    }
    this._enforceCap();
  }

  // setMaxEvents caps how many rendered items stay on screen. It is a view
  // limit only: the log still holds every event, and a refresh replays them
  // all. Zero or negative means no cap.
  setMaxEvents(n) {
    this.maxEvents = Number.isFinite(n) && n > 0 ? n : 0;
    this._enforceCap();
  }

  // _enforceCap discards the oldest items until the screen is within the cap.
  //
  // The correlation maps are pruned along with the DOM, for the same reason
  // clear() empties them: an entry that outlives its element routes the next
  // delta into a node that is no longer on the page, and the text silently
  // vanishes. The hazard is sharper here than in clear(), because trimming
  // happens WHILE a turn is streaming. Only the oldest items are discarded,
  // so a part still streaming is the newest and is never the one dropped.
  _enforceCap() {
    if (!this.maxEvents) return;
    while (this.container.childElementCount > this.maxEvents) {
      this.container.removeChild(this.container.firstElementChild);
    }
    for (const [id, el] of this.artifacts) {
      if (!this.container.contains(el)) {
        this.artifacts.delete(id);
        this.accumulated.delete(id);
      }
    }
    for (const [id, el] of this.toolCards) {
      if (!this.container.contains(el)) this.toolCards.delete(id);
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
    // An opaque part (encrypted reasoning) carries nothing to display. If it
    // never streamed a summary there is no element for it, and creating one
    // leaves an empty artifact on screen for every such turn.
    if (!el && msg.opaque) {
      TTS.flush();
      return;
    }
    if (!el) {
      el = document.createElement('div');
      el.className = 'artifact';
      // Thinking blocks from reconnection carry kind="thinking" so the
      // same styling applies as during live streaming.
      if (msg.kind === 'thinking') el.classList.add('thinking');
      this.container.appendChild(el);
      this.artifacts.set(id, el);
    }

    if (msg.text !== undefined) {
      el.innerHTML = Renderers.markdown(msg.text);

      // A part that never streamed has never been spoken. Deltas are the only other
      // feed into speech, and this.accumulated is populated by deltas alone, so its
      // absence is an exact test for "arrived whole". Without this, disabling
      // streaming, or using a model that cannot stream, leaves the agent SILENT
      // while still looking perfectly healthy on screen.
      if (!this.accumulated.has(id)) TTS.queueChunk(msg.text);
    } else if (msg.tool) {
      el.innerHTML = Renderers.json(`${msg.tool}: ${msg.args || ''}`);
    }

    // Finals may arrive without deltas (including restored history). Keep
    // the newest reply visible even if no turn-ended frame follows it.
    this._scrollToBottom();

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

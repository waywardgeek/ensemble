// TTS — Text-to-speech for artifacts with Chrome wake-up and pause integration.

const TTS = {
  queue: [],         // entries awaiting speech
  current: null,     // entry in flight; already shifted off `queue`
  speaking: false,
  enabled: true,
  rate: 1.2,

  // Record to the transcript but never call speechSynthesis. Used when an automated
  // observer is reading the speech channel: audio is real time, so a driver that
  // waited for it would run at talking speed, and the tab would talk through every
  // test. Text still enters the channel, which is the thing under test.
  bypass: false,

  // Partial text awaiting a phrase boundary. Streaming deltas arrive on arbitrary
  // byte edges, frequently mid-word, so a chunk is not an utterance. Handing "Hel"
  // and then "lo" to the engine produces two wrong noises rather than one word, and
  // every chunk seam that is not a real phrase boundary is heard as a stumble.
  _buffer: '',

  // Everything that has entered the speech channel, in order.
  //
  // Recorded at ENQUEUE rather than at audio completion. Speech is real time, so a
  // transcript that waited for audio would force any observer reading it to run at
  // talking speed. The bug class this exists to expose is "text that reached the
  // screen and never reached the speech channel at all", and entering the queue is
  // the moment that question is decided. `status` still separates text that was cut
  // off from text fully spoken.
  transcript: [],
  _seq: 0,
  _maxTranscript: 500,

  // Fired whenever `speaking` changes. The pause gate subscribes to this; without
  // it the speaking half of the gate is unobservable from outside this file, which
  // is why it went unimplemented for so long.
  onStateChange: null,

  // Fired once per entry as it enters the speech channel. The browser cannot write
  // files, so this is how the record leaves: the GUI forwards it to the server,
  // which appends it to the speech log. Only the client can report this, because
  // the decision about what is speakable is made here.
  onRecord: null,

  // Invalidation token for in-flight utterance callbacks. speechSynthesis.cancel()
  // delivers onend/onerror asynchronously, so a callback from a cancelled utterance
  // can otherwise land after its replacement has already started and clear
  // `speaking` while audio is still playing.
  _gen: 0,

  _setSpeaking(v) {
    if (this.speaking === v) return;
    this.speaking = v;
    if (this.onStateChange) this.onStateChange();
  },

  // ── Making text speakable ────────────────────────────────────────────────────
  //
  // The engine reads exactly what it is handed. Markup that is invisible on screen
  // is loud in the ear: "**ready**" is heard as "star star ready star star". So the
  // marks are removed here rather than being left for a listener to filter out.

  _expandIdentifier(tok) {
    const snake = tok.indexOf('_') >= 0;
    const camel = /[a-z][A-Z]/.test(tok) || /[A-Z]{2,}[a-z]/.test(tok);
    if (!snake && !camel) return tok;      // ordinary word, leave it alone
    return tok
      .replace(/_/g, ' ')
      .replace(/([a-z0-9])([A-Z])/g, '$1 $2')       // queueChunk -> queue Chunk
      .replace(/([A-Z]+)([A-Z][a-z])/g, '$1 $2')    // HTTPServer -> HTTP Server
      .replace(/\s+/g, ' ')
      .trim();
  },

  _speakable(text) {
    return text
      // A fenced block is unlistenable read aloud, so name it instead.
      .replace(/```[\s\S]*?```/g, ' code block. ')
      .replace(/`([^`]*)`/g, '$1')
      .replace(/!?\[([^\]]*)\]\([^)]*\)/g, '$1')    // links and images: keep the text
      .replace(/^\s{0,3}#{1,6}\s+/gm, '')           // headings
      .replace(/^\s*[-*+]\s+/gm, '')                // bullets
      .replace(/^\s*>\s?/gm, '')                    // block quotes
      // Identifiers expand BEFORE emphasis marks are stripped, or snake_case would
      // lose its underscores first and arrive here looking like one long word.
      .replace(/[A-Za-z_][A-Za-z0-9_]{2,}/g, (t) => this._expandIdentifier(t))
      .replace(/[*~_]/g, '')
      .replace(/[ \t]{2,}/g, ' ')
      .trim();
  },

  // The single door into the speech channel. Every enqueue path goes through here,
  // so the transcript cannot silently miss one and no path escapes normalisation.
  _enqueue(raw) {
    const text = this._speakable(raw);
    if (!text) return null;
    const entry = {seq: ++this._seq, text: text, at: Date.now(), status: 'pending'};
    if (raw.trim() !== text) entry.raw = raw.trim();
    this.transcript.push(entry);
    if (this.transcript.length > this._maxTranscript) this.transcript.shift();
    this.queue.push(entry);
    if (this.onRecord) this.onRecord(entry);
    return entry;
  },

  init() {
    // Chrome wake-up: zero-volume utterance on tab focus.
    document.addEventListener('visibilitychange', () => {
      if (!document.hidden && 'speechSynthesis' in window) {
        const wake = new SpeechSynthesisUtterance('');
        wake.volume = 0;
        speechSynthesis.speak(wake);
      }
    });
  },

  // Resolve fenced code blocks against the whole buffer and report what must wait.
  //
  // Markdown is a property of the TEXT, but the phrase splitter breaks on newlines,
  // so by the time a per-phrase filter runs, a fence has already been shattered into
  // lone ``` lines that match no fence pattern. The listener then hears "backtick".
  // Complete fences are therefore named here, and an unterminated one is held back:
  // its contents are not speech, and the closing marker may still be streaming.
  _resolveFences(s) {
    const resolved = s.replace(/```[\s\S]*?```/g, ' code block. ');
    const open = resolved.indexOf('```');
    if (open < 0) return [resolved, ''];
    return [resolved.slice(0, open), resolved.slice(open)];
  },

  queueChunk(text) {
    if (!this.enabled || !('speechSynthesis' in window)) return;
    this._buffer += text;

    const [eligible, held] = this._resolveFences(this._buffer);

    // Newlines are whitespace, not phrase boundaries. Prose wraps, and splitting on
    // the wrap put an unnatural pause inside a single sentence. A BLANK line is a
    // real paragraph break and stays a boundary.
    const flat = eligible.replace(/\n[ \t]*\n\s*/g, '\u0001').replace(/\n/g, ' ');

    // Emit only through the last COMPLETED boundary. split() leaves the trailing
    // fragment as its final element, which is exactly the part that may be half a
    // word, so it goes back in the buffer to wait for more input or a flush.
    const parts = flat.split(/(?<=[.!?])\s+|\u0001/);
    this._buffer = parts.pop() + held;
    for (const p of parts) this._enqueue(p);
    this._processQueue();
  },

  // Speak whatever is buffered, terminal punctuation or not.
  //
  // Streaming routinely ends on a colon or a bare clause, and a buffer waiting for a
  // period that will never arrive is simply never heard. End of stream is itself a
  // phrase boundary, so callers must say so.
  flush() {
    if (!this.enabled || !('speechSynthesis' in window)) return;
    // End of stream closes any fence that never closed itself, or its contents
    // would be held in the buffer forever and silently lost.
    const rest = this._buffer.replace(/```/g, ' code block. ');
    this._buffer = '';
    if (rest.trim()) this._enqueue(rest);
    this._processQueue();
  },


  speakToolDispatch(name, input) {
    if (!this.enabled || !('speechSynthesis' in window)) return;
    let summary = name;
    if (input) {
      try {
        const obj = typeof input === 'string' ? JSON.parse(input) : input;
        if (obj.file || obj.path) summary += ': ' + (obj.file || obj.path);
        if (obj.start_line) summary += ', line ' + obj.start_line;
      } catch(e) { /* ignore */ }
    }
    this._enqueue(summary);
    this._processQueue();
  },

  speakFull(text) {
    if (!('speechSynthesis' in window)) return;
    this._gen++;                 // invalidate callbacks from the utterance we cancel
    speechSynthesis.cancel();
    this._discard();
    this._enqueue(text);
    this._setSpeaking(false);
    this._processQueue();
  },

  // Errors are spoken, not merely displayed. A reader who works by speech would
  // otherwise never learn an error occurred. A complete phrase needs no boundary,
  // so it goes straight out rather than through the buffer.
  speakError(text) {
    if (!this.enabled || !('speechSynthesis' in window)) return;
    this._enqueue('Error. ' + text);
    this._processQueue();
  },

  // Drop everything unspoken, marking it so the transcript shows what was lost
  // rather than quietly forgetting it.
  _discard() {
    if (this.current && this.current.status === 'speaking') {
      this.current.status = 'cancelled';
    }
    for (const e of this.queue) e.status = 'cancelled';
    this.queue = [];
    this.current = null;
  },

  cancel() {
    this._gen++;                 // invalidate in-flight callbacks before clearing
    if ('speechSynthesis' in window) {
      speechSynthesis.cancel();
    }
    this._buffer = '';
    this._discard();
    this._setSpeaking(false);
  },

  _processQueue() {
    if (this.speaking || this.queue.length === 0) return;

    if (this.bypass) {
      // Deliver to the transcript without audio. Marked 'bypassed' rather than
      // 'spoken' so the record never claims audio that did not play. `speaking`
      // stays false, so the pause gate correctly does not close: there is no
      // speech to talk over.
      while (this.queue.length) this.queue.shift().status = 'bypassed';
      this.current = null;
      return;
    }

    this._setSpeaking(true);
    const entry = this.queue.shift();
    this.current = entry;
    entry.status = 'speaking';
    const utt = new SpeechSynthesisUtterance(entry.text);
    utt.rate = this.rate;

    const gen = this._gen;
    const done = () => {
      if (gen !== this._gen) return;   // superseded by cancel() or speakFull()
      entry.status = 'spoken';
      this.current = null;
      this._setSpeaking(false);
      this._processQueue();
    };
    utt.onend = done;
    utt.onerror = done; // onerror MUST mirror onend or the queue deadlocks

    speechSynthesis.speak(utt);
  }
};

TTS.init();

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

  // The single door into the speech channel. Every enqueue path goes through here
  // so the transcript cannot silently miss one.
  _enqueue(text) {
    const entry = {seq: ++this._seq, text: text, at: Date.now(), status: 'pending'};
    this.transcript.push(entry);
    if (this.transcript.length > this._maxTranscript) this.transcript.shift();
    this.queue.push(entry);
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

  queueChunk(text) {
    if (!this.enabled || !('speechSynthesis' in window)) return;
    const sentences = text.split(/(?<=[.!?])\s+/);
    for (const s of sentences) {
      if (s.trim()) {
        this._enqueue(s.trim());
      }
    }
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
  // otherwise never learn an error occurred.
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

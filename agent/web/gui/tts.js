// TTS — Text-to-speech for artifacts with Chrome wake-up and pause integration.

const TTS = {
  queue: [],
  current: null,     // utterance in flight; already shifted off `queue`
  speaking: false,
  enabled: true,
  rate: 1.2,

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
        this.queue.push(s.trim());
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
    this.queue.push(summary);
    this._processQueue();
  },

  speakFull(text) {
    if (!('speechSynthesis' in window)) return;
    this._gen++;                 // invalidate callbacks from the utterance we cancel
    speechSynthesis.cancel();
    this.queue = [text];
    this._setSpeaking(false);
    this._processQueue();
  },

  // Errors are spoken, not merely displayed. A reader who works by speech would
  // otherwise never learn an error occurred.
  speakError(text) {
    if (!this.enabled || !('speechSynthesis' in window)) return;
    this.queue.push('Error. ' + text);
    this._processQueue();
  },

  cancel() {
    this._gen++;                 // invalidate in-flight callbacks before clearing
    if ('speechSynthesis' in window) {
      speechSynthesis.cancel();
    }
    this.queue = [];
    this.current = null;
    this._setSpeaking(false);
  },

  _processQueue() {
    if (this.speaking || this.queue.length === 0) return;
    this._setSpeaking(true);
    const text = this.queue.shift();
    this.current = text;
    const utt = new SpeechSynthesisUtterance(text);
    utt.rate = this.rate;

    const gen = this._gen;
    const done = () => {
      if (gen !== this._gen) return;   // superseded by cancel() or speakFull()
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

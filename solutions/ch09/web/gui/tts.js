// TTS — Text-to-speech for artifacts with Chrome wake-up and pause integration.

const TTS = {
  queue: [],
  speaking: false,
  enabled: true,
  rate: 1.2,

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
    speechSynthesis.cancel();
    this.queue = [text];
    this.speaking = false;
    this._processQueue();
  },

  cancel() {
    if ('speechSynthesis' in window) {
      speechSynthesis.cancel();
    }
    this.queue = [];
    this.speaking = false;
  },

  _processQueue() {
    if (this.speaking || this.queue.length === 0) return;
    this.speaking = true;
    const text = this.queue.shift();
    const utt = new SpeechSynthesisUtterance(text);
    utt.rate = this.rate;

    const done = () => {
      this.speaking = false;
      this._processQueue();
    };
    utt.onend = done;
    utt.onerror = done; // onerror MUST mirror onend or the queue deadlocks

    speechSynthesis.speak(utt);
  }
};

TTS.init();

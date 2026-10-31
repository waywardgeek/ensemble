import { VOICE_PCM } from './voice-pcm.js';

class PcmVoiceProcessor extends AudioWorkletProcessor {
  constructor(options) {
    super();
    const maximum = options.processorOptions.maximumSamples;
    if (sampleRate !== VOICE_PCM.sampleRate || !Number.isSafeInteger(maximum)
        || maximum < 1 || maximum > VOICE_PCM.maximumSamples)
      throw new Error('Unsupported voice recorder format or limit.');
    this.maximum = maximum;
    this.buffer = new ArrayBuffer(maximum * VOICE_PCM.blockAlign);
    this.samples = new DataView(this.buffer);
    this.count = 0;
    this.closed = false;
    this.started = false;
    this.nextProgress = 1600;
    this.port.onmessage = event => {
      if (this.closed) return;
      this.closed = true;
      if (event.data?.type === 'finish') {
        const pcm = this.buffer.slice(0, this.count * VOICE_PCM.blockAlign);
        this.port.postMessage({ type: 'finished', samples: this.count, pcm }, [pcm]);
      } else {
        this.port.postMessage({ type: 'error' });
      }
      this.buffer = null;
      this.samples = null;
    };
  }

  process(inputs) {
    // Outputs stay zero: the graph is pulled by the destination, but microphone
    // samples must never be played back into that microphone.
    if (this.closed) return false;
    const input = inputs[0]?.[0];
    if (!input?.length) return true;
    const count = Math.min(input.length, this.maximum - this.count);
    for (let i = 0; i < count; ++i) {
      const value = Number.isFinite(input[i]) ? Math.max(-1, Math.min(1, input[i])) : 0;
      const integer = Math.trunc(value < 0 ? value * 32768 : value * 32767);
      this.samples.setInt16(this.count * VOICE_PCM.blockAlign, integer, true);
      ++this.count;
    }
    if (!this.started) {
      this.started = true;
      this.port.postMessage({ type: 'started', samples: this.count });
    }
    if (this.count === this.maximum) {
      this.closed = true;
      this.buffer = null;
      this.samples = null;
      // Reaching the limit cancels, never automatically submits partial speech.
      this.port.postMessage({ type: 'limit' });
      return false;
    }
    if (this.count >= this.nextProgress) {
      this.nextProgress = this.count + 1600;
      this.port.postMessage({ type: 'progress', samples: this.count });
    }
    return true;
  }
}
registerProcessor('pcm-voice', PcmVoiceProcessor);

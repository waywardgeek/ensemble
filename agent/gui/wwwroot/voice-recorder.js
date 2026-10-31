import { VOICE_PCM, voiceWave } from './voice-pcm.js';

function deferred() {
  let resolve, reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  // Device cancellation can precede finish(). Preserve rejection for the owner,
  // while observing it immediately so it is not an unhandled browser rejection.
  promise.catch(() => {});
  return { promise, resolve, reject };
}
const aborted = reason => new DOMException(reason || 'Voice recording cancelled.', 'AbortError');

// Construction transfers the stream's initial audio track to this recorder.
// A private membership snapshot prevents later caller edits from adding devices
// or evading cleanup. No permission request or network transcription occurs here.
// Each instance is one capture;
// a cancelled instance cannot be restarted or handed to another Agent.
export class PcmVoiceRecorder extends EventTarget {
  #stream; #maximum; #context = null; #source = null; #node = null;
  #state = 'new'; #samples = 0; #cancelled = false;
  #ready = deferred(); #result = deferred();
  #starting = null; #finishing = null; #closing = null; #cancelling = null;
  #ended = () => this.#abort('The audio source ended.');

  constructor(stream, { maximumSamples = VOICE_PCM.maximumSamples } = {}) {
    super();
    if (!(stream instanceof MediaStream))
      throw new Error('Voice recording requires a MediaStream.');
    const tracks = stream.getTracks();
    if (tracks.length !== 1 || tracks[0].kind !== 'audio' || tracks[0].readyState !== 'live')
      throw new Error('Voice recording requires one live audio-only stream.');
    if (!Number.isSafeInteger(maximumSamples) || maximumSamples < 1
        || maximumSamples > VOICE_PCM.maximumSamples)
      throw new Error('Invalid voice recording limit.');
    this.#stream = new MediaStream(tracks);
    this.#maximum = maximumSamples;
    tracks[0].addEventListener('ended', this.#ended);
  }

  get state() { return this.#state; }
  get sampleCount() { return this.#samples; }

  start() {
    if (this.#state !== 'new') return Promise.reject(aborted('Recorder cannot be restarted.'));
    this.#state = 'starting';
    this.#starting = this.#start();
    this.#starting.catch(() => {});
    return this.#starting;
  }

  async #start() {
    try {
      if (this.#stream.getAudioTracks()[0].readyState !== 'live')
        throw new Error('The audio source ended.');
      const context = new AudioContext({ sampleRate: VOICE_PCM.sampleRate });
      this.#context = context;
      if (context.sampleRate !== VOICE_PCM.sampleRate || !context.audioWorklet)
        throw new Error('This browser cannot record the required voice format.');
      context.onstatechange = () => {
        if (!this.#closing && ['recording', 'finishing'].includes(this.#state)
            && context.state !== 'running') this.#abort('The audio context stopped.');
      };
      await context.resume();
      if (this.#cancelled) throw aborted();
      await context.audioWorklet.addModule(new URL('./voice-worklet.js', import.meta.url));
      if (this.#cancelled) throw aborted();
      if (this.#stream.getAudioTracks()[0].readyState !== 'live')
        throw new Error('The audio source ended.');
      this.#node = new AudioWorkletNode(context, 'pcm-voice', {
        numberOfInputs: 1, numberOfOutputs: 1, outputChannelCount: [1],
        channelCount: 1, channelCountMode: 'explicit',
        processorOptions: { maximumSamples: this.#maximum }
      });
      this.#node.onprocessorerror = () => this.#abort('Voice audio processing failed.');
      this.#node.port.onmessage = event => this.#message(event.data);
      this.#source = context.createMediaStreamSource(this.#stream);
      this.#source.connect(this.#node);
      this.#node.connect(context.destination);
      await this.#ready.promise;
      if (this.#cancelled) throw aborted();
      this.#state = 'recording';
    } catch (error) {
      this.#cancelled = true;
      this.#ready.reject(error);
      this.#result.reject(error);
      try { await this.#release(); }
      catch (cleanup) { throw new AggregateError([error, cleanup], 'Voice startup and cleanup failed.'); }
      finally { this.#state = 'closed'; }
      throw error;
    }
  }

  #message(data) {
    if (this.#cancelled || this.#closing) return;
    if (data?.type === 'limit') { this.#abort('Voice recording reached its duration limit.'); return; }
    if (data?.type === 'error') { this.#abort('Voice audio processing failed.'); return; }
    if (!Number.isSafeInteger(data?.samples) || data.samples < 0 || data.samples > this.#maximum) {
      this.#abort('Invalid voice sample count.'); return;
    }
    this.#samples = data.samples;
    if (data.type === 'started') this.#ready.resolve();
    else if (data.type === 'progress') this.dispatchEvent(new Event('progress'));
    else if (data.type === 'finished' && this.#state === 'finishing') {
      if (!(data.pcm instanceof ArrayBuffer) || data.pcm.byteLength !== data.samples * VOICE_PCM.blockAlign)
        this.#abort('Invalid voice sample payload.');
      else this.#result.resolve(data.pcm);
    } else this.#abort('Unexpected voice recorder message.');
  }

  finish() {
    if (this.#cancelled) return Promise.reject(aborted());
    if (this.#finishing) return this.#finishing;
    if (this.#state !== 'recording') return Promise.reject(new Error('Voice recording is not ready.'));
    if (this.#stream.getAudioTracks().some(track => track.readyState !== 'live')) {
      this.#abort('The audio source ended.');
      return Promise.reject(aborted());
    }
    this.#state = 'finishing';
    this.#finishing = this.#finish();
    this.#finishing.catch(() => {});
    this.#node.port.postMessage({ type: 'finish' });
    return this.#finishing;
  }

  async #finish() {
    try {
      const pcm = await this.#result.promise;
      if (this.#cancelled) throw aborted();
      const wave = voiceWave(pcm);
      await this.#release();
      if (this.#cancelled) throw aborted();
      this.#state = 'closed';
      return wave;
    } catch (error) {
      this.#cancelled = true;
      try { await this.#release(); }
      catch (cleanup) { throw new AggregateError([error, cleanup], 'Voice finish and cleanup failed.'); }
      finally { this.#state = 'closed'; }
      throw error;
    }
  }

  #abort(reason) {
    this.cancel(reason).catch(error => this.dispatchEvent(new CustomEvent('cleanup-error', { detail: error })));
  }

  cancel(reason = 'Voice recording cancelled.') {
    if (this.#cancelling) return this.#cancelling;
    if (this.#state === 'closed') return Promise.resolve();
    this.#cancelled = true;
    this.#state = 'cancelling';
    this.#ready.reject(aborted(reason));
    this.#result.reject(aborted(reason));
    // Stop tracks immediately. Then join preparation, finish and native close;
    // addModule has no AbortSignal, so a late completion must stay inert.
    const closing = this.#release();
    this.#cancelling = (async () => {
      await Promise.allSettled([this.#starting, this.#finishing]);
      try { await closing; }
      finally { this.#state = 'closed'; }
      this.dispatchEvent(new CustomEvent('cancelled', { detail: reason }));
    })();
    return this.#cancelling;
  }

  #release() {
    if (this.#closing) return this.#closing;
    this.#closing = (async () => {
      const errors = [];
      const attempt = action => { try { action(); } catch (error) { errors.push(error); } };
      if (this.#context) this.#context.onstatechange = null;
      if (this.#node) {
        this.#node.port.onmessage = null;
        this.#node.onprocessorerror = null;
        attempt(() => this.#node.port.close());
        attempt(() => this.#node.disconnect());
      }
      if (this.#source) attempt(() => this.#source.disconnect());
      for (const track of this.#stream.getTracks()) {
        track.removeEventListener('ended', this.#ended);
        attempt(() => track.stop());
      }
      if (this.#context && this.#context.state !== 'closed') {
        try { await this.#context.close(); } catch (error) { errors.push(error); }
      }
      if (errors.length) throw new AggregateError(errors, 'Voice device cleanup failed.');
    })();
    return this.#closing;
  }
}

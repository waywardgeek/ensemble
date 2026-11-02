import { PcmVoiceRecorder } from './voice-recorder.js';

// One user hold owns one permission request and recorder. The UI must create a
// fresh hold per gesture and cancel it on blur, selection change or disconnect.
// This layer never submits text or chooses an Agent.
export class VoiceHold extends EventTarget {
  #requestStream; #createRecorder; #recorder = null;
  #state = 'new'; #cancelled = false;
  #starting = null; #releasing = null; #cancelling = null;
  #recorderCancelled = () => {
    this.#cancelled = true;
    this.#state = 'cancelled';
    this.dispatchEvent(new Event('cancelled'));
  };

  constructor({
    requestStream = () => navigator.mediaDevices.getUserMedia({ audio: true, video: false }),
    createRecorder = stream => new PcmVoiceRecorder(stream)
  } = {}) {
    super();
    this.#requestStream = requestStream;
    this.#createRecorder = createRecorder;
  }

  get state() { return this.#state; }

  start() {
    if (this.#state !== 'new') return Promise.reject(new Error('Voice hold already used.'));
    this.#state = 'preparing';
    this.#starting = this.#start();
    this.#starting.catch(() => {});
    return this.#starting;
  }

  async #start() {
    let stream = null;
    try {
      // Browser permission requests cannot be aborted. Cancelled holds still
      // observe their result and immediately stop any tracks delivered later.
      stream = await this.#requestStream();
      if (this.#cancelled) {
        for (const track of stream.getTracks()) track.stop();
        return false;
      }
      const recorder = this.#createRecorder(stream);
      this.#recorder = recorder;
      stream = null; // Ownership transferred only after construction succeeds.
      recorder.addEventListener('cancelled', this.#recorderCancelled);
      await recorder.start();
      if (this.#cancelled) {
        await recorder.cancel();
        return false;
      }
      this.#state = 'recording';
      return true;
    } catch (error) {
      if (stream) {
        for (const track of stream.getTracks()) track.stop();
      }
      if (this.#recorder) await this.#recorder.cancel();
      if (this.#cancelled) return false;
      this.#state = 'failed';
      throw error;
    }
  }

  release() {
    if (this.#releasing) return this.#releasing;
    if (this.#state !== 'recording') {
      // A short gesture must not record after the human has released the key.
      this.#releasing = this.cancel().then(() => null);
    } else {
      this.#state = 'finishing';
      this.#releasing = this.#finish();
    }
    this.#releasing.catch(() => {});
    return this.#releasing;
  }

  async #finish() {
    try {
      const wave = await this.#recorder.finish();
      if (this.#cancelled) return null;
      this.#state = 'finished';
      return wave;
    } catch (error) {
      if (this.#cancelled) return null;
      this.#state = 'failed';
      throw error;
    } finally {
      this.#recorder.removeEventListener('cancelled', this.#recorderCancelled);
    }
  }

  cancel() {
    if (this.#cancelling) return this.#cancelling;
    this.#cancelled = true;
    this.#state = 'cancelled';
    const recorder = this.#recorder;
    if (recorder) recorder.removeEventListener('cancelled', this.#recorderCancelled);
    // Invoke cancellation immediately, before joining an unabortable permission
    // request or a finish already in flight. Never wait before stopping devices.
    const stopping = recorder ? recorder.cancel() : Promise.resolve();
    // Snapshot the already-running finish. An early release creates its own
    // promise from cancel(); reading that later would make cancellation join itself.
    const finishing = this.#releasing;
    this.#cancelling = (async () => {
      const results = await Promise.allSettled([stopping, this.#starting]);
      if (finishing) await Promise.allSettled([finishing]);
      const failures = results.filter(result => result.status === 'rejected');
      if (failures.length) throw new AggregateError(failures.map(result => result.reason), 'Voice hold cleanup failed.');
    })();
    this.#cancelling.catch(() => {});
    return this.#cancelling;
  }
}

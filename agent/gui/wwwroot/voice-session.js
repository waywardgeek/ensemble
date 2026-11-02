import { VoiceHold } from './voice-hold.js';

// One gesture owns one hold and borrows its original socket. The page supplies
// isCurrent and must cancel on blur, selection change, disconnect and unload.
// No transcript is a browser draft: a matching ready reply commits automatically.
export class VoiceSession extends EventTarget {
  #socket; #agent; #sequence; #current; #hold;
  #state = 'new'; #begun = false; #wave = null;
  #releasing = null; #cleanup = null;
  #holdCancelled = () => { this.cancel(); };

  constructor({ socket, agent, sequence, isCurrent, hold = new VoiceHold() }) {
    super();
    if (!socket || typeof agent !== 'string' || !Number.isSafeInteger(sequence)
        || sequence <= 0 || typeof isCurrent !== 'function')
      throw new Error('Invalid voice session binding.');
    this.#socket = socket;
    this.#agent = agent; // Empty string is the canonical unnamed-root address.
    this.#sequence = sequence;
    this.#current = isCurrent;
    this.#hold = hold;
    hold.addEventListener('cancelled', this.#holdCancelled);
  }

  get state() { return this.#state; }
  get terminal() { return ['accepted', 'cancelled', 'failed', 'uncertain'].includes(this.#state); }

  #set(state) {
    this.#state = state;
    this.dispatchEvent(new Event('change'));
  }

  #valid() { return this.#socket.readyState === 1 && this.#current(); }

  #send(type, fields = {}) {
    if (type !== 'voice_cancel' && (this.terminal || !this.#valid()))
      throw new Error('Voice session is no longer current.');
    this.#socket.send(JSON.stringify({ type, agent_id: this.#agent,
      sequence: this.#sequence, ...fields }));
  }

  async start() {
    if (this.#state !== 'new') throw new Error('Voice session already used.');
    if (!this.#valid()) { await this.cancel(); return false; }
    this.#set('preparing');
    try {
      this.#send('voice_begin');
      // Invoke permission within the original gesture, not after the network ack.
      const recording = await this.#hold.start();
      if (this.terminal) return false;
      if (!recording || !this.#valid()) { await this.cancel(); return false; }
      this.#set('recording');
      return !this.terminal;
    } catch (error) {
      if (!this.terminal) this.#stop('failed', true);
      throw error;
    }
  }

  release() {
    if (this.#releasing) return this.#releasing;
    this.#releasing = this.#release();
    this.#releasing.catch(() => {});
    return this.#releasing;
  }

  async #release() {
    if (this.terminal) return;
    if (this.#state !== 'recording') { await this.cancel(); return; }
    this.#set('finishing');
    try {
      const wave = await this.#hold.release();
      if (this.terminal) return;
      if (!wave || !this.#valid()) { await this.cancel(); return; }
      // Canonical structure is verified by the server before paying for transcription.
      if (!(wave instanceof Uint8Array) || wave.byteLength <= 44 || wave.byteLength > 1920044)
        throw new Error('Voice audio exceeds its bounded format.');
      this.#wave = wave;
      this.#set('waiting');
      this.#sendAudio();
    } catch (error) {
      if (!this.terminal) this.#stop('failed', true);
      throw error;
    }
  }

  #sendAudio() {
    if (!this.#begun || this.#state !== 'waiting') return;
    if (!this.#valid()) { this.cancel(); return; }
    let binary = '';
    for (let offset = 0; offset < this.#wave.length; offset += 16384)
      binary += String.fromCharCode(...this.#wave.subarray(offset, offset + 16384));
    this.#wave = null;
    this.#set('transcribing'); // Claim the send before any duplicate reply.
    this.#send('voice_audio', { audio: { mime: 'audio/wav', data: btoa(binary) } });
  }

  receive(socket, message) {
    if (socket !== this.#socket || this.terminal) return false;
    const error = message.type === 'error';
    if ((error ? message.voice_sequence : message.sequence) !== this.#sequence
        || (!error && message.agent_id !== this.#agent)) return false;
    if (!this.#valid()) { this.cancel(); return false; }
    try {
      switch (message.type) {
        case 'voice_begun':
          if (this.#begun || !['preparing', 'recording', 'finishing', 'waiting'].includes(this.#state)) return false;
          this.#begun = true;
          this.#sendAudio();
          break;
        case 'voice_ready':
          if (this.#state !== 'transcribing') return false;
          // After this point even a throwing send may have admitted work. Never
          // claim cancellation retracted it, or retry on a replacement socket.
          this.#set('committing');
          this.#send('voice_commit');
          break;
        case 'voice_accepted':
          if (this.#state !== 'committing') return false;
          this.#stop('accepted', false);
          break;
        case 'voice_failed':
        case 'voice_cancelled':
        case 'error':
          this.#stop(this.#state === 'committing' ? 'uncertain' : 'failed', true);
          break;
        default: return false;
      }
    } catch {
      if (!this.terminal) this.#stop(this.#state === 'committing' ? 'uncertain' : 'failed', true);
    }
    return true;
  }

  cancel() {
    if (!this.terminal) this.#stop(this.#state === 'committing' ? 'uncertain' : 'cancelled', true);
    return this.#cleanup ?? Promise.resolve();
  }

  #stop(state, notify) {
    this.#wave = null;
    this.#hold.removeEventListener('cancelled', this.#holdCancelled);
    // Publish the join before notifications can reenter cancel().
    this.#cleanup ??= Promise.resolve().then(() => this.#hold.cancel());
    this.#cleanup.catch(() => {});
    this.#set(state); // Invalidate late callbacks before stopping/joining capture.
    if (notify && this.#socket.readyState === 1) {
      try { this.#send('voice_cancel'); } catch { /* Local invalidation still wins. */ }
    }
  }
}

import { VoiceSession } from './voice-session.js';

// Page-lifetime gesture owner. Borrows the page's current connection/Agent;
// every individual session freezes that binding before requesting permission.
export class VoiceControl extends EventTarget {
  #button; #cancel; #status; #factory;
  #shortcut = "off";
  #binding = null; #session = null; #joining = false; #gesture = null;
  constructor(button, cancel, status, factory = options => new VoiceSession(options)) {
    super();
    this.#button = button; this.#cancel = cancel; this.#status = status; this.#factory = factory;
    button.addEventListener('pointerdown', event => {
      if (!event.isTrusted || event.button !== 0 || button.disabled) return;
      event.preventDefault();
      button.setPointerCapture(event.pointerId);
      this.#begin({ pointer: event.pointerId });
    });
    button.addEventListener('pointerup', event => {
      if (this.#gesture?.pointer !== event.pointerId) return;
      event.preventDefault(); this.#release();
    });
    for (const name of ['pointercancel', 'lostpointercapture'])
      button.addEventListener(name, event => {
        if (this.#gesture?.pointer === event.pointerId) this.cancel();
      });
    button.addEventListener('keydown', event => {
      if (!event.isTrusted || !['Space', 'Enter'].includes(event.code)) return;
      event.preventDefault();
      if (!event.repeat && !button.disabled && !this.#gesture) this.#begin({ key: event.code });
    });
    button.addEventListener('keyup', event => {
      if (!this.#gesture?.shortcut && this.#gesture?.key === event.code) { event.preventDefault(); this.#release(); }
    });
    button.addEventListener('blur', () => { if (this.#gesture?.key) this.cancel(); });
    cancel.addEventListener('click', () => {
      if (this.#session?.state === 'uncertain' && !this.#joining) {
        this.#session = null; this.#changed(); // Explicit acknowledgement, never retransmission.
      } else this.cancel();
    });
    window.addEventListener('keydown', event => {
      if (!event.isTrusted) return;
      // A second key or unsafe modifier invalidates a held shortcut before effects.
      if (this.#gesture?.shortcut && (event.code !== this.#gesture.key || !this.#safeShortcut(event))) {
        this.cancel(); return;
      }
      if (event.code !== this.#shortcut || event.repeat || this.#gesture || this.#button.disabled
          || event.defaultPrevented || !this.#safeShortcut(event) || document.querySelector('dialog[open]')) return;
      event.preventDefault();
      this.#begin({ key: event.code, shortcut: true });
    });
    window.addEventListener('keyup', event => {
      if (!event.isTrusted || !this.#gesture?.shortcut || event.code !== this.#gesture.key) return;
      event.preventDefault();
      if (this.#safeShortcut(event)) this.#release(); else this.cancel();
    });
    document.addEventListener('focusin', () => { if (this.#gesture?.shortcut) this.cancel(); });
    document.addEventListener('compositionstart', () => { if (this.#gesture?.shortcut) this.cancel(); });
    window.addEventListener('blur', () => this.cancel());
    window.addEventListener('pagehide', () => this.cancel());
    document.addEventListener('visibilitychange', () => { if (document.hidden) this.cancel(); });
    this.#render();
  }
  setShortcut(value) {
    if (!['off', 'AltRight', 'F8'].includes(value)) throw new Error('Unsupported voice shortcut');
    if (value === this.#shortcut) return;
    this.cancel();
    this.#shortcut = value;
    this.#button.setAttribute('aria-keyshortcuts', value === 'off' ? '' : value === 'AltRight' ? 'Alt' : 'F8');
  }
  #safeShortcut(event) {
    if (event.isComposing || event.key === 'AltGraph' || event.getModifierState('AltGraph')
        || event.ctrlKey || event.shiftKey || event.metaKey) return false;
    if (event.code === 'AltRight') return !/Mac|iPhone|iPad/i.test(navigator.platform);
    return !event.altKey;
  }
  get blocked() { return this.#joining || !!this.#session && (!this.#session.terminal || this.#session.state === 'uncertain'); }
  bind(socket, agent, ready, draftBlocked) {
    const previous = this.#binding;
    this.#binding = { socket, agent, ready, draftBlocked };
    if (previous && (socket !== previous.socket || agent !== previous.agent || !ready || draftBlocked)) this.cancel();
    this.#render();
  }
  receive(socket, message) { return this.#session?.receive(socket, message) ?? false; }
  cancel() {
    this.#gesture = null;
    if (!this.#session || this.#session.terminal) return;
    this.#session.cancel().catch(() => { this.#status.textContent = 'Voice cleanup failed; recording was cancelled.'; });
  }
  #begin(gesture) {
    const binding = this.#binding;
    if (this.blocked || !binding?.ready || binding.draftBlocked || !binding.socket?.voiceFunding) return;
    this.#gesture = gesture;
    const session = this.#factory({ socket: binding.socket, agent: binding.agent,
      sequence: binding.socket.voiceSequence = (binding.socket.voiceSequence || 0) + 1,
      isCurrent: () => this.#binding?.socket === binding.socket && this.#binding.agent === binding.agent
        && this.#binding.ready && !this.#binding.draftBlocked });
    this.#session = session;
    session.addEventListener('change', () => {
      if (this.#session !== session) return;
      if (session.terminal) {
        this.#gesture = null;
        this.#joining = true;
        session.cancel().then(() => {
          if (this.#session === session) { this.#joining = false; this.#changed(); }
        }, () => {
          if (this.#session === session) {
            this.#joining = false; this.#changed();
            this.#status.textContent = 'Voice cleanup failed; no recording will be reused.';
          }
        });
      }
      this.#changed();
    });
    session.start().catch(() => {}); // Session publishes failure; no raw device errors or transcript.
  }
  #release() {
    this.#gesture = null; // lostpointercapture after release must not cancel transcription.
    this.#session?.release().catch(() => {});
  }
  #changed() { this.#render(); this.dispatchEvent(new Event('change')); }
  #render() {
    const binding = this.#binding;
    const supported = window.isSecureContext && !!navigator.mediaDevices?.getUserMedia;
    const state = this.#session?.state;
    // Keep a held button enabled: disabling it can prevent the release event.
    this.#button.disabled = !binding?.ready || binding.draftBlocked || !binding.socket?.voiceFunding
      || !supported || (this.blocked && !this.#gesture);
    this.#button.setAttribute('aria-pressed', String(!!this.#gesture));
    this.#cancel.hidden = !this.blocked;
    this.#cancel.disabled = this.#joining;
    this.#cancel.textContent = state === 'uncertain' ? 'Dismiss uncertainty' : 'Cancel voice';
    const descriptions = {
      preparing: 'Waiting for microphone. Keep holding; speak when recording starts.',
      recording: 'Recording. Release to transcribe and send once.',
      finishing: 'Finishing recording…', waiting: 'Waiting for voice admission…',
      transcribing: 'Transcribing with separately billed OpenAI funding…',
      committing: 'Submitting voice message…', accepted: 'Voice message sent.',
      cancelled: 'Voice cancelled; no partial speech sent.', failed: 'Voice failed; no recording will be retried.',
      uncertain: 'Voice may have been sent. Check the original Agent conversation; it will not be retried.'
    };
    this.#status.textContent = descriptions[state] || (!supported ? 'Microphone unavailable in this browser or insecure context.'
      : !binding?.socket?.voiceFunding ? 'Voice is off: separate transcription funding is not enabled.'
      : binding.draftBlocked ? 'Send or clear the existing draft before using voice.'
      : 'Hold the microphone, talk, then release to send. Separate transcription charges apply.');
  }
}

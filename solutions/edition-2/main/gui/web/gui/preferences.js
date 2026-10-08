// Page-owned controls show applied state separately from a pending draft.
export class SettingsPanel {
  constructor(page) {
    this.page = page; this.root = page.root; this.closed = false; this.drafts = new Set(); this.policyDraft = false;
    // A new Page owns pending controls, even when an embedding reuses its DOM.
    for (const control of this.root.querySelectorAll('[data-preference], [data-policy-save], [data-divider]')) control.disabled = false;
    this.media = globalThis.matchMedia?.('(prefers-color-scheme: dark)');
    if (this.media) page.listen(this.media, 'change', () => this.style());
    for (const control of this.root.querySelectorAll('[data-preference]')) {
      page.listen(control, 'input', () => this.drafts.add(control.dataset.preference));
      page.listen(control, 'change', () => {
        const name = control.dataset.preference;
        if (name === 'autoplay') page.speech.activate();
        const value = control.type === 'checkbox' ? control.checked : control.type === 'number' ? Number(control.value) : control.value;
        this.change({[name]: value}, control);
      });
    }
    const policyInput = this.root.querySelector('[data-policy]'); if (policyInput) page.listen(policyInput, 'input', () => { this.policyDraft = true; });
    const policy = this.root.querySelector('[data-policy-save]');
    if (policy) page.listen(policy, 'click', () => this.policyChange());
    for (const tab of this.root.querySelectorAll('[data-tab]')) page.listen(tab, 'click', () => {
      for (const peer of this.root.querySelectorAll('[data-tab]')) peer.setAttribute('aria-selected', String(peer === tab));
      for (const panel of this.root.querySelectorAll('[data-tab-panel]')) panel.hidden = panel.dataset.tabPanel !== tab.dataset.tab;
    });
    for (const divider of this.root.querySelectorAll('[data-divider]')) this.divider(divider);
  }
  apply(snapshot) {
    if (this.closed || this.page.closed) return;
    this.applied = snapshot; this.page.speech.preferences(snapshot);
    for (const control of this.root.querySelectorAll('[data-preference]')) {
      if (this.drafts.has(control.dataset.preference)) continue;
      const value = snapshot.preferences[control.dataset.preference];
      if (control.type === 'checkbox') control.checked = value; else control.value = value;
    }
    this.style(); this.summary();
  }
  style() {
    if (!this.applied || this.closed) return;
    const p = this.applied.preferences;
    this.root.dataset.theme = p.theme === 'system' ? (this.media?.matches ? 'dark' : 'light') : p.theme;
    this.root.style.setProperty('--font-size', p.font_size + 'px');
    this.root.style.setProperty('--sidebar-width', p.sidebar_width + 'px');
    this.root.style.setProperty('--actions-width', p.actions_width + 'px');
    for (const divider of this.root.querySelectorAll('[data-divider]')) {
      divider.setAttribute('aria-valuenow', String(p[divider.dataset.divider]));
      divider.setAttribute('aria-valuetext', p[divider.dataset.divider] + ' CSS pixels');
    }
  }
  summary() {
    const label = this.root.querySelector('[data-preferences-applied]');
    if (label && this.applied) { const p = this.applied.preferences; label.textContent = `Applied revision ${this.applied.revision}: ${p.theme}, ${p.font_size}px, widths ${p.sidebar_width}/${p.actions_width}, autoplay ${p.autoplay ? 'on' : 'off'}, rate ${p.speech_rate}.`; }
  }
  policy(value, active = this.active) {
    if (this.closed || this.page.closed) return;
    this.executionPolicy = value; this.active = active;
    const input = this.root.querySelector('[data-policy]'); if (input && !this.policyDraft) input.value = value.max_model_requests;
    const label = this.root.querySelector('[data-policy-applied]');
    if (label) label.textContent = `Next turn: ${value.max_model_requests === 0 ? 'Default (16)' : value.effective_max_model_requests} model requests; revision ${value.revision}; ${value.persistent ? 'persistent' : 'memory only'}. Active turn: ${active == null ? 'none' : active}.`;
  }
  async change(patch, control = null) {
    if (!this.applied || !this.page.connector.settingsReady) return;
    if (control) control.disabled = true;
    try { await this.page.connector.send('preferences_update', {base_revision: this.applied.revision, patch}); for (const key of Object.keys(patch)) this.drafts.delete(key); this.apply(this.applied); this.page.diagnostic('Preferences applied'); }
    catch (error) {
      if (error.domain === 'preferences' && error.current) this.apply(error.current);
      this.page.diagnostic(`Preferences not saved: ${error.message}. Applied values remain shown; retry deliberately.`);
      for (const key of Object.keys(patch)) this.drafts.delete(key); if (this.applied) this.apply(this.applied);
    } finally { if (control && !this.closed) control.disabled = false; }
  }
  async policyChange() {
    if (!this.executionPolicy || !this.page.connector.settingsReady) return;
    const control = this.root.querySelector('[data-policy-save]'), input = this.root.querySelector('[data-policy]'); const submitted = input.value; control.disabled = true;
    try { await this.page.connector.send('policy_update', {base_revision: this.executionPolicy.revision, patch: {max_model_requests: Number(submitted)}}); if (input.value === submitted) { this.policyDraft = false; this.policy(this.executionPolicy); } this.page.diagnostic('Policy applied for later turns'); }
    catch (error) { if (error.domain === 'policy' && error.current) this.policy(error.current); this.page.diagnostic(`Policy not saved: ${error.message}. Applied policy remains shown above.`); }
    finally { if (!this.closed) control.disabled = false; }
  }
  divider(element) {
    const field = element.dataset.divider, min = Number(element.getAttribute('aria-valuemin')), max = Number(element.getAttribute('aria-valuemax'));
    const clamp = value => Math.min(max, Math.max(min, value));
    let drag = null;
    this.page.listen(element, 'keydown', event => {
      if (!['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown'].includes(event.key) || !this.applied) return;
      event.preventDefault(); this.change({[field]: clamp(this.applied.preferences[field] + (['ArrowRight', 'ArrowUp'].includes(event.key) ? 10 : -10))}, element);
    });
    this.page.listen(element, 'pointerdown', event => {
      if (!this.applied || element.disabled) return;
      drag = {x:event.clientX, width:this.applied.preferences[field], value:this.applied.preferences[field]}; element.setPointerCapture(event.pointerId); event.preventDefault();
    });
    this.page.listen(element, 'pointermove', event => {
      if (!drag) return;
      drag.value = clamp(Math.round(drag.width + (event.clientX - drag.x) * (field === 'actions_width' ? -1 : 1)));
      this.root.style.setProperty(field === 'actions_width' ? '--actions-width' : '--sidebar-width', drag.value + 'px');
    });
    this.page.listen(element, 'pointerup', () => { if (!drag) return; const value = drag.value; drag = null; this.style(); this.change({[field]:value}, element); });
    this.page.listen(element, 'pointercancel', () => { drag = null; this.style(); });
  }
  close() { this.closed = true; }
}

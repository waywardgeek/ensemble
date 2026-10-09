// Read-only inspection of actual owners. No synthesized speech or hearing claim.
export function speechState(app, nativeAdmissions) {
  if (!app) throw new Error('instrumented application unavailable');
  return {nativeAdmissions, serviceAdmissions:app.__ch10ServiceAdmissions??0, servicePending: app.service.pending.length,
    serviceActive: app.service.active !== null,
    pages: [...app.pages].map(p => ({closed:p.closed, connected:p.connected,
      queue:p.speech.queue.length, current:p.speech.current !== null,
      causes:p.lastCauses === null || p.lastCauses === undefined ? null : JSON.parse(p.lastCauses),
      partials:p.speech.cursors.size, playbackReady:p.speech.playbackReady}))};
}
export function assertNoRestoredSpeech(before, after) {
  if (after.nativeAdmissions !== before.nativeAdmissions || after.serviceAdmissions !== (before.serviceAdmissions??0) || after.servicePending || after.serviceActive ||
      after.pages.some(p => p.queue || p.current || p.causes?.speaking)) {
    throw new Error('restored speech admission/queue/pause detected');
  }
}

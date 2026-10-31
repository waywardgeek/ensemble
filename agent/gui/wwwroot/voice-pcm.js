// Shared browser-side wire constants; the server independently validates them.
export const VOICE_PCM = Object.freeze({ sampleRate: 16000, channels: 1,
  bitsPerSample: 16, blockAlign: 2, headerBytes: 44, maximumSamples: 960000 });

export function voiceWave(pcm) {
  if (!(pcm instanceof ArrayBuffer) || pcm.byteLength === 0
      || pcm.byteLength % VOICE_PCM.blockAlign !== 0
      || pcm.byteLength > VOICE_PCM.maximumSamples * VOICE_PCM.blockAlign)
    throw new Error('Voice samples are empty, misaligned, or too large.');
  const wave = new Uint8Array(VOICE_PCM.headerBytes + pcm.byteLength);
  const view = new DataView(wave.buffer);
  const text = (offset, value) => {
    for (let i = 0; i < value.length; ++i) wave[offset + i] = value.charCodeAt(i);
  };
  text(0, 'RIFF'); view.setUint32(4, wave.length - 8, true);
  text(8, 'WAVEfmt '); view.setUint32(16, 16, true);
  view.setUint16(20, 1, true); view.setUint16(22, VOICE_PCM.channels, true);
  view.setUint32(24, VOICE_PCM.sampleRate, true);
  view.setUint32(28, VOICE_PCM.sampleRate * VOICE_PCM.blockAlign, true);
  view.setUint16(32, VOICE_PCM.blockAlign, true);
  view.setUint16(34, VOICE_PCM.bitsPerSample, true);
  text(36, 'data'); view.setUint32(40, pcm.byteLength, true);
  wave.set(new Uint8Array(pcm), VOICE_PCM.headerBytes);
  return wave;
}

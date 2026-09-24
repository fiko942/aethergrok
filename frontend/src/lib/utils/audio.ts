/**
 * Dual-stage realistic mechanical camera shutter sound ("click-clack / cekrek")
 * synthesized via Web Audio API without requiring any external audio files.
 */
let sharedAudioCtx: AudioContext | null = null;

function getAudioContext(): AudioContext | null {
  if (typeof window === 'undefined') return null;
  const AudioCtx = window.AudioContext || (window as unknown as { webkitAudioContext: typeof AudioContext }).webkitAudioContext;
  if (!AudioCtx) return null;
  if (!sharedAudioCtx || sharedAudioCtx.state === 'closed') {
    try {
      sharedAudioCtx = new AudioCtx();
    } catch {
      return null;
    }
  }
  return sharedAudioCtx;
}

export function playCameraShutterSound(volume: number = 0.5): void {
  const ctx = getAudioContext();
  if (!ctx) return;

  if (ctx.state === 'suspended') {
    ctx.resume().catch(() => {});
  }

  try {
    const t0 = ctx.currentTime;
    const masterGain = ctx.createGain();
    masterGain.gain.setValueAtTime(Math.max(0, Math.min(1, volume)), t0);
    masterGain.connect(ctx.destination);

    // Stage 1: Quick high-frequency burst representing the camera mirror flip-up (t0 to t0+20ms)
    const osc1 = ctx.createOscillator();
    const gain1 = ctx.createGain();
    osc1.type = 'triangle';
    osc1.frequency.setValueAtTime(1100, t0);
    osc1.frequency.exponentialRampToValueAtTime(400, t0 + 0.02);

    gain1.gain.setValueAtTime(0.35, t0);
    gain1.gain.exponentialRampToValueAtTime(0.01, t0 + 0.02);

    osc1.connect(gain1);
    gain1.connect(masterGain);
    osc1.start(t0);
    osc1.stop(t0 + 0.025);

    // Stage 2: Resonant acoustic click & mechanical body release ("cekrek" clack, t1 to t1+65ms)
    const t1 = t0 + 0.045;
    const osc2 = ctx.createOscillator();
    const gain2 = ctx.createGain();
    osc2.type = 'sine';
    osc2.frequency.setValueAtTime(320, t1);
    osc2.frequency.exponentialRampToValueAtTime(90, t1 + 0.065);

    gain2.gain.setValueAtTime(0.5, t1);
    gain2.gain.exponentialRampToValueAtTime(0.01, t1 + 0.065);

    osc2.connect(gain2);
    gain2.connect(masterGain);
    osc2.start(t1);
    osc2.stop(t1 + 0.07);

    // Stage 3: Low metallic body reverberation (t1+10ms)
    const osc3 = ctx.createOscillator();
    const gain3 = ctx.createGain();
    osc3.type = 'sine';
    osc3.frequency.setValueAtTime(180, t1 + 0.01);
    osc3.frequency.exponentialRampToValueAtTime(60, t1 + 0.05);

    gain3.gain.setValueAtTime(0.2, t1 + 0.01);
    gain3.gain.exponentialRampToValueAtTime(0.01, t1 + 0.05);

    osc3.connect(gain3);
    gain3.connect(masterGain);
    osc3.start(t1 + 0.01);
    osc3.stop(t1 + 0.055);
  } catch {
    // Graceful silent fallback if audio is not supported in current environment
  }
}

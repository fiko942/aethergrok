/**
 * voiceRecorder.ts
 * Manages audio recording via Web Audio API / MediaRecorder,
 * handles macOS and browser permissions, automatic retry (up to 3x) on transcription failure,
 * network disconnection resilience (waiting for internet recovery),
 * and guaranteed cleanup of temporary audio files and sessions.
 */

export interface AudioInputDevice {
  deviceId: string;
  label: string;
  isDefault?: boolean;
  transport?: 'built-in' | 'bluetooth' | 'usb' | 'virtual' | 'continuity' | 'unknown';
  manufacturer?: string;
}

export type VoiceRecorderState =
  | 'idle'
  | 'checking_permission'
  | 'recording'
  | 'waiting_network'
  | 'transcribing'
  | 'error';

export interface VoiceTranscriptionProgress {
  state: VoiceRecorderState;
  attempt?: number;
  maxAttempts?: number;
  message?: string;
  elapsedSeconds?: number;
}

export class VoiceRecorderManager {
  private mediaRecorder: MediaRecorder | null = null;
  private audioStream: MediaStream | null = null;
  private audioChunks: Blob[] = [];
  private recordingTimer: number | null = null;
  private startTime: number = 0;
  private audioCtx: AudioContext | null = null;
  private analyser: AnalyserNode | null = null;
  private dataArray: Uint8Array<ArrayBuffer> | null = null;

  /**
   * Enumerate connected audio input devices (microphones) with dual-layer native & browser metadata
   */
  async getAudioInputDevices(): Promise<AudioInputDevice[]> {
    const win = (typeof window !== 'undefined' ? window : {}) as any;
    let nativeDevices: Array<{ name: string; isDefault: boolean; transport: string; manufacturer: string }> = [];

    if (win.go?.main?.App?.GetSystemAudioInputDevices) {
      try {
        const res = await win.go.main.App.GetSystemAudioInputDevices();
        if (Array.isArray(res)) {
          nativeDevices = res;
        }
      } catch (err) {
        console.warn('Failed to retrieve native audio input devices:', err);
      }
    }

    let webDevices: MediaDeviceInfo[] = [];
    if (typeof navigator !== 'undefined' && navigator.mediaDevices?.enumerateDevices) {
      try {
        const all = await navigator.mediaDevices.enumerateDevices();
        webDevices = all.filter((d) => d && d.kind === 'audioinput');
      } catch (err) {
        console.warn('Failed to enumerate web media devices:', err);
      }
    }

    if (webDevices.length > 0) {
      const results: AudioInputDevice[] = [];
      for (let i = 0; i < webDevices.length; i++) {
        const wd = webDevices[i];
        let label = wd.label || '';
        let matchedNative = nativeDevices.find((nd) => label && (nd.name.toLowerCase().includes(label.toLowerCase()) || label.toLowerCase().includes(nd.name.toLowerCase())));

        if (!matchedNative && nativeDevices[i]) {
          matchedNative = nativeDevices[i];
        }

        if (!label) {
          label = matchedNative?.name || `Microphone ${i + 1}`;
        }

        let transport = (matchedNative?.transport as any) || 'unknown';
        if (transport === 'unknown') {
          const l = label.toLowerCase();
          if (l.includes('realtek') || l.includes('array') || l.includes('built-in') || l.includes('internal')) transport = 'built-in';
          else if (l.includes('bluetooth') || l.includes('wireless') || l.includes('hands-free') || l.includes('airpods')) transport = 'bluetooth';
          else if (l.includes('usb') || l.includes('yeti') || l.includes('scarlett') || l.includes('hyperx')) transport = 'usb';
          else if (l.includes('virtual') || l.includes('voicemeeter') || l.includes('cable')) transport = 'virtual';
          else transport = 'built-in';
        }

        results.push({
          deviceId: wd.deviceId || (matchedNative?.isDefault ? 'default' : `dev-${i}`),
          label,
          isDefault: matchedNative?.isDefault || wd.deviceId === 'default' || i === 0,
          transport,
          manufacturer: matchedNative?.manufacturer || 'Windows Audio'
        });
      }
      return results;
    }

    if (nativeDevices.length > 0) {
      return nativeDevices.map((nd, idx) => ({
        deviceId: nd.isDefault ? 'default' : `native-dev-${idx}`,
        label: nd.name,
        isDefault: nd.isDefault,
        transport: (nd.transport as any) || 'built-in',
        manufacturer: nd.manufacturer || 'Windows Audio'
      }));
    }

    return [{
      deviceId: 'default',
      label: 'Default System Microphone',
      isDefault: true,
      transport: 'built-in',
      manufacturer: 'Windows Audio'
    }];
  }

  /**
   * Check microphone permission with native macOS AVFoundation bridge first, then navigator.permissions
   */
  async checkPermission(): Promise<{ granted: boolean; message: string }> {
    const win = (typeof window !== 'undefined' ? window : {}) as any;
    if (win.go?.main?.App?.CheckMicrophonePermission) {
      try {
        const status = await win.go.main.App.CheckMicrophonePermission();
        if (status && status.granted) {
          return { granted: true, message: status.message };
        }
      } catch {
        // Fallback to browser permissions API
      }
    }

    if (typeof navigator !== 'undefined' && navigator.permissions?.query) {
      try {
        const result = await navigator.permissions.query({ name: 'microphone' as PermissionName });
        if (result && result.state === 'granted') {
          return { granted: true, message: 'Microphone permission granted' };
        }
      } catch {
        // query not supported for microphone on some WebViews
      }
    }

    return { granted: false, message: 'Microphone permission not granted' };
  }

  /**
   * Request microphone permission (triggers macOS prompt or browser prompt)
   */
  async requestPermission(): Promise<boolean> {
    const win = (typeof window !== 'undefined' ? window : {}) as any;
    if (win.go?.main?.App?.RequestMicrophonePermission) {
      try {
        const status = await win.go.main.App.RequestMicrophonePermission();
        if (status && status.granted) return true;
      } catch {
        // Fallback
      }
    }

    if (typeof navigator !== 'undefined' && navigator.mediaDevices?.getUserMedia) {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
        if (stream) {
          // Immediately stop track after permission verification
          stream.getTracks().forEach((track) => track.stop());
          return true;
        }
      } catch {
        return false;
      }
    }

    return false;
  }

  /**
   * Start recording audio with selected input device (Instant 1-step initialization)
   */
  async startRecording(
    deviceId?: string,
    onProgress?: (progress: VoiceTranscriptionProgress) => void
  ): Promise<void> {
    if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
      try {
        this.mediaRecorder.stop();
      } catch {
        // ignore
      }
      this.cleanupStream();
    }

    if (typeof navigator === 'undefined' || !navigator.mediaDevices?.getUserMedia) {
      throw new Error('Audio recording is not supported in this environment');
    }

    onProgress?.({ state: 'checking_permission', message: 'Starting microphone...' });

    const constraints: MediaStreamConstraints = {
      audio: {
        ...(deviceId ? { deviceId: { exact: deviceId } } : {}),
        channelCount: 1, // Mono audio is much smaller and optimal for voice STT
        sampleRate: 16000, // 16kHz speech recognition standard
        echoCancellation: true,
        noiseSuppression: true,
        autoGainControl: true,
      },
    };

    try {
      this.audioStream = await navigator.mediaDevices.getUserMedia(constraints);
    } catch (err: any) {
      if (err?.name === 'NotAllowedError' || err?.name === 'PermissionDeniedError') {
        // Attempt native permission request prompt if available
        await this.requestPermission();
        throw new Error('Microphone permission denied. Please allow microphone access in System Settings.');
      }
      throw err;
    }

    this.audioChunks = [];

    // Select supported mimeType with optimized speech bitrate (24-32kbps)
    let mimeType = '';
    if (typeof MediaRecorder !== 'undefined') {
      if (typeof MediaRecorder.isTypeSupported === 'function') {
        if (MediaRecorder.isTypeSupported('audio/webm;codecs=opus')) {
          mimeType = 'audio/webm;codecs=opus';
        } else if (MediaRecorder.isTypeSupported('audio/mp4')) {
          mimeType = 'audio/mp4';
        } else if (MediaRecorder.isTypeSupported('audio/wav')) {
          mimeType = 'audio/wav';
        }
      }
    }

    if (typeof MediaRecorder === 'undefined') {
      throw new Error('MediaRecorder is not available');
    }

    const recorderOptions: MediaRecorderOptions = {
      ...(mimeType ? { mimeType } : {}),
      audioBitsPerSecond: 32000, // 32kbps opus mono (super lightweight & fast upload)
    };

    this.mediaRecorder = new MediaRecorder(this.audioStream, recorderOptions);

    this.mediaRecorder.ondataavailable = (event) => {
      if (event && event.data && event.data.size > 0) {
        this.audioChunks.push(event.data);
      }
    };

    // Initialize Web Audio API AnalyserNode for real-time 4-bar equalizer
    try {
      const AudioCtxClass = window.AudioContext || (window as any).webkitAudioContext;
      if (AudioCtxClass && this.audioStream) {
        this.audioCtx = new AudioCtxClass();
        const source = this.audioCtx.createMediaStreamSource(this.audioStream);
        this.analyser = this.audioCtx.createAnalyser();
        this.analyser.fftSize = 64;
        this.analyser.smoothingTimeConstant = 0.5;
        source.connect(this.analyser);
        this.dataArray = new Uint8Array(this.analyser.frequencyBinCount);
      }
    } catch (err) {
      console.warn('Failed to initialize voice analyzer:', err);
    }

    this.startTime = Date.now();
    this.mediaRecorder.start(250); // Slice every 250ms

    if (this.recordingTimer) {
      clearInterval(this.recordingTimer);
    }

    this.recordingTimer = window.setInterval(() => {
      const elapsedSeconds = Math.floor((Date.now() - this.startTime) / 1000);
      onProgress?.({
        state: 'recording',
        elapsedSeconds,
        message: 'Recording audio...',
      });
    }, 500);

    onProgress?.({
      state: 'recording',
      elapsedSeconds: 0,
      message: 'Recording audio...',
    });
  }

  /**
   * Stop recording and return audio Blob
   */
  async stopRecording(): Promise<{ blob: Blob; ext: string }> {
    if (this.recordingTimer) {
      clearInterval(this.recordingTimer);
      this.recordingTimer = null;
    }

    return new Promise((resolve) => {
      // If recorder hasn't started or is already inactive
      if (!this.mediaRecorder || this.mediaRecorder.state === 'inactive') {
        const mime = 'audio/webm';
        const ext = 'webm';
        const audioBlob = new Blob(this.audioChunks, { type: mime });
        this.cleanupStream();
        resolve({ blob: audioBlob, ext });
        return;
      }

      const mime = this.mediaRecorder.mimeType || 'audio/webm';
      const ext = mime.includes('mp4') ? 'mp4' : mime.includes('wav') ? 'wav' : 'webm';

      this.mediaRecorder.onstop = () => {
        const audioBlob = new Blob(this.audioChunks, { type: mime });
        this.cleanupStream();
        resolve({ blob: audioBlob, ext });
      };

      try {
        // Request any remaining audio data slice before stopping
        if (typeof (this.mediaRecorder as any).requestData === 'function') {
          try {
            (this.mediaRecorder as any).requestData();
          } catch {
            // ignore
          }
        }
        this.mediaRecorder.stop();
      } catch (err) {
        this.cleanupStream();
        resolve({ blob: new Blob(this.audioChunks, { type: mime }), ext });
      }
    });
  }

  /**
   * Cancel recording without returning audio
   */
  cancelRecording(): void {
    if (this.recordingTimer) {
      clearInterval(this.recordingTimer);
      this.recordingTimer = null;
    }
    if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
      try {
        this.mediaRecorder.stop();
      } catch {
        // ignore
      }
    }
    this.cleanupStream();
  }

  private cleanupStream(): void {
    if (this.audioStream) {
      this.audioStream.getTracks().forEach((track) => track.stop());
      this.audioStream = null;
    }
    if (this.audioCtx) {
      try {
        this.audioCtx.close();
      } catch {
        // ignore
      }
      this.audioCtx = null;
    }
    this.analyser = null;
    this.dataArray = null;
    this.mediaRecorder = null;
    this.audioChunks = [];
  }

  /**
   * Samples current audio stream and returns 4 normalized frequency bar heights (0.0 to 1.0)
   */
  getVolumeLevels(): [number, number, number, number] {
    if (!this.analyser || !this.dataArray) {
      return [0.15, 0.2, 0.2, 0.15];
    }

    try {
      this.analyser.getByteFrequencyData(this.dataArray as any);
      const bins = this.dataArray;
      const binCount = bins.length; // 32 bins with fftSize 64

      // Divide into 4 bands: Sub/Bass, Low-Mid Voice, Core Voice, High Presence
      const q = Math.floor(binCount / 4);
      const b1 = this.avgBand(bins, 0, q);
      const b2 = this.avgBand(bins, q, q * 2);
      const b3 = this.avgBand(bins, q * 2, q * 3);
      const b4 = this.avgBand(bins, q * 3, binCount);

      return [
        Math.max(0.1, Math.min(1.0, b1 / 180)),
        Math.max(0.15, Math.min(1.0, b2 / 160)),
        Math.max(0.15, Math.min(1.0, b3 / 160)),
        Math.max(0.1, Math.min(1.0, b4 / 180)),
      ];
    } catch {
      return [0.15, 0.2, 0.2, 0.15];
    }
  }

  private avgBand(bins: Uint8Array, start: number, end: number): number {
    let sum = 0;
    const len = Math.max(1, end - start);
    for (let i = start; i < end && i < bins.length; i++) {
      sum += bins[i];
    }
    return sum / len;
  }

  /**
   * Convert Blob to Base64 data string
   */
  private blobToBase64(blob: Blob): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onloadend = () => resolve(reader.result as string);
      reader.onerror = reject;
      reader.readAsDataURL(blob);
    });
  }

  /**
   * Wait until internet connection is recovered if offline
   */
  private waitForInternet(onProgress?: (progress: VoiceTranscriptionProgress) => void): Promise<void> {
    if (navigator.onLine) {
      return Promise.resolve();
    }

    return new Promise((resolve) => {
      onProgress?.({
        state: 'waiting_network',
        message: 'No internet connection. Waiting for network recovery...',
      });

      const handleOnline = () => {
        window.removeEventListener('online', handleOnline);
        resolve();
      };
      window.addEventListener('online', handleOnline);
    });
  }

  /**
   * Transcribe recorded audio with auto-retry (3x), network resilience, and guaranteed file cleanup.
   */
  async transcribeAudioBlob(
    blob: Blob,
    ext: string,
    workspacePath: string,
    onProgress?: (progress: VoiceTranscriptionProgress) => void
  ): Promise<string> {
    const win = window as any;
    if (!win.go?.main?.App?.SaveVoiceAudioRecording || !win.go?.main?.App?.TranscribeAudioWithGrok) {
      throw new Error('Backend transcription bridge is not available');
    }

    // 1. Validate audio blob before attempting transcription
    if (!blob || blob.size < 64) {
      throw new Error('Recording was too short or contained no audio');
    }

    // 2. Convert to Base64
    const base64Data = await this.blobToBase64(blob);

    // 2. Save scratch audio recording file
    const audioFilePath: string = await win.go.main.App.SaveVoiceAudioRecording(base64Data, ext);

    const maxAttempts = 3;
    let lastError: Error | null = null;

    try {
      for (let attempt = 1; attempt <= maxAttempts; attempt++) {
        try {
          // Check network status before attempt
          await this.waitForInternet(onProgress);

          onProgress?.({
            state: 'transcribing',
            attempt,
            maxAttempts,
            message: `Transcribing voice with Grok (Attempt ${attempt}/${maxAttempts})...`,
          });

          const transcript: string = await win.go.main.App.TranscribeAudioWithGrok(workspacePath, audioFilePath);
          if (transcript && transcript.trim()) {
            return transcript.trim();
          }
          throw new Error('Empty transcript returned from Grok');
        } catch (err: any) {
          lastError = err instanceof Error ? err : new Error(String(err));
          // If internet was lost during request, loop will catch wait on next attempt
          if (attempt < maxAttempts) {
            await new Promise((res) => setTimeout(res, attempt * 1200));
          }
        }
      }

      throw lastError || new Error('Failed to transcribe audio after 3 attempts');
    } finally {
      // Guaranteed Cleanup: Ensure scratch audio recording file is deleted on disk
      try {
        if (win.go?.main?.App?.DeleteVoiceAudioRecording) {
          await win.go.main.App.DeleteVoiceAudioRecording(audioFilePath);
        }
      } catch {
        // Ignore deletion errors in finally
      }
    }
  }
}

export const voiceRecorder = new VoiceRecorderManager();

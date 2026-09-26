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

  /**
   * Enumerate connected audio input devices (microphones)
   */
  async getAudioInputDevices(): Promise<AudioInputDevice[]> {
    if (typeof navigator === 'undefined' || !navigator.mediaDevices?.enumerateDevices) {
      return [];
    }
    try {
      const devices = await navigator.mediaDevices.enumerateDevices();
      return devices
        .filter((d) => d && d.kind === 'audioinput')
        .map((d, index) => ({
          deviceId: d.deviceId || '',
          label: d.label || `Microphone ${index + 1}`,
        }));
    } catch {
      return [];
    }
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
   * Start recording audio with selected input device
   */
  async startRecording(
    deviceId?: string,
    onProgress?: (progress: VoiceTranscriptionProgress) => void
  ): Promise<void> {
    if (this.mediaRecorder && this.mediaRecorder.state !== 'inactive') {
      this.stopRecording();
    }

    if (typeof navigator === 'undefined' || !navigator.mediaDevices?.getUserMedia) {
      throw new Error('Audio recording is not supported in this environment');
    }

    onProgress?.({ state: 'checking_permission', message: 'Verifying microphone access...' });

    const hasPermission = await this.requestPermission();
    if (!hasPermission) {
      throw new Error('Microphone permission denied. Please allow microphone access in System Settings.');
    }

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

    this.audioStream = await navigator.mediaDevices.getUserMedia(constraints);
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

    return new Promise((resolve, reject) => {
      if (!this.mediaRecorder || this.mediaRecorder.state === 'inactive') {
        this.cleanupStream();
        reject(new Error('No active audio recording session'));
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
        this.mediaRecorder.stop();
      } catch (err) {
        this.cleanupStream();
        reject(err);
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
    this.mediaRecorder = null;
    this.audioChunks = [];
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

    // 1. Convert to Base64
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

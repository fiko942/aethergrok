import { logger } from './logger.svelte';
import type { ReleaseInfo, ReleaseAsset, UpdateCheckResult, UpdateProgress } from '../../app';

export class UpdaterStore {
  checking = $state<boolean>(false);
  updateAvailable = $state<boolean>(false);
  currentVersion = $state<string>(typeof __APP_VERSION__ !== 'undefined' ? __APP_VERSION__ : '1.0.1');
  latestVersion = $state<string>('1.0.1');
  latestRelease = $state<ReleaseInfo | null>(null);
  allReleases = $state<ReleaseInfo[]>([]);
  matchedAsset = $state<ReleaseAsset | null>(null);
  lastChecked = $state<string | null>(null);
  error = $state<string | null>(null);

  // Download & Installation state
  isInstalling = $state<boolean>(false);
  installProgress = $state<UpdateProgress>({
    downloadedBytes: 0,
    totalBytes: 0,
    percent: 0,
    speedFormatted: '0 KB/s',
    stage: 'idle',
    message: ''
  });
  installError = $state<string | null>(null);

  private intervalId: any = null;
  private unlistenProgress: (() => void) | null = null;

  constructor() {
    if (typeof __APP_VERSION__ !== 'undefined' && __APP_VERSION__) {
      this.currentVersion = __APP_VERSION__;
      this.latestVersion = __APP_VERSION__;
    }
    this.setupEventListener();
  }

  private setupEventListener(): void {
    if (typeof window !== 'undefined' && (window as any).runtime?.EventsOn) {
      this.unlistenProgress = (window as any).runtime.EventsOn('updater:progress', (progress: UpdateProgress) => {
        if (!progress) return;
        this.installProgress = { ...progress };
        if (progress.stage === 'error') {
          this.isInstalling = false;
          this.installError = progress.message;
          logger.error('UPDATER', `Update error: ${progress.message}`);
        } else if (progress.stage === 'ready') {
          this.isInstalling = true;
          logger.info('UPDATER', `Update completed. System restarting...`);
        } else if (progress.stage === 'downloading' || progress.stage === 'verifying' || progress.stage === 'installing') {
          this.isInstalling = true;
          this.installError = null;
        }
      });
    }
  }

  async checkForUpdates(silent: boolean = false): Promise<UpdateCheckResult | null> {
    if (this.checking) {
      return null;
    }

    this.checking = true;
    this.error = null;

    if (!silent) {
      logger.info('SYSTEM', `Checking for updates (current version: v${this.currentVersion})...`);
    }

    try {
      const win = typeof window !== 'undefined' ? (window as any) : null;
      if (!win?.go?.main?.App?.CheckForUpdates) {
        throw new Error('CheckForUpdates IPC bridge not available');
      }

      const result: UpdateCheckResult = await win.go.main.App.CheckForUpdates(this.currentVersion);
      
      this.updateAvailable = !!result.updateAvailable;
      this.currentVersion = result.currentVersion || this.currentVersion;
      this.latestVersion = result.latestVersion || this.latestVersion;
      this.latestRelease = result.latestRelease || null;
      this.allReleases = result.allReleases || [];
      this.matchedAsset = result.platformAsset || null;
      this.lastChecked = result.checkedAt || new Date().toISOString();
      this.checking = false;

      if (this.updateAvailable) {
        logger.info('SYSTEM', `Update available: v${this.latestVersion} (installed: v${this.currentVersion})`, {
          releaseTitle: this.latestRelease?.title,
          assetName: this.matchedAsset?.name
        });
      } else if (!silent) {
        logger.info('SYSTEM', `App is up-to-date (v${this.currentVersion})`);
      }

      return result;
    } catch (err: any) {
      const errMsg = err?.message || String(err);
      this.error = errMsg;
      this.checking = false;
      this.lastChecked = new Date().toISOString();
      logger.error('SYSTEM', `Update check failed: ${errMsg}`, err);
      return null;
    }
  }

  async fetchChangelogHistory(): Promise<ReleaseInfo[]> {
    try {
      const win = typeof window !== 'undefined' ? (window as any) : null;
      if (!win?.go?.main?.App?.GetChangelogHistory) {
        throw new Error('GetChangelogHistory IPC bridge not available');
      }

      const releases: ReleaseInfo[] = await win.go.main.App.GetChangelogHistory();
      if (Array.isArray(releases) && releases.length > 0) {
        this.allReleases = releases;
        if (!this.latestRelease && releases[0]) {
          this.latestRelease = releases[0];
          this.latestVersion = releases[0].version;
        }
      }
      return this.allReleases;
    } catch (err: any) {
      const errMsg = err?.message || String(err);
      logger.error('SYSTEM', `Failed to fetch changelog history: ${errMsg}`, err);
      return [];
    }
  }

  async startDownloadAndInstall(customAssetUrl?: string): Promise<void> {
    const targetUrl = customAssetUrl || this.matchedAsset?.downloadUrl;
    if (!targetUrl) {
      this.installError = 'No downloadable binary asset found for your platform';
      return;
    }

    this.isInstalling = true;
    this.installError = null;
    this.installProgress = {
      downloadedBytes: 0,
      totalBytes: this.matchedAsset?.size || 0,
      percent: 0,
      speedFormatted: '0 KB/s',
      stage: 'downloading',
      message: 'Initiating download...'
    };

    logger.info('UPDATER', `Starting in-app download and installation: ${targetUrl}`);

    try {
      const win = typeof window !== 'undefined' ? (window as any) : null;
      if (!win?.go?.main?.App?.DownloadAndInstallUpdate) {
        throw new Error('DownloadAndInstallUpdate IPC bridge not available');
      }

      // If a checksum asset exists (.sha256 or .sha256sum) in assets, extract its URL
      let sha256Url = '';
      if (this.latestRelease?.assets) {
        const shaAsset = this.latestRelease.assets.find(a => a.name.endsWith('.sha256') || a.name.endsWith('.sha256sum') || a.name.includes('checksum'));
        if (shaAsset) {
          sha256Url = shaAsset.downloadUrl;
        }
      }

      await win.go.main.App.DownloadAndInstallUpdate(targetUrl, sha256Url);
    } catch (err: any) {
      const errMsg = err?.message || String(err);
      this.isInstalling = false;
      this.installError = errMsg;
      logger.error('UPDATER', `In-app update failed: ${errMsg}`, err);
    }
  }

  async cancelDownload(): Promise<void> {
    try {
      const win = typeof window !== 'undefined' ? (window as any) : null;
      if (win?.go?.main?.App?.CancelUpdateDownload) {
        await win.go.main.App.CancelUpdateDownload();
      }
      this.isInstalling = false;
      this.installProgress = {
        downloadedBytes: 0,
        totalBytes: 0,
        percent: 0,
        speedFormatted: '0 KB/s',
        stage: 'idle',
        message: 'Download cancelled'
      };
      logger.info('UPDATER', 'Download was cancelled by user');
    } catch (err: any) {
      logger.error('UPDATER', `Failed to cancel download: ${err?.message || err}`, err);
    }
  }

  async openDownload(url?: string): Promise<void> {
    const targetUrl = url || this.matchedAsset?.downloadUrl || (this.latestRelease ? `https://github.com/fiko942/grok-build/releases/tag/${this.latestRelease.tagName}` : undefined);
    
    if (!targetUrl) {
      logger.warn('UI', 'No download URL available to open');
      return;
    }

    try {
      const win = typeof window !== 'undefined' ? (window as any) : null;
      if (win?.go?.main?.App?.OpenExternalURL) {
        await win.go.main.App.OpenExternalURL(targetUrl);
      } else if (win?.runtime?.BrowserOpenURL) {
        win.runtime.BrowserOpenURL(targetUrl);
      } else if (typeof window !== 'undefined') {
        window.open(targetUrl, '_blank');
      }
      logger.info('UI', `Opened download URL: ${targetUrl}`);
    } catch (err: any) {
      logger.error('UI', `Failed to open download URL: ${err?.message || err}`, err);
    }
  }

  initPeriodicCheck(intervalMs: number = 3600000): void {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }

    // Run initial silent check after 5 seconds delay so app startup is unaffected
    if (typeof window !== 'undefined') {
      setTimeout(() => {
        this.checkForUpdates(true).catch(() => {});
      }, 5000);
    }

    this.intervalId = setInterval(() => {
      this.checkForUpdates(true).catch(() => {});
    }, intervalMs);
  }
}

export const updaterStore = new UpdaterStore();

import { writable } from 'svelte/store';

export interface DownloadModalState {
  isOpen: boolean;
  platform: 'macOS' | 'Windows';
  filename: string;
  downloadUrl: string;
}

export const downloadModalStore = writable<DownloadModalState>({
  isOpen: false,
  platform: 'macOS',
  filename: '',
  downloadUrl: '',
});

export function openDownloadModal(platform: 'macOS' | 'Windows', filename: string, downloadUrl: string) {
  // Trigger file download in browser
  if (downloadUrl && downloadUrl !== '#') {
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.download = filename || '';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }

  downloadModalStore.set({
    isOpen: true,
    platform,
    filename,
    downloadUrl,
  });
}

export function closeDownloadModal() {
  downloadModalStore.update((s) => ({ ...s, isOpen: false }));
}

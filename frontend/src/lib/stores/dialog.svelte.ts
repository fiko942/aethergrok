import { type WorkspaceFolder, type Session } from './session.svelte';

export interface ConfirmDialogOptions {
  title: string;
  content: string;
  confirmText?: string;
  cancelText?: string;
  type?: 'danger' | 'warning' | 'info';
  onConfirm: () => void;
  onCancel?: () => void;
}

export class DialogStore {
  confirmState = $state<{
    open: boolean;
    title: string;
    content: string;
    confirmText: string;
    cancelText: string;
    type: 'danger' | 'warning' | 'info';
    onConfirm: () => void;
    onCancel: () => void;
  }>({
    open: false,
    title: '',
    content: '',
    confirmText: 'Confirm',
    cancelText: 'Cancel',
    type: 'danger',
    onConfirm: () => {},
    onCancel: () => {}
  });

  openConfirm(options: ConfirmDialogOptions): void {
    this.confirmState = {
      open: true,
      title: options.title,
      content: options.content,
      confirmText: options.confirmText || 'Confirm',
      cancelText: options.cancelText || 'Cancel',
      type: options.type || 'danger',
      onConfirm: () => {
        this.confirmState.open = false;
        options.onConfirm();
      },
      onCancel: () => {
        this.confirmState.open = false;
        if (options.onCancel) options.onCancel();
      }
    };
  }

  closeConfirm(): void {
    this.confirmState.open = false;
  }
}

export const dialogStore = new DialogStore();

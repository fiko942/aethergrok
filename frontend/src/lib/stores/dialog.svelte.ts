import { type WorkspaceFolder, type Session, type ToolCall } from './session.svelte';
import { type SubagentParsedMeta, extractSubagentMeta } from '../utils/toolUtils';

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
  fileViewerState = $state<{
    open: boolean;
    filePath: string;
    workspacePath: string;
  }>({
    open: false,
    filePath: '',
    workspacePath: ''
  });

  openFileViewer(filePath: string, workspacePath: string = ''): void {
    this.fileViewerState = {
      open: true,
      filePath,
      workspacePath
    };
  }

  closeFileViewer(): void {
    this.fileViewerState.open = false;
  }

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

  subagentModalState = $state<{
    open: boolean;
    toolCall: ToolCall | null;
    meta: SubagentParsedMeta | null;
  }>({
    open: false,
    toolCall: null,
    meta: null
  });

  openSubagentModal(toolCall: ToolCall, meta?: SubagentParsedMeta): void {
    const safeMeta = meta || extractSubagentMeta(toolCall.params, toolCall.result);
    this.subagentModalState = {
      open: true,
      toolCall,
      meta: safeMeta
    };
  }

  closeSubagentModal(): void {
    this.subagentModalState = {
      open: false,
      toolCall: null,
      meta: null
    };
  }
}

export const dialogStore = new DialogStore();

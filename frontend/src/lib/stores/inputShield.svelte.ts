/**
 * inputShield.svelte.ts
 *
 * Global UI shield state manager that controls temporary read-only state for:
 * - Text inputs & textareas (Composer prompt box, file search, skill search, etc.)
 * - Terminals (xterm instances stdin disabling)
 *
 * When screenshot or dictation triggers are pressed/held or running,
 * inputShield guarantees that all inputs and terminals become read-only,
 * preventing accidental keystrokes, stray characters (like `\`), or shell interference.
 *
 * Designed with multi-layered safety mechanisms against getting permanently stuck.
 */

export class InputShieldStore {
  // Set of active lock reasons, e.g. 'snapshot_key_hold', 'snapshot_capturing', 'dictation_key_hold', 'dictation_recording', 'dictation_transcribing'
  private activeLocks = $state<Set<string>>(new Set());

  // Reactive boolean: true if ANY lock is currently active
  get isReadOnly(): boolean {
    return this.activeLocks.size > 0;
  }

  // Safety timers per lock reason to avoid orphaned locks
  private lockTimers: Map<string, any> = new Map();

  /**
   * Acquire a read-only lock for a specific reason with an optional auto-expiry safety timeout.
   */
  acquireLock(reason: string, timeoutMs: number = 10000): void {
    if (!reason) return;
    const updated = new Set(this.activeLocks);
    updated.add(reason);
    this.activeLocks = updated;

    // Clear existing timer for this reason if any
    if (this.lockTimers.has(reason)) {
      clearTimeout(this.lockTimers.get(reason));
      this.lockTimers.delete(reason);
    }

    // Set safety auto-release timeout
    if (timeoutMs > 0) {
      const timer = setTimeout(() => {
        this.releaseLock(reason);
      }, timeoutMs);
      this.lockTimers.set(reason, timer);
    }
  }

  /**
   * Release a specific read-only lock.
   */
  releaseLock(reason: string): void {
    if (!reason) return;
    if (this.lockTimers.has(reason)) {
      clearTimeout(this.lockTimers.get(reason));
      this.lockTimers.delete(reason);
    }

    if (this.activeLocks.has(reason)) {
      const updated = new Set(this.activeLocks);
      updated.delete(reason);
      this.activeLocks = updated;
    }
  }

  /**
   * Emergency reset: clears all active locks.
   * Called on window blur, pointer click recovery, or escape.
   */
  resetAll(): void {
    for (const timer of this.lockTimers.values()) {
      clearTimeout(timer);
    }
    this.lockTimers.clear();
    this.activeLocks = new Set();
  }

  /**
   * Checks whether a specific lock reason is currently held.
   */
  hasLock(reason: string): boolean {
    return this.activeLocks.has(reason);
  }

  /**
   * Gets list of all currently active lock reasons for debugging/diagnostics.
   */
  getActiveLockReasons(): string[] {
    return Array.from(this.activeLocks);
  }
}

export const inputShieldStore = new InputShieldStore();

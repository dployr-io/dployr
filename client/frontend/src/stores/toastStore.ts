import { writable } from 'svelte/store';

export type ToastType = 'success' | 'info' | 'warning' | 'error';

export interface Toast {
  id: number;
  message: string;
  type: ToastType;
}

// Internal writable store
const _toasts = writable<Toast[]>([]);
let counter = 0;

/**
 * Show a new toast.
 * @param message The text to display.
 * @param type    One of 'success' | 'info' | 'warning' | 'error'.
 * @param duration How long before auto‑dismiss (ms).
 */
export function addToast(
  message: string,
  type: ToastType = 'info',
  duration = 4000
) {
  const id = ++counter;
  _toasts.update((all) => [{ id, message, type }, ...all]);

  // Auto‑remove after the given duration
  setTimeout(() => removeToast(id), duration);
}

/** Dismiss a toast early by its ID */
export function removeToast(id: number) {
  _toasts.update((all) => all.filter((t) => t.id !== id));
}

/** Public, read‑only store for your components to subscribe to */
export const toasts = {
  subscribe: _toasts.subscribe
};


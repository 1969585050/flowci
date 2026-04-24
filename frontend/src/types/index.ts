/**
 * Frontend-internal types. Types that travel over IPC should come from
 * wailsjs/go/models (auto-generated). This file is for purely UI state.
 */

export type ThemeMode = 'dark' | 'light' | 'system'

export interface ToastItem {
  id: number
  type: 'success' | 'error' | 'info' | 'warning'
  message: string
}

export interface ConfirmOptions {
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  variant?: 'default' | 'danger'
}

/** Async operation states for UI rendering. */
export type AsyncState<T> =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: string }

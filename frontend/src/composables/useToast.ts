import { ref, readonly } from 'vue'
import type { ToastItem } from '../types'

/**
 * Global toast store. Mounted once by a top-level <ToastHost /> component.
 * Anywhere in the app, call useToast().success(msg) / error(msg) / etc.
 */

const toasts = ref<ToastItem[]>([])
let nextId = 1

function push(type: ToastItem['type'], message: string, durationMs = 3000) {
  const id = nextId++
  toasts.value.push({ id, type, message })
  if (durationMs > 0) {
    window.setTimeout(() => dismiss(id), durationMs)
  }
  return id
}

function dismiss(id: number) {
  toasts.value = toasts.value.filter((t) => t.id !== id)
}

export function useToast() {
  return {
    success: (msg: string, ms?: number) => push('success', msg, ms),
    error:   (msg: string, ms?: number) => push('error',   msg, ms ?? 5000),
    warning: (msg: string, ms?: number) => push('warning', msg, ms ?? 4000),
    info:    (msg: string, ms?: number) => push('info',    msg, ms),
    dismiss,
    toasts: readonly(toasts),
  }
}

/**
 * Convenience wrapper: turn a rejected Promise's reason into a user-visible
 * error toast with a helpful prefix. Returns the same promise for chaining.
 */
export async function withErrorToast<T>(
  p: Promise<T>,
  prefix: string,
): Promise<T> {
  try {
    return await p
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    useToast().error(`${prefix}: ${msg}`)
    throw e
  }
}

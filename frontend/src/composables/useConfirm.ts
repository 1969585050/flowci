import { ref, readonly } from 'vue'
import type { ConfirmOptions } from '../types'

/**
 * Programmatic replacement for window.confirm().
 * A top-level <ConfirmDialog /> component watches `current` and resolves the
 * in-flight promise when user clicks confirm/cancel.
 *
 *   const ok = await useConfirm().ask({title:'删除', message:'确定吗？', variant:'danger'})
 *   if (ok) doDelete()
 */

interface PendingConfirm {
  options: ConfirmOptions
  resolve: (ok: boolean) => void
}

const current = ref<PendingConfirm | null>(null)

function ask(options: ConfirmOptions): Promise<boolean> {
  return new Promise((resolve) => {
    current.value = { options, resolve }
  })
}

function answer(ok: boolean) {
  if (!current.value) return
  current.value.resolve(ok)
  current.value = null
}

export function useConfirm() {
  return {
    ask,
    current: readonly(current),
    /** Internal — called by ConfirmDialog.vue when user chooses. */
    _answer: answer,
  }
}

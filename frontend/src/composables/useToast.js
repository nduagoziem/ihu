import { reactive } from 'vue'

const toasts = reactive([])
let counter = 0

export function useToast() {
  function notify(message, type = 'error') {
    const id = ++counter
    toasts.push({ id, message, type })
    setTimeout(() => dismiss(id), 5000)
  }
  function dismiss(id) {
    const i = toasts.findIndex((t) => t.id === id)
    if (i >= 0) toasts.splice(i, 1)
  }
  return { toasts, notify, dismiss }
}

import { ref } from 'vue'

const toasts = ref([])
const activeIds = new Set()
let nextId = 0

export function useToasts() {
  function show(toastId, title, msg, type, duration = 5000) {
    if (activeIds.has(toastId)) return
    activeIds.add(toastId)
    const id = nextId++
    toasts.value.push({ id, toastId, title, msg, type, dismissing: false })
    setTimeout(() => dismiss(id), duration)
  }

  function dismiss(id) {
    const t = toasts.value.find(t => t.id === id)
    if (!t || t.dismissing) return
    t.dismissing = true
    setTimeout(() => {
      const idx = toasts.value.findIndex(t => t.id === id)
      if (idx !== -1) {
        activeIds.delete(toasts.value[idx].toastId)
        toasts.value.splice(idx, 1)
      }
    }, 260)
  }

  return { toasts, show, dismiss }
}

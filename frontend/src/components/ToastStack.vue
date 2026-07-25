<script setup>
import { X } from '@lucide/vue'
import { useToast } from '../composables/useToast'

const { toasts, dismiss } = useToast()
</script>

<template>
  <div class="toast-stack">
    <Transition-group name="toast">
      <div v-for="t in toasts" :key="t.id" class="toast" :class="'toast--' + t.type">
        <span class="toast__msg">{{ t.message }}</span>
        <button class="toast__close" @click="dismiss(t.id)"><X :size="14" /></button>
      </div>
    </Transition-group>
  </div>
</template>

<style scoped>
.toast-stack {
  position: fixed;
  top: 70px;
  right: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  z-index: 200;
  pointer-events: none;
}
.toast {
  pointer-events: auto;
  display: flex;
  align-items: center;
  gap: 12px;
  max-width: 380px;
  padding: 12px 14px;
  border-radius: var(--r-md);
  background: var(--surface-3);
  border: 1px solid var(--surface-5);
  box-shadow: var(--shadow-lg);
  animation: scaleIn var(--t-med) var(--ease-out);
}
.toast--error { border-color: rgba(251, 113, 133, 0.4); }
.toast--success { border-color: rgba(52, 211, 153, 0.4); }
.toast__msg { font-size: 13px; color: var(--text-primary); flex: 1; line-height: 1.4; }
.toast__close {
  width: 24px; height: 24px; border-radius: var(--r-sm);
  border: none; background: transparent; color: var(--text-muted);
  display: grid; place-items: center; flex-shrink: 0;
}
.toast__close:hover { background: rgba(255,255,255,0.08); color: var(--text-primary); }

.toast-enter-active, .toast-leave-active {
  transition: all var(--t-med) var(--ease);
}
.toast-enter-from { opacity: 0; transform: translateX(40px); }
.toast-leave-to { opacity: 0; transform: translateX(40px); }
</style>

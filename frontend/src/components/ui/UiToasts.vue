<script setup lang="ts">
import { dismissToast, toasts } from '@/composables/useToast'
</script>

<template>
  <TransitionGroup tag="div" name="toast" class="toasts" aria-live="polite">
    <button v-for="t in toasts" :key="t.id" type="button" :class="['toast', t.tone]" @click="dismissToast(t.id)">
      {{ t.text }}
    </button>
  </TransitionGroup>
</template>

<style scoped>
.toasts {
  position: fixed;
  right: 16px;
  bottom: 16px;
  left: 16px;
  z-index: 100;
  display: grid;
  gap: 8px;
  justify-items: end;
  pointer-events: none;
}

.toast {
  max-width: 420px;
  padding: 12px 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-left: 4px solid var(--success);
  border-radius: var(--radius);
  box-shadow: var(--shadow-lg);
  text-align: left;
  pointer-events: auto;
  cursor: pointer;
}

.toast.danger {
  border-left-color: var(--danger);
}

.toast-enter-active,
.toast-leave-active {
  transition:
    opacity 0.2s,
    transform 0.2s;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(8px);
}
</style>

<script setup lang="ts">
import UiSpinner from './UiSpinner.vue'

const {
  variant = 'secondary',
  size = 'md',
  type = 'button',
  loading = false,
  disabled = false,
  square = false,
} = defineProps<{
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md'
  type?: 'button' | 'submit'
  loading?: boolean
  disabled?: boolean
  square?: boolean
}>()
</script>

<template>
  <button :type="type" :class="['btn', variant, size, { square }]" :disabled="disabled || loading">
    <UiSpinner v-if="loading" small />
    <slot />
  </button>
</template>

<style scoped>
.btn {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  justify-content: center;
  min-height: 38px;
  padding: 0 16px;
  border: 1px solid transparent;
  border-radius: var(--radius-sm);
  font-weight: 500;
  white-space: nowrap;
  cursor: pointer;
  transition:
    background-color 0.15s,
    border-color 0.15s,
    opacity 0.15s;
}

.btn:disabled {
  cursor: default;
  opacity: 0.55;
}

.sm {
  min-height: 30px;
  padding: 0 10px;
  font-size: 0.875rem;
}

.square {
  width: 38px;
  padding: 0;
}

.sm.square {
  width: 30px;
}

@media (pointer: coarse) {
  .btn {
    min-height: 44px;
  }

  .sm {
    min-height: 36px;
  }

  .square {
    width: 44px;
  }

  .sm.square {
    width: 36px;
  }
}

.primary {
  background: var(--accent);
  color: var(--on-accent);
}

.primary:not(:disabled):hover {
  background: var(--accent-hover);
}

.secondary {
  background: var(--surface);
  border-color: var(--border);
}

.secondary:not(:disabled):hover {
  background: var(--surface-muted);
}

.ghost {
  background: transparent;
  color: var(--text-muted);
}

.ghost:not(:disabled):hover {
  background: var(--surface-muted);
  color: var(--text);
}

.danger {
  background: var(--danger-soft);
  color: var(--danger);
}

.danger:not(:disabled):hover {
  border-color: var(--danger);
}
</style>

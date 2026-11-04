<script setup lang="ts">
const { paid } = defineProps<{ paid: boolean }>()
const emit = defineEmits<{ toggle: [] }>()
</script>

<template>
  <button type="button" :class="['paid', { on: paid }]" :aria-pressed="paid" @click="emit('toggle')">
    {{ paid ? '✓ оплачено' : 'не оплачено' }}
  </button>
</template>

<style scoped>
.paid {
  padding: 2px 10px;
  color: var(--text-muted);
  background: var(--surface-muted);
  border: 1px solid var(--border);
  border-radius: 999px;
  font: inherit;
  font-size: 0.75rem;
  font-weight: 500;
  white-space: nowrap;
  cursor: pointer;
}

.paid.on {
  color: var(--success);
  background: var(--success-soft);
  border-color: transparent;
}
</style>

<script setup lang="ts">
const { tone = 'neutral' } = defineProps<{
  tone?: 'neutral' | 'accent' | 'success' | 'danger' | 'warning'
}>()
</script>

<template>
  <span :class="['badge', tone]"><slot /></span>
</template>

<style scoped>
.badge {
  display: inline-flex;
  gap: 4px;
  align-items: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 0.8125rem;
  font-weight: 500;
  white-space: nowrap;
}

.neutral {
  background: var(--surface-muted);
  color: var(--text-muted);
}

.accent {
  background: var(--accent-soft);
  color: var(--accent);
}

.success {
  background: var(--success-soft);
  color: var(--success);
}

.danger {
  background: var(--danger-soft);
  color: var(--danger);
}

.warning {
  background: var(--warning-soft);
  color: var(--warning);
}
</style>

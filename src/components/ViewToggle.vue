<script setup lang="ts">
import { viewMode, type ViewMode } from '@/composables/useViewMode'

const options: { value: ViewMode; label: string; icon: string }[] = [
  { value: 'calendar', label: 'Календари', icon: 'M4 6h16v14H4zM4 10h16M8 3v4m8-4v4' },
  { value: 'table', label: 'Таблица', icon: 'M4 5h16v14H4zM4 10h16M4 14.5h16M10 5v14' },
]
</script>

<template>
  <div class="toggle" role="radiogroup" aria-label="Вид">
    <button
      v-for="o in options"
      :key="o.value"
      type="button"
      role="radio"
      :aria-checked="viewMode === o.value"
      :class="['option', { active: viewMode === o.value }]"
      @click="viewMode = o.value"
    >
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path :d="o.icon" />
      </svg>
      {{ o.label }}
    </button>
  </div>
</template>

<style scoped>
.toggle {
  display: inline-flex;
  gap: 2px;
  padding: 2px;
  background: var(--surface-muted);
  border-radius: var(--radius-sm);
}

.option {
  display: inline-flex;
  gap: 6px;
  align-items: center;
  padding: 5px 12px;
  color: var(--text-muted);
  background: none;
  border: 0;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
}

.option:hover {
  color: var(--text);
}

.option.active {
  color: var(--text);
  background: var(--surface);
  box-shadow: var(--shadow);
}
</style>

<script setup lang="ts">
import { theme, type Theme } from '@/composables/useTheme'

const options: { value: Theme; label: string; icon: string }[] = [
  { value: 'light', label: 'Светлая тема', icon: 'M12 4V2m0 20v-2m8-8h2M2 12h2m13.66-5.66 1.41-1.41M4.93 19.07l1.41-1.41m0-11.32L4.93 4.93m14.14 14.14-1.41-1.41M16 12a4 4 0 1 1-8 0 4 4 0 0 1 8 0Z' },
  { value: 'system', label: 'Как в системе', icon: 'M4 5h16v11H4zM9 20h6m-3-4v4' },
  { value: 'dark', label: 'Тёмная тема', icon: 'M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5Z' },
]
</script>

<template>
  <div class="toggle" role="radiogroup" aria-label="Тема">
    <button
      v-for="o in options"
      :key="o.value"
      type="button"
      role="radio"
      :aria-checked="theme === o.value"
      :aria-label="o.label"
      :title="o.label"
      :class="['option', { active: theme === o.value }]"
      @click="theme = o.value"
    >
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path :d="o.icon" />
      </svg>
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
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  padding: 0;
  color: var(--text-muted);
  background: none;
  border: 0;
  border-radius: 6px;
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

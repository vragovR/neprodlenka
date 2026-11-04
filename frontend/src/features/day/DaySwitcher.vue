<script setup lang="ts">
import { addDays, capitalize, dayTitle } from '@/utils/format'

const date = defineModel<string>({ required: true })

function pick(e: Event) {
  const value = (e.target as HTMLInputElement).value
  if (value) {
    date.value = value
  }
}
</script>

<template>
  <div class="switcher">
    <button type="button" class="arrow" aria-label="Предыдущий день" @click="date = addDays(date, -1)">‹</button>
    <label class="title">
      {{ capitalize(dayTitle(date)) }}
      <input type="date" class="picker" :value="date" aria-label="Выбрать дату" @change="pick" />
    </label>
    <button type="button" class="arrow" aria-label="Следующий день" @click="date = addDays(date, 1)">›</button>
  </div>
</template>

<style scoped>
.switcher {
  display: inline-flex;
  gap: 4px;
  align-items: center;
}

.title {
  position: relative;
  min-width: 15rem;
  font-weight: 600;
  text-align: center;
  cursor: pointer;
}

.picker {
  position: absolute;
  inset: 0;
  width: 100%;
  opacity: 0;
  cursor: pointer;
}

.picker::-webkit-calendar-picker-indicator {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  cursor: pointer;
}

.arrow {
  width: 32px;
  height: 32px;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
}

.arrow:hover {
  background: var(--surface-muted);
}

@media (pointer: coarse) {
  .arrow {
    width: 40px;
    height: 40px;
  }
}
</style>

<script setup lang="ts">
import { weekdaysShort } from '@/utils/format'

const { label = 'Дни недели' } = defineProps<{ label?: string }>()
const days = defineModel<number[]>({ required: true })

const toggle = (day: number) => {
  days.value = days.value.includes(day) ? days.value.filter((d) => d !== day) : [...days.value, day].sort((a, b) => a - b)
}
</script>

<template>
  <span class="days" role="group" :aria-label="label">
    <button
      v-for="(d, i) in weekdaysShort"
      :key="d"
      type="button"
      :class="['day', { on: days.includes(i + 1) }]"
      :aria-pressed="days.includes(i + 1)"
      @click="toggle(i + 1)"
    >
      {{ d }}
    </button>
  </span>
</template>

<style scoped>
.days {
  display: inline-flex;
  gap: 4px;
}

.day {
  width: 30px;
  height: 30px;
  padding: 0;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 50%;
  color: inherit;
  font-size: 0.75rem;
  cursor: pointer;
}

.day.on {
  background: var(--accent-soft);
  border-color: var(--accent-soft);
  color: var(--accent);
  font-weight: 600;
}

@media (pointer: coarse) {
  .day {
    width: 36px;
    height: 36px;
  }
}
</style>

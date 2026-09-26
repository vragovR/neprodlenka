<script setup lang="ts">
import { computed } from 'vue'
import { dayOfMonth, monthDates, weekday, weekdaysShort } from '@/utils/format'

const { month, dayClass, disabled } = defineProps<{
  month: string
  dayClass?: (date: string) => string | Record<string, boolean>
  disabled?: (date: string) => boolean
}>()
const emit = defineEmits<{ select: [date: string] }>()

const dates = computed(() => monthDates(month))
const offset = computed(() => (dates.value[0] ? weekday(dates.value[0]) : 0))
</script>

<template>
  <div class="grid">
    <span v-for="w in weekdaysShort" :key="w" class="weekday">{{ w }}</span>
    <span v-for="i in offset" :key="`gap-${i}`" />
    <button
      v-for="date in dates"
      :key="date"
      type="button"
      :class="['day', dayClass?.(date)]"
      :disabled="disabled?.(date)"
      @click="emit('select', date)"
    >
      <slot name="day" :date="date">{{ dayOfMonth(date) }}</slot>
    </button>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 6px;
}

.weekday {
  padding-bottom: 4px;
  color: var(--text-muted);
  font-size: 0.75rem;
  font-weight: 600;
  text-align: center;
  text-transform: uppercase;
}

.day {
  width: 100%;
  aspect-ratio: 1;
  max-height: 64px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.15s, border-color 0.15s, color 0.15s;
}

.day:not(:disabled):hover {
  border-color: var(--accent);
}

.day:disabled {
  cursor: default;
  opacity: 0.4;
}
</style>

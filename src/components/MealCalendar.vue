<script setup lang="ts">
import { computed } from 'vue'
import type { ReportDay, Status } from '@/api/types'
import { dayOfMonth, today, weekday, weekdaysShort } from '@/utils/format'

const { days, status, editable, large = false } = defineProps<{
  days: ReportDay[]
  status: (date: string) => Status | null
  editable?: (date: string) => boolean
  large?: boolean
}>()
const emit = defineEmits<{ toggle: [date: string] }>()

const now = today()
const offset = computed(() => (days[0] ? weekday(days[0].date) : 0))
const titles: Record<Status, string> = { present: 'Обедает', absent: 'Не обедает' }

function dayTitle(d: ReportDay) {
  const s = status(d.date)
  if (s) {
    return titles[s]
  }
  return d.school ? 'Нет отметки' : 'Неучебный день'
}
</script>

<template>
  <div :class="['grid', { large }]">
    <span v-for="w in weekdaysShort" :key="w" class="weekday">{{ w }}</span>
    <span v-for="i in offset" :key="`gap-${i}`" />
    <button
      v-for="d in days"
      :key="d.date"
      type="button"
      :class="['day', status(d.date) ?? (d.school ? 'none' : 'off'), { today: d.date === now }]"
      :title="dayTitle(d)"
      :disabled="!editable?.(d.date)"
      @click="emit('toggle', d.date)"
    >
      {{ dayOfMonth(d.date) }}
    </button>
  </div>
</template>

<style scoped>
.grid {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr));
  gap: 4px;
}

.weekday {
  padding-bottom: 2px;
  color: var(--text-muted);
  font-size: 0.75rem;
  font-weight: 600;
  text-align: center;
}

.day {
  aspect-ratio: 1;
  padding: 0;
  color: var(--text);
  background: transparent;
  border: 2px solid transparent;
  border-radius: var(--radius-sm);
  font: inherit;
  font-size: 0.875rem;
  font-variant-numeric: tabular-nums;
  cursor: pointer;
}

.large {
  gap: 8px;
}

.large .weekday {
  padding-bottom: 4px;
  font-size: 0.875rem;
}

.large .day {
  border-radius: var(--radius);
  font-size: 1rem;
}

.day:disabled {
  cursor: default;
}

.day:not(:disabled):hover {
  border-color: var(--border);
}

.none {
  background: var(--surface-muted);
}

.present {
  color: var(--success);
  background: var(--success-soft);
  font-weight: 700;
}

.absent {
  color: var(--danger);
  background: var(--danger-soft);
  font-weight: 700;
}

.off {
  color: var(--text-muted);
  opacity: 0.6;
}

.day.today,
.day.today:hover {
  border-color: var(--accent);
}
</style>

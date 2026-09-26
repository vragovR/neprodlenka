<script setup lang="ts">
import { computed, ref, watch, watchEffect } from 'vue'
import { getAttendance } from '@/api'
import type { AttendanceDay } from '@/api/types'
import { useRequest } from '@/composables/useRequest'
import { isWeekend, monthOf } from '@/utils/format'
import DatePicker from '@/components/DatePicker.vue'
import DebugLegend from './DebugLegend.vue'

const { mode } = defineProps<{ mode: 'message' | 'status' }>()
const date = defineModel<string>({ required: true })
const state = defineModel<AttendanceDay | null | undefined>('day')
const month = ref(monthOf(date.value))

const { data, error, run } = useRequest(() => getAttendance(month.value))
watch(month, run, { immediate: true })
defineExpose({ reload: run })

const days = computed(() => new Map<string, AttendanceDay>(data.value?.days.map((d) => [d.date, d])))

watchEffect(() => {
  const loaded = data.value?.month === monthOf(date.value)
  state.value = loaded ? (days.value.get(date.value) ?? null) : undefined
})

function dayClass(d: string): Record<string, boolean> {
  const day = days.value.get(d)
  const off = isWeekend(d) || (!!data.value && !day)
  const dot = !!day?.published
  if (mode === 'message') {
    return { off, dot, message: dot }
  }
  return { off, dot, open: dot && !day?.closed, closed: !!day?.closed }
}

const legend = computed(() => [
  { label: 'Неучебный', swatch: 'off' },
  ...(mode === 'message'
    ? [{ label: 'Есть сообщение', swatch: 'message' }]
    : [
        { label: 'Открыт', swatch: 'open' },
        { label: 'Закрыт', swatch: 'closed' },
      ]),
])
</script>

<template>
  <div class="day-picker">
    <DatePicker v-model="date" v-model:month="month" :day-class="dayClass" />
    <DebugLegend :items="legend" />
    <p v-if="error" class="error">Не удалось загрузить состояние дней: {{ error.message }}</p>
  </div>
</template>

<style scoped>
.day-picker {
  display: grid;
  gap: 12px;
}

.error {
  color: var(--danger);
  font-size: 0.875rem;
}

:deep(.day) {
  position: relative;
}

:deep(.day.off:not(.selected)) {
  background: var(--danger-soft);
  border-color: transparent;
  color: var(--danger);
}

:deep(.day.dot)::after {
  position: absolute;
  bottom: 4px;
  left: 50%;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  content: '';
  translate: -50% 0;
}

:deep(.day.message)::after {
  background: var(--accent);
}

:deep(.day.open)::after {
  background: var(--warning);
}

:deep(.day.closed)::after {
  background: var(--success);
}

:deep(.day.selected.dot)::after {
  background: var(--on-accent);
}
</style>

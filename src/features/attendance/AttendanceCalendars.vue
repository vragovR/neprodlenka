<script setup lang="ts">
import type { AttendanceDay, AttendanceRow, Status } from '@/api/types'
import { dayOfMonth, formatMoney, fullName, plural, today } from '@/utils/format'
import ChildCalendar from '@/components/ChildCalendar.vue'

const { days, rows, status } = defineProps<{
  days: AttendanceDay[]
  rows: AttendanceRow[]
  status: (date: string, childId: number) => Status | null
}>()
const emit = defineEmits<{ toggle: [date: string, childId: number] }>()

const now = today()
const editable = (date: string) => !!days[dayOfMonth(date) - 1]?.school && date <= now
const meals = (r: AttendanceRow) => Object.values(r.marks).filter((s) => s === 'present').length
</script>

<template>
  <div class="calendars">
    <ChildCalendar
      v-for="r in rows"
      :key="r.child_id"
      :name="fullName(r)"
      :days="days"
      :status="(date) => status(date, r.child_id)"
      :editable="editable"
      @toggle="(date) => emit('toggle', date, r.child_id)"
    >
      {{ meals(r) }} {{ plural(meals(r), ['обед', 'обеда', 'обедов']) }} ·
      <strong class="sum">{{ formatMoney(r.total) }}</strong>
      <template v-if="!r.active"> · деактивирован</template>
    </ChildCalendar>
  </div>
</template>

<style scoped>
.calendars {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(max(280px, calc((100% - 32px) / 3)), 1fr));
  gap: 16px;
}

.sum {
  color: var(--text);
}
</style>

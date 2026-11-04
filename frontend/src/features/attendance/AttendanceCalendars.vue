<script setup lang="ts">
import type { AttendanceDay, AttendanceRow, Status } from '@/api/types'
import { dayOfMonth, formatMoney, fullName, plural, weekday } from '@/utils/format'
import ChildCalendar from '@/components/ChildCalendar.vue'
import PaidToggle from './PaidToggle.vue'

const { days, rows, status, paid } = defineProps<{
  days: AttendanceDay[]
  rows: AttendanceRow[]
  status: (date: string, childId: number) => Status | null
  paid: (childId: number) => boolean
}>()
const emit = defineEmits<{
  toggle: [date: string, childId: number]
  togglePaid: [childId: number]
}>()

const editable = (date: string) => !!days[dayOfMonth(date) - 1]?.school
const rest = (r: AttendanceRow, date: string) => !r.weekdays.includes(weekday(date) + 1)
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
      :rest="(date) => rest(r, date)"
      @toggle="(date) => emit('toggle', date, r.child_id)">
      {{ meals(r) }} {{ plural(meals(r), ['обед', 'обеда', 'обедов']) }} ·
      <strong class="sum">{{ formatMoney(r.total) }}</strong>
      <template v-if="!r.active"> · деактивирован</template>
      <PaidToggle class="paid" :paid="paid(r.child_id)" @toggle="emit('togglePaid', r.child_id)" />
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

.paid {
  margin-left: 6px;
}
</style>

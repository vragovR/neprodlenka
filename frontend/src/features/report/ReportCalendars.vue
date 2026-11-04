<script setup lang="ts">
import type { Report } from '@/api/types'
import { dayOfMonth, formatMoney, fullName, plural } from '@/utils/format'
import ChildCalendar from '@/components/ChildCalendar.vue'

const { uid, report } = defineProps<{ uid: string; report: Report }>()
</script>

<template>
  <div class="calendars">
    <ChildCalendar
      v-for="(row, i) in report.children"
      :key="i"
      :name="fullName(row)"
      :days="report.days"
      :status="(date) => row.marks[dayOfMonth(date) - 1] ?? null"
      :to="{ name: 'child', params: { uid, id: row.id } }">
      {{ row.present }} {{ plural(row.present, ['обед', 'обеда', 'обедов']) }} ·
      <strong class="sum">{{ formatMoney(row.total) }}</strong>
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

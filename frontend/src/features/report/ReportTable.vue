<script setup lang="ts">
import type { Report } from '@/api/types'
import { dayOfMonth, formatMoney, shortName, today, weekday, weekdaysShort } from '@/utils/format'
import StatusMark from '@/components/StatusMark.vue'

const { uid, report } = defineProps<{ uid: string; report: Report }>()
const now = today()
</script>

<template>
  <div class="scroll">
    <table class="table">
      <thead>
        <tr>
          <th class="name">Ребёнок</th>
          <th
            v-for="d in report.days"
            :key="d.date"
            :class="['day', { off: !d.school, today: d.date === now }]"
            :title="d.school ? undefined : 'Неучебный день'">
            <span class="num">{{ dayOfMonth(d.date) }}</span>
            <span class="wd">{{ weekdaysShort[weekday(d.date)] }}</span>
          </th>
          <th class="sum">Обедал</th>
          <th class="sum">Не обедал</th>
          <th class="sum pay">К оплате</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(row, i) in report.children" :key="i">
          <th class="name" scope="row">
            <RouterLink :to="{ name: 'child', params: { uid, id: row.id } }" class="link">
              {{ shortName(row) }}
            </RouterLink>
          </th>
          <td v-for="(status, j) in row.marks" :key="j" :class="['day', { off: !report.days[j]?.school }]">
            <StatusMark :status="status" :blank="!report.days[j]?.school" />
          </td>
          <td class="sum">{{ row.present }}</td>
          <td class="sum">{{ row.absent }}</td>
          <td class="sum pay">{{ formatMoney(row.total) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.scroll {
  overflow-x: auto;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
}

.table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 0.875rem;
}

th,
td {
  padding: 8px 6px;
  border-bottom: 1px solid var(--border);
  text-align: center;
  white-space: nowrap;
}

tbody tr:last-child > * {
  border-bottom: 0;
}

thead th {
  color: var(--text-muted);
  font-weight: 600;
}

.name {
  position: sticky;
  left: 0;
  z-index: 1;
  min-width: 140px;
  padding-left: 16px;
  background: var(--surface);
  border-right: 1px solid var(--border);
  font-weight: 500;
  text-align: left;
}

thead .name {
  font-weight: 600;
}

.link {
  color: inherit;
  text-decoration: none;
}

.link:hover {
  color: var(--accent);
  text-decoration: underline;
}

.day {
  min-width: 30px;
  padding-inline: 2px;
}

.num,
.wd {
  display: block;
}

.num {
  color: var(--text);
}

.wd {
  font-size: 0.6875rem;
  font-weight: 500;
}

.off .num,
.off .wd {
  color: var(--text-muted);
  opacity: 0.6;
}

thead .today {
  background: var(--accent-soft);
}

.sum {
  padding: 8px 12px;
  font-variant-numeric: tabular-nums;
}

.pay {
  position: sticky;
  right: 0;
  z-index: 1;
  padding-right: 16px;
  background: var(--surface);
  border-left: 1px solid var(--border);
  text-align: right;
}

tbody .pay {
  font-weight: 600;
}
</style>

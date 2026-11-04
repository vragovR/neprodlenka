<script setup lang="ts">
import { useTemplateRef, watch } from 'vue'
import type { AttendanceDay, AttendanceRow, Status } from '@/api/types'
import { dayOfMonth, formatMoney, fullName, shortName, today, weekday, weekdaysShort } from '@/utils/format'
import AttendanceCell from './AttendanceCell.vue'
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

const now = today()
const rest = (r: AttendanceRow, date: string) => !r.weekdays.includes(weekday(date) + 1)
const scroll = useTemplateRef('scroll')

watch(
  () => [scroll.value, days[0]?.date] as const,
  ([el]) => el?.querySelector('thead .today')?.scrollIntoView({ block: 'nearest', inline: 'center' }),
  { flush: 'post' },
)

function dayTitle(day: AttendanceDay) {
  if (!day.school) {
    return 'Неучебный день'
  }
  if (day.date > now) {
    return 'Ещё не наступил'
  }
  if (!day.published) {
    return 'Без публикации'
  }
  return day.closed ? 'Отметка закрыта' : 'Идёт отметка'
}
</script>

<template>
  <div ref="scroll" class="scroll">
    <table class="table">
      <thead>
        <tr>
          <th class="name">Ребёнок</th>
          <th v-for="d in days" :key="d.date" :class="['day', { today: d.date === now, off: !d.school }]" :title="dayTitle(d)">
            <span class="num">{{ dayOfMonth(d.date) }}</span>
            <span class="wd">{{ weekdaysShort[weekday(d.date)] }}</span>
            <span :class="['post', { published: d.published, open: d.published && !d.closed }]" />
          </th>
          <th class="pay">К оплате</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.child_id">
          <th class="name" scope="row">
            <span class="full">{{ fullName(r) }}</span>
            <span class="short">{{ shortName(r) }}</span>
            <span v-if="!r.active" class="inactive muted">деактивирован</span>
          </th>
          <td
            v-for="d in days"
            :key="d.date"
            :class="['day', { today: d.date === now, off: !d.school }]"
            :title="d.school && rest(r, d.date) ? 'Не ходит в этот день' : undefined">
            <AttendanceCell
              :status="status(d.date, r.child_id)"
              :disabled="!d.school"
              :blank="!d.school || rest(r, d.date)"
              @toggle="emit('toggle', d.date, r.child_id)" />
          </td>
          <td class="pay">
            <span class="sum">{{ formatMoney(r.total) }}</span>
            <PaidToggle :paid="paid(r.child_id)" @toggle="emit('togglePaid', r.child_id)" />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.scroll {
  max-height: calc(100dvh - 190px);
  margin: 0 -16px;
  overflow: auto;
  background: var(--surface);
  border-block: 1px solid var(--border);
  overscroll-behavior-x: contain;
}

.table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  font-size: 0.875rem;
}

th,
td {
  padding: 6px 2px;
  border-bottom: 1px solid var(--border);
  text-align: center;
  white-space: nowrap;
}

tbody tr:last-child > * {
  border-bottom: 0;
}

thead th {
  position: sticky;
  top: 0;
  z-index: 2;
  background: var(--surface);
  color: var(--text-muted);
  font-weight: 600;
}

thead .today {
  background: var(--accent-soft);
}

thead .name {
  z-index: 3;
}

.name {
  position: sticky;
  left: 0;
  z-index: 1;
  max-width: 120px;
  padding: 6px 8px 6px 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  background: var(--surface);
  border-right: 1px solid var(--border);
  font-weight: 500;
  text-align: left;
}

thead .name {
  font-weight: 600;
}

.full {
  display: none;
}

.inactive {
  display: block;
  font-size: 0.75rem;
  font-weight: 400;
}

.day {
  min-width: 38px;
}

.today {
  background: var(--accent-soft);
}

.off .num,
.off .wd {
  color: var(--text-muted);
  opacity: 0.6;
}

.num,
.wd {
  display: block;
}

.num {
  color: var(--text);
}

thead .today .num {
  color: var(--accent);
}

.wd {
  font-size: 0.6875rem;
  font-weight: 500;
}

.post {
  display: block;
  width: 6px;
  height: 6px;
  margin: 3px auto 0;
  border-radius: 50%;
}

.post.published {
  background: var(--text-muted);
}

.post.open {
  background: var(--accent);
}

.pay {
  padding: 6px 12px;
  background: var(--surface);
  font-variant-numeric: tabular-nums;
  text-align: right;
}

tbody .pay .sum {
  display: block;
  margin-bottom: 2px;
  font-weight: 600;
}

@media (min-width: 720px) {
  .scroll {
    margin: 0;
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow);
  }

  .name {
    max-width: none;
    min-width: 140px;
    padding: 6px 12px 6px 16px;
  }

  .full {
    display: inline;
  }

  .short {
    display: none;
  }

  .day {
    min-width: 34px;
  }

  .pay {
    position: sticky;
    right: 0;
    z-index: 1;
    padding: 6px 16px 6px 12px;
    border-left: 1px solid var(--border);
  }

  thead .pay {
    z-index: 3;
  }
}
</style>

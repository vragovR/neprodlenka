<script setup lang="ts">
import type { DayClub, DayRow, Status } from '@/api/types'
import { fullName } from '@/utils/format'
import UiBadge from '@/components/ui/UiBadge.vue'
import AttendanceCell from '@/features/attendance/AttendanceCell.vue'
import DayClubsEditor from './DayClubsEditor.vue'

const { row, status, clubs, names, marksDisabled, clubsReadonly } = defineProps<{
  row: DayRow
  status: Status | null
  clubs: DayClub[]
  names: string[]
  marksDisabled: boolean
  clubsReadonly: boolean
}>()
const emit = defineEmits<{ toggle: []; clubs: [clubs: DayClub[]] }>()
const editing = defineModel<boolean>('editing', { default: false })

const labels: Record<Status, string> = { pending: 'Не отметился', present: 'Обедает', absent: 'Не обедает' }
</script>

<template>
  <article :class="['child', { inactive: !row.active }]">
    <div class="head">
      <span class="name">{{ fullName(row) }}</span>
      <UiBadge v-if="!row.active">деактивирован</UiBadge>
      <span class="mark">
        <span class="muted">{{ status ? labels[status] : 'Нет отметки' }}</span>
        <AttendanceCell class="cell" :status="status" :disabled="marksDisabled" @toggle="emit('toggle')" />
      </span>
    </div>
    <div v-if="clubs.length || !clubsReadonly" class="clubs">
      <DayClubsEditor
        v-model:editing="editing"
        :clubs="clubs"
        :names="names"
        :readonly="clubsReadonly"
        @change="emit('clubs', $event)"
      />
    </div>
  </article>
</template>

<style scoped>
.child {
  display: grid;
  gap: 12px;
  padding: 16px 20px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
}

.child.inactive {
  opacity: 0.7;
}

.head {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  align-items: center;
}

.name {
  font-weight: 600;
}

.mark {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  margin-left: auto;
  font-size: 0.875rem;
}

.cell {
  width: 40px;
}

.clubs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

@media (min-width: 720px) {
  .child {
    padding: 18px 24px;
  }
}
</style>

<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from 'vue'
import { putAttendance, putDayClubs } from '@/api'
import { errorMessage } from '@/api/client'
import type { DayClub, DayRow, Status } from '@/api/types'
import { nextStatus } from '@/composables/useMarks'
import { toast } from '@/composables/useToast'
import { fullName, monthOf } from '@/utils/format'
import UiBadge from '@/components/ui/UiBadge.vue'
import UiButton from '@/components/ui/UiButton.vue'
import AttendanceCell from '@/features/attendance/AttendanceCell.vue'
import DayClubsEditor from './DayClubsEditor.vue'

const { date, row, names, marksDisabled, clubsReadonly } = defineProps<{
  date: string
  row: DayRow
  names: string[]
  marksDisabled: boolean
  clubsReadonly: boolean
}>()
const emit = defineEmits<{ saved: []; dirty: [dirty: boolean] }>()

const labels: Record<Status, string> = { pending: 'Не отметился', present: 'Обедает', absent: 'Не обедает' }

const status = ref(row.status)
const clubs = ref(row.clubs)
const editing = ref(false)
const saving = ref(false)

function reset() {
  status.value = row.status
  clubs.value = row.clubs
  editing.value = false
}
watch(() => JSON.stringify(row), reset)

const snapshot = (list: DayClub[]) => list.map((c) => `${c.start} ${c.name}`).sort().join('\n')
const statusDirty = computed(() => status.value !== row.status)
const clubsDirty = computed(() => snapshot(clubs.value) !== snapshot(row.clubs))
const dirty = computed(() => statusDirty.value || clubsDirty.value)
watch(dirty, (d) => emit('dirty', d))
onUnmounted(() => emit('dirty', false))

async function save() {
  saving.value = true
  try {
    if (statusDirty.value) {
      await putAttendance(monthOf(date), [{ date, child_id: row.child_id, status: status.value }])
    }
    if (clubsDirty.value) {
      await putDayClubs(date, row.child_id, clubs.value)
    }
    toast('Сохранено')
    emit('saved')
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <article :class="['child', { inactive: !row.active, dirty }]">
    <div class="head">
      <span class="name">{{ fullName(row) }}</span>
      <UiBadge v-if="!row.active">деактивирован</UiBadge>
      <span class="mark">
        <span class="muted">{{ status ? labels[status] : 'Нет отметки' }}</span>
        <AttendanceCell
          class="cell"
          :status="status"
          :disabled="marksDisabled || saving"
          @toggle="status = nextStatus(status, row.status)"
        />
      </span>
    </div>
    <div v-if="clubs.length || !clubsReadonly" class="clubs">
      <DayClubsEditor
        v-model:editing="editing"
        :clubs="clubs"
        :names="names"
        :readonly="clubsReadonly"
        @change="clubs = $event"
      />
    </div>
    <div v-if="dirty" class="save">
      <UiButton size="sm" variant="ghost" :disabled="saving" @mousedown.prevent @click="reset">Отмена</UiButton>
      <UiButton size="sm" variant="primary" :loading="saving" :disabled="editing" @mousedown.prevent @click="save">
        Сохранить
      </UiButton>
      <span v-if="editing" class="muted hint">Сначала нажмите «Добавить» или «Готово» в строке кружка</span>
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

.child.dirty {
  border-color: var(--accent);
  opacity: 1;
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

.clubs,
.save {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.hint {
  font-size: 0.8125rem;
}

@media (min-width: 720px) {
  .child {
    padding: 18px 24px;
  }
}
</style>

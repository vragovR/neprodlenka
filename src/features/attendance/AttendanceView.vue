<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router'
import { getAttendance, putAttendance } from '@/api'
import { errorMessage } from '@/api/client'
import type { Status } from '@/api/types'
import { useMarks } from '@/composables/useMarks'
import { useRequest } from '@/composables/useRequest'
import { viewMode } from '@/composables/useViewMode'
import { toast } from '@/composables/useToast'
import { isMonthKey, monthOf, today } from '@/utils/format'
import { routeParam } from '@/utils/route'
import MonthSwitcher from '@/components/MonthSwitcher.vue'
import UiButton from '@/components/ui/UiButton.vue'
import UiEmpty from '@/components/ui/UiEmpty.vue'
import UiModal from '@/components/ui/UiModal.vue'
import MessagePreview from '@/components/MessagePreview.vue'
import UiSpinner from '@/components/ui/UiSpinner.vue'
import MonthSummary from '@/components/MonthSummary.vue'
import ViewToggle from '@/components/ViewToggle.vue'
import AttendanceCalendars from './AttendanceCalendars.vue'
import AttendanceTable from './AttendanceTable.vue'

const route = useRoute()
const router = useRouter()

const month = computed({
  get: (prev?: string): string => {
    if (prev && route.name !== 'month') {
      return prev
    }
    const m = routeParam(route, 'month')
    return isMonthKey(m) ? m : monthOf(today())
  },
  set: (m: string) => router.push({ name: 'month', params: { month: m } }),
})

const { data, error, loading, run } = useRequest(() => getAttendance(month.value))
watch(month, run, { immediate: true })


const filled = computed(() => !!data.value?.days.some((d) => !d.school))

const meals = computed(
  () => data.value?.children.reduce((sum, c) => sum + Object.values(c.marks).filter((s) => s === 'present').length, 0) ?? 0,
)

const saved = (date: string, childId: number): Status | null =>
  data.value?.children.find((c) => c.child_id === childId)?.marks[date] ?? null
const { changes, status, toggle, reset } = useMarks(saved)

const paidChanges = ref(new Map<number, boolean>())
const savedPaid = (childId: number) => !!data.value?.children.find((c) => c.child_id === childId)?.paid
const paid = (childId: number) => paidChanges.value.get(childId) ?? savedPaid(childId)
function togglePaid(childId: number) {
  const value = !paid(childId)
  if (value === savedPaid(childId)) {
    paidChanges.value.delete(childId)
  } else {
    paidChanges.value.set(childId, value)
  }
}
const dirty = computed(() => changes.value.size + paidChanges.value.size > 0)

watch(data, () => {
  reset()
  paidChanges.value = new Map()
})

const saving = ref(false)
const messageOpen = ref(false)

async function save() {
  saving.value = true
  try {
    await putAttendance(
      month.value,
      [...changes.value.values()],
      [...paidChanges.value].map(([child_id, paid]) => ({ child_id, paid })),
    )
    toast('Отметки сохранены')
    await run()
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    saving.value = false
  }
}

const confirmLeave = () => !dirty.value || confirm('Отметки не сохранены. Уйти?')
onBeforeRouteUpdate(confirmLeave)
onBeforeRouteLeave(confirmLeave)
</script>

<template>
  <main class="page attendance">
    <div class="page-head head">
      <h1>Месяц</h1>
      <MonthSwitcher v-model="month" />
      <div class="tools">
        <UiButton v-if="data?.message" :disabled="loading" @click="messageOpen = true">Сообщение месяца</UiButton>
        <ViewToggle />
      </div>
    </div>

    <UiEmpty v-if="error" title="Не удалось загрузить" :text="error.message">
      <UiButton :loading="loading" @click="run">Повторить</UiButton>
    </UiEmpty>
    <div v-else-if="!data" class="center"><UiSpinner /></div>
    <UiEmpty
      v-else-if="!filled"
      title="Календарь месяца не заполнен"
      text="Отметки появятся, когда в календаре будут отмечены неучебные дни."
    >
      <UiButton @click="router.push({ name: 'calendar', params: { month } })">Открыть календарь</UiButton>
    </UiEmpty>
    <UiEmpty v-else-if="!data.children.length" title="Нет детей" text="Добавьте детей на вкладке «Дети»." />
    <template v-else>
      <MonthSummary
        :children="data.children.length"
        :days="data.days.filter((d) => d.school).length"
        :meals="meals"
        :total="data.children.reduce((sum, c) => sum + c.total, 0)"
      />
      <div class="toolbar">
        <p v-if="viewMode === 'calendar'" class="muted hint">
          Нажмите на день: обедает (зелёный) → не обедает (красный) → без отметки. Жёлтый «?» — родитель ещё не отметился, при закрытии отметки станет «обедает». Бледные дни — неучебные. Сумма — по сохранённым отметкам, рядом — оплачено ли за месяц.
        </p>
        <p v-else class="muted hint">
          Нажмите на ячейку: ✓ обедает → ✕ не обедает → серая — без отметки. Жёлтый «?» — родитель ещё не отметился, при закрытии отметки станет «обедает». Точка под датой — в группе было сообщение (синяя — отметка ещё идёт). Бледные даты — неучебные дни. «К оплате» — по сохранённым отметкам, под суммой — оплачено ли за месяц.
        </p>
        <UiButton variant="primary" :disabled="!dirty" :loading="saving" @click="save">Сохранить</UiButton>
      </div>
      <AttendanceCalendars
        v-if="viewMode === 'calendar'"
        :days="data.days"
        :rows="data.children"
        :status="status"
        :paid="paid"
        @toggle="toggle"
        @toggle-paid="togglePaid"
      />
      <AttendanceTable
        v-else
        :days="data.days"
        :rows="data.children"
        :status="status"
        :paid="paid"
        @toggle="toggle"
        @toggle-paid="togglePaid"
      />
    </template>
    <UiModal v-model:open="messageOpen" title="Сообщение месяца">
      <MessagePreview v-if="data?.message" :message="data.message" :sent="data.sent" :dirty="dirty" />
    </UiModal>
  </main>
</template>

<style scoped>
.attendance {
  max-width: none;
}

.tools {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
}

@media (min-width: 720px) {
  .head {
    display: grid;
    grid-template-columns: 1fr auto 1fr;
  }

  .head .tools {
    justify-self: end;
  }
}

.center {
  display: grid;
  place-items: center;
  padding: 48px;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 24px;
  align-items: center;
  justify-content: space-between;
}

.hint {
  flex: 1 1 320px;
  margin: 0;
  font-size: 0.875rem;
}
</style>

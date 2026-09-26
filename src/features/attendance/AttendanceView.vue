<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router'
import { getAttendance, putAttendance } from '@/api'
import { errorMessage } from '@/api/client'
import type { MarkChange, Status } from '@/api/types'
import { useRequest } from '@/composables/useRequest'
import { viewMode } from '@/composables/useViewMode'
import { toast } from '@/composables/useToast'
import { isMonthKey, monthOf, today } from '@/utils/format'
import { routeParam } from '@/utils/route'
import MonthSwitcher from '@/components/MonthSwitcher.vue'
import UiButton from '@/components/ui/UiButton.vue'
import UiEmpty from '@/components/ui/UiEmpty.vue'
import UiSpinner from '@/components/ui/UiSpinner.vue'
import ViewToggle from '@/components/ViewToggle.vue'
import AttendanceCalendars from './AttendanceCalendars.vue'
import AttendanceTable from './AttendanceTable.vue'

const route = useRoute()
const router = useRouter()

const month = computed({
  get: (prev?: string): string => {
    if (prev && route.name !== 'attendance') {
      return prev
    }
    const m = routeParam(route, 'month')
    return isMonthKey(m) ? m : monthOf(today())
  },
  set: (m: string) => router.push({ name: 'attendance', params: { month: m } }),
})

const { data, error, loading, run } = useRequest(() => getAttendance(month.value))
watch(month, run, { immediate: true })

const changes = ref(new Map<string, MarkChange>())
watch(data, () => (changes.value = new Map()))

const filled = computed(() => !!data.value?.days.some((d) => !d.school))

const saved = (date: string, childId: number): Status | null =>
  data.value?.children.find((c) => c.child_id === childId)?.marks[date] ?? null

const status = (date: string, childId: number): Status | null => {
  const change = changes.value.get(`${date}:${childId}`)
  return change ? change.status : saved(date, childId)
}

function toggle(date: string, childId: number) {
  const current = status(date, childId)
  const next = current === null ? 'present' : current === 'present' ? 'absent' : null
  const key = `${date}:${childId}`
  if (next === saved(date, childId)) {
    changes.value.delete(key)
  } else {
    changes.value.set(key, { date, child_id: childId, status: next })
  }
}

const saving = ref(false)

async function save() {
  saving.value = true
  try {
    await putAttendance(month.value, [...changes.value.values()])
    toast('Отметки сохранены')
    await run()
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    saving.value = false
  }
}

const confirmLeave = () => !changes.value.size || confirm('Отметки не сохранены. Уйти?')
onBeforeRouteUpdate(confirmLeave)
onBeforeRouteLeave(confirmLeave)
</script>

<template>
  <main class="page attendance">
    <div class="page-head">
      <h1>Отметки</h1>
      <MonthSwitcher v-model="month" />
      <ViewToggle />
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
      <div class="toolbar">
        <p v-if="viewMode === 'calendar'" class="muted hint">
          Нажмите на день: обедает (зелёный) → не обедает (красный) → без отметки. Бледные дни — неучебные. Сумма — по сохранённым отметкам.
        </p>
        <p v-else class="muted hint">
          Нажмите на ячейку: ✓ обедает → ✕ не обедает → серая — без отметки. Точка под датой — в группе было сообщение (синяя — отметка ещё идёт). Бледные даты — неучебные дни. «К оплате» — по сохранённым отметкам.
        </p>
        <UiButton variant="primary" :disabled="!changes.size" :loading="saving" @click="save">Сохранить</UiButton>
      </div>
      <AttendanceCalendars
        v-if="viewMode === 'calendar'"
        :days="data.days"
        :rows="data.children"
        :status="status"
        @toggle="toggle"
      />
      <AttendanceTable v-else :days="data.days" :rows="data.children" :status="status" @toggle="toggle" />
    </template>
  </main>
</template>

<style scoped>
.attendance {
  max-width: none;
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

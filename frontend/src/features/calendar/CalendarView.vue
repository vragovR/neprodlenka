<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router'
import { getMonth, putMonth } from '@/api'
import { errorMessage } from '@/api/client'
import { useRequest } from '@/composables/useRequest'
import { toast } from '@/composables/useToast'
import { isMonthKey, isWeekend, monthDates, monthOf, today } from '@/utils/format'
import { routeParam } from '@/utils/route'
import MonthGrid from '@/components/MonthGrid.vue'
import MonthSwitcher from '@/components/MonthSwitcher.vue'
import UiButton from '@/components/ui/UiButton.vue'
import UiCard from '@/components/ui/UiCard.vue'
import UiEmpty from '@/components/ui/UiEmpty.vue'
import UiSpinner from '@/components/ui/UiSpinner.vue'
import CalendarLegend from './CalendarLegend.vue'
import ReportLinkCard from './ReportLinkCard.vue'

const route = useRoute()
const router = useRouter()

const month = computed({
  get: (prev?: string): string => {
    if (prev && route.name !== 'calendar') {
      return prev
    }
    const m = routeParam(route, 'month')
    return isMonthKey(m) ? m : monthOf(today())
  },
  set: (m: string) => router.push({ name: 'calendar', params: { month: m } }),
})

const { data, error, loading, run } = useRequest(() => getMonth(month.value))
watch(month, run, { immediate: true })

const draft = ref(new Set<string>())
watch(data, (d) => (draft.value = new Set(d?.no_school_days)))

const dirty = computed(() => {
  const saved = data.value?.no_school_days ?? []
  return saved.length !== draft.value.size || saved.some((d) => !draft.value.has(d))
})

function toggle(date: string) {
  const next = new Set(draft.value)
  if (!next.delete(date)) {
    next.add(date)
  }
  draft.value = next
}

function markWeekends() {
  draft.value = new Set([...draft.value, ...monthDates(month.value).filter(isWeekend)])
}

const saving = ref(false)

async function save() {
  saving.value = true
  try {
    await putMonth(month.value, [...draft.value].sort())
    toast('Календарь сохранён')
    await run()
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    saving.value = false
  }
}

const confirmLeave = () => !dirty.value || confirm('Изменения в календаре не сохранены. Уйти?')
onBeforeRouteUpdate(confirmLeave)
onBeforeRouteLeave(confirmLeave)

const dayClass = (date: string) => ({ off: draft.value.has(date), today: date === today() })
</script>

<template>
  <main class="page">
    <div class="page-head">
      <h1>Календарь</h1>
      <MonthSwitcher v-model="month" />
    </div>

    <UiEmpty v-if="error" title="Не удалось загрузить" :text="error.message">
      <UiButton :loading="loading" @click="run">Повторить</UiButton>
    </UiEmpty>
    <div v-else-if="!data" class="center"><UiSpinner /></div>
    <template v-else>
      <UiCard title="Учебные дни">
        <template #actions>
          <UiButton @click="markWeekends">Отметить все сб и вс</UiButton>
          <UiButton variant="primary" :disabled="!dirty" :loading="saving" @click="save">Сохранить</UiButton>
        </template>
        <p class="muted">Нажмите на день, чтобы сделать его учебным или неучебным.</p>
        <MonthGrid :month="month" :day-class="dayClass" @select="toggle" />
        <CalendarLegend />
      </UiCard>
      <ReportLinkCard :url="data.report_url" />
    </template>
  </main>
</template>

<style scoped>
.center {
  display: grid;
  place-items: center;
  padding: 48px;
}

:deep(.day.off) {
  background: var(--danger-soft);
  border-color: transparent;
  color: var(--danger);
}

:deep(.day.today) {
  outline: 2px solid var(--accent);
  outline-offset: -2px;
}
</style>

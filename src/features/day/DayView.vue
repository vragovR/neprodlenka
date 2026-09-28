<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router'
import { getDay, putAttendance, putDayClubs } from '@/api'
import { errorMessage } from '@/api/client'
import type { DayClub, Status } from '@/api/types'
import { useMarks } from '@/composables/useMarks'
import { useRequest } from '@/composables/useRequest'
import { toast } from '@/composables/useToast'
import { capitalize, isISODate, monthOf, today } from '@/utils/format'
import { routeParam } from '@/utils/route'
import UiButton from '@/components/ui/UiButton.vue'
import UiEmpty from '@/components/ui/UiEmpty.vue'
import UiModal from '@/components/ui/UiModal.vue'
import MessagePreview from '@/components/MessagePreview.vue'
import UiSpinner from '@/components/ui/UiSpinner.vue'
import DayChildCard from './DayChildCard.vue'
import DaySwitcher from './DaySwitcher.vue'

const route = useRoute()
const router = useRouter()

const date = computed({
  get: (prev?: string): string => {
    if (prev && route.name !== 'day') {
      return prev
    }
    const d = routeParam(route, 'date')
    return isISODate(d) ? d : today()
  },
  set: (d: string) => router.push({ name: 'day', params: { date: d } }),
})

const { data, error, loading, run } = useRequest(() => getDay(date.value))
watch(date, run, { immediate: true })

const saved = (_date: string, childId: number): Status | null =>
  data.value?.children.find((c) => c.child_id === childId)?.status ?? null
const { changes, status, toggle, reset: resetMarks } = useMarks(saved)

const clubs = ref(new Map<number, DayClub[]>())
const editing = ref(new Set<number>())

function reset() {
  resetMarks()
  clubs.value = new Map(data.value?.children.map((c) => [c.child_id, c.clubs]))
  editing.value = new Set()
}
watch(data, reset)

const snapshot = (list: DayClub[] = []) => list.map((c) => `${c.start} ${c.name}`).sort().join('\n')
const clubsDirty = computed(
  () => !!data.value?.children.some((c) => snapshot(c.clubs) !== snapshot(clubs.value.get(c.child_id))),
)
const dirty = computed(() => changes.value.size > 0 || clubsDirty.value)

const names = computed(() =>
  [...new Set([...clubs.value.values()].flat().map((c) => c.name))].sort((a, b) => a.localeCompare(b, 'ru')),
)
const marked = computed(() => !!data.value?.children.some((c) => c.status))
const emptyText = computed(() => {
  if (date.value > today()) {
    return 'Этот день ещё не наступил.'
  }
  if (date.value === today()) {
    return 'Сообщение за сегодня ещё не отправлено — отметки появятся после отправки.'
  }
  return 'Сообщение за этот день не отправлялось. Отметки можно поставить на странице «Месяц».'
})
const marksDisabled = computed(() => !data.value?.school || date.value > today())
const summary = computed(() => {
  const counts: Record<Status, number> = { pending: 0, present: 0, absent: 0 }
  for (const c of data.value?.children ?? []) {
    const s = status(date.value, c.child_id)
    if (s) {
      counts[s]++
    }
  }
  const parts = [`обедают — ${counts.present}`, `не обедают — ${counts.absent}`]
  if (counts.pending) {
    parts.push(`не отметились — ${counts.pending}`)
  }
  return capitalize(parts.join(', ')) + '.'
})

function setClubs(childId: number, list: DayClub[]) {
  clubs.value = new Map(clubs.value).set(childId, list)
}

function setEditing(childId: number, on: boolean) {
  const next = new Set(editing.value)
  if (on) {
    next.add(childId)
  } else {
    next.delete(childId)
  }
  editing.value = next
}

const saving = ref(false)
const messageOpen = ref(false)

async function save() {
  saving.value = true
  try {
    if (changes.value.size) {
      await putAttendance(monthOf(date.value), [...changes.value.values()])
    }
    if (clubsDirty.value) {
      await putDayClubs(
        date.value,
        [...clubs.value].flatMap(([childId, list]) => list.map((c) => ({ child_id: childId, ...c }))),
      )
    }
    toast('Сохранено')
    await run()
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    saving.value = false
  }
}

const confirmLeave = () => !dirty.value || confirm('Изменения за день не сохранены. Уйти?')
onBeforeRouteUpdate(confirmLeave)
onBeforeRouteLeave(confirmLeave)

function warnUnload(e: BeforeUnloadEvent) {
  if (dirty.value) {
    e.preventDefault()
  }
}
onMounted(() => window.addEventListener('beforeunload', warnUnload))
onUnmounted(() => window.removeEventListener('beforeunload', warnUnload))
</script>

<template>
  <main class="page">
    <div class="page-head">
      <h1>День</h1>
      <div class="tools">
        <UiButton v-if="data?.message" :disabled="loading" @click="messageOpen = true">Сообщение дня</UiButton>
        <DaySwitcher v-model="date" />
      </div>
    </div>

    <UiEmpty v-if="error" title="Не удалось загрузить" :text="error.message">
      <UiButton :loading="loading" @click="run">Повторить</UiButton>
    </UiEmpty>
    <div v-else-if="!data" class="center"><UiSpinner /></div>
    <UiEmpty v-else-if="!data.school" title="Неучебный день" text="В этот день сообщение в группу не отправляется." />
    <UiEmpty v-else-if="!data.children.length" title="Нет детей" text="Добавьте детей на вкладке «Дети»." />
    <UiEmpty v-else-if="!marked" title="Отметок нет" :text="emptyText">
      <UiButton v-if="date < today()" @click="router.push({ name: 'month', params: { month: monthOf(date) } })">
        Открыть месяц
      </UiButton>
    </UiEmpty>
    <template v-else>
      <div class="toolbar">
        <p class="muted hint">
          <template v-if="!data.published">
            Сообщения за этот день не было: кружки показаны по расписанию, править их нельзя.
          </template>
          <template v-else-if="data.closed">Отметка закрыта.</template>
          <template v-else>Отметка идёт: «Не отметился» при закрытии станет «Обедает».</template>
          {{ summary }}
        </p>
        <div class="actions">
          <UiButton v-if="dirty" variant="ghost" :disabled="saving" @click="reset">Отмена</UiButton>
          <UiButton variant="primary" :disabled="!dirty || editing.size > 0" :loading="saving" @click="save">
            Сохранить
          </UiButton>
        </div>
      </div>
      <p v-if="editing.size" class="muted hint">Сначала нажмите «Добавить» или «Готово» в строке кружка</p>
      <div class="list">
        <DayChildCard
          v-for="c in data.children"
          :key="c.child_id"
          :row="c"
          :status="status(date, c.child_id)"
          :clubs="clubs.get(c.child_id) ?? []"
          :names="names"
          :marks-disabled="marksDisabled"
          :clubs-readonly="!data.published"
          :editing="editing.has(c.child_id)"
          @update:editing="setEditing(c.child_id, $event)"
          @toggle="toggle(date, c.child_id)"
          @clubs="setClubs(c.child_id, $event)"
        />
      </div>
    </template>
    <UiModal v-model:open="messageOpen" title="Сообщение дня">
      <MessagePreview v-if="data?.message" :message="data.message" :sent="data.published" :dirty="dirty" />
    </UiModal>
  </main>
</template>

<style scoped>
.tools {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  margin-left: auto;
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

.actions {
  display: flex;
  gap: 8px;
}

.list {
  display: grid;
  gap: 12px;
}
</style>

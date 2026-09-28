<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { onBeforeRouteLeave, onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router'
import { getDay } from '@/api'
import type { Status } from '@/api/types'
import { useRequest } from '@/composables/useRequest'
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

const dirtyIds = ref(new Set<number>())
function setDirty(id: number, dirty: boolean) {
  const next = new Set(dirtyIds.value)
  if (dirty) {
    next.add(id)
  } else {
    next.delete(id)
  }
  dirtyIds.value = next
}
const dirty = computed(() => dirtyIds.value.size > 0)

const names = computed(() =>
  [...new Set(data.value?.children.flatMap((c) => c.clubs.map((club) => club.name)))].sort((a, b) =>
    a.localeCompare(b, 'ru'),
  ),
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
    if (c.status) {
      counts[c.status]++
    }
  }
  const parts = [`обедают — ${counts.present}`, `не обедают — ${counts.absent}`]
  if (counts.pending) {
    parts.push(`не отметились — ${counts.pending}`)
  }
  return capitalize(parts.join(', ')) + '.'
})

const messageOpen = ref(false)

const confirmLeave = () => !dirty.value || confirm('Изменения в карточках детей не сохранены. Уйти?')
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
      </div>
      <div class="list">
        <DayChildCard
          v-for="c in data.children"
          :key="`${data.date}:${c.child_id}`"
          :date="data.date"
          :row="c"
          :names="names"
          :marks-disabled="marksDisabled"
          :clubs-readonly="!data.published"
          @saved="run"
          @dirty="setDirty(c.child_id, $event)"
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

.list {
  display: grid;
  gap: 12px;
}
</style>

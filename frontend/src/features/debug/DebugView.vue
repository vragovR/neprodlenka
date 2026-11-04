<script setup lang="ts">
import { computed, ref, useTemplateRef } from 'vue'
import { debugCloseDay, debugDeleteDay, debugDeleteMonth, debugSendDay, debugSendMonth, debugSummaries } from '@/api'
import type { AttendanceDay } from '@/api/types'
import { useRequest } from '@/composables/useRequest'
import { formatDate, monthOf, monthTitle, today } from '@/utils/format'
import MonthPicker from '@/components/MonthPicker.vue'
import DebugActionCard, { type DebugAction } from './DebugActionCard.vue'
import DebugDayPicker from './DebugDayPicker.vue'
import DebugLegend from './DebugLegend.vue'

const sendDate = ref(today())
const sendState = ref<AttendanceDay | null>()
const closeDate = ref(today())
const closeState = ref<AttendanceDay | null>()
const month = ref(monthOf(today()))
const sendPicker = useTemplateRef('sendPicker')
const closePicker = useTemplateRef('closePicker')

const summaries = useRequest(debugSummaries)
summaries.run()
const summaryMonths = computed(() => new Set(summaries.data.value?.months))

const dayAction = computed((): DebugAction | undefined => {
  const date = formatDate(sendDate.value)
  if (sendState.value?.published) {
    return {
      label: `Удалить за ${date}`,
      run: deleteDay,
      danger: true,
      confirm: `Удалить сообщение за ${date} из группы?`,
    }
  }
  if (sendState.value && sendDate.value <= today()) {
    return { label: `Отправить за ${date}`, run: sendDay }
  }
  return undefined
})
const dayNote = computed(() => {
  if (!sendState.value) {
    return sendState.value === null ? 'Неучебный день' : ''
  }
  return sendDate.value > today() ? 'Будущий день' : ''
})

const monthAction = computed((): DebugAction => {
  const title = monthTitle(month.value)
  if (summaryMonths.value.has(month.value)) {
    return {
      label: `Удалить за ${title}`,
      run: deleteMonth,
      danger: true,
      confirm: `Удалить итоги за ${title} из группы?`,
    }
  }
  return { label: `Отправить за ${title}`, run: sendMonth }
})

const closeAction = computed((): DebugAction | undefined =>
  closeState.value?.published && !closeState.value.closed ? { label: `Закрыть ${formatDate(closeDate.value)}`, run: closeDay } : undefined,
)
const closeNote = computed(() => {
  if (!closeState.value) {
    return closeState.value === null ? 'Неучебный день' : ''
  }
  return closeState.value.closed ? 'День уже закрыт' : 'За этот день нет сообщения'
})

function reloadDays() {
  sendPicker.value?.reload()
  closePicker.value?.reload()
}

async function sendDay() {
  await debugSendDay(sendDate.value)
  reloadDays()
  return `Сообщение за ${formatDate(sendDate.value)} отправлено`
}

async function sendMonth() {
  await debugSendMonth(month.value)
  summaries.run()
  return `Итоги за ${monthTitle(month.value)} отправлены`
}

async function deleteDay() {
  await debugDeleteDay(sendDate.value)
  reloadDays()
  return `Сообщение за ${formatDate(sendDate.value)} удалено`
}

async function deleteMonth() {
  await debugDeleteMonth(month.value)
  summaries.run()
  return `Итоги за ${monthTitle(month.value)} удалены`
}

async function closeDay() {
  await debugCloseDay(closeDate.value)
  reloadDays()
  return `Отметка на ${formatDate(closeDate.value)} закрыта`
}
</script>

<template>
  <main class="page">
    <div class="page-head">
      <h1>Отладка</h1>
    </div>
    <p class="muted">
      Сообщения отсюда публикуются так же, как их публикует бот по расписанию: кнопка работает, сообщение обновляется при изменении отметок.
      Бот сам закрывает только сообщение за сегодня, за другой день — закройте здесь. Закрытие дня работает как у бота: неотмеченным
      ставится «обедает», кнопка из сообщения убирается. Удаление убирает сообщение из группы и запись о нём, отметки остаются — сообщение
      можно отправить заново.
    </p>
    <div class="cards">
      <DebugActionCard title="Сообщение за день" :action="dayAction" :note="dayNote">
        <DebugDayPicker ref="sendPicker" v-model="sendDate" v-model:day="sendState" mode="message" />
      </DebugActionCard>
      <DebugActionCard title="Итоги месяца" :action="monthAction">
        <MonthPicker v-model="month" :marked="(m) => summaryMonths.has(m)" />
        <DebugLegend :items="[{ label: 'Есть сообщение с итогами', swatch: 'message' }]" />
        <p v-if="summaries.error.value" class="error"> Не удалось загрузить итоги: {{ summaries.error.value.message }} </p>
      </DebugActionCard>
      <DebugActionCard title="Закрыть день" :action="closeAction" :note="closeNote">
        <DebugDayPicker ref="closePicker" v-model="closeDate" v-model:day="closeState" mode="status" />
      </DebugActionCard>
    </div>
  </main>
</template>

<style scoped>
.cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(min(100%, 320px), 1fr));
  gap: 16px;
  align-items: start;
}

.error {
  color: var(--danger);
  font-size: 0.875rem;
}
</style>

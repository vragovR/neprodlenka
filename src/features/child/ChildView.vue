<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { getChildReport } from '@/api'
import { useRequest } from '@/composables/useRequest'
import { dayOfMonth, formatMoney, fullName, monthTitle, plural } from '@/utils/format'
import MealCalendar from '@/components/MealCalendar.vue'
import MonthSwitcher from '@/components/MonthSwitcher.vue'
import UiButton from '@/components/ui/UiButton.vue'
import UiEmpty from '@/components/ui/UiEmpty.vue'
import UiSpinner from '@/components/ui/UiSpinner.vue'

const { uid, id } = defineProps<{ uid: string; id: string }>()

const router = useRouter()

const { data, error, loading, run } = useRequest(() => getChildReport(uid, id))
watch(() => uid, run, { immediate: true })

const month = computed({
  get: () => data.value?.month ?? '',
  set: (m: string) => {
    const target = m < month.value ? data.value?.prev_uid : data.value?.next_uid
    if (target) {
      router.replace({ name: 'child', params: { uid: target, id } })
    }
  },
})

const name = computed(() => (data.value ? fullName(data.value.child) : ''))
const filled = computed(() => !!data.value?.days.some((d) => !d.school))
const status = (date: string) => data.value?.child.marks[dayOfMonth(date) - 1] ?? null

const legend = [
  { label: 'обедал', swatch: 'present' },
  { label: 'не обедал', swatch: 'absent' },
  { label: 'ещё не отмечен', swatch: 'pending' },
  { label: 'нет отметки', swatch: 'none' },
  { label: 'школа не работает', swatch: 'off' },
]
</script>

<template>
  <main class="page child">
    <UiEmpty v-if="error?.status === 404" title="Страница не найдена" text="Проверьте ссылку." />
    <UiEmpty v-else-if="error" title="Не удалось загрузить" :text="error.message">
      <UiButton :loading="loading" @click="run">Повторить</UiButton>
    </UiEmpty>
    <div v-else-if="!data" class="center"><UiSpinner /></div>
    <template v-else>
      <RouterLink :to="{ name: 'report', params: { uid } }" class="back">
        ← Отчёт по классу
      </RouterLink>
      <header class="hero">
        <span class="avatar">{{ name.charAt(0) }}</span>
        <div>
          <h1>{{ name }}</h1>
          <p class="muted">Календарь школьного питания</p>
        </div>
      </header>

      <div><MonthSwitcher v-model="month" :has-prev="!!data.prev_uid" :has-next="!!data.next_uid" /></div>

      <UiEmpty
        v-if="!filled"
        title="Календарь месяца не заполнен"
        text="Отметки появятся, когда школа отметит неучебные дни этого месяца."
      />
      <div v-else class="layout">
        <section class="panel">
          <MealCalendar large :days="data.days" :status="status" />
          <ul class="legend">
            <li v-for="l in legend" :key="l.swatch">
              <span :class="['swatch', l.swatch]" />{{ l.label }}
            </li>
          </ul>
        </section>

        <aside class="panel">
          <div class="tiles">
            <div class="tile">
              <strong>{{ data.child.present }}</strong>
              <span class="muted">{{ plural(data.child.present, ['обед', 'обеда', 'обедов']) }} в месяце</span>
            </div>
            <div v-if="data.price" class="tile">
              <strong>{{ formatMoney(data.price) }}</strong>
              <span class="muted">за обед</span>
            </div>
          </div>
          <div class="pay">
            <strong>{{ formatMoney(data.child.total) }}</strong>
            <span class="muted">к оплате за {{ monthTitle(data.month) }}</span>
          </div>
        </aside>
      </div>
    </template>
  </main>
</template>

<style scoped>
.center {
  display: grid;
  place-items: center;
  padding: 48px;
}

.back {
  justify-self: start;
  color: var(--accent);
  font-weight: 500;
  text-decoration: none;
}

.back:hover {
  text-decoration: underline;
}

.hero {
  display: flex;
  gap: 16px;
  align-items: center;
}

.hero h1 {
  margin: 0;
}

.hero p {
  margin: 4px 0 0;
}

.avatar {
  display: grid;
  flex: none;
  place-items: center;
  width: 64px;
  height: 64px;
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 50%;
  font-size: 1.75rem;
  font-weight: 700;
}

.layout {
  display: grid;
  gap: 16px;
  align-items: start;
}

.panel {
  display: grid;
  gap: 16px;
  padding: 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
}

.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin: 0;
  padding: 0;
  color: var(--text-muted);
  font-size: 0.875rem;
  list-style: none;
}

.legend li {
  display: flex;
  gap: 6px;
  align-items: center;
}

.swatch {
  width: 14px;
  height: 14px;
  border-radius: 4px;
}

.swatch.present {
  background: var(--success-soft);
  border: 1px solid var(--success);
}

.swatch.absent {
  background: var(--danger-soft);
  border: 1px solid var(--danger);
}

.swatch.pending {
  background: var(--warning-soft);
  border: 1px solid var(--warning);
}

.swatch.none {
  background: var(--surface-muted);
}

.swatch.off {
  border: 1px solid var(--border);
}

.tiles {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 12px;
}

.tile {
  display: grid;
  gap: 4px;
  align-content: start;
  padding: 12px 16px;
  background: var(--surface-muted);
  border-radius: var(--radius);
}

.tile strong {
  font-size: 1.375rem;
  font-variant-numeric: tabular-nums;
}

.pay {
  display: grid;
  gap: 4px;
  padding-top: 16px;
  border-top: 1px solid var(--border);
}

.pay strong {
  font-size: 2rem;
  font-variant-numeric: tabular-nums;
}

@media (min-width: 900px) {
  .layout {
    grid-template-columns: minmax(0, 2fr) minmax(260px, 1fr);
  }

  .panel {
    padding: 24px;
  }
}
</style>

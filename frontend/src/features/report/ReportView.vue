<script setup lang="ts">
import { onMounted } from 'vue'
import { getReport } from '@/api'
import { useRequest } from '@/composables/useRequest'
import { monthTitle } from '@/utils/format'
import { viewMode } from '@/composables/useViewMode'
import UiButton from '@/components/ui/UiButton.vue'
import UiEmpty from '@/components/ui/UiEmpty.vue'
import UiSpinner from '@/components/ui/UiSpinner.vue'
import MonthSummary from '@/components/MonthSummary.vue'
import ViewToggle from '@/components/ViewToggle.vue'
import ReportCalendars from './ReportCalendars.vue'
import ReportTable from './ReportTable.vue'

const { uid } = defineProps<{ uid: string }>()

const { data, error, loading, run } = useRequest(() => getReport(uid))
onMounted(run)
</script>

<template>
  <main class="page report">
    <UiEmpty v-if="error?.status === 404" title="Отчёт не найден" text="Проверьте ссылку." />
    <UiEmpty v-else-if="error" title="Не удалось загрузить" :text="error.message">
      <UiButton :loading="loading" @click="run">Повторить</UiButton>
    </UiEmpty>
    <div v-else-if="!data" class="center"><UiSpinner /></div>
    <template v-else>
      <div class="page-head">
        <h1>Отчёт за {{ monthTitle(data.month) }}</h1>
        <ViewToggle v-if="data.children.length" />
      </div>
      <UiEmpty v-if="!data.children.length" title="В этом месяце отметок нет" />
      <template v-else>
        <MonthSummary
          :children="data.children.length"
          :days="data.days.filter((d) => d.school).length"
          :meals="data.children.reduce((sum, c) => sum + c.present, 0)"
          :total="data.total" />
        <ReportCalendars v-if="viewMode === 'calendar'" :uid="uid" :report="data" />
        <ReportTable v-else :uid="uid" :report="data" />
      </template>
    </template>
  </main>
</template>

<style scoped>
.report {
  max-width: none;
}

.center {
  display: grid;
  place-items: center;
  padding: 48px;
}
</style>

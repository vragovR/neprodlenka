<script setup lang="ts">
import { onMounted } from 'vue'
import { getReport } from '@/api'
import { useRequest } from '@/composables/useRequest'
import { monthTitle } from '@/utils/format'
import UiButton from '@/components/ui/UiButton.vue'
import UiEmpty from '@/components/ui/UiEmpty.vue'
import UiSpinner from '@/components/ui/UiSpinner.vue'
import ReportTable from './ReportTable.vue'
import ReportTotal from './ReportTotal.vue'

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
      </div>
      <UiEmpty v-if="!data.children.length" title="В этом месяце отметок нет" />
      <template v-else>
        <ReportTable :report="data" />
        <ReportTotal :total="data.total" />
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

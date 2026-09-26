<script setup lang="ts">
import { computed } from 'vue'
import type { Report } from '@/api/types'
import { formatMoney, plural } from '@/utils/format'

const { report } = defineProps<{ report: Report }>()

const tiles = computed(() => {
  const children = report.children.length
  const days = report.days.filter((d) => d.school).length
  const meals = report.children.reduce((sum, c) => sum + c.present, 0)
  return [
    { value: String(children), label: plural(children, ['ребёнок', 'ребёнка', 'детей']) },
    { value: String(days), label: `${plural(days, ['учебный день', 'учебных дня', 'учебных дней'])} в месяце` },
    { value: String(meals), label: `${plural(meals, ['обед', 'обеда', 'обедов'])} всего` },
    { value: formatMoney(report.total), label: 'сумма по классу' },
  ]
})
</script>

<template>
  <section class="summary">
    <div v-for="t in tiles" :key="t.label" class="tile">
      <strong>{{ t.value }}</strong>
      <span class="muted">{{ t.label }}</span>
    </div>
  </section>
</template>

<style scoped>
.summary {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  padding: 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
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
  font-size: 1.25rem;
  font-variant-numeric: tabular-nums;
}

.tile span {
  font-size: 0.875rem;
}
</style>

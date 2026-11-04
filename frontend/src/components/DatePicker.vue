<script setup lang="ts">
import { today } from '@/utils/format'
import MonthGrid from '@/components/MonthGrid.vue'
import MonthSwitcher from '@/components/MonthSwitcher.vue'

const { dayClass } = defineProps<{ dayClass?: (date: string) => Record<string, boolean> }>()
const date = defineModel<string>({ required: true })
const month = defineModel<string>('month', { required: true })

const classes = (d: string) => ({
  ...dayClass?.(d),
  selected: d === date.value,
  today: d === today(),
})
</script>

<template>
  <div class="picker">
    <MonthSwitcher v-model="month" />
    <MonthGrid :month="month" :day-class="classes" @select="date = $event" />
  </div>
</template>

<style scoped>
.picker {
  display: grid;
  gap: 12px;
  justify-items: center;
}

.picker :deep(.grid) {
  width: 100%;
}

.picker :deep(.day.today) {
  outline: 2px solid var(--accent);
  outline-offset: -2px;
}

.picker :deep(.day.selected) {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--on-accent);
}
</style>

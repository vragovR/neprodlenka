<script setup lang="ts">
import { addMonths, capitalize, monthTitle } from '@/utils/format'

const month = defineModel<string>({ required: true })
</script>

<template>
  <div class="switcher">
    <button type="button" class="arrow" aria-label="Предыдущий месяц" @click="month = addMonths(month, -1)">‹</button>
    <span class="title">{{ capitalize(monthTitle(month)) }}</span>
    <button type="button" class="arrow" aria-label="Следующий месяц" @click="month = addMonths(month, 1)">›</button>
  </div>
</template>

<style scoped>
.switcher {
  display: inline-flex;
  gap: 4px;
  align-items: center;
}

.title {
  min-width: 140px;
  font-weight: 600;
  text-align: center;
}

.arrow {
  width: 32px;
  height: 32px;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
}

@media (pointer: coarse) {
  .arrow {
    width: 40px;
    height: 40px;
  }
}

.arrow:hover {
  background: var(--surface-muted);
}
</style>

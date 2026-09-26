<script setup lang="ts">
import { formatMoney } from '@/utils/format'

const { total } = defineProps<{ total: number }>()
</script>

<template>
  <div class="total">
    <span class="muted">Итого по классу</span>
    <strong class="value">{{ formatMoney(total) }}</strong>
  </div>
</template>

<style scoped>
.total {
  display: flex;
  gap: 16px;
  align-items: baseline;
  justify-content: flex-end;
  padding: 16px 20px;
  background: var(--accent-soft);
  border-radius: var(--radius-lg);
}

.value {
  color: var(--accent);
  font-size: 1.375rem;
  font-variant-numeric: tabular-nums;
}
</style>

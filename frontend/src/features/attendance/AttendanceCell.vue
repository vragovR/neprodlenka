<script setup lang="ts">
import type { Status } from '@/api/types'
import StatusMark from '@/components/StatusMark.vue'

const {
  status,
  disabled = false,
  blank = false,
} = defineProps<{
  status: Status | null
  disabled?: boolean
  blank?: boolean
}>()
const emit = defineEmits<{ toggle: [] }>()
</script>

<template>
  <button type="button" class="cell" :disabled="disabled" @click="emit('toggle')">
    <StatusMark :status="status" :blank="blank" />
  </button>
</template>

<style scoped>
.cell {
  display: grid;
  place-items: center;
  width: 100%;
  min-height: 36px;
  padding: 0;
  background: transparent;
  border: 0;
  border-radius: var(--radius-sm);
  cursor: pointer;
}

.cell:not(:disabled):hover {
  background: var(--surface-muted);
}

.cell:disabled {
  cursor: default;
}
</style>

<script setup lang="ts">
import type { Status } from '@/api/types'

const { status, blank = false } = defineProps<{ status: Status | null; blank?: boolean }>()
</script>

<template>
  <span v-if="status === 'present'" class="mark present" title="Придёт">✓</span>
  <span v-else-if="status === 'absent'" class="mark absent" title="Не придёт">✕</span>
  <span v-else-if="!blank" class="mark none" title="Нет отметки">—</span>
  <span v-else class="mark" />
</template>

<style scoped>
.mark {
  display: inline-grid;
  place-items: center;
  width: 24px;
  height: 24px;
  border-radius: 6px;
  font-size: 0.8125rem;
  font-weight: 700;
}

.present {
  background: var(--success-soft);
  color: var(--success);
}

.absent {
  background: var(--danger-soft);
  color: var(--danger);
}

.none {
  color: var(--text-muted);
}
</style>

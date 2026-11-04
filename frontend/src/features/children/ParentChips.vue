<script setup lang="ts">
import { computed } from 'vue'
import type { Member } from '@/api/types'

const { ids, members } = defineProps<{ ids: number[]; members: Member[] }>()
const emit = defineEmits<{ remove: [id: number] }>()

const parents = computed(() =>
  ids.map((id) => {
    const member = members.find((m) => m.user_id === id)
    return { id, name: member?.name ?? `не в группе (id ${id})`, missing: !member }
  }),
)
</script>

<template>
  <span v-if="!parents.length" class="muted none">Родители не привязаны</span>
  <span v-for="p in parents" :key="p.id" :class="['chip', { missing: p.missing }]">
    {{ p.name }}
    <button type="button" class="remove" :aria-label="`Отвязать ${p.name}`" @click="emit('remove', p.id)">×</button>
  </span>
</template>

<style scoped>
.none {
  font-size: 0.875rem;
}

.chip {
  display: inline-flex;
  gap: 2px;
  align-items: center;
  height: 30px;
  padding: 0 4px 0 12px;
  background: var(--surface-muted);
  border-radius: 999px;
  font-size: 0.875rem;
}

.chip.missing {
  background: var(--warning-soft);
  color: var(--warning);
}

.remove {
  width: 22px;
  height: 22px;
  background: transparent;
  border: 0;
  border-radius: 50%;
  color: inherit;
  line-height: 1;
  opacity: 0.6;
  cursor: pointer;
}

@media (pointer: coarse) {
  .chip {
    height: 36px;
  }

  .remove {
    width: 30px;
    height: 30px;
  }
}

.remove:hover {
  background: var(--border);
  opacity: 1;
}
</style>

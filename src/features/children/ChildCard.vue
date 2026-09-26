<script setup lang="ts">
import { computed } from 'vue'
import type { Child, ChildPatch, Member } from '@/api/types'
import { formatMoney, fullName } from '@/utils/format'
import UiBadge from '@/components/ui/UiBadge.vue'
import UiButton from '@/components/ui/UiButton.vue'
import ParentChips from './ParentChips.vue'
import ParentPicker from './ParentPicker.vue'

const { child, members } = defineProps<{
  child: Child
  members: Member[]
}>()
const emit = defineEmits<{ edit: []; patch: [patch: ChildPatch] }>()

const freeMembers = computed(() => members.filter((m) => !child.parent_ids.includes(m.user_id)))
const setParents = (ids: number[]) => emit('patch', { parent_ids: ids })
const toggleLabel = computed(() => (child.active ? 'Деактивировать' : 'Активировать'))
const icon = {
  viewBox: '0 0 24 24',
  width: 18,
  height: 18,
  fill: 'none',
  stroke: 'currentColor',
  'stroke-width': 2,
  'stroke-linecap': 'round',
  'stroke-linejoin': 'round',
} as const
</script>

<template>
  <article :class="['child', { inactive: !child.active }]">
    <div class="main">
      <div class="title">
        <span class="name">{{ fullName(child) }}</span>
        <UiBadge tone="accent">{{ formatMoney(child.price) }} / день</UiBadge>
        <UiBadge v-if="!child.active">деактивирован</UiBadge>
      </div>
      <div class="parents">
        <ParentChips
          :ids="child.parent_ids"
          :members="members"
          @remove="setParents(child.parent_ids.filter((id) => id !== $event))"
        />
        <ParentPicker :members="freeMembers" @pick="setParents([...child.parent_ids, $event])" />
      </div>
    </div>
    <div class="actions">
      <UiButton size="sm" variant="ghost" square aria-label="Изменить" title="Изменить" @click="emit('edit')">
        <svg v-bind="icon"><path d="M4 20h4L19 9l-4-4L4 16v4ZM13.5 6.5l4 4" /></svg>
      </UiButton>
      <UiButton
        size="sm"
        square
        :variant="child.active ? 'danger' : 'secondary'"
        :aria-label="toggleLabel"
        :title="toggleLabel"
        @click="emit('patch', { active: !child.active })"
      >
        <svg v-bind="icon">
          <path v-if="child.active" d="M4 7h16v3H4zM5 10v9h14v-9M10 14h4" />
          <path v-else d="M4 12a8 8 0 1 0 2.3-5.7M4 4v4h4" />
        </svg>
      </UiButton>
    </div>
  </article>
</template>

<style scoped>
.child {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 16px;
  align-items: flex-start;
  justify-content: space-between;
  padding: 14px 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
}

.child.inactive {
  opacity: 0.7;
}

.main {
  display: grid;
  gap: 10px;
  min-width: 0;
}

.title {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.name {
  font-weight: 600;
}

.parents {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.actions {
  display: flex;
  gap: 6px;
  margin-left: auto;
}

@media (min-width: 720px) {
  .child {
    padding: 16px 20px;
  }
}
</style>

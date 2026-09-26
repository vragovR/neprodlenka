<script setup lang="ts">
import type { Child, ChildPatch, Member } from '@/api/types'
import ChildCard from './ChildCard.vue'

const { children, members } = defineProps<{
  children: Child[]
  members: Member[]
}>()
const emit = defineEmits<{ edit: [child: Child]; patch: [child: Child, patch: ChildPatch] }>()
</script>

<template>
  <ul v-if="children.length" class="list">
    <li v-for="c in children" :key="c.id">
      <ChildCard
        :child="c"
        :members="members"
        @edit="emit('edit', c)"
        @patch="emit('patch', c, $event)"
      />
    </li>
  </ul>
  <slot v-else />
</template>

<style scoped>
.list {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}
</style>

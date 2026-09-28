<script setup lang="ts">
import type { Child, Member } from '@/api/types'
import ChildCard from './ChildCard.vue'

const { children, members, clubNames } = defineProps<{
  children: Child[]
  members: Member[]
  clubNames: string[]
}>()
const emit = defineEmits<{ update: [child: Child]; dirty: [id: number, dirty: boolean] }>()
</script>

<template>
  <ul v-if="children.length" class="list">
    <li v-for="c in children" :key="c.id">
      <ChildCard
        :child="c"
        :members="members"
        :club-names="clubNames"
        @update="emit('update', $event)"
        @dirty="emit('dirty', c.id, $event)"
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

<script setup lang="ts">
import type { RouteLocationRaw } from 'vue-router'
import type { ReportDay, Status } from '@/api/types'
import MealCalendar from './MealCalendar.vue'

const { name, days, status, editable, rest, to } = defineProps<{
  name: string
  days: ReportDay[]
  status: (date: string) => Status | null
  editable?: (date: string) => boolean
  rest?: (date: string) => boolean
  to?: RouteLocationRaw
}>()
const emit = defineEmits<{ toggle: [date: string] }>()
</script>

<template>
  <article class="card">
    <header class="head">
      <span class="avatar">{{ name.charAt(0) }}</span>
      <div class="info">
        <RouterLink v-if="to" :to="to" class="name link">{{ name }}</RouterLink>
        <div v-else class="name">{{ name }}</div>
        <div class="muted sub"><slot /></div>
      </div>
    </header>
    <MealCalendar :days="days" :status="status" :editable="editable" :rest="rest" @toggle="(date) => emit('toggle', date)" />
  </article>
</template>

<style scoped>
.card {
  display: grid;
  gap: 16px;
  padding: 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
}

.head {
  display: flex;
  gap: 12px;
  align-items: center;
  min-width: 0;
}

.avatar {
  display: grid;
  flex: none;
  place-items: center;
  width: 40px;
  height: 40px;
  color: var(--accent);
  background: var(--accent-soft);
  border-radius: 50%;
  font-weight: 700;
}

.info {
  min-width: 0;
}

.name {
  display: block;
  overflow: hidden;
  color: var(--text);
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link {
  text-decoration: none;
}

.link:hover {
  color: var(--accent);
  text-decoration: underline;
}

.sub {
  font-size: 0.875rem;
}
</style>

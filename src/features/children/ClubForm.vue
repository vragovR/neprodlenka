<script setup lang="ts">
import { onMounted, reactive, useId, useTemplateRef } from 'vue'
import { weekdaysShort } from '@/utils/format'
import type { ClubGroup } from './clubs'

const { group, names, submitLabel } = defineProps<{ group: ClubGroup; names: string[]; submitLabel: string }>()
const emit = defineEmits<{ submit: [group: ClubGroup]; cancel: [] }>()

const form = reactive({ name: group.name, weekdays: [...group.weekdays], start: group.start })
const listId = useId()
const nameInput = useTemplateRef('nameInput')

onMounted(() => nameInput.value?.focus())

const toggleDay = (day: number) => {
  form.weekdays = form.weekdays.includes(day) ? form.weekdays.filter((d) => d !== day) : [...form.weekdays, day]
}

function submit() {
  const name = form.name.trim()
  if (!name || !form.start || !form.weekdays.length) {
    return
  }
  emit('submit', { name, start: form.start, weekdays: [...form.weekdays].sort((a, b) => a - b) })
}
</script>

<template>
  <form class="form" @submit.prevent="submit" @keydown.esc="emit('cancel')">
    <input
      ref="nameInput"
      v-model="form.name"
      class="field name"
      placeholder="Название"
      aria-label="Название кружка"
      :list="listId"
      required
    />
    <datalist :id="listId">
      <option v-for="n in names" :key="n" :value="n" />
    </datalist>
    <span class="days" role="group" aria-label="Дни недели">
      <button
        v-for="(d, i) in weekdaysShort"
        :key="d"
        type="button"
        :class="['day', { on: form.weekdays.includes(i + 1) }]"
        :aria-pressed="form.weekdays.includes(i + 1)"
        @click="toggleDay(i + 1)"
      >
        {{ d }}
      </button>
    </span>
    <input v-model="form.start" type="time" class="field" aria-label="Время начала" required />
    <button type="submit" class="btn primary" :disabled="!form.name.trim() || !form.start || !form.weekdays.length">
      {{ submitLabel }}
    </button>
    <button type="button" class="btn" aria-label="Отмена" @click="emit('cancel')">×</button>
  </form>
</template>

<style scoped>
.form {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.field,
.btn {
  height: 30px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 999px;
  color: inherit;
  font-size: 0.875rem;
}

.field {
  padding: 0 10px;
}

.field:focus {
  border-color: var(--accent);
  outline: none;
}

.name {
  width: 12rem;
  max-width: 100%;
}

.days {
  display: inline-flex;
  gap: 4px;
}

.day {
  width: 30px;
  height: 30px;
  padding: 0;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 50%;
  color: inherit;
  font-size: 0.75rem;
  cursor: pointer;
}

.day.on {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--on-accent);
}

.btn {
  min-width: 30px;
  padding: 0 10px;
  cursor: pointer;
}

.btn.primary {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--on-accent);
}

.btn:disabled {
  opacity: 0.5;
  cursor: default;
}

@media (pointer: coarse) {
  .field,
  .btn {
    height: 36px;
    font-size: 16px;
  }

  .day {
    width: 36px;
    height: 36px;
  }
}
</style>

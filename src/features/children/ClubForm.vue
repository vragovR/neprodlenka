<script setup lang="ts">
import { onMounted, reactive, useId, useTemplateRef } from 'vue'
import type { ClubGroup } from './clubs'
import WeekdayPicker from './WeekdayPicker.vue'

const {
  group,
  names,
  submitLabel,
  withDays = true,
} = defineProps<{ group: ClubGroup; names: string[]; submitLabel: string; withDays?: boolean }>()
const emit = defineEmits<{ submit: [group: ClubGroup]; cancel: [] }>()

const form = reactive({ name: group.name, weekdays: [...group.weekdays], start: group.start })
const listId = useId()
const nameInput = useTemplateRef('nameInput')

onMounted(() => nameInput.value?.focus())

const ready = () => !!form.name.trim() && !!form.start && (!withDays || !!form.weekdays.length)

function submit() {
  if (!ready()) {
    return
  }
  const name = form.name.trim()
  emit('submit', { name, start: form.start, weekdays: [...form.weekdays] })
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
    <WeekdayPicker v-if="withDays" v-model="form.weekdays" />
    <input v-model="form.start" type="time" class="field" aria-label="Время начала" required />
    <button type="submit" class="btn primary" :disabled="!ready()">
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
}
</style>

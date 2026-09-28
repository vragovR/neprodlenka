<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { DayClub } from '@/api/types'
import ClubForm from '@/features/children/ClubForm.vue'

const { clubs, names, readonly = false } = defineProps<{ clubs: DayClub[]; names: string[]; readonly?: boolean }>()
const emit = defineEmits<{ change: [clubs: DayClub[]] }>()
const editing = defineModel<boolean>('editing', { default: false })

const key = (c: DayClub) => `${c.start}\n${c.name}`
const label = (c: DayClub) => `${c.start} ${c.name}`
const sorted = computed(() => [...clubs].sort((a, b) => a.start.localeCompare(b.start) || a.name.localeCompare(b.name, 'ru')))

const form = ref<{ old: DayClub | null } | null>(null)
watch(form, (f) => (editing.value = !!f))
watch(editing, (e) => {
  if (!e) {
    form.value = null
  }
})
const isEditing = (c: DayClub) => !!form.value?.old && key(form.value.old) === key(c)

function replace(old: DayClub | null, next: DayClub | null) {
  const rest = old ? clubs.filter((c) => key(c) !== key(old)) : clubs
  const added = next && !rest.some((c) => key(c) === key(next)) ? [next] : []
  emit('change', [...rest, ...added])
  form.value = null
}
</script>

<template>
  <template v-if="readonly">
    <span v-for="c in sorted" :key="key(c)" class="chip readonly">{{ label(c) }}</span>
  </template>
  <template v-else>
    <span v-for="c in sorted" :key="key(c)" :class="['chip', { active: isEditing(c) }]">
      <button type="button" class="edit" :title="`Изменить ${label(c)}`" @click="form = { old: c }">{{ label(c) }}</button>
      <button type="button" class="remove" :aria-label="`Удалить ${label(c)}`" @click="replace(c, null)">×</button>
    </span>
    <ClubForm
      v-if="form"
      :key="form.old ? key(form.old) : ''"
      class="form"
      :group="{ name: form.old?.name ?? '', start: form.old?.start ?? '', weekdays: [] }"
      :names="names"
      :with-days="false"
      :submit-label="form.old ? 'Готово' : 'Добавить'"
      @submit="replace(form.old, { start: $event.start, name: $event.name })"
      @cancel="form = null"
    />
    <button v-else type="button" class="toggle" @click="form = { old: null }">+ Кружок</button>
  </template>
</template>

<style scoped>
.form {
  flex-basis: 100%;
  margin-top: 4px;
}

.chip {
  display: inline-flex;
  gap: 2px;
  align-items: center;
  height: 30px;
  padding: 0 4px 0 0;
  background: var(--accent-soft);
  border-radius: 999px;
  font-size: 0.875rem;
}

.chip.readonly {
  padding: 0 12px;
}

.chip.active {
  box-shadow: 0 0 0 1px var(--accent);
}

.edit {
  height: 100%;
  padding: 0 4px 0 12px;
  background: transparent;
  border: 0;
  border-radius: 999px 0 0 999px;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.edit:hover {
  text-decoration: underline;
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

.remove:hover {
  background: var(--border);
  opacity: 1;
}

.toggle {
  height: 30px;
  padding: 0 12px;
  background: var(--surface);
  border: 1px dashed var(--border);
  border-radius: 999px;
  color: inherit;
  font-size: 0.875rem;
  cursor: pointer;
}

.toggle:hover {
  border-color: var(--accent);
}

@media (pointer: coarse) {
  .chip,
  .toggle {
    height: 36px;
  }

  .remove {
    width: 30px;
    height: 30px;
  }
}
</style>

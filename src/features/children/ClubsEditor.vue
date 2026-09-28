<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Club } from '@/api/types'
import ClubForm from './ClubForm.vue'
import { type ClubGroup, expandGroup, groupClubs, groupKey, groupLabel } from './clubs'

const { clubs, names } = defineProps<{ clubs: Club[]; names: string[] }>()
const emit = defineEmits<{ change: [clubs: Club[]] }>()
const editing = defineModel<boolean>('editing', { default: false })

const empty: ClubGroup = { name: '', start: '', weekdays: [] }
const groups = computed(() => groupClubs(clubs))
const form = ref<{ old: ClubGroup | null } | null>(null)
watch(form, (f) => (editing.value = !!f))
watch(editing, (e) => {
  if (!e) {
    form.value = null
  }
})
const isEditing = (g: ClubGroup) => !!form.value?.old && groupKey(form.value.old) === groupKey(g)

function replace(old: ClubGroup | null, next: ClubGroup | null) {
  const rest = old ? clubs.filter((c) => groupKey(c) !== groupKey(old)) : clubs
  emit('change', next ? [...rest, ...expandGroup(next)] : rest)
  form.value = null
}
</script>

<template>
  <span v-for="g in groups" :key="groupKey(g)" :class="['chip', { active: isEditing(g) }]">
    <button type="button" class="edit" :title="`Изменить ${groupLabel(g)}`" @click="form = { old: g }">
      {{ groupLabel(g) }}
    </button>
    <button type="button" class="remove" :aria-label="`Удалить ${groupLabel(g)}`" @click="replace(g, null)">×</button>
  </span>
  <ClubForm
    v-if="form"
    :key="form.old ? groupKey(form.old) : ''"
    class="form"
    :group="form.old ?? empty"
    :names="names"
    :submit-label="form.old ? 'Готово' : 'Добавить'"
    @submit="replace(form.old, $event)"
    @cancel="form = null"
  />
  <button v-else type="button" class="toggle" @click="form = { old: null }">+ Кружок</button>
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

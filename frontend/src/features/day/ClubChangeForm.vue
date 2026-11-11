<script setup lang="ts">
import { computed, reactive, ref, useId } from 'vue'
import type { Child, ClubChange, DayClubChange } from '@/api/types'
import { fullName } from '@/utils/format'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'
import UiSelect from '@/components/ui/UiSelect.vue'
import UiSwitch from '@/components/ui/UiSwitch.vue'

const { change, children, saving = false } = defineProps<{ change: DayClubChange | null; children: Child[]; saving?: boolean }>()
const emit = defineEmits<{ submit: [childId: number, change: ClubChange]; cancel: [] }>()

const form = reactive({
  childId: change?.child_id ?? null,
  name: change?.name ?? '',
  start: change?.start ?? '',
  canceled: change?.start === null,
})
const touched = ref(false)
const selectId = useId()
const listId = useId()

const options = computed(() => children.map((c) => ({ value: c.id, label: fullName(c) })))
const names = computed(() =>
  [...new Set(children.find((c) => c.id === form.childId)?.clubs.map((c) => c.name))].sort((a, b) => a.localeCompare(b, 'ru')),
)
const errors = computed(() => ({
  child: form.childId === null ? 'Выберите ребёнка' : '',
  name: form.name.trim() ? '' : 'Укажите кружок',
  start: form.canceled || form.start ? '' : 'Укажите время',
}))

function submit() {
  touched.value = true
  if (form.childId === null || errors.value.name || errors.value.start) {
    return
  }
  emit('submit', form.childId, { name: form.name.trim(), start: form.canceled ? null : form.start })
}
</script>

<template>
  <form class="form" novalidate @submit.prevent="submit">
    <div class="field">
      <label :for="selectId" class="label">Ребёнок</label>
      <UiSelect
        :id="selectId"
        v-model="form.childId"
        class="select"
        :options="options"
        placeholder="Выберите ребёнка"
        :disabled="!!change" />
      <span v-if="touched && errors.child" class="error">{{ errors.child }}</span>
    </div>
    <UiInput v-model="form.name" label="Кружок" :list="listId" :error="touched ? errors.name : ''" autofocus />
    <datalist :id="listId">
      <option v-for="n in names" :key="n" :value="n" />
    </datalist>
    <UiSwitch v-model="form.canceled" label="Кружок отменён" />
    <UiInput v-if="!form.canceled" v-model="form.start" type="time" label="Время" :error="touched ? errors.start : ''" />
    <div class="actions">
      <UiButton variant="ghost" @click="emit('cancel')">Отмена</UiButton>
      <UiButton type="submit" variant="primary" :loading="saving">Сохранить</UiButton>
    </div>
  </form>
</template>

<style scoped>
.form {
  display: grid;
  gap: 14px;
}

.field {
  display: grid;
  gap: 6px;
}

.label {
  font-size: 0.875rem;
  font-weight: 500;
}

.select {
  width: 100%;
}

.error {
  color: var(--danger);
  font-size: 0.8125rem;
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  padding-top: 6px;
}
</style>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { getChildren, putClubChanges } from '@/api'
import { errorMessage } from '@/api/client'
import type { ClubChange, DayClubChange } from '@/api/types'
import { useRequest } from '@/composables/useRequest'
import { toast } from '@/composables/useToast'
import UiButton from '@/components/ui/UiButton.vue'
import UiCard from '@/components/ui/UiCard.vue'
import UiModal from '@/components/ui/UiModal.vue'
import ClubChangeForm from './ClubChangeForm.vue'

const { date, changes, reloading = false } = defineProps<{ date: string; changes: DayClubChange[]; reloading?: boolean }>()
const emit = defineEmits<{ saved: [] }>()

const { data: loaded, error: loadError, loading: childrenLoading, run: loadChildren } = useRequest(getChildren)
const children = computed(() => loaded.value?.children.filter((c) => c.active) ?? [])

const form = ref<{ change: DayClubChange | null } | null>(null)
const formOpen = computed({
  get: () => !!form.value,
  set: (open) => {
    if (!open) {
      form.value = null
    }
  },
})
const saving = ref(false)
const busy = computed(() => saving.value || reloading)

async function openForm(change: DayClubChange | null) {
  if (!loaded.value) {
    await loadChildren()
    if (loadError.value) {
      toast(loadError.value.message, 'danger')
      return
    }
  }
  form.value = { change }
}

const key = (c: DayClubChange) => `${c.child_id}\n${c.name}`
const label = (c: DayClubChange) => `${c.last_name} ${c.first_name} — ${c.name} — ${c.start ?? 'отменён'}`

const otherChanges = (childId: number, replacedNames: string[]): ClubChange[] =>
  changes.filter((c) => c.child_id === childId && !replacedNames.includes(c.name)).map((c) => ({ name: c.name, start: c.start }))

async function save(childId: number, list: ClubChange[], done: string) {
  saving.value = true
  try {
    await putClubChanges(date, childId, list)
    toast(done)
    form.value = null
    emit('saved')
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    saving.value = false
  }
}

function submit(childId: number, change: ClubChange) {
  const replaced = [form.value?.change?.name ?? '', change.name]
  void save(childId, [...otherChanges(childId, replaced), change], 'Сохранено')
}

function remove(c: DayClubChange) {
  void save(c.child_id, otherChanges(c.child_id, [c.name]), 'Изменение удалено')
}

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
  <UiCard title="Изменения кружков">
    <template #actions>
      <UiButton :loading="childrenLoading" @click="openForm(null)">Добавить изменение</UiButton>
    </template>
    <p class="muted hint">
      Разовые изменения применятся к сообщению за этот день и удалятся после его отправки. Постоянное расписание не меняется.
    </p>
    <ul v-if="changes.length" class="list">
      <li v-for="c in changes" :key="key(c)" class="item">
        <span class="label">{{ label(c) }}</span>
        <UiButton
          size="sm"
          square
          variant="ghost"
          :aria-label="`Изменить: ${label(c)}`"
          title="Изменить"
          :disabled="busy || childrenLoading"
          @click="openForm(c)">
          <svg v-bind="icon"><path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" /></svg>
        </UiButton>
        <UiButton size="sm" square variant="ghost" :aria-label="`Удалить: ${label(c)}`" title="Удалить" :disabled="busy" @click="remove(c)">
          <svg v-bind="icon"><path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6M10 11v6M14 11v6" /></svg>
        </UiButton>
      </li>
    </ul>
    <p v-else class="muted">Изменений нет — кружки пройдут по расписанию.</p>
  </UiCard>
  <UiModal v-model:open="formOpen" :title="form?.change ? 'Изменение кружка' : 'Новое изменение'">
    <ClubChangeForm v-if="form" :change="form.change" :children="children" :saving="busy" @submit="submit" @cancel="form = null" />
  </UiModal>
</template>

<style scoped>
.hint {
  font-size: 0.875rem;
}

.list {
  display: grid;
  gap: 4px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.item {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 8px;
  align-items: center;
  padding: 6px 0;
  border-bottom: 1px solid var(--border);
}

.item:last-child {
  border-bottom: 0;
}

.label {
  flex: 1 1 240px;
}
</style>

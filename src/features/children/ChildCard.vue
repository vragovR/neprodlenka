<script setup lang="ts">
import { computed, nextTick, onUnmounted, reactive, ref, useTemplateRef, watch } from 'vue'
import { updateChild } from '@/api'
import { errorMessage } from '@/api/client'
import type { Child, ChildPatch, Member } from '@/api/types'
import { toast } from '@/composables/useToast'
import { formatMoney, fullName } from '@/utils/format'
import UiBadge from '@/components/ui/UiBadge.vue'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'
import ClubsEditor from './ClubsEditor.vue'
import ParentChips from './ParentChips.vue'
import ParentPicker from './ParentPicker.vue'
import PriceInput from './PriceInput.vue'
import WeekdayPicker from './WeekdayPicker.vue'

const { child, members, clubNames } = defineProps<{
  child: Child
  members: Member[]
  clubNames: string[]
}>()
const emit = defineEmits<{ update: [child: Child]; dirty: [dirty: boolean] }>()

const fromChild = (c: Child) => ({
  last_name: c.last_name,
  first_name: c.first_name,
  middle_name: c.middle_name,
  price: c.price as number | null,
  weekdays: [...c.weekdays],
  parent_ids: [...c.parent_ids],
  clubs: c.clubs.map((club) => ({ ...club })),
})

const draft = reactive(fromChild(child))
const editingName = ref(false)
const editingPrice = ref(false)
const clubsEditing = ref(false)
const saving = ref(false)
const touched = ref(false)

function reset() {
  Object.assign(draft, fromChild(child))
  editingName.value = false
  editingPrice.value = false
  clubsEditing.value = false
  touched.value = false
}
watch(() => child, reset)

const snapshot = (d: ReturnType<typeof fromChild>) =>
  JSON.stringify({ ...d, clubs: d.clubs.map((c) => `${c.weekday} ${c.start} ${c.name}`).sort() })
const dirty = computed(() => snapshot(draft) !== snapshot(fromChild(child)))
watch(dirty, (d) => emit('dirty', d))
onUnmounted(() => emit('dirty', false))
const freeMembers = computed(() => members.filter((m) => !draft.parent_ids.includes(m.user_id)))
const nameError = (value: string, message: string) => (touched.value && !value.trim() ? message : '')
const nameValid = () => !!draft.last_name.trim() && !!draft.first_name.trim()

const nameFields = useTemplateRef('nameFields')
const priceField = useTemplateRef('priceField')

async function edit(field: 'name' | 'price') {
  const [editing, el] = field === 'name' ? [editingName, nameFields] : [editingPrice, priceField]
  editing.value = true
  await nextTick()
  el.value?.querySelector('input')?.focus()
}

function leave(e: FocusEvent, field: 'name' | 'price') {
  if ((e.currentTarget as HTMLElement).contains(e.relatedTarget as Node | null)) {
    return
  }
  const valid = field === 'name' ? nameValid() : draft.price !== null
  if (!valid) {
    touched.value = true
    return
  }
  if (field === 'name') {
    editingName.value = false
  } else {
    editingPrice.value = false
  }
}

async function send(patch: ChildPatch) {
  saving.value = true
  try {
    emit('update', await updateChild(child.id, patch))
    return true
  } catch (e) {
    toast(errorMessage(e), 'danger')
    return false
  } finally {
    saving.value = false
  }
}

async function save() {
  touched.value = true
  if (!nameValid() || draft.price === null) {
    editingName.value ||= !nameValid()
    editingPrice.value ||= draft.price === null
    return
  }
  if (
    await send({
      last_name: draft.last_name.trim(),
      first_name: draft.first_name.trim(),
      middle_name: draft.middle_name.trim(),
      price: draft.price,
      weekdays: draft.weekdays,
      parent_ids: draft.parent_ids,
      clubs: draft.clubs,
    })
  ) {
    toast('Сохранено')
  }
}

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
  <article :class="['child', { inactive: !child.active, dirty }]">
    <div class="main">
      <div v-if="editingName" ref="nameFields" class="fields" @focusout="leave($event, 'name')">
        <UiInput
          v-model="draft.last_name"
          placeholder="Фамилия"
          aria-label="Фамилия"
          :error="nameError(draft.last_name, 'Укажите фамилию')"
        />
        <UiInput v-model="draft.first_name" placeholder="Имя" aria-label="Имя" :error="nameError(draft.first_name, 'Укажите имя')" />
        <UiInput v-model="draft.middle_name" placeholder="Отчество" aria-label="Отчество" />
      </div>
      <div v-if="!editingName || !editingPrice || !child.active" class="title">
        <button v-if="!editingName" type="button" class="name" title="Изменить ФИО" @click="edit('name')">
          {{ fullName(draft) }}
        </button>
        <button v-if="!editingPrice" type="button" class="price-badge" title="Изменить цену" @click="edit('price')">
          <UiBadge tone="accent">{{ formatMoney(draft.price ?? 0) }} / день</UiBadge>
        </button>
        <UiBadge v-if="!child.active">деактивирован</UiBadge>
      </div>
      <div v-if="editingPrice" ref="priceField" class="price" @focusout="leave($event, 'price')">
        <PriceInput v-model="draft.price" :error="touched && draft.price === null ? 'Укажите цену' : ''" />
        <span class="muted hint">Цена за день. Новая цена действует только на новые отметки.</span>
      </div>
      <div class="weekdays">
        <span class="muted hint">Ходит</span>
        <WeekdayPicker v-model="draft.weekdays" label="Дни посещения" />
      </div>
      <div class="parents">
        <ParentChips
          :ids="draft.parent_ids"
          :members="members"
          @remove="draft.parent_ids = draft.parent_ids.filter((id) => id !== $event)"
        />
        <ParentPicker :members="freeMembers" @pick="draft.parent_ids = [...draft.parent_ids, $event].sort((a, b) => a - b)" />
      </div>
      <div class="clubs">
        <ClubsEditor v-model:editing="clubsEditing" :clubs="draft.clubs" :names="clubNames" @change="draft.clubs = $event" />
      </div>
      <div v-if="dirty" class="save">
        <UiButton size="sm" variant="ghost" :disabled="saving" @mousedown.prevent @click="reset">Отмена</UiButton>
        <UiButton
          size="sm"
          variant="primary"
          :loading="saving"
          :disabled="clubsEditing || !draft.weekdays.length"
          @mousedown.prevent
          @click="save"
        >
          Сохранить
        </UiButton>
        <span v-if="clubsEditing" class="muted hint">Сначала нажмите «Добавить» или «Готово» в строке кружка</span>
        <span v-else-if="!draft.weekdays.length" class="muted hint">Выберите хотя бы один день посещения</span>
      </div>
    </div>
    <div class="actions">
      <UiButton
        size="sm"
        square
        :variant="child.active ? 'danger' : 'secondary'"
        :aria-label="toggleLabel"
        :title="toggleLabel"
        :disabled="saving || dirty"
        @click="send({ active: !child.active })"
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
  padding: 18px 20px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius);
  box-shadow: var(--shadow);
}

.child.inactive {
  opacity: 0.7;
}

.child.dirty {
  border-color: var(--accent);
  opacity: 1;
}

.main {
  display: grid;
  flex: 1;
  gap: 14px;
  min-width: 0;
}

.title {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.name,
.price-badge {
  padding: 0;
  background: transparent;
  border: 0;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.name {
  font-weight: 600;
  text-align: left;
}

.name:hover {
  text-decoration: underline;
}

.fields {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
  gap: 8px;
}

.price {
  display: grid;
  gap: 4px;
  justify-items: start;
}

.price :deep(.input) {
  width: 7rem;
}

.hint {
  font-size: 0.8125rem;
}

.save {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.weekdays,
.parents {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.clubs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
}

.actions {
  display: flex;
  gap: 6px;
  margin-left: auto;
}

@media (min-width: 720px) {
  .child {
    padding: 22px 26px;
  }
}
</style>

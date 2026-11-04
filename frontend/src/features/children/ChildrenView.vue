<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { onBeforeRouteLeave } from 'vue-router'
import { addChild, getChildren } from '@/api'
import { errorMessage } from '@/api/client'
import type { Child, ChildInput, Member } from '@/api/types'
import { useRequest } from '@/composables/useRequest'
import { toast } from '@/composables/useToast'
import { fullName, sortByName } from '@/utils/format'
import UiButton from '@/components/ui/UiButton.vue'
import UiEmpty from '@/components/ui/UiEmpty.vue'
import UiModal from '@/components/ui/UiModal.vue'
import UiSpinner from '@/components/ui/UiSpinner.vue'
import ChildForm from './ChildForm.vue'
import ChildList from './ChildList.vue'

const children = ref<Child[]>([])
const members = ref<Member[]>([])
const {
  data: loaded,
  error,
  loading,
  run,
} = useRequest(async () => {
  const res = await getChildren()
  children.value = sortByName(res.children)
  members.value = res.members
  return true
})
onMounted(run)

const active = computed(() => children.value.filter((c) => c.active))
const inactive = computed(() => children.value.filter((c) => !c.active))
const clubNames = computed(() =>
  [...new Set(children.value.flatMap((c) => c.clubs.map((club) => club.name)))].sort((a, b) => a.localeCompare(b, 'ru')),
)

const dirtyIds = ref(new Set<number>())
function setDirty(id: number, dirty: boolean) {
  const next = new Set(dirtyIds.value)
  if (dirty) {
    next.add(id)
  } else {
    next.delete(id)
  }
  dirtyIds.value = next
}

const confirmLeave = () => !dirtyIds.value.size || confirm('Изменения в карточках детей не сохранены. Уйти?')
onBeforeRouteLeave(confirmLeave)

function warnUnload(e: BeforeUnloadEvent) {
  if (dirtyIds.value.size) {
    e.preventDefault()
  }
}
onMounted(() => window.addEventListener('beforeunload', warnUnload))
onUnmounted(() => window.removeEventListener('beforeunload', warnUnload))

const formOpen = ref(false)
const saving = ref(false)

function replace(child: Child) {
  const rest = children.value.filter((c) => c.id !== child.id)
  children.value = sortByName([...rest, child])
}

async function save(input: ChildInput) {
  saving.value = true
  try {
    const child = await addChild(input)
    replace(child)
    formOpen.value = false
    toast(`${fullName(child)} добавлен(а)`)
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <main class="page">
    <div class="page-head">
      <h1>Дети</h1>
      <UiButton variant="primary" @click="formOpen = true">Добавить ребёнка</UiButton>
    </div>

    <UiEmpty v-if="error" title="Не удалось загрузить" :text="error.message">
      <UiButton :loading="loading" @click="run">Повторить</UiButton>
    </UiEmpty>
    <div v-else-if="!loaded" class="center"><UiSpinner /></div>
    <template v-else>
      <ChildList :children="active" :members="members" :club-names="clubNames" @update="replace" @dirty="setDirty">
        <UiEmpty title="Пока никого нет" text="Добавьте первого ребёнка." />
      </ChildList>

      <details v-if="inactive.length" class="inactive">
        <summary>Деактивированные · {{ inactive.length }}</summary>
        <ChildList :children="inactive" :members="members" :club-names="clubNames" @update="replace" @dirty="setDirty" />
      </details>
    </template>

    <UiModal v-model:open="formOpen" title="Новый ребёнок">
      <ChildForm :saving="saving" @submit="save" @cancel="formOpen = false" />
    </UiModal>
  </main>
</template>

<style scoped>
.center {
  display: grid;
  place-items: center;
  padding: 48px;
}

.inactive {
  display: grid;
  gap: 12px;
}

.inactive summary {
  padding: 8px 0;
  color: var(--text-muted);
  font-weight: 600;
  cursor: pointer;
}
</style>

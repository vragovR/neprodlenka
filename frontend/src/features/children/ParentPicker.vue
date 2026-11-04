<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Member } from '@/api/types'
import UiSelect from '@/components/ui/UiSelect.vue'

const { members } = defineProps<{ members: Member[] }>()
const emit = defineEmits<{ pick: [id: number] }>()

const options = computed(() => members.map((m) => ({ value: m.user_id, label: m.name })))
const selected = ref<number | null>(null)

watch(selected, (id) => {
  if (id !== null) {
    emit('pick', id)
    selected.value = null
  }
})
</script>

<template>
  <UiSelect v-if="options.length" v-model="selected" :options="options" placeholder="+ Родитель" />
</template>

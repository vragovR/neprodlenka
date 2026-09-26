<script setup lang="ts">
import { ref, watch } from 'vue'
import UiInput from '@/components/ui/UiInput.vue'

const { label = '', error = '' } = defineProps<{ label?: string; error?: string }>()
const model = defineModel<number | null>({ required: true })

const toText = (kopecks: number | null) => (kopecks === null ? '' : String(kopecks / 100))
const text = ref(toText(model.value))

watch(text, (s) => {
  const n = Number(s.replace(',', '.').replace(/\s/g, ''))
  model.value = s.trim() !== '' && Number.isFinite(n) && n >= 0 ? Math.round(n * 100) : null
})
</script>

<template>
  <UiInput v-model="text" :label="label" :error="error" inputmode="decimal">
    <template #suffix><span>₽</span></template>
  </UiInput>
</template>

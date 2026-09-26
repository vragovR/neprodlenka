<script setup lang="ts">
import { reactive, ref } from 'vue'
import type { Child, ChildInput } from '@/api/types'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'
import PriceInput from './PriceInput.vue'

const defaultPrice = 13000

const { child, saving = false } = defineProps<{ child: Child | null; saving?: boolean }>()
const emit = defineEmits<{ submit: [input: ChildInput]; cancel: [] }>()

const form = reactive({
  last_name: child?.last_name ?? '',
  first_name: child?.first_name ?? '',
  middle_name: child?.middle_name ?? '',
  price: (child?.price ?? defaultPrice) as number | null,
})
const touched = ref(false)

function submit() {
  touched.value = true
  if (!form.last_name.trim() || !form.first_name.trim() || form.price === null) {
    return
  }
  emit('submit', {
    last_name: form.last_name.trim(),
    first_name: form.first_name.trim(),
    middle_name: form.middle_name.trim(),
    price: form.price,
  })
}
</script>

<template>
  <form class="form" novalidate @submit.prevent="submit">
    <UiInput
      v-model="form.last_name"
      label="Фамилия"
      :error="touched && !form.last_name.trim() ? 'Укажите фамилию' : ''"
      autofocus
    />
    <UiInput v-model="form.first_name" label="Имя" :error="touched && !form.first_name.trim() ? 'Укажите имя' : ''" />
    <UiInput v-model="form.middle_name" label="Отчество (необязательно)" />
    <PriceInput
      v-model="form.price"
      label="Цена за день"
      :error="touched && form.price === null ? 'Укажите цену' : ''"
    />
    <p v-if="child" class="muted hint">Новая цена действует только на новые отметки.</p>
    <div class="actions">
      <UiButton variant="ghost" @click="emit('cancel')">Отмена</UiButton>
      <UiButton type="submit" variant="primary" :loading="saving">{{ child ? 'Сохранить' : 'Добавить' }}</UiButton>
    </div>
  </form>
</template>

<style scoped>
.form {
  display: grid;
  gap: 14px;
}

.hint {
  font-size: 0.8125rem;
}

.actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  padding-top: 6px;
}
</style>

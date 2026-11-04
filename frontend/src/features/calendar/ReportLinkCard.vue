<script setup lang="ts">
import CopyField from '@/components/CopyField.vue'
import UiButton from '@/components/ui/UiButton.vue'
import UiCard from '@/components/ui/UiCard.vue'

const { url } = defineProps<{ url: string }>()
const openReport = () => window.open(url, '_blank', 'noopener')
</script>

<template>
  <UiCard title="Отчёт за месяц">
    <p class="muted">Ссылка открывается без входа — отправьте её родителям.</p>
    <CopyField :value="url">
      <UiButton @click="openReport">Открыть</UiButton>
    </CopyField>
  </UiCard>
</template>

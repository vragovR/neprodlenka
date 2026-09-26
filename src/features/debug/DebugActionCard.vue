<script setup lang="ts">
import { ref } from 'vue'
import { errorMessage } from '@/api/client'
import { toast } from '@/composables/useToast'
import UiButton from '@/components/ui/UiButton.vue'
import UiCard from '@/components/ui/UiCard.vue'

export interface DebugAction {
  label: string
  run: () => Promise<string>
  danger?: boolean
  confirm?: string
}

const { title, action, note = '' } = defineProps<{
  title: string
  action?: DebugAction
  note?: string
}>()
const running = ref(false)

async function run(action: DebugAction) {
  if (action.confirm && !window.confirm(action.confirm)) {
    return
  }
  running.value = true
  try {
    toast(await action.run())
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    running.value = false
  }
}
</script>

<template>
  <UiCard :title="title">
    <slot />
    <div v-if="action" class="foot">
      <UiButton :variant="action.danger ? 'danger' : 'primary'" :loading="running" @click="run(action)">
        {{ action.label }}
      </UiButton>
    </div>
    <p v-else-if="note" class="muted">{{ note }}</p>
  </UiCard>
</template>

<style scoped>
.foot {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  align-items: center;
}
</style>

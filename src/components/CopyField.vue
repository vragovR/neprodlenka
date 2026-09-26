<script setup lang="ts">
import { ref } from 'vue'
import UiButton from './ui/UiButton.vue'

const { value } = defineProps<{ value: string }>()
const copied = ref(false)

async function copy() {
  try {
    await navigator.clipboard.writeText(value)
  } catch {
    return
  }
  copied.value = true
  setTimeout(() => (copied.value = false), 2000)
}
</script>

<template>
  <div class="copy">
    <input class="value" :value="value" readonly @focus="($event.target as HTMLInputElement).select()" />
    <UiButton variant="primary" @click="copy">{{ copied ? 'Скопировано' : 'Скопировать' }}</UiButton>
    <slot />
  </div>
</template>

<style scoped>
.copy {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.copy > :deep(.btn) {
  flex: 1;
}

.value {
  flex: 1 1 100%;
  min-width: 0;
  height: 38px;
  padding: 0 12px;
  background: var(--surface-muted);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.875rem;
}

@media (min-width: 720px) {
  .value {
    flex-basis: 0;
  }

  .copy > :deep(.btn) {
    flex: none;
  }
}
</style>

<script setup lang="ts">
import { useTemplateRef, watch } from 'vue'

const { title } = defineProps<{ title: string }>()
const open = defineModel<boolean>('open', { default: false })
const dialog = useTemplateRef('dialog')

watch([open, dialog], ([isOpen, el]) => {
  if (!el) {
    return
  }
  if (isOpen && !el.open) {
    el.showModal()
  } else if (!isOpen && el.open) {
    el.close()
  }
})

function onClick(e: MouseEvent) {
  if (e.target === dialog.value) {
    open.value = false
  }
}
</script>

<template>
  <dialog ref="dialog" class="modal" @close="open = false" @click="onClick">
    <div v-if="open" class="body">
      <header class="head">
        <h2>{{ title }}</h2>
        <button type="button" class="close" aria-label="Закрыть" @click="open = false">×</button>
      </header>
      <slot />
    </div>
  </dialog>
</template>

<style scoped>
.modal {
  width: min(460px, calc(100vw - 32px));
  max-height: calc(100dvh - 32px);
  padding: 0;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  color: var(--text);
}

.modal::backdrop {
  background: rgb(10 12 20 / 45%);
  backdrop-filter: blur(2px);
}

.body {
  display: grid;
  gap: 20px;
  padding: 20px;
}

@media (min-width: 720px) {
  .body {
    padding: 24px;
  }
}

.head {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.close {
  width: 32px;
  height: 32px;
  background: transparent;
  border: 0;
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  font-size: 1.5rem;
  line-height: 1;
  cursor: pointer;
}

.close:hover {
  background: var(--surface-muted);
}
</style>

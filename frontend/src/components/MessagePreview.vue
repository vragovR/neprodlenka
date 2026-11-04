<script setup lang="ts">
import type { BotMessage } from '@/api/types'

const { message, sent, dirty } = defineProps<{
  message: BotMessage
  sent: boolean
  dirty: boolean
}>()
</script>

<template>
  <div class="message">
    <p class="muted hint">
      {{ sent ? 'Так сообщение выглядит в группе.' : 'Сообщение ещё не отправлено — так оно выглядело бы по текущим данным.' }}
      <template v-if="dirty">Несохранённые правки в нём не учтены.</template>
    </p>
    <div class="bubble">
      <div class="text" v-html="message.text" />
      <div v-if="message.buttons.length" class="buttons">
        <span v-for="b in message.buttons" :key="b" class="button">{{ b }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message {
  display: grid;
  gap: 12px;
}

.hint {
  margin: 0;
  font-size: 0.875rem;
}

.bubble {
  display: grid;
  gap: 12px;
  max-height: 60vh;
  padding: 14px 16px;
  overflow-y: auto;
  background: var(--surface-muted);
  border-radius: var(--radius);
}

.text {
  line-height: 1.45;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.text :deep(a) {
  color: var(--accent);
}

.buttons {
  display: grid;
}

.button {
  padding: 8px 12px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-weight: 500;
  text-align: center;
}
</style>

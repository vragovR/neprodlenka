<script setup lang="ts">
import { useId } from 'vue'

defineOptions({ inheritAttrs: false })

const { label = '', error = '' } = defineProps<{ label?: string; error?: string }>()
const model = defineModel<string>({ default: '' })
const id = useId()
</script>

<template>
  <div class="field">
    <label v-if="label" :for="id" class="label">{{ label }}</label>
    <div class="control" :class="{ invalid: error }">
      <input :id="id" v-model="model" v-bind="$attrs" class="input" />
      <slot name="suffix" />
    </div>
    <span v-if="error" class="error">{{ error }}</span>
  </div>
</template>

<style scoped>
.field {
  display: grid;
  gap: 6px;
}

.label {
  font-size: 0.875rem;
  font-weight: 500;
}

.control {
  display: flex;
  align-items: center;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  transition:
    border-color 0.15s,
    box-shadow 0.15s;
}

.control:focus-within {
  border-color: var(--accent);
  box-shadow: var(--focus);
}

.control.invalid {
  border-color: var(--danger);
}

.input {
  flex: 1;
  min-width: 0;
  height: 38px;
  padding: 0 12px;
  background: transparent;
  border: 0;
  outline: none;
}

.input:focus-visible {
  box-shadow: none;
}

.control :slotted(*) {
  padding-right: 12px;
  color: var(--text-muted);
}

@media (pointer: coarse) {
  .input {
    height: 44px;
    font-size: 16px;
  }
}

.error {
  color: var(--danger);
  font-size: 0.8125rem;
}
</style>

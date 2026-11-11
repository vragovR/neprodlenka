<script setup lang="ts">
const { label } = defineProps<{ label: string }>()
const model = defineModel<boolean>({ default: false })
</script>

<template>
  <label class="switch">
    <input v-model="model" type="checkbox" role="switch" class="input" />
    {{ label }}
  </label>
</template>

<style scoped>
.switch {
  display: inline-flex;
  gap: 10px;
  align-items: center;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
}

.input {
  appearance: none;
  display: inline-flex;
  flex-shrink: 0;
  align-items: center;
  width: 36px;
  height: 20px;
  margin: 0;
  padding: 2px;
  background: var(--surface-muted);
  border: 1px solid var(--border);
  border-radius: 999px;
  cursor: pointer;
  transition:
    background 0.15s,
    border-color 0.15s;
}

.input::before {
  content: '';
  width: 14px;
  height: 14px;
  background: var(--text-muted);
  border-radius: 50%;
  transition:
    transform 0.15s,
    background 0.15s;
}

.input:checked {
  background: var(--accent);
  border-color: var(--accent);
}

.input:checked::before {
  background: var(--on-accent);
  transform: translateX(16px);
}

.input:focus-visible {
  box-shadow: var(--focus);
}
</style>

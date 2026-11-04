<script setup lang="ts" generic="T extends string | number">
const { options, placeholder = '' } = defineProps<{
  options: { value: T; label: string }[]
  placeholder?: string
}>()
const model = defineModel<T | null>({ default: null })
</script>

<template>
  <select v-model="model" class="select">
    <option v-if="placeholder" :value="null" disabled>{{ placeholder }}</option>
    <option v-for="o in options" :key="o.value" :value="o.value">{{ o.label }}</option>
  </select>
</template>

<style scoped>
.select {
  height: 30px;
  max-width: 100%;
  padding: 0 28px 0 10px;
  background: var(--surface)
    url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='none' stroke='%23888' stroke-width='2'%3E%3Cpath d='m3 4.5 3 3 3-3'/%3E%3C/svg%3E")
    no-repeat right 10px center;
  border: 1px dashed var(--border);
  border-radius: 999px;
  font-size: 0.875rem;
  appearance: none;
  cursor: pointer;
}

@media (pointer: coarse) {
  .select {
    height: 36px;
    font-size: 16px;
  }
}

.select:hover {
  border-color: var(--accent);
}
</style>

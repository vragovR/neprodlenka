<script setup lang="ts">
const { small = false } = defineProps<{ small?: boolean }>()
</script>

<template>
  <span :class="['spinner', { small }]" role="status" aria-label="Загрузка" />
</template>

<style scoped>
.spinner {
  display: inline-block;
  width: 28px;
  height: 28px;
  border: 3px solid var(--border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.small {
  width: 14px;
  height: 14px;
  border-width: 2px;
  border-color: currentcolor;
  border-top-color: transparent;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>

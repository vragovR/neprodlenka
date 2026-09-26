<script setup lang="ts">
const { title = '' } = defineProps<{ title?: string }>()
</script>

<template>
  <section class="card">
    <header v-if="title || $slots.actions" class="head">
      <h2 v-if="title">{{ title }}</h2>
      <div class="actions"><slot name="actions" /></div>
    </header>
    <slot />
  </section>
</template>

<style scoped>
.card {
  display: grid;
  gap: 16px;
  padding: 16px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow);
}

.head {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
}

@media (min-width: 720px) {
  .card {
    padding: 20px;
  }
}

.actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>

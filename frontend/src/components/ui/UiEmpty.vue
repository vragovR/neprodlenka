<script setup lang="ts">
const { title, text = '' } = defineProps<{ title: string; text?: string }>()
</script>

<template>
  <div class="empty">
    <p class="title">{{ title }}</p>
    <p v-if="text" class="muted">{{ text }}</p>
    <slot />
  </div>
</template>

<style scoped>
.empty {
  display: grid;
  gap: 8px;
  justify-items: center;
  padding: 40px 16px;
  text-align: center;
}

.title {
  font-weight: 600;
}
</style>

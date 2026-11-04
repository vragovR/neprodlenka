<script setup lang="ts">
defineProps<{ items: { label: string; swatch: string }[] }>()
</script>

<template>
  <div class="legend">
    <span v-for="i in items" :key="i.label" class="item"><span :class="['swatch', i.swatch]" />{{ i.label }}</span>
  </div>
</template>

<style scoped>
.legend {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  color: var(--text-muted);
  font-size: 0.875rem;
}

.item {
  display: inline-flex;
  gap: 6px;
  align-items: center;
}

.swatch {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.swatch.off {
  width: 14px;
  height: 14px;
  background: var(--danger-soft);
  border-radius: 4px;
}

.swatch.message {
  background: var(--accent);
}

.swatch.open {
  background: var(--warning);
}

.swatch.closed {
  background: var(--success);
}
</style>

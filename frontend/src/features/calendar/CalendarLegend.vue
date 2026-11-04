<template>
  <div class="legend">
    <span class="item"><span class="swatch school" />Учебный</span>
    <span class="item"><span class="swatch off" />Неучебный</span>
  </div>
</template>

<style scoped>
.legend {
  display: flex;
  gap: 16px;
  color: var(--text-muted);
  font-size: 0.875rem;
}

.item {
  display: inline-flex;
  gap: 6px;
  align-items: center;
}

.swatch {
  width: 14px;
  height: 14px;
  border: 1px solid var(--border);
  border-radius: 4px;
}

.swatch.school {
  background: var(--surface);
}

.swatch.off {
  background: var(--danger-soft);
  border-color: transparent;
}
</style>

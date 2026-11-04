<script setup lang="ts">
import { ref } from 'vue'
import { capitalize, monthNames, monthOf, today } from '@/utils/format'

defineProps<{ marked?: (month: string) => boolean }>()
const month = defineModel<string>({ required: true })
const year = ref(Number(month.value.slice(0, 4)))

const key = (i: number) => `${year.value}-${String(i + 1).padStart(2, '0')}`
</script>

<template>
  <div class="picker">
    <div class="switcher">
      <button type="button" class="arrow" aria-label="Предыдущий год" @click="year--">‹</button>
      <span class="title">{{ year }}</span>
      <button type="button" class="arrow" aria-label="Следующий год" @click="year++">›</button>
    </div>
    <div class="grid">
      <button
        v-for="(name, i) in monthNames"
        :key="name"
        type="button"
        :class="[
          'month',
          {
            selected: key(i) === month,
            current: key(i) === monthOf(today()),
            marked: marked?.(key(i)),
          },
        ]"
        @click="month = key(i)">
        {{ capitalize(name) }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.picker {
  display: grid;
  gap: 12px;
  justify-items: center;
}

.switcher {
  display: inline-flex;
  gap: 4px;
  align-items: center;
}

.title {
  min-width: 140px;
  font-weight: 600;
  text-align: center;
}

.arrow {
  width: 32px;
  height: 32px;
  background: transparent;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-size: 1.25rem;
  line-height: 1;
  cursor: pointer;
}

.arrow:hover {
  background: var(--surface-muted);
}

.grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
  width: 100%;
}

.month {
  position: relative;
  height: 44px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  font-weight: 500;
  cursor: pointer;
  transition:
    background-color 0.15s,
    border-color 0.15s,
    color 0.15s;
}

.month:hover {
  border-color: var(--accent);
}

.month.current {
  outline: 2px solid var(--accent);
  outline-offset: -2px;
}

.month.selected {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--on-accent);
}

.month.marked::after {
  position: absolute;
  bottom: 5px;
  left: 50%;
  width: 6px;
  height: 6px;
  background: var(--accent);
  border-radius: 50%;
  content: '';
  translate: -50% 0;
}

.month.selected.marked::after {
  background: var(--on-accent);
}

@media (pointer: coarse) {
  .arrow {
    width: 40px;
    height: 40px;
  }
}
</style>

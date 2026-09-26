<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { logout } from '@/api'
import { errorMessage } from '@/api/client'
import { session } from '@/composables/useSession'
import { toast } from '@/composables/useToast'
import { debugEnabled } from '@/utils/debug'
import ThemeToggle from '@/components/ThemeToggle.vue'
import UiButton from '@/components/ui/UiButton.vue'

const tabs = [
  { name: 'children', label: 'Дети' },
  { name: 'attendance', label: 'Отметки' },
  { name: 'calendar', label: 'Календарь' },
  ...(debugEnabled ? [{ name: 'debug', label: 'Отладка' }] : []),
]
const route = useRoute()
const leaving = ref(false)
const menuOpen = ref(false)

async function signOut() {
  leaving.value = true
  try {
    await logout()
    session.signOut()
  } catch (e) {
    toast(errorMessage(e), 'danger')
  } finally {
    leaving.value = false
  }
}
</script>

<template>
  <header class="header">
    <div class="inner">
      <span class="brand">Непродлёнка</span>
      <button type="button" class="burger" :aria-expanded="menuOpen" aria-label="Меню" @click="menuOpen = !menuOpen">
        <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <path :d="menuOpen ? 'M6 6l12 12M18 6 6 18' : 'M4 7h16M4 12h16M4 17h16'" />
        </svg>
      </button>
      <div :class="['menu', { open: menuOpen }]">
        <nav class="tabs">
          <RouterLink
            v-for="t in tabs"
            :key="t.name"
            :to="{ name: t.name }"
            class="tab"
            :class="{ active: route.name === t.name }"
            @click="menuOpen = false"
          >
            {{ t.label }}
          </RouterLink>
        </nav>
        <div class="user">
          <ThemeToggle />
          <UiButton variant="ghost" size="sm" square :loading="leaving" aria-label="Выйти" title="Выйти" @click="signOut">
            <svg v-if="!leaving" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M15 4h3a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-3M10 17l5-5-5-5M15 12H4" />
            </svg>
          </UiButton>
        </div>
      </div>
    </div>
  </header>
</template>

<style scoped>
.header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: color-mix(in srgb, var(--surface) 85%, transparent);
  border-bottom: 1px solid var(--border);
  backdrop-filter: blur(12px);
}

.inner {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 24px;
  align-items: center;
  max-width: 1100px;
  margin: 0 auto;
  padding: 10px 16px;
}

.brand {
  font-weight: 700;
}

.burger {
  display: grid;
  place-items: center;
  width: 40px;
  height: 40px;
  margin-left: auto;
  padding: 0;
  background: transparent;
  border: 0;
  border-radius: var(--radius-sm);
  color: var(--text);
  cursor: pointer;
}

.burger:hover {
  background: var(--surface-muted);
}

.menu {
  display: none;
  gap: 8px;
  width: 100%;
  padding: 4px 0 6px;
}

.menu.open {
  display: grid;
}

.tabs {
  display: grid;
  gap: 2px;
}

.tab {
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  color: var(--text-muted);
  font-weight: 500;
  text-decoration: none;
  white-space: nowrap;
}

.tab:hover {
  background: var(--surface-muted);
  color: var(--text);
}

.tab.active {
  background: var(--accent-soft);
  color: var(--accent);
}

.user {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
  padding-top: 8px;
  border-top: 1px solid var(--border);
}

@media (min-width: 720px) {
  .burger {
    display: none;
  }

  .menu,
  .menu.open {
    display: flex;
    flex: 1;
    gap: 24px;
    align-items: center;
    width: auto;
    padding: 0;
  }

  .tabs {
    display: flex;
    gap: 4px;
  }

  .tab {
    padding: 6px 12px;
  }

  .user {
    margin-left: auto;
    padding: 0;
    border: 0;
  }
}
</style>

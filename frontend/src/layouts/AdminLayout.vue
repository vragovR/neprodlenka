<script setup lang="ts">
import AppHeader from './AppHeader.vue'
import LoginForm from '@/features/auth/LoginForm.vue'
import { session } from '@/composables/useSession'
</script>

<template>
  <LoginForm v-if="session.needLogin" />
  <template v-else>
    <AppHeader />
    <RouterView />
  </template>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { login as apiLogin } from '@/api'
import { errorMessage } from '@/api/client'
import { session } from '@/composables/useSession'
import UiButton from '@/components/ui/UiButton.vue'
import UiInput from '@/components/ui/UiInput.vue'

const login = ref(session.login)
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  loading.value = true
  error.value = ''
  try {
    const res = await apiLogin(login.value.trim(), password.value)
    session.signIn(res.login)
  } catch (e) {
    error.value = errorMessage(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="wrap">
    <form class="form" @submit.prevent="submit">
      <div class="head">
        <h1>Вход</h1>
        <p class="muted">Непродлёнка — для администратора</p>
      </div>
      <UiInput v-model="login" label="Логин" autocomplete="username" required />
      <UiInput v-model="password" label="Пароль" type="password" autocomplete="current-password" required />
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <UiButton type="submit" variant="primary" :loading="loading">Войти</UiButton>
    </form>
  </main>
</template>

<style scoped>
.wrap {
  display: grid;
  place-items: center;
  min-height: 100vh;
  padding: 16px;
}

.form {
  display: grid;
  gap: 16px;
  width: min(380px, 100%);
  padding: 32px 24px;
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
}

.head {
  display: grid;
  gap: 4px;
}

.error {
  padding: 10px 12px;
  background: var(--danger-soft);
  border-radius: var(--radius-sm);
  color: var(--danger);
  font-size: 0.875rem;
}
</style>

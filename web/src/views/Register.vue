<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const inviteCode = ref('')
const loading = ref(false)
const error = ref('')

const handleRegister = async () => {
  if (!email.value || !password.value) {
    error.value = '请填写邮箱和密码'
    return
  }

  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  if (password.value.length < 6) {
    error.value = '密码长度至少6位'
    return
  }

  loading.value = true
  error.value = ''

  try {
    await userStore.register(email.value, password.value, inviteCode.value || undefined)
    router.push('/')
  } catch (e: any) {
    error.value = e.response?.data?.error || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4 bg-gradient-to-br from-surface-50 via-macaron-mint/20 to-macaron-blue/20">
    <div class="w-full max-w-md animate-scale-in">
      <!-- Logo -->
      <div class="text-center mb-8">
        <div class="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-primary-500 to-primary-600 text-white text-2xl font-bold shadow-lg mb-4">
          X
        </div>
        <h1 class="text-3xl font-bold gradient-text">XBoard</h1>
        <p class="text-gray-500 mt-2">创建新账户</p>
      </div>

      <!-- Form -->
      <div class="card">
        <form @submit.prevent="handleRegister" class="space-y-5">
          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">邮箱</label>
            <input
              v-model="email"
              type="email"
              placeholder="your@email.com"
              class="input"
              autocomplete="email"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">密码</label>
            <input
              v-model="password"
              type="password"
              placeholder="至少6位"
              class="input"
              autocomplete="new-password"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">确认密码</label>
            <input
              v-model="confirmPassword"
              type="password"
              placeholder="再次输入密码"
              class="input"
              autocomplete="new-password"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 mb-2">邀请码 (可选)</label>
            <input
              v-model="inviteCode"
              type="text"
              placeholder="输入邀请码"
              class="input"
            />
          </div>

          <div v-if="error" class="p-3 rounded-xl bg-red-50 text-red-600 text-sm">
            {{ error }}
          </div>

          <button
            type="submit"
            :disabled="loading"
            class="w-full btn btn-primary py-3 text-base"
          >
            {{ loading ? '注册中...' : '注册' }}
          </button>
        </form>

        <div class="mt-6 text-center text-sm text-gray-500">
          已有账户？
          <RouterLink to="/login" class="text-primary-600 hover:text-primary-700 font-medium">
            立即登录
          </RouterLink>
        </div>
      </div>

      <!-- Footer -->
      <p class="text-center text-sm text-gray-400 mt-8">
        © 2024 XBoard. All rights reserved.
      </p>
    </div>
  </div>
</template>

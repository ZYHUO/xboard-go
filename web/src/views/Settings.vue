<script setup lang="ts">
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'
import api from '@/api'

const userStore = useUserStore()

const passwordForm = ref({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

const changingPassword = ref(false)
const resettingToken = ref(false)
const resettingUUID = ref(false)

const changePassword = async () => {
  if (!passwordForm.value.old_password || !passwordForm.value.new_password) {
    alert('请填写完整')
    return
  }

  if (passwordForm.value.new_password !== passwordForm.value.confirm_password) {
    alert('两次输入的密码不一致')
    return
  }

  if (passwordForm.value.new_password.length < 6) {
    alert('新密码长度至少6位')
    return
  }

  changingPassword.value = true
  try {
    await api.post('/api/v1/user/change_password', {
      old_password: passwordForm.value.old_password,
      new_password: passwordForm.value.new_password,
    })
    alert('密码修改成功')
    passwordForm.value = { old_password: '', new_password: '', confirm_password: '' }
  } catch (e: any) {
    alert(e.response?.data?.error || '修改失败')
  } finally {
    changingPassword.value = false
  }
}

const resetToken = async () => {
  if (!confirm('重置后，您需要重新导入订阅链接，确定继续？')) return

  resettingToken.value = true
  try {
    const res = await api.post('/api/v1/user/reset_token')
    await userStore.fetchUser()
    alert('订阅链接已重置')
  } catch (e: any) {
    alert(e.response?.data?.error || '重置失败')
  } finally {
    resettingToken.value = false
  }
}

const resetUUID = async () => {
  if (!confirm('重置后，您需要重新导入订阅链接，确定继续？')) return

  resettingUUID.value = true
  try {
    const res = await api.post('/api/v1/user/reset_uuid')
    await userStore.fetchUser()
    alert('UUID 已重置')
  } catch (e: any) {
    alert(e.response?.data?.error || '重置失败')
  } finally {
    resettingUUID.value = false
  }
}
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">账户设置</h1>
      <p class="text-gray-500 mt-1">管理您的账户信息</p>
    </div>

    <!-- Account Info -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4">📧 账户信息</h2>
      <div class="space-y-4">
        <div class="flex items-center justify-between p-4 rounded-xl bg-surface-50">
          <div>
            <p class="text-sm text-gray-500">邮箱</p>
            <p class="font-medium">{{ userStore.user?.email }}</p>
          </div>
        </div>
        <div class="flex items-center justify-between p-4 rounded-xl bg-surface-50">
          <div>
            <p class="text-sm text-gray-500">UUID</p>
            <p class="font-mono text-sm">{{ userStore.user?.uuid }}</p>
          </div>
          <button @click="resetUUID" :disabled="resettingUUID" class="btn btn-secondary text-sm">
            {{ resettingUUID ? '重置中...' : '重置' }}
          </button>
        </div>
        <div class="flex items-center justify-between p-4 rounded-xl bg-surface-50">
          <div>
            <p class="text-sm text-gray-500">订阅 Token</p>
            <p class="font-mono text-sm">{{ userStore.user?.token }}</p>
          </div>
          <button @click="resetToken" :disabled="resettingToken" class="btn btn-secondary text-sm">
            {{ resettingToken ? '重置中...' : '重置' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Change Password -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4">🔐 修改密码</h2>
      <form @submit.prevent="changePassword" class="space-y-4 max-w-md">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">当前密码</label>
          <input
            v-model="passwordForm.old_password"
            type="password"
            class="input"
            autocomplete="current-password"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">新密码</label>
          <input
            v-model="passwordForm.new_password"
            type="password"
            class="input"
            autocomplete="new-password"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-2">确认新密码</label>
          <input
            v-model="passwordForm.confirm_password"
            type="password"
            class="input"
            autocomplete="new-password"
          />
        </div>
        <button type="submit" :disabled="changingPassword" class="btn btn-primary">
          {{ changingPassword ? '修改中...' : '修改密码' }}
        </button>
      </form>
    </div>

    <!-- Danger Zone -->
    <div class="card border-2 border-red-100">
      <h2 class="text-lg font-semibold text-red-600 mb-4">⚠️ 危险操作</h2>
      <p class="text-sm text-gray-500 mb-4">以下操作不可逆，请谨慎操作</p>
      <button class="btn bg-red-50 text-red-600 hover:bg-red-100">
        注销账户
      </button>
    </div>
  </div>
</template>

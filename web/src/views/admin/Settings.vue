<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

const settings = ref<Record<string, string>>({})
const loading = ref(false)
const saving = ref(false)

const settingGroups = [
  {
    name: '基础设置',
    icon: '⚙️',
    items: [
      { key: 'app_name', label: '站点名称', type: 'text' },
      { key: 'app_url', label: '站点地址', type: 'text' },
      { key: 'subscribe_url', label: '订阅地址', type: 'text' },
    ]
  },
  {
    name: '邮件设置',
    icon: '📧',
    items: [
      { key: 'mail_host', label: 'SMTP 服务器', type: 'text' },
      { key: 'mail_port', label: 'SMTP 端口', type: 'text' },
      { key: 'mail_username', label: 'SMTP 用户名', type: 'text' },
      { key: 'mail_password', label: 'SMTP 密码', type: 'password' },
      { key: 'mail_from_address', label: '发件人地址', type: 'text' },
      { key: 'mail_from_name', label: '发件人名称', type: 'text' },
    ]
  },
  {
    name: 'Telegram 设置',
    icon: '📱',
    items: [
      { key: 'telegram_bot_token', label: 'Bot Token', type: 'password' },
      { key: 'telegram_webhook_url', label: 'Webhook URL', type: 'text' },
    ]
  },
  {
    name: '节点设置',
    icon: '🌐',
    items: [
      { key: 'server_push_interval', label: '推送间隔(秒)', type: 'number' },
      { key: 'server_pull_interval', label: '拉取间隔(秒)', type: 'number' },
    ]
  },
]

const fetchSettings = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/settings')
    settings.value = res.data.data || {}
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const saveSettings = async () => {
  saving.value = true
  try {
    await api.post('/api/v2/admin/settings', settings.value)
    alert('保存成功')
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(fetchSettings)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">系统设置</h1>
        <p class="text-gray-500 mt-1">配置系统参数</p>
      </div>
      <button @click="saveSettings" :disabled="saving" class="btn btn-primary">
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-500">
      加载中...
    </div>

    <div v-else class="space-y-6">
      <div v-for="group in settingGroups" :key="group.name" class="bg-white rounded-xl shadow-sm p-6">
        <h2 class="text-lg font-semibold mb-4 flex items-center gap-2">
          <span>{{ group.icon }}</span>
          {{ group.name }}
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div v-for="item in group.items" :key="item.key">
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ item.label }}</label>
            <input
              v-model="settings[item.key]"
              :type="item.type"
              class="input"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

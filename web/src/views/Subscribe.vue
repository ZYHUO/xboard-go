<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import api from '@/api'

const userStore = useUserStore()
const servers = ref<any[]>([])
const loading = ref(false)
const copied = ref(false)

const subscribeUrl = computed(() => {
  return `${window.location.origin}/api/v1/client/subscribe?token=${userStore.user?.token}`
})

const clients = [
  { name: 'Clash', icon: '🔥', format: 'clash', desc: 'Windows / macOS / Linux' },
  { name: 'Clash Meta', icon: '⚡', format: 'clashmeta', desc: 'mihomo 内核' },
  { name: 'sing-box', icon: '📦', format: 'singbox', desc: '全平台通用' },
  { name: 'Shadowrocket', icon: '🚀', format: 'shadowrocket', desc: 'iOS' },
  { name: 'Quantumult X', icon: '🎯', format: 'quantumultx', desc: 'iOS' },
  { name: 'Surge', icon: '🌊', format: 'surge', desc: 'iOS / macOS' },
  { name: 'Loon', icon: '🎈', format: 'loon', desc: 'iOS' },
  { name: 'Surfboard', icon: '🏄', format: 'surfboard', desc: 'Android' },
]

const copyUrl = async (format?: string) => {
  let url = subscribeUrl.value
  if (format) {
    url += `&format=${format}`
  }
  
  try {
    await navigator.clipboard.writeText(url)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  } catch (e) {
    // Fallback
    const input = document.createElement('input')
    input.value = url
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  }
}

const fetchServers = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/user/subscribe')
    servers.value = res.data.data.servers || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchServers)
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">订阅管理</h1>
      <p class="text-gray-500 mt-1">获取订阅链接并导入到客户端</p>
    </div>

    <!-- Subscribe URL -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4">📋 通用订阅链接</h2>
      <div class="flex gap-2">
        <input 
          type="text" 
          :value="subscribeUrl" 
          readonly 
          class="input flex-1 bg-surface-50 font-mono text-sm"
        />
        <button @click="copyUrl()" class="btn btn-primary whitespace-nowrap">
          {{ copied ? '✓ 已复制' : '复制' }}
        </button>
      </div>
      <p class="text-sm text-gray-500 mt-3">
        ⚠️ 请勿泄露此链接，如已泄露请在设置中重置订阅链接
      </p>
    </div>

    <!-- Clients -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4">📱 客户端订阅</h2>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <button
          v-for="client in clients"
          :key="client.name"
          @click="copyUrl(client.format)"
          class="flex flex-col items-center gap-2 p-4 rounded-xl bg-surface-50 hover:bg-surface-100 hover:shadow-md transition-all duration-200 group"
        >
          <span class="text-3xl group-hover:scale-110 transition-transform">{{ client.icon }}</span>
          <span class="font-medium">{{ client.name }}</span>
          <span class="text-xs text-gray-500">{{ client.desc }}</span>
        </button>
      </div>
    </div>

    <!-- Server List -->
    <div class="card">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-lg font-semibold">🌐 节点列表</h2>
        <button @click="fetchServers" class="btn btn-ghost text-sm" :disabled="loading">
          {{ loading ? '刷新中...' : '刷新' }}
        </button>
      </div>
      
      <div v-if="loading" class="text-center py-8 text-gray-500">
        加载中...
      </div>
      
      <div v-else-if="servers.length === 0" class="text-center py-8 text-gray-500">
        暂无可用节点
      </div>
      
      <div v-else class="space-y-3">
        <div 
          v-for="server in servers" 
          :key="server.id"
          class="flex items-center justify-between p-4 rounded-xl bg-surface-50 hover:bg-surface-100 transition-colors"
        >
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white text-sm font-medium">
              {{ server.type?.charAt(0).toUpperCase() }}
            </div>
            <div>
              <p class="font-medium">{{ server.name }}</p>
              <p class="text-sm text-gray-500">{{ server.type }} · {{ server.rate }}x</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span 
              v-for="tag in (server.tags || [])" 
              :key="tag"
              class="badge badge-info"
            >
              {{ tag }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

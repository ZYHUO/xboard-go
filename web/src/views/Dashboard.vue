<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '@/stores/user'
import dayjs from 'dayjs'

const userStore = useUserStore()

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const expireDate = computed(() => {
  if (!userStore.user?.expired_at) return '永久'
  return dayjs.unix(userStore.user.expired_at).format('YYYY-MM-DD')
})

const daysLeft = computed(() => {
  if (!userStore.user?.expired_at) return -1
  const now = dayjs()
  const expire = dayjs.unix(userStore.user.expired_at)
  return expire.diff(now, 'day')
})
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <!-- Welcome -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">欢迎回来 👋</h1>
        <p class="text-gray-500 mt-1">查看您的账户概览</p>
      </div>
      <RouterLink to="/subscribe" class="btn btn-primary">
        获取订阅
      </RouterLink>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Traffic Card -->
      <div class="macaron-card macaron-mint text-gray-800">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium opacity-80">流量使用</span>
          <span class="text-2xl">📊</span>
        </div>
        <div class="text-2xl font-bold mb-2">{{ userStore.trafficPercent }}%</div>
        <div class="text-sm opacity-80">
          {{ formatBytes(userStore.usedTraffic) }} / {{ formatBytes(userStore.totalTraffic) }}
        </div>
        <div class="mt-3 h-2 bg-white/50 rounded-full overflow-hidden">
          <div 
            class="h-full bg-white rounded-full transition-all duration-500"
            :style="{ width: `${userStore.trafficPercent}%` }"
          ></div>
        </div>
      </div>

      <!-- Expire Card -->
      <div class="macaron-card macaron-lavender text-gray-800">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium opacity-80">到期时间</span>
          <span class="text-2xl">📅</span>
        </div>
        <div class="text-2xl font-bold mb-2">{{ expireDate }}</div>
        <div class="text-sm opacity-80">
          <template v-if="daysLeft >= 0">
            剩余 {{ daysLeft }} 天
          </template>
          <template v-else>
            永久有效
          </template>
        </div>
      </div>

      <!-- Balance Card -->
      <div class="macaron-card macaron-yellow text-gray-800">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium opacity-80">账户余额</span>
          <span class="text-2xl">💰</span>
        </div>
        <div class="text-2xl font-bold mb-2">
          ¥{{ ((userStore.user?.balance ?? 0) / 100).toFixed(2) }}
        </div>
        <div class="text-sm opacity-80">可用于购买套餐</div>
      </div>

      <!-- Upload/Download Card -->
      <div class="macaron-card macaron-pink text-gray-800">
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm font-medium opacity-80">上传/下载</span>
          <span class="text-2xl">📈</span>
        </div>
        <div class="space-y-2">
          <div class="flex justify-between">
            <span class="text-sm opacity-80">⬆️ 上传</span>
            <span class="font-medium">{{ formatBytes(userStore.user?.u ?? 0) }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-sm opacity-80">⬇️ 下载</span>
            <span class="font-medium">{{ formatBytes(userStore.user?.d ?? 0) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4">快捷操作</h2>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <RouterLink to="/subscribe" class="flex flex-col items-center gap-2 p-4 rounded-xl bg-surface-50 hover:bg-surface-100 transition-colors">
          <span class="text-3xl">🔗</span>
          <span class="text-sm font-medium">获取订阅</span>
        </RouterLink>
        <RouterLink to="/plans" class="flex flex-col items-center gap-2 p-4 rounded-xl bg-surface-50 hover:bg-surface-100 transition-colors">
          <span class="text-3xl">💎</span>
          <span class="text-sm font-medium">购买套餐</span>
        </RouterLink>
        <RouterLink to="/tickets" class="flex flex-col items-center gap-2 p-4 rounded-xl bg-surface-50 hover:bg-surface-100 transition-colors">
          <span class="text-3xl">💬</span>
          <span class="text-sm font-medium">提交工单</span>
        </RouterLink>
        <RouterLink to="/settings" class="flex flex-col items-center gap-2 p-4 rounded-xl bg-surface-50 hover:bg-surface-100 transition-colors">
          <span class="text-3xl">⚙️</span>
          <span class="text-sm font-medium">账户设置</span>
        </RouterLink>
      </div>
    </div>

    <!-- Announcements -->
    <div class="card">
      <h2 class="text-lg font-semibold mb-4">📢 公告</h2>
      <div class="space-y-3">
        <div class="p-4 rounded-xl bg-surface-50 border-l-4 border-primary-500">
          <h3 class="font-medium">欢迎使用 XBoard</h3>
          <p class="text-sm text-gray-500 mt-1">感谢您的支持，祝您使用愉快！</p>
        </div>
      </div>
    </div>
  </div>
</template>

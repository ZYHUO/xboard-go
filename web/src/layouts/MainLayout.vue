<script setup lang="ts">
import { ref, computed } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const route = useRoute()
const isSidebarOpen = ref(false)

const navItems = [
  { path: '/', name: '仪表盘', icon: '📊' },
  { path: '/subscribe', name: '订阅', icon: '🔗' },
  { path: '/plans', name: '套餐', icon: '💎' },
  { path: '/orders', name: '订单', icon: '📋' },
  { path: '/tickets', name: '工单', icon: '💬' },
  { path: '/invite', name: '邀请', icon: '🎁' },
  { path: '/knowledge', name: '帮助', icon: '📚' },
  { path: '/settings', name: '设置', icon: '⚙️' },
]

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}
</script>

<template>
  <div class="min-h-screen bg-gradient-to-br from-surface-50 to-primary-50/30">
    <!-- Mobile Header -->
    <header class="lg:hidden fixed top-0 left-0 right-0 z-50 bg-white/80 backdrop-blur-lg border-b border-surface-200">
      <div class="flex items-center justify-between px-4 h-16">
        <button @click="isSidebarOpen = true" class="p-2 rounded-xl hover:bg-surface-100">
          <span class="text-xl">☰</span>
        </button>
        <span class="font-semibold gradient-text">XBoard</span>
        <div class="w-10"></div>
      </div>
    </header>

    <!-- Sidebar -->
    <aside 
      :class="[
        'fixed top-0 left-0 z-50 h-full w-64 bg-white shadow-soft transition-transform duration-300 lg:translate-x-0',
        isSidebarOpen ? 'translate-x-0' : '-translate-x-full'
      ]"
    >
      <div class="flex flex-col h-full">
        <!-- Logo -->
        <div class="flex items-center gap-3 px-6 h-20 border-b border-surface-100">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center text-white font-bold">
            X
          </div>
          <span class="text-xl font-bold gradient-text">XBoard</span>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 px-4 py-6 space-y-2 overflow-y-auto">
          <RouterLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            :class="[
              'flex items-center gap-3 px-4 py-3 rounded-xl transition-all duration-200',
              isActive(item.path) 
                ? 'bg-primary-500 text-white shadow-md' 
                : 'text-gray-600 hover:bg-surface-100'
            ]"
            @click="isSidebarOpen = false"
          >
            <span class="text-lg">{{ item.icon }}</span>
            <span class="font-medium">{{ item.name }}</span>
          </RouterLink>
        </nav>

        <!-- User Info -->
        <div class="p-4 border-t border-surface-100">
          <div class="flex items-center gap-3 p-3 rounded-xl bg-surface-50">
            <div class="w-10 h-10 rounded-full bg-gradient-to-br from-macaron-pink to-macaron-peach flex items-center justify-center text-white font-medium">
              {{ userStore.user?.email?.charAt(0).toUpperCase() }}
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium truncate">{{ userStore.user?.email }}</p>
              <p class="text-xs text-gray-500">{{ userStore.isAdmin ? '管理员' : '用户' }}</p>
            </div>
          </div>
          <button 
            @click="userStore.logout()" 
            class="w-full mt-3 px-4 py-2 text-sm text-gray-600 hover:text-red-500 hover:bg-red-50 rounded-xl transition-colors"
          >
            退出登录
          </button>
        </div>
      </div>
    </aside>

    <!-- Overlay -->
    <div 
      v-if="isSidebarOpen" 
      class="fixed inset-0 z-40 bg-black/20 backdrop-blur-sm lg:hidden"
      @click="isSidebarOpen = false"
    ></div>

    <!-- Main Content -->
    <main class="lg:ml-64 pt-16 lg:pt-0 min-h-screen">
      <div class="p-4 lg:p-8">
        <RouterView />
      </div>
    </main>
  </div>
</template>

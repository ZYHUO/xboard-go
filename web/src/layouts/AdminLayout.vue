<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const route = useRoute()
const isSidebarOpen = ref(false)

const navItems = [
  { path: '/admin', name: '仪表盘', icon: '📊' },
  { path: '/admin/users', name: '用户管理', icon: '👥' },
  { path: '/admin/hosts', name: '主机管理', icon: '🖥️' },
  { path: '/admin/servers', name: '节点管理', icon: '🌐' },
  { path: '/admin/plans', name: '套餐管理', icon: '💎' },
  { path: '/admin/orders', name: '订单管理', icon: '📋' },
  { path: '/admin/tickets', name: '工单管理', icon: '💬' },
  { path: '/admin/coupons', name: '优惠券', icon: '🎟️' },
  { path: '/admin/notices', name: '公告管理', icon: '📢' },
  { path: '/admin/settings', name: '系统设置', icon: '⚙️' },
]

const isActive = (path: string) => {
  if (path === '/admin') return route.path === '/admin'
  return route.path.startsWith(path)
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Sidebar -->
    <aside 
      :class="[
        'fixed top-0 left-0 z-50 h-full w-64 bg-gray-900 transition-transform duration-300 lg:translate-x-0',
        isSidebarOpen ? 'translate-x-0' : '-translate-x-full'
      ]"
    >
      <div class="flex flex-col h-full">
        <!-- Logo -->
        <div class="flex items-center gap-3 px-6 h-16 border-b border-gray-800">
          <div class="w-8 h-8 rounded-lg bg-primary-500 flex items-center justify-center text-white font-bold text-sm">
            X
          </div>
          <span class="text-lg font-bold text-white">XBoard Admin</span>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
          <RouterLink
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            :class="[
              'flex items-center gap-3 px-4 py-2.5 rounded-lg transition-all duration-200',
              isActive(item.path) 
                ? 'bg-primary-500 text-white' 
                : 'text-gray-400 hover:bg-gray-800 hover:text-white'
            ]"
            @click="isSidebarOpen = false"
          >
            <span>{{ item.icon }}</span>
            <span class="text-sm font-medium">{{ item.name }}</span>
          </RouterLink>
        </nav>

        <!-- Back to User -->
        <div class="p-3 border-t border-gray-800">
          <RouterLink to="/" class="flex items-center gap-3 px-4 py-2.5 rounded-lg text-gray-400 hover:bg-gray-800 hover:text-white transition-colors">
            <span>←</span>
            <span class="text-sm">返回用户端</span>
          </RouterLink>
        </div>
      </div>
    </aside>

    <!-- Mobile Header -->
    <header class="lg:hidden fixed top-0 left-0 right-0 z-40 bg-white border-b border-gray-200">
      <div class="flex items-center justify-between px-4 h-14">
        <button @click="isSidebarOpen = true" class="p-2 rounded-lg hover:bg-gray-100">
          <span>☰</span>
        </button>
        <span class="font-semibold">Admin</span>
        <div class="w-10"></div>
      </div>
    </header>

    <!-- Overlay -->
    <div 
      v-if="isSidebarOpen" 
      class="fixed inset-0 z-40 bg-black/50 lg:hidden"
      @click="isSidebarOpen = false"
    ></div>

    <!-- Main Content -->
    <main class="lg:ml-64 pt-14 lg:pt-0 min-h-screen">
      <div class="p-6">
        <RouterView />
      </div>
    </main>
  </div>
</template>

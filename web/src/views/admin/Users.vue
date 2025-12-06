<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import dayjs from 'dayjs'

interface User {
  id: number
  email: string
  balance: number
  plan_id: number | null
  transfer_enable: number
  u: number
  d: number
  expired_at: number | null
  banned: boolean
  is_admin: boolean
  created_at: number
}

const users = ref<User[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 20

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (ts: number | null) => {
  if (!ts) return '永久'
  return dayjs.unix(ts).format('YYYY-MM-DD')
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/users', {
      params: { page: page.value, page_size: pageSize }
    })
    users.value = res.data.data || []
    total.value = res.data.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchUsers)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">用户管理</h1>
        <p class="text-gray-500 mt-1">管理系统用户</p>
      </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">
        加载中...
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 border-b border-gray-200">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">余额</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">流量</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">到期</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="user in users" :key="user.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white text-sm font-medium">
                    {{ user.email.charAt(0).toUpperCase() }}
                  </div>
                  <div>
                    <div class="font-medium text-gray-900">{{ user.email }}</div>
                    <div class="text-xs text-gray-500">ID: {{ user.id }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 text-sm">
                ¥{{ (user.balance / 100).toFixed(2) }}
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ formatBytes(user.u + user.d) }} / {{ formatBytes(user.transfer_enable) }}
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ formatDate(user.expired_at) }}
              </td>
              <td class="px-6 py-4">
                <span v-if="user.banned" class="badge badge-danger">封禁</span>
                <span v-else-if="user.is_admin" class="badge badge-info">管理员</span>
                <span v-else class="badge badge-success">正常</span>
              </td>
              <td class="px-6 py-4 text-right">
                <RouterLink :to="`/admin/user/${user.id}`" class="text-primary-600 hover:text-primary-700 text-sm">
                  查看
                </RouterLink>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

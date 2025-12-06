<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import dayjs from 'dayjs'

interface Ticket {
  id: number
  user_id: number
  user_email: string
  subject: string
  level: number
  status: number
  reply_status: number
  created_at: number
  updated_at: number
}

const tickets = ref<Ticket[]>([])
const loading = ref(false)

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '开启', class: 'badge-success' },
  1: { text: '已关闭', class: 'badge-danger' },
}

const replyStatusMap: Record<number, { text: string; class: string }> = {
  0: { text: '待回复', class: 'badge-warning' },
  1: { text: '已回复', class: 'badge-info' },
}

const levelMap: Record<number, { text: string; class: string }> = {
  0: { text: '低', class: 'text-gray-500' },
  1: { text: '中', class: 'text-yellow-600' },
  2: { text: '高', class: 'text-red-600' },
}

const formatDate = (ts: number) => dayjs.unix(ts).format('YYYY-MM-DD HH:mm')

const fetchTickets = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/tickets')
    tickets.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchTickets)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">工单管理</h1>
      <p class="text-gray-500 mt-1">处理用户工单</p>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">
        加载中...
      </div>

      <div v-else-if="tickets.length === 0" class="text-center py-12 text-gray-500">
        暂无工单
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 border-b border-gray-200">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">主题</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">优先级</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">更新时间</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="ticket in tickets" :key="ticket.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div class="font-medium text-gray-900">{{ ticket.subject }}</div>
                <div class="text-xs text-gray-500">#{{ ticket.id }}</div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ ticket.user_email || `用户 ${ticket.user_id}` }}
              </td>
              <td class="px-6 py-4">
                <span :class="['text-sm font-medium', levelMap[ticket.level]?.class]">
                  {{ levelMap[ticket.level]?.text }}
                </span>
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-2">
                  <span :class="['badge', statusMap[ticket.status]?.class]">
                    {{ statusMap[ticket.status]?.text }}
                  </span>
                  <span :class="['badge', replyStatusMap[ticket.reply_status]?.class]">
                    {{ replyStatusMap[ticket.reply_status]?.text }}
                  </span>
                </div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ formatDate(ticket.updated_at) }}
              </td>
              <td class="px-6 py-4 text-right">
                <button class="text-primary-600 hover:text-primary-700 text-sm">
                  查看
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

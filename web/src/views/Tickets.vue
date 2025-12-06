<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import dayjs from 'dayjs'

interface Ticket {
  id: number
  subject: string
  level: number
  status: number
  reply_status: number
  created_at: number
  updated_at: number
}

const tickets = ref<Ticket[]>([])
const loading = ref(false)
const showCreateModal = ref(false)
const creating = ref(false)

const newTicket = ref({
  subject: '',
  message: '',
  level: 1,
})

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '开启', class: 'badge-success' },
  1: { text: '已关闭', class: 'badge-danger' },
}

const replyStatusMap: Record<number, { text: string; class: string }> = {
  0: { text: '待回复', class: 'badge-warning' },
  1: { text: '已回复', class: 'badge-info' },
}

const levelMap: Record<number, string> = {
  0: '低',
  1: '中',
  2: '高',
}

const formatDate = (ts: number) => dayjs.unix(ts).format('YYYY-MM-DD HH:mm')

const createTicket = async () => {
  if (!newTicket.value.subject || !newTicket.value.message) {
    alert('请填写主题和内容')
    return
  }

  creating.value = true
  try {
    await api.post('/api/v1/user/ticket/create', newTicket.value)
    showCreateModal.value = false
    newTicket.value = { subject: '', message: '', level: 1 }
    fetchTickets()
  } catch (e: any) {
    alert(e.response?.data?.error || '创建失败')
  } finally {
    creating.value = false
  }
}

const fetchTickets = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/user/tickets')
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
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">工单系统</h1>
        <p class="text-gray-500 mt-1">有问题？提交工单获取帮助</p>
      </div>
      <button @click="showCreateModal = true" class="btn btn-primary">
        新建工单
      </button>
    </div>

    <div class="card">
      <div v-if="loading" class="text-center py-12 text-gray-500">
        加载中...
      </div>

      <div v-else-if="tickets.length === 0" class="text-center py-12">
        <span class="text-5xl mb-4 block">💬</span>
        <p class="text-gray-500">暂无工单</p>
        <button @click="showCreateModal = true" class="btn btn-primary mt-4">
          新建工单
        </button>
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="ticket in tickets"
          :key="ticket.id"
          class="p-4 rounded-xl bg-surface-50 hover:bg-surface-100 transition-colors cursor-pointer"
        >
          <div class="flex items-start justify-between">
            <div class="space-y-2">
              <h3 class="font-medium text-gray-900">{{ ticket.subject }}</h3>
              <div class="flex items-center gap-2">
                <span :class="['badge', statusMap[ticket.status]?.class]">
                  {{ statusMap[ticket.status]?.text }}
                </span>
                <span :class="['badge', replyStatusMap[ticket.reply_status]?.class]">
                  {{ replyStatusMap[ticket.reply_status]?.text }}
                </span>
                <span class="text-sm text-gray-500">优先级: {{ levelMap[ticket.level] }}</span>
              </div>
            </div>
            <div class="text-right text-sm text-gray-500">
              <p>{{ formatDate(ticket.updated_at) }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showCreateModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg p-6 animate-scale-in">
          <h3 class="text-xl font-bold mb-4">新建工单</h3>
          
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">主题</label>
              <input
                v-model="newTicket.subject"
                type="text"
                placeholder="简要描述您的问题"
                class="input"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">优先级</label>
              <select v-model="newTicket.level" class="input">
                <option :value="0">低</option>
                <option :value="1">中</option>
                <option :value="2">高</option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">内容</label>
              <textarea
                v-model="newTicket.message"
                rows="5"
                placeholder="详细描述您遇到的问题..."
                class="input resize-none"
              ></textarea>
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showCreateModal = false" class="flex-1 btn btn-secondary">
              取消
            </button>
            <button @click="createTicket" :disabled="creating" class="flex-1 btn btn-primary">
              {{ creating ? '提交中...' : '提交' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import dayjs from 'dayjs'

interface Order {
  id: number
  user_id: number
  trade_no: string
  total_amount: number
  status: number
  type: number
  period: string
  created_at: number
}

const orders = ref<Order[]>([])
const loading = ref(false)

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '待支付', class: 'badge-warning' },
  1: { text: '开通中', class: 'badge-info' },
  2: { text: '已取消', class: 'badge-danger' },
  3: { text: '已完成', class: 'badge-success' },
  4: { text: '已折抵', class: 'badge-info' },
}

const typeMap: Record<number, string> = {
  1: '新购',
  2: '续费',
  3: '升级',
  4: '流量重置',
}

const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`
const formatDate = (ts: number) => dayjs.unix(ts).format('YYYY-MM-DD HH:mm')

const fetchOrders = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/orders')
    orders.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchOrders)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">订单管理</h1>
      <p class="text-gray-500 mt-1">查看所有订单</p>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">
        加载中...
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 border-b border-gray-200">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">订单号</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">金额</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="order in orders" :key="order.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 font-mono text-sm text-gray-500">{{ order.trade_no }}</td>
              <td class="px-6 py-4 text-sm">{{ order.user_id }}</td>
              <td class="px-6 py-4 text-sm">{{ typeMap[order.type] }}</td>
              <td class="px-6 py-4 text-sm font-medium">{{ formatPrice(order.total_amount) }}</td>
              <td class="px-6 py-4">
                <span :class="['badge', statusMap[order.status]?.class]">
                  {{ statusMap[order.status]?.text }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">{{ formatDate(order.created_at) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

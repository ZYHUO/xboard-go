<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import dayjs from 'dayjs'

interface Order {
  id: number
  trade_no: string
  plan_id: number
  period: string
  total_amount: number
  status: number
  type: number
  created_at: number
  paid_at: number | null
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

const periodMap: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  half_yearly: '半年付',
  yearly: '年付',
  two_yearly: '两年付',
  three_yearly: '三年付',
  onetime: '一次性',
}

const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`
const formatDate = (ts: number) => dayjs.unix(ts).format('YYYY-MM-DD HH:mm')

const cancelOrder = async (order: Order) => {
  if (!confirm('确定要取消此订单吗？')) return

  try {
    await api.post('/api/v1/user/order/cancel', { order_id: order.id })
    order.status = 2
  } catch (e: any) {
    alert(e.response?.data?.error || '取消失败')
  }
}

const fetchOrders = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/user/orders')
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
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">我的订单</h1>
        <p class="text-gray-500 mt-1">查看您的订单记录</p>
      </div>
      <RouterLink to="/plans" class="btn btn-primary">
        购买套餐
      </RouterLink>
    </div>

    <div class="card">
      <div v-if="loading" class="text-center py-12 text-gray-500">
        加载中...
      </div>

      <div v-else-if="orders.length === 0" class="text-center py-12">
        <span class="text-5xl mb-4 block">📋</span>
        <p class="text-gray-500">暂无订单记录</p>
        <RouterLink to="/plans" class="btn btn-primary mt-4">
          去购买
        </RouterLink>
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="order in orders"
          :key="order.id"
          class="p-4 rounded-xl bg-surface-50 hover:bg-surface-100 transition-colors"
        >
          <div class="flex items-start justify-between">
            <div class="space-y-2">
              <div class="flex items-center gap-2">
                <span :class="['badge', statusMap[order.status]?.class]">
                  {{ statusMap[order.status]?.text }}
                </span>
                <span class="text-sm text-gray-500">{{ typeMap[order.type] }}</span>
              </div>
              <p class="font-mono text-sm text-gray-500">{{ order.trade_no }}</p>
              <p class="text-sm text-gray-500">
                {{ periodMap[order.period] || order.period }} · {{ formatDate(order.created_at) }}
              </p>
            </div>
            <div class="text-right">
              <p class="text-xl font-bold text-gray-900">{{ formatPrice(order.total_amount) }}</p>
              <div class="mt-2 space-x-2">
                <button
                  v-if="order.status === 0"
                  class="btn btn-primary text-sm py-1"
                >
                  去支付
                </button>
                <button
                  v-if="order.status === 0"
                  @click="cancelOrder(order)"
                  class="btn btn-ghost text-sm py-1 text-red-500"
                >
                  取消
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

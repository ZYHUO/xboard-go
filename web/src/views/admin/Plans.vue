<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface Plan {
  id: number
  name: string
  transfer_enable: number
  speed_limit: number | null
  prices: Record<string, number>
  show: boolean
  sell: boolean
}

const plans = ref<Plan[]>([])
const loading = ref(false)

const formatBytes = (gb: number) => {
  if (gb >= 1024) return `${(gb / 1024).toFixed(0)} TB`
  return `${gb} GB`
}

const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

const getLowestPrice = (plan: Plan) => {
  if (!plan.prices) return 0
  const prices = Object.values(plan.prices).filter(p => p > 0)
  return prices.length > 0 ? Math.min(...prices) : 0
}

const fetchPlans = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/plans')
    plans.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchPlans)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">套餐管理</h1>
        <p class="text-gray-500 mt-1">管理订阅套餐</p>
      </div>
      <button class="btn btn-primary">添加套餐</button>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">
        加载中...
      </div>

      <table v-else class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">名称</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">流量</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">限速</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">起步价</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="plan in plans" :key="plan.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 font-medium text-gray-900">{{ plan.name }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ formatBytes(plan.transfer_enable) }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">
              {{ plan.speed_limit ? `${plan.speed_limit} Mbps` : '不限速' }}
            </td>
            <td class="px-6 py-4 text-sm font-medium text-primary-600">
              {{ formatPrice(getLowestPrice(plan)) }}
            </td>
            <td class="px-6 py-4">
              <span :class="['badge', plan.show && plan.sell ? 'badge-success' : 'badge-danger']">
                {{ plan.show && plan.sell ? '销售中' : '已下架' }}
              </span>
            </td>
            <td class="px-6 py-4 text-right space-x-2">
              <button class="text-primary-600 hover:text-primary-700 text-sm">编辑</button>
              <button class="text-red-600 hover:text-red-700 text-sm">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
